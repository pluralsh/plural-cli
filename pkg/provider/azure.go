package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/azure/azure-sdk-for-go/services/compute/mgmt/2021-07-01/compute"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	pluralerr "github.com/pluralsh/plural/pkg/utils/errors"
	v1 "k8s.io/api/core/v1"
)

// ResourceGroupClient is the subset of functions we need from armresources.VirtualResourceGroupsClient;
// this interface is purely here for allowing unit tests.
type ResourceGroupClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error)
	Get(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientGetOptions) (armresources.ResourceGroupsClientGetResponse, error)
}

type AccountsClient interface {
	GetProperties(ctx context.Context, resourceGroupName string, accountName string, expand storage.AccountExpand) (result storage.Account, err error)
	Create(ctx context.Context, resourceGroupName string, accountName string, parameters storage.AccountCreateParameters) (result storage.AccountsCreateFuture, err error)
	ListKeys(ctx context.Context, resourceGroupName string, accountName string, expand storage.ListKeyExpand) (result storage.AccountListKeysResult, err error)
}

type ContainerClient interface {
	GetProperties(ctx context.Context, ac azblob.LeaseAccessConditions) (*azblob.ContainerGetPropertiesResponse, error)
	Create(ctx context.Context, metadata azblob.Metadata, publicAccessType azblob.PublicAccessType) (*azblob.ContainerCreateResponse, error)
}

type ClientSet struct {
	Groups         ResourceGroupClient
	Accounts       AccountsClient
	Containers     ContainerClient
	AutorestClient autorest.Client
	AccountClient  storage.AccountsClient
}

func GetClientSet(subscriptionId string) (*ClientSet, error) {
	resourceGroupClient, err := getResourceGroupClient(subscriptionId)
	if err != nil {
		return nil, err
	}

	storageAccountsClient, err := getStorageAccountsClient(subscriptionId)
	if err != nil {
		return nil, err
	}
	return &ClientSet{
		Groups:         resourceGroupClient,
		Accounts:       storageAccountsClient,
		AutorestClient: storageAccountsClient.Client,
		AccountClient:  storageAccountsClient,
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

var (
	azureRegions = []string{
		"eastus",
		"eastus2",
		"southcentralus",
		"westus2",
		"westus3",
		"australiaeast",
		"southeastasia",
		"northeurope",
		"swedencentral",
		"uksouth",
		"westeurope",
		"centralus",
		"southafricanorth",
		"centralindia",
		"eastasia",
		"japaneast",
		"koreacentral",
		"canadacentral",
		"francecentral",
		"germanywestcentral",
		"norwayeast",
		"brazilsouth",
	}
)

var azureSurvey = []*survey.Question{
	{
		Name:     "cluster",
		Prompt:   &survey.Input{Message: "Enter the name of your cluster:"},
		Validate: validCluster,
	},
	{
		Name:     "storage",
		Prompt:   &survey.Input{Message: "Enter the name of the storage account to use for your stage, must be globally unique or already owned by your subscription: "},
		Validate: utils.ValidateAlphaNumeric,
	},
	{
		Name:     "region",
		Prompt:   &survey.Select{Message: "Enter the region you want to deploy to:", Default: "eastus", Options: azureRegions},
		Validate: survey.Required,
	},
	{
		Name:     "resource",
		Prompt:   &survey.Input{Message: "Enter the name of the resource group to use as default: "},
		Validate: utils.ValidateAlphaNumExtended,
	},
}

func mkAzure(conf config.Config) (prov *AzureProvider, err error) {
	var resp struct {
		Cluster  string
		Storage  string
		Region   string
		Resource string
	}
	err = survey.Ask(azureSurvey, &resp)
	if err != nil {
		return
	}

	subId, tenID, err := GetAzureAccount()
	if err != nil {
		return
	}
	clients, err := GetClientSet(subId)
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
		Provider: AZURE,
		Region:   prov.Region(),
		Context:  prov.Context(),
		Owner:    &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}
	prov.writer = projectManifest.Configure()
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

func (azure *AzureProvider) CreateBackend(prefix string, version string, ctx map[string]interface{}) (string, error) {
	if err := azure.CreateResourceGroup(azure.Project()); err != nil {
		return "", pluralerr.ErrorWrap(err, fmt.Sprintf("Failed to create terraform state resource group %s", azure.Project()))
	}

	if err := azure.CreateBucket(azure.bucket); err != nil {
		return "", pluralerr.ErrorWrap(err, fmt.Sprintf("Failed to create terraform state bucket %s", azure.bucket))
	}

	ctx["Region"] = azure.Region()
	ctx["Bucket"] = azure.Bucket()
	ctx["Prefix"] = prefix
	ctx["ResourceGroup"] = azure.Project()
	ctx["__CLUSTER__"] = azure.Cluster()
	ctx["Context"] = azure.Context()
	if cluster, ok := ctx["cluster"]; ok {
		ctx["Cluster"] = cluster
		ctx["ClusterCreated"] = true
	} else {
		ctx["Cluster"] = fmt.Sprintf(`"%s"`, azure.Cluster())
	}

	scaffold, err := GetProviderScaffold("AZURE", version)
	if err != nil {
		return "", err
	}
	return template.RenderString(scaffold, ctx)
}

func (az *AzureProvider) CreateBucket(bucket string) (err error) {
	acc, err := az.upsertStorageAccount(utils.ToString(az.Context()["StorageAccount"]))
	if err != nil {
		return
	}

	err = az.upsertStorageContainer(acc, bucket)
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
		param := armresources.ResourceGroup{
			Location: to.StringPtr(az.region),
		}

		_, err := az.clients.Groups.CreateOrUpdate(ctx, resourceGroup, param, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (azure *AzureProvider) KubeConfig() error {
	if kubernetes.InKubernetes() {
		return nil
	}

	cmd := exec.Command(
		"az", "aks", "get-credentials", "--overwrite-existing", "--name", azure.cluster, "--resource-group", azure.resourceGroup)
	return utils.Execute(cmd)
}

func (az *AzureProvider) Name() string {
	return AZURE
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

func (azure *AzureProvider) Flush() error {
	if azure.writer == nil {
		return nil
	}
	return azure.writer()
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
	err, resourceGroup := getPathElement(node.Spec.ProviderID, "resourceGroups")
	if err != nil {
		return err
	}
	err, virtualMachineScaleSet := getPathElement(node.Spec.ProviderID, "virtualMachineScaleSets")
	if err != nil {
		return err
	}
	err, InstanceID := getPathElement(node.Spec.ProviderID, "virtualMachines")
	if err != nil {
		return err
	}

	// This method scale down the virtualMachineScaleSet otherwise the VM will be recreated
	pollerDeallocate, err := client.BeginDeallocate(ctx, resourceGroup, virtualMachineScaleSet, &armcompute.VirtualMachineScaleSetsClientBeginDeallocateOptions{
		VMInstanceIDs: &armcompute.VirtualMachineScaleSetVMInstanceIDs{
			InstanceIDs: []*string{to.StringPtr(InstanceID)},
		},
	})
	if err != nil {
		return err
	}
	if _, err = pollerDeallocate.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 1 * time.Second,
	}); err != nil {
		return err
	}

	pollerDelete, err := client.BeginDeleteInstances(ctx, resourceGroup, virtualMachineScaleSet, armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIDs: []*string{
			to.StringPtr(InstanceID)},
	}, &armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions{ForceDeletion: to.BoolPtr(true)})
	if err != nil {
		return err
	}
	if _, err := pollerDelete.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 1 * time.Second,
	}); err != nil {
		return err
	}

	return nil
}

