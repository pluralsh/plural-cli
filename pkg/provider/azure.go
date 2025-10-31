package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	"github.com/pluralsh/plural-cli/pkg/utils"
	pluralerr "github.com/pluralsh/plural-cli/pkg/utils/errors"
)

type ClientSet struct {
	Subscriptions *armsubscription.SubscriptionsClient
	Groups        *armresources.ResourceGroupsClient
	Accounts      *armstorage.AccountsClient
	Zones         *armdns.ZonesClient
	Containers    *azblob.ContainerURL
}

func GetClientSet(subscriptionId string) (*ClientSet, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	subscriptionsClient, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	storageAccountsClient, err := armstorage.NewAccountsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	zonesClient, err := armdns.NewZonesClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	return &ClientSet{
		Subscriptions: subscriptionsClient,
		Groups:        resourceGroupClient,
		Accounts:      storageAccountsClient,
		Zones:         zonesClient,
	}, nil
}

type AzureProvider struct {
	cluster       string
	resourceGroup string
	bucket        string
	region        string
	ctx           map[string]interface{}
	writer        manifest.Writer
	clients       *ClientSet
}

func mkAzure(conf config.Config) (prov *AzureProvider, err error) {
	subId, tenID, err := GetAzureAccount()
	if err != nil {
		return
	}

	clients, err := GetClientSet(subId)
	if err != nil {
		return
	}

	locations, err := AzureLocations(context.Background(), clients.Subscriptions, subId)
	if err != nil {
		return
	}

	groups, err := AzureResourceGroups(context.Background(), clients.Groups)
	if err != nil {
		return
	}

	accounts, err := AzureStorageAccounts(context.Background(), clients.Accounts)
	if err != nil {
		return
	}

	var resp struct{ Cluster, Region, Resource, Storage string }
	var azureSurvey = []*survey.Question{
		{
			Name:     "cluster",
			Prompt:   &survey.Input{Message: "Enter the name of your cluster:", Default: clusterFlag},
			Validate: validCluster,
		},
		{
			Name:     "region",
			Prompt:   &survey.Select{Message: "Select the region you want to deploy to:", Default: "eastus", Options: locations},
			Validate: survey.Required,
		},
		{
			Name:     "resource",
			Prompt:   &survey.Select{Message: "Select the resource group to use:", Options: groups},
			Validate: survey.Required,
		},
		{
			Name:     "storage",
			Prompt:   &survey.Select{Message: "Select the storage account to use:", Options: accounts},
			Validate: survey.Required,
		},
	}

	err = survey.Ask(azureSurvey, &resp)
	if err != nil {
		return
	}

	prov = &AzureProvider{
		resp.Cluster,
		resp.Resource,
		"",
		resp.Region,
		map[string]interface{}{
			"SubscriptionId": subId,
			"TenantId":       tenID,
			"StorageAccount": resp.Storage,
		},
		nil,
		clients,
	}

	projectManifest := manifest.ProjectManifest{
		Cluster:  prov.Cluster(),
		Project:  prov.Project(),
		Provider: api.ProviderAzure,
		Region:   prov.Region(),
		Context:  prov.Context(),
		Owner:    &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}
	prov.writer = projectManifest.Configure(cloudFlag, prov.Cluster())
	prov.bucket = projectManifest.Bucket
	return
}

func AzureFromManifest(man *manifest.ProjectManifest, clientSet *ClientSet) (*AzureProvider, error) {
	var err error
	clients := clientSet
	if clientSet == nil {
		clients, err = GetClientSet(utils.ToString(man.Context["SubscriptionId"]))
		if err != nil {
			return nil, err
		}
	}

	return &AzureProvider{man.Cluster, man.Project, man.Bucket, man.Region, man.Context, nil, clients}, nil
}

func (az *AzureProvider) CreateBucket() error {
	if err := az.CreateResourceGroup(az.Project()); err != nil {
		return pluralerr.ErrorWrap(err, fmt.Sprintf("Failed to create terraform state resource group %s", az.Project()))
	}

	if err := az.createContainer(az.bucket); err != nil {
		return pluralerr.ErrorWrap(err, fmt.Sprintf("Failed to create terraform state bucket %s", az.bucket))
	}

	return nil
}

