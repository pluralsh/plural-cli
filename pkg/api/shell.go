package api

func (client *Client) GetShell() (CloudShell, error) {
	resp, err := client.pluralClient.GetShell(client.ctx)
	if err != nil {
		return CloudShell{}, err
	}

	if resp.Shell != nil {
		return CloudShell{
			Id:     resp.Shell.ID,
			AesKey: resp.Shell.AesKey,
			GitUrl: resp.Shell.GitURL,
		}, err
	}
	return CloudShell{}, err
}

func (client *Client) DeleteShell() error {
	_, err := client.pluralClient.DeleteShell(client.ctx)
	if err != nil {
		return err
	}

	return nil
}
