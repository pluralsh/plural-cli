package provider

import (
	"context"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

const createNewOption = "Create new..."

func azureLocations(ctx context.Context, client *armsubscription.SubscriptionsClient, subscriptionID string) ([]string, error) {
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
	options, err := azureLocations(ctx, client, subscriptionID)
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

func azureResourceGroups(ctx context.Context, client *armresources.ResourceGroupsClient) ([]string, error) {
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

func askAzureResourceGroup(ctx context.Context, client *armresources.ResourceGroupsClient) (string, error) {
	options, err := azureResourceGroups(ctx, client)
	if err != nil {
		return "", err
	}
	options = append(options, createNewOption)

	group := ""
	if err = survey.AskOne(
		&survey.Select{Message: "Select the resource group to use:", Options: options},
		&group, survey.WithValidator(survey.Required),
	); err != nil {
		return "", err
	}

	if group == createNewOption {
		if err = survey.AskOne(&survey.Input{Message: "Enter resource group name:"}, &group); err != nil {
			return "", err
		}
	}

	return group, nil
}

func azureStorageAccounts(ctx context.Context, client *armstorage.AccountsClient) ([]string, error) {
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

func askAzureStorageAccount(ctx context.Context, client *armstorage.AccountsClient) (string, error) {
	options, err := azureStorageAccounts(ctx, client)
	if err != nil {
		return "", err
	}
	options = append(options, createNewOption)

	account := ""
	if err = survey.AskOne(
		&survey.Select{Message: "Select the storage account to use:", Options: options},
		&account, survey.WithValidator(survey.Required),
	); err != nil {
		return "", err
	}

	if account == createNewOption {
		if err = survey.AskOne(&survey.Input{Message: "Enter globally unique storage account name:"}, &account); err != nil {
			return "", err
		}
	}

	return account, nil
}
