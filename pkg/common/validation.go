package common

import (
	"fmt"
	"os"

	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/utils/errors"
	"github.com/pluralsh/polly/algorithms"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/executor"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/urfave/cli"
)

func init() {
	BootstrapMode = false
}

var BootstrapMode bool

func RequireArgs(fn func(*cli.Context) error, args []string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		nargs := c.NArg()
		if nargs > len(args) {
			return fmt.Errorf("Too many args passed to %s.  Try running --help to see usage.", c.Command.FullName())
		}

		if nargs < len(args) {
			return fmt.Errorf("Not enough arguments provided: needs %s. Try running --help to see usage.", args[nargs])
		}

		return fn(c)
	}
}

func Rooted(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := RepoRoot(); err != nil {
			return err
		}

		return fn(c)
	}
}

func Owned(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := ValidateOwner(); err != nil {
			return err
		}

		return fn(c)
	}
}

func Affirmed(fn func(*cli.Context) error, msg string, envKey string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if !Affirm(msg, envKey) {
			return nil
		}

		return fn(c)
	}
}

func Highlighted(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		return utils.HighlightError(fn(c))
	}
}

func Tracked(fn func(*cli.Context) error, event string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		event := api.UserEventAttributes{Data: "", Event: event, Status: "OK"}
		err := fn(c)
		if err != nil {
			event.Status = "ERROR"
			if we, ok := err.(*executor.WrappedError); ok { //nolint:errorlint
				event.Data = we.Output
			} else {
				event.Data = fmt.Sprint(err)
			}
		}

		conf := config.Read()
		if conf.ReportErrors {
			client := api.FromConfig(&conf)
			if err := client.CreateEvent(&event); err != nil {
				return api.GetErrorResponse(err, "CreateEvent")
			}
		}
		return err
	}
}

func ValidateOwner() error {
	path := manifest.ProjectManifestPath()
	project, err := manifest.ReadProject(path)
	if err != nil {
		return fmt.Errorf("Your workspace hasn't been configured. Try running `plural init`.")
	}

	if owner := project.Owner; owner != nil {
		conf := config.Read()
		if owner.Endpoint != conf.Endpoint {
			return fmt.Errorf(
				"The owner of this project is actually %s; plural environment = %s",
				owner.Email,
				config.PluralUrl(owner.Endpoint),
			)
		}
	}

	return nil
}

func Confirm(msg string, envKey string) bool {
	res := true
	conf, ok := utils.GetEnvBoolValue(envKey)
	if ok {
		return conf
	}
	prompt := &survey.Confirm{Message: msg}
	if err := survey.AskOne(prompt, &res, survey.WithValidator(survey.Required)); err != nil {
		return false
	}
	return res
}

func Affirm(msg string, envKey string) bool {
	res := true
	conf, ok := utils.GetEnvBoolValue(envKey)
	if ok {
		return conf
	}
	prompt := &survey.Confirm{Message: msg, Default: true}
	if err := survey.AskOne(prompt, &res, survey.WithValidator(survey.Required)); err != nil {
		return false
	}
	return res
}

func RepoRoot() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	// santiize the filepath, respecting the OS
	dir = pathing.SanitizeFilepath(dir)

	root, err := git.Root()
	if err != nil {
		return err
	}

	if root != dir {
		return fmt.Errorf("You must run this command at the root of your git repository")
	}

	return nil
}

func LatestVersion(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if os.Getenv("PLURAL_CONSOLE") != "1" && os.Getenv("CLOUD_SHELL") != "1" && algorithms.Coinflip(1, 5) {
			utils.CheckLatestVersion(Version)
		}

		return fn(c)
	}
}

func InitKubeconfig(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		_, found := utils.ProjectRoot()
		if found {
			prov, err := provider.GetProvider()
			if err != nil {
				return err
			}
			if BootstrapMode {
				prov = &provider.KINDProvider{Clust: "bootstrap"}
			}
			if err := prov.KubeConfig(); err != nil {
				return err
			}
			utils.LogInfo().Println("init", prov.Name(), "provider")
		} else {
			utils.LogInfo().Println("not found provider")
		}

		return fn(c)
	}
}

func RequireKind(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		exists, _ := utils.Which("kind")
		if !exists {
			return fmt.Errorf("The kind CLI is not installed")
		}

		return fn(c)
	}
}

func CommitMsg(c *cli.Context) string {
	if commit := c.String("commit"); commit != "" {
		return commit
	}

	if !c.Bool("silence") {
		var commit string
		if err := survey.AskOne(&survey.Input{Message: "Enter a commit message (empty to not commit right now)"}, &commit); err != nil {
			return ""
		}
		return commit
	}

	return ""
}

func UpstreamSynced(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		changed, sha, err := git.HasUpstreamChanges()
		if err != nil {
			utils.LogError().Println(err)
			return errors.ErrorWrap(ErrNoGit, "Failed to get git information")
		}

		force := c.Bool("force")
		if !changed && !force {
			return errors.ErrorWrap(ErrRemoteDiff, fmt.Sprintf("Expecting HEAD at commit=%s", sha))
		}

		return fn(c)
	}
}
