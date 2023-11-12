package cd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	gqlclient "github.com/pluralsh/console-client-go"

	"github.com/pluralsh/plural/pkg/api"
)

func AskCloudProviderSettings(provider string) (*gqlclient.CloudProviderSettingsAttributes, error) {
	switch provider {
	case api.ProviderAWS:
		if acs, err := askAWSCloudProviderSettings(); err != nil {
			return nil, err
		} else {
			return &gqlclient.CloudProviderSettingsAttributes{Aws: acs}, nil
		}
	case api.ProviderAzure:
		if acs, err := askAzureCloudProviderSettings(); err != nil {
			return nil, err
		} else {
			return &gqlclient.CloudProviderSettingsAttributes{Azure: acs}, nil
		}
	case api.ProviderGCP:
		if gcs, err := askGCPCloudProviderSettings(); err != nil {
			return nil, err
		} else {
			return &gqlclient.CloudProviderSettingsAttributes{Gcp: gcs}, nil
		}
	}

	return nil, fmt.Errorf("unknown provider")
}

func askAWSCloudProviderSettings() (*gqlclient.AwsSettingsAttributes, error) {
	awsSurvey := []*survey.Question{
		{
			Name:   "key",
			Prompt: &survey.Input{Message: "Enter the Access Key ID:"},
		},
		{
			Name:   "secret",
			Prompt: &survey.Input{Message: "Enter Secret Access Key:"},
		},
	}
	var resp struct {
		Key    string
		Secret string
	}
	if err := survey.Ask(awsSurvey, &resp); err != nil {
		return nil, err
	}
	return &gqlclient.AwsSettingsAttributes{
		AccessKeyID:     resp.Key,
		SecretAccessKey: resp.Secret,
	}, nil
}

func askAzureCloudProviderSettings() (*gqlclient.AzureSettingsAttributes, error) {
	azureSurvey := []*survey.Question{
		{
			Name:   "tenant",
			Prompt: &survey.Input{Message: "Enter the tenant ID:"},
		},
		{
			Name:   "client",
			Prompt: &survey.Input{Message: "Enter the client ID:"},
		},
		{
			Name:   "secret",
			Prompt: &survey.Input{Message: "Enter the client secret:"},
		},
		{
			Name:   "subscription",
			Prompt: &survey.Input{Message: "Enter the subscription ID:"},
		},
	}
	var resp struct {
		Tenant       string
		Client       string
		Secret       string
		Subscription string
	}
	if err := survey.Ask(azureSurvey, &resp); err != nil {
		return nil, err
	}
	return &gqlclient.AzureSettingsAttributes{
		TenantID:       resp.Tenant,
		ClientID:       resp.Client,
		ClientSecret:   resp.Secret,
		SubscriptionID: resp.Subscription,
	}, nil
}

func askGCPCloudProviderSettings() (*gqlclient.GcpSettingsAttributes, error) {
	applicationCredentialsFilePath := ""

	prompt := &survey.Input{
		Message: "Enter GCP application credentials file path:",
	}
	if err := survey.AskOne(prompt, &applicationCredentialsFilePath, survey.WithValidator(validServiceAccountCredentials)); err != nil {
		return nil, err
	}

	return &gqlclient.GcpSettingsAttributes{
		ApplicationCredentials: toCredentialsJSON(applicationCredentialsFilePath),
	}, nil
}

type credentials struct {
	Email string          `json:"client_email"`
	ID    string          `json:"client_id"`
	Type  credentialsType `json:"type"`
}

type credentialsType = string

const (
	ServiceAccountType credentialsType = "service_account"
)

func validServiceAccountCredentials(val interface{}) error {
	path, _ := val.(string)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	creds := new(credentials)
	if err = json.Unmarshal(bytes, creds); err != nil {
		return err
	}

	if creds.Type != ServiceAccountType || len(creds.Email) == 0 || len(creds.ID) == 0 {
		return fmt.Errorf("provided credentials file is not a valid service account. Must have type 'service_account' and both 'client_id' and 'client_email' set")
	}

	return nil
}

func toCredentialsJSON(path string) string {
	bytes, _ := os.ReadFile(path)
	return string(bytes)
}
