package console

import (
	gqlclient "github.com/pluralsh/console/go/client"
)

func (c *consoleClient) ListNotificationSinks(after *string, first *int64) (*gqlclient.ListNotificationSinks_NotificationSinks, error) {
	response, err := c.client.ListNotificationSinks(c.ctx, after, first, nil, nil)
	if err != nil {
		return nil, err
	}
	return response.NotificationSinks, nil
}

func (c *consoleClient) CreateNotificationSinks(attr gqlclient.NotificationSinkAttributes) (*gqlclient.NotificationSinkFragment, error) {
	response, err := c.client.UpsertNotificationSink(c.ctx, attr)
	if err != nil {
		return nil, err
	}
	return response.UpsertNotificationSink, nil
}