func (az *AzureProvider) createContainer(bucket string) (err error) {
	acc, err := az.upsertStorageAccount(utils.ToString(az.Context()["StorageAccount"]))
	if err != nil {
		return
	}

	err = az.upsertStorageContainer(*acc, bucket)
	if err != nil {
		return
	}
	return
}

func (az *AzureProvider) CreateResourceGroup(resourceGroup string) error {
	ctx := context.Background()
	_, err := az.clients.Groups.Get(ctx, resourceGroup, nil)
	if err != nil && !isNotFoundResourceGroup(err) {
		return err
	}

	if isNotFoundResourceGroup(err) {
		utils.LogInfo().Printf("The resource group %s is not found, creating ...", resourceGroup)
		param := armresources.ResourceGroup{Location: to.Ptr(az.region)}
		_, err := az.clients.Groups.CreateOrUpdate(ctx, resourceGroup, param, nil)
		if err != nil {
			return err
		}
		utils.LogInfo().Printf("The resource group %s created successfully", resourceGroup)
	}

	return nil
}

func (az *AzureProvider) KubeConfig() error {
	if kubernetes.InKubernetes() {
		return nil
	}

	cmd := exec.Command(
		"az", "aks", "get-credentials", "--overwrite-existing", "--name", az.cluster, "--resource-group", az.resourceGroup)
	return utils.Execute(cmd)
}

func (az *AzureProvider) KubeContext() string {
	return az.cluster
}

func (az *AzureProvider) Name() string {
	return api.ProviderAzure
}

func (az *AzureProvider) Cluster() string {
	return az.cluster
}

func (az *AzureProvider) Project() string {
	return az.resourceGroup
}

func (az *AzureProvider) Bucket() string {
	return az.bucket
}

func (az *AzureProvider) Region() string {
	return az.region
}

func (az *AzureProvider) Context() map[string]interface{} {
	return az.ctx
}

func (az *AzureProvider) Preflights() []*Preflight {
	return nil
}

func (*AzureProvider) Permissions() (permissions.Checker, error) {
	return permissions.NullChecker(), nil
}

func (az *AzureProvider) Flush() error {
	if az.writer == nil {
		return nil
	}
	return az.writer()
}

func (az *AzureProvider) Decommision(node *v1.Node) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	ctx := context.Background()
	client, err := armcompute.NewVirtualMachineScaleSetsClient(utils.ToString(az.ctx["SubscriptionId"]), cred, nil)
	if err != nil {
		return err
	}

	// azure:///subscriptions/xxx/resourceGroups/yyy/providers/Microsoft.Compute/virtualMachineScaleSets/zzz/virtualMachines/0
	resourceGroup, err := getPathElement(node.Spec.ProviderID, "resourceGroups")
	if err != nil {
		return err
	}
	virtualMachineScaleSet, err := getPathElement(node.Spec.ProviderID, "virtualMachineScaleSets")
	if err != nil {
		return err
	}
	InstanceID, err := getPathElement(node.Spec.ProviderID, "virtualMachines")
	if err != nil {
		return err
	}

	// This method scale down the virtualMachineScaleSet otherwise the VM will be recreated
	pollerDeallocate, err := client.BeginDeallocate(ctx, resourceGroup, virtualMachineScaleSet, &armcompute.VirtualMachineScaleSetsClientBeginDeallocateOptions{
		VMInstanceIDs: &armcompute.VirtualMachineScaleSetVMInstanceIDs{
			InstanceIDs: []*string{to.Ptr(InstanceID)},
		},
	})
	if err != nil {
		return err
	}
	if _, err = pollerDeallocate.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: time.Second}); err != nil {
		return err
	}

	pollerDelete, err := client.BeginDeleteInstances(ctx, resourceGroup, virtualMachineScaleSet, armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIDs: []*string{to.Ptr(InstanceID)}}, &armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions{ForceDeletion: to.Ptr(true)})
	if err != nil {
		return err
	}
	if _, err := pollerDelete.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: time.Second}); err != nil {
		return err
	}

	return nil
}

func (az *AzureProvider) getStorageAccount(account string) (*armstorage.Account, error) {
	ctx := context.Background()
	pager := az.clients.Accounts.NewListPager(nil)

	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %w", err)
		}

		for _, sa := range nextResult.Value {
			if *sa.Name == account {
				resourceGroup, err := getPathElement(*sa.ID, "resourceGroups")
				if err != nil {
					return nil, fmt.Errorf("failed to read Storage Account's Resource Group: %w", err)
				}

				if resourceGroup != az.resourceGroup {
					return nil, fmt.Errorf("the '%s' Storage Account already exists and belongs to the '%s' Resource Group", account, resourceGroup)
				}
				break
			}
		}
	}

	res, err := az.clients.Accounts.GetProperties(ctx, az.resourceGroup, account, nil)
	if err != nil {
		return nil, err
	}

	return &res.Account, nil
}

