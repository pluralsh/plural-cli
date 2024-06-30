package plural

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/up"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

const (
	affirmUp   = "Are you ready to set up your initial management cluster?  You can check the generated terraform/helm to confirm everything looks good first"
	affirmDown = "Are you ready to destroy your plural infrastructure?  This will destroy all k8s clusters and any data stored within"
)

func (p *Plural) handleUp(c *cli.Context) error {
	// provider.IgnoreProviders([]string{"GENERIC", "KIND"})
	if err := p.handleInit(c); err != nil {
		return err
	}
	p.InitPluralClient()

	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	ctx, err := up.Build()
	if err != nil {
		return err
	}

	if err := ctx.Backfill(); err != nil {
		return err
	}

	if err := ctx.Generate(); err != nil {
		return err
	}

	if !affirm(affirmUp, "PLURAL_UP_AFFIRM_DEPLOY") {
		return fmt.Errorf("cancelled deploy")
	}

	if err := ctx.Deploy(func() error {
		utils.Highlight("\n==> Commit and push your configuration\n\n")
		if commit := commitMsg(c); commit != "" {
			utils.Highlight("Pushing upstream...\n")
			return git.Sync(repoRoot, commit, c.Bool("force"))
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (p *Plural) handleDown(c *cli.Context) error {
	if !affirm(affirmDown, "PLURAL_DOWN_AFFIRM_DESTROY") {
		return fmt.Errorf("cancelled destroy")
	}

	ctx, err := up.Build()
	if err != nil {
		return err
	}

	return ctx.Destroy()
}
