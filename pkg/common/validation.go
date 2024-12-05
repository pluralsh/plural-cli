package common

import (
	"fmt"
	"os"

	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/polly/algorithms"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/executor"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

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
