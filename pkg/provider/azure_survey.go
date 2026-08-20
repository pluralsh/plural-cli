package provider

import (
	"context"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

// CreateNewOption is the survey sentinel for typing a new Azure name.
const CreateNewOption = "Create new..."

func filterSurveyOptions(filter string, value string, index int) (include bool) {
	if value == CreateNewOption {
		return true
	}

	filter = strings.ToLower(filter)

	return strings.Contains(strings.ToLower(value), filter)
}

func askCluster() (string, error) {
	cluster := ""
	if err := survey.AskOne(
		&survey.Input{Message: "Enter the name of your cluster:"},
		&cluster, survey.WithValidator(validCluster),
	); err != nil {
		return "", err
	}

	return cluster, nil
}

// AzureLocations lists subscription locations used by plural up.
func AzureLocations(ctx context.Context, client *armsubscription.SubscriptionsClient, subscriptionID string) ([]string, error) {
	locations := make([]string, 0)
	pager := client.NewListLocationsPager(subscriptionID, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Value {
			if v != nil {
				locations = append(locations, *v.Name)
			}
		}
	}

	return locations, nil
}

func askAzureLocation(ctx context.Context, client *armsubscription.SubscriptionsClient, subscriptionID string) (string, error) {
	options, err := AzureLocations(ctx, client, subscriptionID)
	if err != nil {
		return "", err
	}

	location := ""
	if err = survey.AskOne(
		&survey.Select{Message: "Select the location you want to deploy to:", Options: options, Default: "eastus"},
		&location, survey.WithValidator(survey.Required),
	); err != nil {
		return "", err
	}

	return location, nil
}

// AzureResourceGroups lists resource groups for the signed-in subscription.
func AzureResourceGroups(ctx context.Context, client *armresources.ResourceGroupsClient) ([]string, error) {
	groups := make([]string, 0)
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Value {
			if v != nil {
				groups = append(groups, *v.Name)
			}
		}
	}

	return groups, nil
}

// AzureResourceGroupChoices is the plural-up resource-group select list (existing + Create new…).
func AzureResourceGroupChoices(ctx context.Context, client *armresources.ResourceGroupsClient) ([]string, error) {
	options, err := AzureResourceGroups(ctx, client)
	if err != nil {
		return nil, err
	}
	return append(options, CreateNewOption), nil
}

func askAzureResourceGroup(ctx context.Context, client *armresources.ResourceGroupsClient) (string, error) {
	options, err := AzureResourceGroupChoices(ctx, client)
	if err != nil {
		return "", err
	}

	group := ""
	if err = survey.AskOne(
		&survey.Select{Message: "Select the resource group to use:", Options: options},
		&group, survey.WithValidator(survey.Required), survey.WithFilter(filterSurveyOptions),
	); err != nil {
		return "", err
	}

	if group == CreateNewOption {
		if err = survey.AskOne(&survey.Input{Message: "Enter resource group name:"}, &group, survey.WithValidator(utils.ValidateResourceGroupName)); err != nil {
			return "", err
		}
	}

	return group, nil
}

// AzureStorageAccounts lists storage accounts for the signed-in subscription.
func AzureStorageAccounts(ctx context.Context, client *armstorage.AccountsClient) ([]string, error) {
	accounts := make([]string, 0)
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Value {
			if v != nil {
				accounts = append(accounts, *v.Name)
			}
		}
	}

	return accounts, nil
}

// AzureStorageAccountChoices is the plural-up storage-account select list (existing + Create new…).
func AzureStorageAccountChoices(ctx context.Context, client *armstorage.AccountsClient) ([]string, error) {
	options, err := AzureStorageAccounts(ctx, client)
	if err != nil {
		return nil, err
	}
	return append(options, CreateNewOption), nil
}

func askAzureStorageAccount(ctx context.Context, client *armstorage.AccountsClient) (string, error) {
	options, err := AzureStorageAccountChoices(ctx, client)
	if err != nil {
		return "", err
	}

	account := ""
	if err = survey.AskOne(
		&survey.Select{Message: "Select the storage account to use:", Options: options},
		&account, survey.WithValidator(survey.Required), survey.WithFilter(filterSurveyOptions),
	); err != nil {
		return "", err
	}

	if account == CreateNewOption {
		if err = survey.AskOne(&survey.Input{Message: "Enter globally unique storage account name:"}, &account, survey.WithValidator(utils.ValidateStorageAccountName)); err != nil {
			return "", err
		}
	}

	return account, nil
}
