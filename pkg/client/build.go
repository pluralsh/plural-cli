package client

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
)

func GetSortedInstallations(client Plural, repo string) ([]*api.Installation, error) {
	client.InitPluralClient()
	installations, err := client.GetInstallations()
	if err != nil {
		return installations, api.GetErrorResponse(err, "GetInstallations")
	}

	if len(installations) == 0 {
		return installations, fmt.Errorf("no installations present, run `plural bundle install <repo> <bundle-name>` to install your first app")
	}

	sorted, err := wkspace.UntilRepo(client.Client, repo, installations)
	if err != nil {
		sorted = installations // we don't know all the dependencies yet
	}

	return sorted, nil
}
