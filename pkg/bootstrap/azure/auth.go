package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	azwi "github.com/Azure/azure-workload-identity/pkg/cloud"
	"github.com/Azure/go-autorest/autorest/azure"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type AuthService struct {
	subscriptionID string

	azwiClient    *azwi.AzureClient
	msgraphClient *msgraph.GraphServiceClient
	context       context.Context

	app models.Applicationable
	sp  models.ServicePrincipalable
}

func GetAuthService(subscriptionID string) (*AuthService, error) {
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

	return &AuthService{
		subscriptionID: subscriptionID,
		msgraphClient:  msgraphClient,
		azwiClient:     azwiClient,
		context:        context.Background(),
	}, nil
}

func (as *AuthService) addServicePrincipalPassword(servicePrincipalId string) (models.PasswordCredentialable, error) {
	pwd := serviceprincipals.NewItemAddPasswordPostRequestBody()
	pwd.SetPasswordCredential(models.NewPasswordCredential())

	return as.msgraphClient.ServicePrincipalsById(servicePrincipalId).AddPassword().
		Post(as.context, pwd, nil)
}

func (as *AuthService) Setup(name string) (clientId string, clientSecret string, err error) {
	app, err := as.azwiClient.CreateApplication(as.context, name)
	if err != nil {
		return
	}
	as.app = app
	utils.Success("Created %s application\n", *app.GetDisplayName())

	sp, err := as.azwiClient.CreateServicePrincipal(as.context, *app.GetAppId(), nil)
	if err != nil {
		return
	}
	as.sp = sp
	utils.Success("Created %s service principal\n", *sp.GetDisplayName())

	role := "Contributor"
	scope := fmt.Sprintf("/subscriptions/%s/", as.subscriptionID)
	_, err = as.azwiClient.CreateRoleAssignment(as.context, scope, role, *sp.GetId())
	if err != nil {
		return
	}
	utils.Success("Assigned %s role to %s service principal\n", role, *sp.GetDisplayName())

	pwd, err := as.addServicePrincipalPassword(*sp.GetId())
	if err != nil {
		return
	}
	utils.Success("Added password for %s service principal\n", *sp.GetDisplayName())

	clientId = *sp.GetAppId()
	clientSecret = *pwd.GetSecretText()

	return
}

func (as *AuthService) Cleanup() error {
	if as.sp != nil {
		err := as.azwiClient.DeleteServicePrincipal(as.context, *as.sp.GetId())
		if err != nil {
			return err
		}
		utils.Success("Deleted %s service principal\n", *as.sp.GetDisplayName())
	}

	if as.app != nil {
		err := as.azwiClient.DeleteApplication(as.context, *as.app.GetId())
		if err != nil {
			return err
		}
		utils.Success("Deleted %s application\n", *as.app.GetDisplayName())
	}

	return nil
}
