package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pluralsh/plural/pkg/provider"
)

type fakeAccountsClient struct {
	armstorage.AccountsClient
	Account           *armstorage.Account
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

func getFakeClientSetWithAccountsClient(existingAccount *armstorage.Account) *provider.ClientSet {
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

func (a *fakeAccountsClient) BeginCreate(_ context.Context, _ string, _ string, _ armstorage.AccountCreateParameters, _ *armstorage.AccountsClientBeginCreateOptions) (*runtime.Poller[armstorage.AccountsClientCreateResponse], error) {
	a.CreateCalledCount++

	js, err := json.Marshal(a.Account)
	if err != nil {
		return nil, err
	}

	return runtime.NewPoller(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(js)),
			Request: &http.Request{
				Method: http.MethodGet,
			},
		},
		runtime.Pipeline{}, &runtime.NewPollerOptions[armstorage.AccountsClientCreateResponse]{
			Response: &armstorage.AccountsClientCreateResponse{Account: *a.Account},
		})
}

func (a *fakeAccountsClient) GetProperties(_ context.Context, _, name string, _ *armstorage.AccountsClientGetPropertiesOptions) (armstorage.AccountsClientGetPropertiesResponse, error) {
	if a.Account != nil && a.Account.Name != nil && *a.Account.Name == name {
		return armstorage.AccountsClientGetPropertiesResponse{Account: *a.Account}, nil
	}

	return armstorage.AccountsClientGetPropertiesResponse{}, &azcore.ResponseError{StatusCode: http.StatusNotFound}
}

func (a *fakeAccountsClient) NewListPager(_ *armstorage.AccountsClientListOptions) *runtime.Pager[armstorage.AccountsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armstorage.AccountsClientListResponse]{
		More: func(_ armstorage.AccountsClientListResponse) bool {
			return false
		},
		Fetcher: func(ctx context.Context, a *armstorage.AccountsClientListResponse) (armstorage.AccountsClientListResponse, error) {
			return armstorage.AccountsClientListResponse{
				AccountListResult: armstorage.AccountListResult{
					NextLink: nil,
					Value:    nil,
				},
			}, nil
		},
	})
}

func (a *fakeAccountsClient) ListKeys(_ context.Context, _ string, _ string, _ *armstorage.AccountsClientListKeysOptions) (armstorage.AccountsClientListKeysResponse, error) {
	keyName := "name"
	keyValue := "value"
	keys := []*armstorage.AccountKey{{
		KeyName: &keyName,
		Value:   &keyValue,
	}}
	return armstorage.AccountsClientListKeysResponse{
		AccountListKeysResult: armstorage.AccountListKeysResult{Keys: keys},
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
