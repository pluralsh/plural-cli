package provider

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/pluralsh/plural/pkg/utils"
)

type AzureGraphService struct {
	client  msgraphsdk.GraphServiceClient
	context context.Context
}

func GetAzureGraphService() (*AzureGraphService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	graphServiceClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, nil)
	if err != nil {
		return nil, err
	}

	return &AzureGraphService{
		client:  *graphServiceClient,
		context: context.Background(),
	}, nil
}

func (ags *AzureGraphService) CreateAzureApplication(name string) (models.Applicationable, error) {
	app := models.NewApplication()
	app.SetDisplayName(&name)

	return ags.client.Applications().Post(ags.context, app, nil)
}

func (ags *AzureGraphService) CreateServicePrincipal(name, applicationId string) (models.ServicePrincipalable, error) {
	sp := models.NewServicePrincipal()
	sp.SetAppId(&applicationId)
	sp.SetDisplayName(&name)

	return ags.client.ServicePrincipals().Post(ags.context, sp, nil)
}

func (ags *AzureGraphService) CreateServicePrincipalPasswordCredential(name, servicePrincipalId string) (models.PasswordCredentialable, error) {
	passwordCredential := models.NewPasswordCredential()
	passwordCredential.SetDisplayName(&name)

	requestBody := serviceprincipals.NewItemAddPasswordPostRequestBody()
	requestBody.SetPasswordCredential(passwordCredential)

	return ags.client.ServicePrincipals().ByServicePrincipalId(servicePrincipalId).AddPassword().
		Post(context.Background(), requestBody, nil)
}

func (ags *AzureGraphService) SetupServicePrincipal(name string) (clientId string, clientSecret string, err error) {
	app, err := ags.CreateAzureApplication(name)
	if err != nil {
		return
	}
	utils.Success("Created %s application\n", *app.GetDisplayName())

	sp, err := ags.CreateServicePrincipal(name, *app.GetAppId())
	if err != nil {
		return
	}
	utils.Success("Created %s service principal\n", *sp.GetDisplayName())

	pwd, err := ags.CreateServicePrincipalPasswordCredential(name, *sp.GetId())
	if err != nil {
		return
	}
	utils.Success("Created password for %s service principal\n", *sp.GetDisplayName())

	clientId = pwd.GetKeyId().String()
	clientSecret = *pwd.GetSecretText()

	return
}