func (az *AzureProvider) upsertStorageAccount(account string) (*armstorage.Account, error) {
	acc, err := az.getStorageAccount(account)
	if err != nil && !inNotFoundStorageAccount(err) {
		return nil, err
	}

	if inNotFoundStorageAccount(err) {
		utils.LogInfo().Printf("The storage account %s is not found, creating ...", account)
		ctx := context.Background()
		poller, err := az.clients.Accounts.BeginCreate(ctx, az.resourceGroup, account,
			armstorage.AccountCreateParameters{
				SKU:        &armstorage.SKU{Name: to.Ptr(armstorage.SKUNameStandardLRS)},
				Kind:       to.Ptr(armstorage.KindStorageV2),
				Location:   to.Ptr(az.region),
				Properties: &armstorage.AccountPropertiesCreateParameters{},
			}, nil)
		if err != nil {
			return nil, err
		}

		res, err := poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: time.Second})
		if err != nil {
			return nil, err
		}

		return &res.Account, nil
	}

	return acc, nil
}

func (az *AzureProvider) upsertStorageContainer(acc armstorage.Account, name string) error {
	ctx := context.Background()
	accountName := *acc.Name

	resp, err := az.clients.Accounts.ListKeys(ctx, az.resourceGroup, accountName, to.Ptr(armstorage.AccountsClientListKeysOptions{Expand: to.Ptr("kerb")}))
	if err != nil {
		return err
	}
	key := *resp.Keys[0].Value

	if az.clients.Containers == nil {
		c, _ := azblob.NewSharedKeyCredential(accountName, key)
		p := azblob.NewPipeline(c, azblob.PipelineOptions{})
		u, _ := url.Parse(fmt.Sprintf(`https://%s.blob.core.windows.net`, accountName))
		service := azblob.NewServiceURL(*u, p)
		containerClient := service.NewContainerURL(name)
		az.clients.Containers = &containerClient
	}

	_, err = az.clients.Containers.GetProperties(ctx, azblob.LeaseAccessConditions{})
	if err == nil {
		return err
	}

	_, err = az.clients.Containers.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	return err
}

func GetAzureAccount() (string, string, error) {
	cmd := exec.Command("az", "account", "show")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(string(out))
		return "", "", err
	}

	var res struct {
		TenantId string
		Id       string
	}

	if err := json.Unmarshal(out, &res); err != nil {
		return "", "", err
	}
	return res.Id, res.TenantId, nil
}

func isNotFoundResourceGroup(err error) bool {
	var aerr *azcore.ResponseError
	if err != nil && errors.As(err, &aerr) {
		return aerr.StatusCode == http.StatusNotFound
	}

	return false
}

func inNotFoundStorageAccount(err error) bool {
	var aerr *azcore.ResponseError
	if err != nil && errors.As(err, &aerr) {
		return aerr.StatusCode == http.StatusNotFound
	}

	return false
}

func getPathElement(path, indexName string) (string, error) {
	pattern := fmt.Sprintf(`.*\/%s\/(?P<element>([\w'-]+))`, indexName)
	captureGroupRegex := regexp.MustCompile(pattern)
	match := captureGroupRegex.FindStringSubmatch(path)
	if match != nil {
		index := captureGroupRegex.SubexpIndex("element")
		if index >= 0 {
			return match[index], nil
		}
	}

	return "", fmt.Errorf("%s not found", indexName)
}

func ValidateAzureDomainRegistration(ctx context.Context, domain, resourceGroup string) error {
	subId, _, err := GetAzureAccount()
	if err != nil {
		return err
	}

	clients, err := GetClientSet(subId)
	if err != nil {
		return err
	}

	d := strings.TrimSuffix(domain, ".")

	pager := clients.Zones.NewListByResourceGroupPager(resourceGroup, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, zone := range resp.Value {
			if lo.FromPtr(zone.Name) == d {
				return nil // Domain is registered, return without error.
			}
		}
	}

	return fmt.Errorf("domain %s not found", domain)
}
