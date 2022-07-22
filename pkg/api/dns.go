package api

func (client *client) CreateDomain(name string) error {
	_, err := client.pluralClient.CreateDomain(client.ctx, name)
	if err != nil {
		return err
	}

	return nil
}
