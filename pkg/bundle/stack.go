package bundle

import (
	"fmt"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
)

func Stack(client api.Client, name, provider string, refresh bool) error {
	s, err := client.GetStack(name, provider)
	if err != nil {
		return api.GetErrorResponse(err, "GetStack")
	}

	utils.Highlight("You're attempting to install stack: %s\n>> ", s.Name)
	fmt.Println(s.Description)
	fmt.Println()

	repos := make([]string, 0)
	for _, r := range s.Bundles {
		repos = append(repos, r.Repository.Name)
	}

	if !utils.Confirm(fmt.Sprintf("This will install all of {%s}, do you want to proceed?", strings.Join(repos, ", "))) {
		return nil
	}

	for _, recipe := range s.Bundles {
		if err := doInstall(client, recipe, recipe.Repository.Name, provider, refresh); err != nil {
			return err
		}
	}

	return nil
}
