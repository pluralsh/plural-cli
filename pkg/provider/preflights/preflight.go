package preflights

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

type Preflight struct {
	Name     string
	Callback func() error
}

func (pf *Preflight) Validate() error {
	utils.Highlight("Executing preflight check :: %s ", pf.Name)
	if err := pf.Callback(); err != nil {
		fmt.Println("\nFound error:")
		return err
	}

	utils.Success("\u2713\n")
	return nil
}