func (az *AzureProvider) getStorageAccount(account string) (storage.Account, error) {
	return az.clients.Accounts.GetProperties(context.Background(), az.resourceGroup, account, storage.AccountExpandBlobRestoreStatus)
}

func (az *AzureProvider) upsertStorageAccount(account string) (storage.Account, error) {
	acc, err := az.getStorageAccount(account)
	if err != nil && !inNotFoundStorageAccount(err) {
		return storage.Account{}, err
	}

	if inNotFoundStorageAccount(err) {
		ctx := context.Background()
		future, err := az.clients.Accounts.Create(
			ctx,
			az.resourceGroup,
			account,
			storage.AccountCreateParameters{
				Sku:                               &storage.Sku{Name: storage.StandardLRS},
				Kind:                              storage.StorageV2,
				Location:                          to.StringPtr(az.region),
				AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
			})

		if err != nil {
			return storage.Account{}, err
		}

		err = future.WaitForCompletionRef(ctx, az.clients.AutorestClient)
		if err != nil {
			return storage.Account{}, err
		}

		return future.Result(az.clients.AccountClient)
	}

	return acc, nil
}

func (az *AzureProvider) upsertStorageContainer(acc storage.Account, name string) error {
	ctx := context.Background()
	accountName := *acc.Name

	resp, err := az.clients.Accounts.ListKeys(ctx, az.resourceGroup, accountName, storage.Kerb)
	if err != nil {
		return err
	}
	key := *(((*resp.Keys)[0]).Value)

	if az.clients.Containers == nil {
		c, _ := azblob.NewSharedKeyCredential(accountName, key)
		p := azblob.NewPipeline(c, azblob.PipelineOptions{})
		u, _ := url.Parse(fmt.Sprintf(`https://%s.blob.core.windows.net`, accountName))
		service := azblob.NewServiceURL(*u, p)
		containerClient := service.NewContainerURL(name)
		az.clients.Containers = containerClient
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
		fmt.Println(out)
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
	var aerr *azure.RequestError
	if err != nil && errors.As(err, &aerr) {
		return aerr.StatusCode == http.StatusNotFound
	}

	return false
}

func authorizer() (autorest.Authorizer, error) {
	if os.Getenv("ARM_USE_MSI") != "" {
		return auth.NewAuthorizerFromEnvironment()
	}

	return auth.NewAuthorizerFromCLI()
}

func getStorageAccountsClient(subscriptionId string) (storage.AccountsClient, error) {
	storageAccountsClient := storage.NewAccountsClient(subscriptionId)
	authorizer, err := authorizer()
	if err != nil {
		return storage.AccountsClient{}, err
	}
	storageAccountsClient.Authorizer = authorizer
	return storageAccountsClient, nil
}

func getResourceGroupClient(subscriptionId string) (*armresources.ResourceGroupsClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, pluralerr.ErrorWrap(err, "getting resource group client failed with")
	}
	groupClient, err := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	return groupClient, nil
}

func getPathElement(path, indexName string) (error, string) {
	pattern := fmt.Sprintf(`.*\/%s\/(?P<element>([\w'-]+))`, indexName)
	captureGroupRegex := regexp.MustCompile(pattern)
	match := captureGroupRegex.FindStringSubmatch(path)
	if match != nil {
		index := captureGroupRegex.SubexpIndex("element")
		if index >= 0 {
			return nil, match[index]
		}
	}

	return fmt.Errorf("%s not found", indexName), ""
}
