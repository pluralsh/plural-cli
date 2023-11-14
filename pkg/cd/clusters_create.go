package cd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func AskCloudSettings(provider string) (*gqlclient.CloudSettingsAttributes, error) {
	switch provider {
	case api.ProviderAWS:
		if acs, err := askAWSCloudSettings(); err != nil {
			return nil, err
		} else {
			return &gqlclient.CloudSettingsAttributes{Aws: acs}, nil
		}
	case api.ProviderAzure:
		if acs, err := askAzureCloudSettings(); err != nil {
			return nil, err
		} else {
			return &gqlclient.CloudSettingsAttributes{Azure: acs}, nil
		}
	case api.ProviderGCP:
		if gcs, err := askGCPCloudSettings(); err != nil {
			return nil, err
		} else {
			return &gqlclient.CloudSettingsAttributes{Gcp: gcs}, nil
		}
	}

	return nil, fmt.Errorf("unknown provider")
}

func askAWSCloudSettings() (*gqlclient.AwsCloudAttributes, error) {
	region := ""
	prompt := &survey.Input{
		Message: "Enter AWS region:",
	}
	if err := survey.AskOne(prompt, &region, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}
	return &gqlclient.AwsCloudAttributes{
		Region: &region,
	}, nil
}

func askAzureCloudSettings() (*gqlclient.AzureCloudAttributes, error) {
	azureSurvey := []*survey.Question{
		{
			Name:   "location",
			Prompt: &survey.Input{Message: "Enter the location:"},
		},
		{
			Name:   "subscription",
			Prompt: &survey.Input{Message: "Enter the subscription ID:"},
		},
		{
			Name:   "resource",
			Prompt: &survey.Input{Message: "Enter the resource group:"},
		},
		{
			Name:   "network",
			Prompt: &survey.Input{Message: "Enter the network name:"},
		},
	}
	var resp struct {
		Location     string
		Subscription string
		Resource     string
		Network      string
	}
	if err := survey.Ask(azureSurvey, &resp); err != nil {
		return nil, err
	}
	return &gqlclient.AzureCloudAttributes{
		Location:       &resp.Location,
		SubscriptionID: &resp.Subscription,
		ResourceGroup:  &resp.Resource,
		Network:        &resp.Network,
	}, nil
}

func askGCPCloudSettings() (*gqlclient.GcpCloudAttributes, error) {
	awsSurvey := []*survey.Question{
		{
			Name:   "project",
			Prompt: &survey.Input{Message: "Enter the project name:"},
		},
		{
			Name:   "network",
			Prompt: &survey.Input{Message: "Enter the network name:"},
		},
		{
			Name:   "region",
			Prompt: &survey.Input{Message: "Enter the region:"},
		},
	}
	var resp struct {
		Project string
		Network string
		Region  string
	}
	if err := survey.Ask(awsSurvey, &resp); err != nil {
		return nil, err
	}
	return &gqlclient.GcpCloudAttributes{
		Project: &resp.Project,
		Network: &resp.Network,
		Region:  &resp.Region,
	}, nil
}
