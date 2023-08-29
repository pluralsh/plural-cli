package bootstrap

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

type AzureCredentialsService struct {
	subscriptionID string

	azwiClient    *azwi.AzureClient
	msgraphClient *msgraph.GraphServiceClient
	context       context.Context

	app models.Applicationable
	sp  models.ServicePrincipalable
}

func GetAzureCredentialsService(subscriptionID string) (*AzureCredentialsService, error) {
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

	return &AzureCredentialsService{
		subscriptionID: subscriptionID,
		msgraphClient:  msgraphClient,
		azwiClient:     azwiClient,
		context:        context.Background(),
	}, nil
}

func (acs *AzureCredentialsService) addServicePrincipalPassword(servicePrincipalId string) (models.PasswordCredentialable, error) {
	pwd := serviceprincipals.NewItemAddPasswordPostRequestBody()
	pwd.SetPasswordCredential(models.NewPasswordCredential())

	return acs.msgraphClient.ServicePrincipalsById(servicePrincipalId).AddPassword().
		Post(acs.context, pwd, nil)
}

func (acs *AzureCredentialsService) Setup(name string) (clientId string, clientSecret string, err error) {
	app, err := acs.azwiClient.CreateApplication(acs.context, name)
	if err != nil {
		return
	}
	acs.app = app
	utils.Success("Created %s application\n", *app.GetDisplayName())

	sp, err := acs.azwiClient.CreateServicePrincipal(acs.context, *app.GetAppId(), nil)
	if err != nil {
		return
	}
	acs.sp = sp
	utils.Success("Created %s service principal\n", *sp.GetDisplayName())

	role := "Contributor"
	scope := fmt.Sprintf("/subscriptions/%s/", acs.subscriptionID)
	_, err = acs.azwiClient.CreateRoleAssignment(acs.context, scope, role, *sp.GetId())
	if err != nil {
		return
	}
	utils.Success("Assigned %s role to %s service principal\n", role, *sp.GetDisplayName())

	pwd, err := acs.addServicePrincipalPassword(*sp.GetId())
	if err != nil {
		return
	}
	utils.Success("Added password for %s service principal\n", *sp.GetDisplayName())

	clientId = *sp.GetAppId()
	clientSecret = *pwd.GetSecretText()

	return
}

func (acs *AzureCredentialsService) Cleanup() error {
	if acs.sp != nil {
		err := acs.azwiClient.DeleteServicePrincipal(acs.context, *acs.sp.GetId())
		if err != nil {
			return err
		}
		utils.Success("Deleted %s service principal\n", *acs.sp.GetDisplayName())
	}

	if acs.app != nil {
		err := acs.azwiClient.DeleteApplication(acs.context, *acs.app.GetId())
		if err != nil {
			return err
		}
		utils.Success("Deleted %s application\n", *acs.app.GetDisplayName())
	}

	return nil
}
