package provider

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	azwi "github.com/Azure/azure-workload-identity/pkg/cloud"
	"github.com/Azure/go-autorest/autorest/azure"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/pluralsh/plural/pkg/utils"
)

type AzureService struct {
	subscriptionID string
	azwiClient     *azwi.AzureClient
	msgraphClient  *msgraph.GraphServiceClient
	context        context.Context
}

func GetAzureService(subscriptionID string) (*AzureService, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	azwiClient, err := azwi.NewAzureClientWithCLI(azure.PublicCloud, subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	msgraphClient, err := msgraph.NewGraphServiceClientWithCredentials(credential, nil)
	if err != nil {
		return nil, err
	}

	return &AzureService{
		subscriptionID: subscriptionID,
		msgraphClient:  msgraphClient,
		azwiClient:     azwiClient,
		context:        context.Background(),
	}, nil
}

func (as *AzureService) AddServicePrincipalPassword(name, servicePrincipalId string) (models.PasswordCredentialable, error) {
	passwordCredential := models.NewPasswordCredential()
	passwordCredential.SetDisplayName(&name)

	body := serviceprincipals.NewItemAddPasswordPostRequestBody()
	body.SetPasswordCredential(passwordCredential)

	return as.msgraphClient.ServicePrincipalsById(servicePrincipalId).AddPassword().
		Post(as.context, body, nil)
}

func (as *AzureService) SetupServicePrincipal(name string) (clientId string, clientSecret string, err error) {
	app, err := as.azwiClient.CreateApplication(as.context, name)
	if err != nil {
		return
	}
	utils.Success("Created %s application\n", *app.GetDisplayName())

	sp, err := as.azwiClient.CreateServicePrincipal(as.context, *app.GetAppId(), nil)
	if err != nil {
		return
	}
	utils.Success("Created %s service principal\n", *sp.GetDisplayName())

	role := "Contributor"
	scope := fmt.Sprintf("/subscriptions/%s/", as.subscriptionID)
	_, err = as.azwiClient.CreateRoleAssignment(as.context, scope, role, *sp.GetId())
	if err != nil {
		return
	}
	utils.Success("Assigned %s role to %s service principal\n", role, *sp.GetDisplayName())

	pwd, err := as.AddServicePrincipalPassword(name, *sp.GetId())
	if err != nil {
		return
	}
	utils.Success("Created password for %s service principal\n", *sp.GetDisplayName())

	clientId = pwd.GetKeyId().String()
	clientSecret = *pwd.GetSecretText()

	return
}
