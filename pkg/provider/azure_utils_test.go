package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/pluralsh/plural/pkg/provider"
)

type fakeAccountsClient struct {
	storage.AccountsClient
	Account           *storage.Account
	CreateCalledCount int
}

type fakeGroupsClient struct {
	armresources.ResourceGroupsClient
	Group                     *armresources.ResourceGroup
	CreateOrUpdateCalledCount int
}

type fakeContainersClient struct {
	CreateCalledCount int
}

func getFakeClientSetWithGroupsClient(existingGroup *armresources.ResourceGroup) *provider.ClientSet {
	return &provider.ClientSet{
		Groups: &fakeGroupsClient{
			Group: existingGroup,
		},
	}
}

func getFakeClientSetWithAccountsClient(existingAccount *storage.Account) *provider.ClientSet {
	return &provider.ClientSet{
		Accounts: &fakeAccountsClient{
			Account:           existingAccount,
			CreateCalledCount: 0,
		},
		Containers: &fakeContainersClient{},
	}
}

func (f *fakeContainersClient) GetProperties(_ context.Context, _ azblob.LeaseAccessConditions) (*azblob.ContainerGetPropertiesResponse, error) {
	return nil, fmt.Errorf("some error")
}

func (f *fakeContainersClient) Create(_ context.Context, _ azblob.Metadata, _ azblob.PublicAccessType) (*azblob.ContainerCreateResponse, error) {
	f.CreateCalledCount++
	return nil, nil
}

func (a *fakeAccountsClient) GetProperties(_ context.Context, _ string, name string, _ storage.AccountExpand) (result storage.Account, err error) {
	if a.Account != nil && a.Account.Name != nil && *a.Account.Name == name {
		return *a.Account, nil
	}

	return storage.Account{}, &azure.RequestError{
		DetailedError: autorest.DetailedError{StatusCode: http.StatusNotFound},
	}
}

func (a *fakeAccountsClient) Create(_ context.Context, _ string, _ string, _ storage.AccountCreateParameters) (result storage.AccountsCreateFuture, err error) {
	a.CreateCalledCount++
	return storage.AccountsCreateFuture{
		FutureAPI: &fakeFutureAPI{},
		Result: func(storage.AccountsClient) (storage.Account, error) {
			return *a.Account, nil
		},
	}, nil
}

func (a *fakeAccountsClient) ListKeys(_ context.Context, _ string, _ string, _ storage.ListKeyExpand) (result storage.AccountListKeysResult, err error) {
	keyName := "name"
	keyValue := "value"
	keys := []storage.AccountKey{{
		KeyName:     &keyName,
		Value:       &keyValue,
		Permissions: "",
	}}
	return storage.AccountListKeysResult{
		Response: autorest.Response{},
		Keys:     &keys,
	}, nil
}

func (c *fakeGroupsClient) CreateOrUpdate(_ context.Context, _ string, parameters armresources.ResourceGroup, _ *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	c.CreateOrUpdateCalledCount++
	c.Group = &parameters

	return armresources.ResourceGroupsClientCreateOrUpdateResponse{
		ResourceGroup: *c.Group,
	}, nil
}

func (c *fakeGroupsClient) Get(_ context.Context, resourceGroupName string, _ *armresources.ResourceGroupsClientGetOptions) (armresources.ResourceGroupsClientGetResponse, error) {

	if c.Group != nil && c.Group.Name != nil && resourceGroupName == *c.Group.Name {
		return armresources.ResourceGroupsClientGetResponse{
			ResourceGroup: *c.Group,
		}, nil
	}

	return armresources.ResourceGroupsClientGetResponse{}, &azcore.ResponseError{
		StatusCode: http.StatusNotFound,
	}

}

type fakeFutureAPI struct {
}

func (f *fakeFutureAPI) Response() *http.Response {
	return nil
}

func (f *fakeFutureAPI) Status() string {
	return ""
}

func (f *fakeFutureAPI) PollingMethod() azure.PollingMethodType {
	return ""
}

func (f *fakeFutureAPI) DoneWithContext(_ context.Context, _ autorest.Sender) (bool, error) {
	return false, nil
}

func (f *fakeFutureAPI) GetPollingDelay() (time.Duration, bool) {
	return time.Nanosecond, false
}

func (f *fakeFutureAPI) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (f *fakeFutureAPI) UnmarshalJSON(_ []byte) error {
	return nil
}

func (f *fakeFutureAPI) PollingURL() string {
	return ""
}

func (f *fakeFutureAPI) GetResult(_ autorest.Sender) (*http.Response, error) {
	return nil, nil
}

func (f *fakeFutureAPI) WaitForCompletionRef(context.Context, autorest.Client) error {
	return nil
}
