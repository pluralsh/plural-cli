package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/azure/azure-sdk-for-go/services/compute/mgmt/2021-07-01/compute"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/errors"
	v1 "k8s.io/api/core/v1"
)

type AzureProvider struct {
	cluster       string
	resourceGroup string
	bucket        string
	region        string
	ctx           map[string]interface{}
	writer        manifest.Writer
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

	subId, tenID, err := getAzureAccount()
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

func azureFromManifest(man *manifest.ProjectManifest) (*AzureProvider, error) {
	return &AzureProvider{man.Cluster, man.Project, man.Bucket, man.Region, man.Context, nil}, nil
}

func (azure *AzureProvider) CreateBackend(prefix string, ctx map[string]interface{}) (string, error) {
	if err := azure.CreateBucket(azure.bucket); err != nil {
		return "", errors.ErrorWrap(err, fmt.Sprintf("Failed to create terraform state bucket %s", azure.bucket))
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

	scaffold, err := GetProviderScaffold("AZURE")
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

func (azure *AzureProvider) KubeConfig() error {
	if utils.InKubernetes() {
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
	ctx := context.Background()
	vms := compute.NewVirtualMachinesClient(utils.ToString(az.ctx["SubscriptionId"]))
	fut, err := vms.Delete(ctx, az.Project(), node.Name, to.BoolPtr(true))
	if err != nil {
		return errors.ErrorWrap(err, "failed to call deletion api")
	}

	err = fut.WaitForCompletionRef(ctx, vms.Client)
	return errors.ErrorWrap(err, "vm deletion failed with")
}

func (az *AzureProvider) Authorizer() (autorest.Authorizer, error) {
	if os.Getenv("ARM_USE_MSI") != "" {
		return auth.NewAuthorizerFromEnvironment()
	}

	return auth.NewAuthorizerFromCLI()
}

func (az *AzureProvider) getStorageAccountsClient() storage.AccountsClient {
	storageAccountsClient := storage.NewAccountsClient(utils.ToString(az.ctx["SubscriptionId"]))
	authorizer, _ := az.Authorizer()
	storageAccountsClient.Authorizer = authorizer
	return storageAccountsClient
}

func (az *AzureProvider) getStorageAccount(account string) (storage.Account, error) {
	client := az.getStorageAccountsClient()
	return client.GetProperties(context.Background(), az.resourceGroup, account, storage.AccountExpandBlobRestoreStatus)
}

func (az *AzureProvider) upsertStorageAccount(account string) (acc storage.Account, err error) {
	acc, err = az.getStorageAccount(account)
	if err == nil {
		return
	}

	client := az.getStorageAccountsClient()
	ctx := context.Background()
	future, err := client.Create(
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
		return
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return
	}

	acc, err = future.Result(client)
	return
}

func (az *AzureProvider) upsertStorageContainer(acc storage.Account, name string) error {
	ctx := context.Background()
	accountName := *acc.Name

	client := az.getStorageAccountsClient()
	resp, err := client.ListKeys(ctx, az.resourceGroup, accountName, storage.Kerb)
	if err != nil {
		return err
	}
	key := *(((*resp.Keys)[0]).Value)

	c, _ := azblob.NewSharedKeyCredential(accountName, key)
	p := azblob.NewPipeline(c, azblob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(`https://%s.blob.core.windows.net`, accountName))
	service := azblob.NewServiceURL(*u, p)

	container := service.NewContainerURL(name)
	_, err = container.GetProperties(ctx, azblob.LeaseAccessConditions{})
	if err == nil {
		return err
	}

	_, err = container.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	return err
}

func getAzureAccount() (string, string, error) {
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
