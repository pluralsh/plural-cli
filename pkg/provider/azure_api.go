package provider

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

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
