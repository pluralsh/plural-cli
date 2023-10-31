package cd

import (
	"fmt"

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
	}
	var resp struct {
		Tenant string
		Client string
		Secret string
	}
	if err := survey.Ask(azureSurvey, &resp); err != nil {
		return nil, err
	}
	return &gqlclient.AzureSettingsAttributes{
		TenantID:     resp.Tenant,
		ClientID:     resp.Client,
		ClientSecret: resp.Secret,
	}, nil
}

func askGCPCloudProviderSettings() (*gqlclient.GcpSettingsAttributes, error) {
	applicationCredentials := ""
	prompt := &survey.Input{
		Message: "Enter GCP application credentials:",
	}
	if err := survey.AskOne(prompt, &applicationCredentials, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}
	return &gqlclient.GcpSettingsAttributes{
		ApplicationCredentials: applicationCredentials,
	}, nil
}
