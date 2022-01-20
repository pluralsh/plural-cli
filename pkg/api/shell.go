package api

import (
	"fmt"
)

var shellQuery = fmt.Sprintf(`
	query {
		shell { ...CloudShellFragment }
	}
	%s
`, CloudShellFragment)

var deleteShell = fmt.Sprintf(`
	mutation {
		deleteShell { ...CloudShellFragment }
	}
	%s
`, CloudShellFragment)

func (client *Client) GetShell() (CloudShell, error) {
	var resp struct {
		Shell CloudShell
	}

	req := client.Build(shellQuery)
	err := client.Run(req, &resp)
	return resp.Shell, err
}

func (client *Client) DeleteShell() error {
	var resp struct {
		DeleteShell CloudShell
	}

	req := client.Build(deleteShell)
	return client.Run(req, &resp)
}