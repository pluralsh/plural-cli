package provider

import (
	"context"
	"fmt"
)

type azureSetup struct{}

func (azureSetup) Name() string { return "azure" }

func (azureSetup) Schema() []SetupField {
	return []SetupField{
		{Key: "cluster", Label: "Cluster name", Placeholder: "max 15 chars", Required: true},
		{Key: "location", Label: "Location", Placeholder: "e.g. eastus", Default: "eastus", Required: true},
		{Key: "resourceGroup", Label: "Resource group", Placeholder: "existing or new name", Required: true},
		{Key: "storageAccount", Label: "Storage account", Placeholder: "globally unique name", Required: true},
	}
}

func (s azureSetup) Probe(ctx context.Context) (SetupResult, error) {
	subID, tenID, subName, err := GetAzureAccount()
	if err != nil {
		return SetupResult{}, fmt.Errorf("Azure login (az account show): %w", err)
	}
	user, err := GetAzureUser()
	if err != nil {
		return SetupResult{}, fmt.Errorf("Azure user (az ad signed-in-user show): %w", err)
	}
	clients, err := GetClientSet(subID)
	if err != nil {
		return SetupResult{}, fmt.Errorf("Azure clients: %w", err)
	}

	locations, err := AzureLocations(ctx, clients.Subscriptions, subID)
	if err != nil {
		return SetupResult{}, fmt.Errorf("Azure locations: %w", err)
	}
	groups, err := AzureResourceGroupChoices(ctx, clients.Groups)
	if err != nil {
		return SetupResult{}, fmt.Errorf("Azure resource groups: %w", err)
	}
	accounts, err := AzureStorageAccountChoices(ctx, clients.Accounts)
	if err != nil {
		return SetupResult{}, fmt.Errorf("Azure storage accounts: %w", err)
	}

	fields := s.Schema()
	fields = withOptions(fields, "location", locations)
	fields = withOptions(fields, "resourceGroup", groups)
	fields = withOptions(fields, "storageAccount", accounts)

	return SetupResult{
		Summary: fmt.Sprintf("%s · subscription %s (%s) · tenant %s",
			user, subName, truncateMiddle(subID, 12), truncateMiddle(tenID, 12)),
		Fields: fields,
	}, nil
}

func (azureSetup) Options(ctx context.Context, fieldKey string, _ map[string]string) ([]string, error) {
	subID, _, _, err := GetAzureAccount()
	if err != nil {
		return nil, err
	}
	clients, err := GetClientSet(subID)
	if err != nil {
		return nil, err
	}
	switch fieldKey {
	case "location":
		return AzureLocations(ctx, clients.Subscriptions, subID)
	case "resourceGroup":
		return AzureResourceGroupChoices(ctx, clients.Groups)
	case "storageAccount":
		return AzureStorageAccountChoices(ctx, clients.Accounts)
	default:
		return nil, nil
	}
}

// Preflights: AzureProvider.Preflights is empty — nothing to run.
func (azureSetup) Preflights(context.Context, map[string]string) error { return nil }
