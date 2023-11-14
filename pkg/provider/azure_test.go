package provider_test

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func TestCreateResourceGroup(t *testing.T) {
	resourceGroupName := "test"
	tests := []struct {
		name                              string
		expectedError                     string
		expectedCreateOrUpdateCalledCount int
		resourceGroupName                 string
		existingResourceGroup             *armresources.ResourceGroup
		manifest                          *manifest.ProjectManifest
	}{
		{
			name:                              `create resource group when doesn't exist`,
			resourceGroupName:                 "test",
			manifest:                          &manifest.ProjectManifest{},
			expectedCreateOrUpdateCalledCount: 1,
		},
		{
			name:              `resource group exists`,
			resourceGroupName: resourceGroupName,
			existingResourceGroup: &armresources.ResourceGroup{
				Name: &resourceGroupName,
			},
			manifest:                          &manifest.ProjectManifest{},
			expectedCreateOrUpdateCalledCount: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientSet := getFakeClientSetWithGroupsClient(test.existingResourceGroup)

			prov, err := provider.AzureFromManifest(test.manifest, clientSet)
			assert.NoError(t, err)

			err = prov.CreateResourceGroup(test.resourceGroupName)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
			}

			fakeClient, ok := clientSet.Groups.(*fakeGroupsClient)
			if !ok {
				t.Fatalf("failed to access underlying fake GroupsClient")
			}

			assert.Equal(t, fakeClient.CreateOrUpdateCalledCount, test.expectedCreateOrUpdateCalledCount)
		})
	}
}

func TestCreateBucket(t *testing.T) {
	accountName := "test"
	secondAccountName := "second"
	tests := []struct {
		name                               string
		expectedError                      string
		expectedAccountCreateCalledCount   int
		expectedContainerCreateCalledCount int
		accountName                        string
		existingAccount                    *armstorage.Account
		manifest                           *manifest.ProjectManifest
	}{
		{
			name:        `create account when doesn't exist`,
			accountName: accountName,
			existingAccount: &armstorage.Account{
				Name: &secondAccountName,
			},
			manifest: &manifest.ProjectManifest{
				Context: map[string]interface{}{
					"StorageAccount": "test",
				},
			},
			expectedAccountCreateCalledCount:   1,
			expectedContainerCreateCalledCount: 1,
		},
		{
			name:        `storage account exists`,
			accountName: accountName,
			existingAccount: &armstorage.Account{
				Name: &accountName,
			},
			manifest: &manifest.ProjectManifest{
				Context: map[string]interface{}{
					"StorageAccount": "test",
				},
			},
			expectedAccountCreateCalledCount:   0,
			expectedContainerCreateCalledCount: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientSet := getFakeClientSetWithAccountsClient(test.existingAccount)

			prov, err := provider.AzureFromManifest(test.manifest, clientSet)
			assert.NoError(t, err)

			err = prov.CreateBucket("test")
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
			}

			fakeAccountsClient, ok := clientSet.Accounts.(*fakeAccountsClient)
			if !ok {
				t.Fatalf("failed to access underlying fake AccountsClient")
			}
			fakeContainerClient, ok := clientSet.Containers.(*fakeContainersClient)
			if !ok {
				t.Fatalf("failed to access underlying fake ContainersClient")
			}

			assert.Equal(t, fakeAccountsClient.CreateCalledCount, test.expectedAccountCreateCalledCount, "expected Create call for Accounts")
			assert.Equal(t, fakeContainerClient.CreateCalledCount, test.expectedContainerCreateCalledCount, "expected Create call for Containers")
		})
	}
}
