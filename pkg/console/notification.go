package console

import (
	"fmt"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) ListNotificationSinks(after *string, first *int64) (*gqlclient.ListNotificationSinks_NotificationSinks, error) {
	response, err := c.client.ListNotificationSinks(c.ctx, after, first, nil, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListNotificationSinks")
	}
	if response == nil {
		return nil, fmt.Errorf("the result from ListNotificationSinks is null")
	}
	return response.NotificationSinks, nil
}

func (c *consoleClient) GetNotificationSink(id string) (*gqlclient.NotificationSinkFragment, error) {
	response, err := c.client.GetNotificationSink(c.ctx, id)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetNotificationSink")
	}
	if response == nil || response.NotificationSink == nil {
		return nil, fmt.Errorf("notification sink %s was not found", id)
	}
	return response.NotificationSink, nil
}

func (c *consoleClient) CreateNotificationSinks(attr gqlclient.NotificationSinkAttributes) (*gqlclient.NotificationSinkFragment, error) {
	response, err := c.client.UpsertNotificationSink(c.ctx, attr)
	if err != nil {
		return nil, api.GetErrorResponse(err, "UpsertNotificationSink")
	}
	return response.UpsertNotificationSink, nil
}
