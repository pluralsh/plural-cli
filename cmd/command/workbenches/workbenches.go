package workbenches

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"

	pluralclient "github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type Workbenches struct {
	pluralclient.Plural
	consoleToken string
	consoleURL   string
}

func NewWorkbenches(clients pluralclient.Plural) *Workbenches {
	return &Workbenches{Plural: clients}
}

func Command(clients pluralclient.Plural) cli.Command {
	return NewWorkbenches(clients).Command()
}

func (w *Workbenches) Command() cli.Command {
	return cli.Command{
		Name:     "workbenches",
		Aliases:  []string{"wb"},
		Usage:    "manage Plural workbenches",
		Category: "AI",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "token",
				Usage:       "console token",
				EnvVar:      "PLURAL_CONSOLE_TOKEN",
				Destination: &w.consoleToken,
			},
			cli.StringFlag{
				Name:        "console-url",
				Usage:       "console URL address",
				EnvVar:      "PLURAL_CONSOLE_URL",
				Destination: &w.consoleURL,
			},
		},
		Subcommands: []cli.Command{w.prFollowupCommand()},
	}
}

func (w *Workbenches) prFollowupCommand() cli.Command {
	return cli.Command{
		Name:   "pr-followup",
		Usage:  "send a follow-up prompt to the workbench job associated with a pull request",
		Action: common.LatestVersion(w.handlePRFollowup),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "prompt",
				Usage:    "follow-up prompt",
				Required: true,
			},
			cli.StringFlag{
				Name:  "url",
				Usage: "pull request URL; bypasses commit inference",
			},
			cli.StringFlag{
				Name:  "commit",
				Usage: "commit or ref whose subject identifies the pull request (defaults to HEAD)",
			},
			cli.StringFlag{
				Name:  "base-url",
				Usage: "repository web URL used to construct the pull request URL",
			},
			cli.StringFlag{
				Name:  "provider",
				Usage: "source control provider (auto, github, gitlab, or bitbucket)",
				Value: string(ProviderAuto),
			},
			cli.StringFlag{
				Name:  "defer",
				Usage: "Defer the follow-up for a duration, eg 1s, 1m, 2h, etc",
				Value: "0s",
			},
			common.StringEnumFlag("output, o", "output format", common.OutputFormatRaw, common.OutputFormats...),
			cli.BoolFlag{
				Name:  "skip-missing",
				Usage: "exit successfully when the pull request is not associated with a workbench job",
			},
		},
	}
}

func (w *Workbenches) handlePRFollowup(ctx *cli.Context) error {
	if err := w.InitConsoleClient(w.consoleToken, w.consoleURL); err != nil {
		return err
	}

	deferDuration, err := time.ParseDuration(ctx.String("defer"))
	if err != nil {
		return fmt.Errorf("invalid defer duration: %w", err)
	}

	if deferDuration < 0 {
		return fmt.Errorf("defer duration must be non-negative")
	}

	output := ctx.String("output")
	if err := common.ValidateStringEnum("output", output, common.OutputFormats...); err != nil {
		return err
	}

	service := NewPRFollowupService(w.ConsoleClient, NewPullRequestResolver(nil))
	result, err := service.Create(PRFollowupOptions{
		Prompt:      ctx.String("prompt"),
		Defer:       deferDuration,
		SkipMissing: ctx.Bool("skip-missing"),
		PullRequest: PullRequestOptions{
			URL:      ctx.String("url"),
			Commit:   ctx.String("commit"),
			BaseURL:  ctx.String("base-url"),
			Provider: ctx.String("provider"),
		},
	})
	if err != nil {
		return err
	}

	return writePRFollowupResult(output, result)
}

func writePRFollowupResult(output string, result PRFollowupResult) error {
	switch output {
	case common.OutputFormatRaw:
		writeRawPRFollowupResult(result)
	case common.OutputFormatJSON:
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	return nil
}

func writeRawPRFollowupResult(result PRFollowupResult) {
	if result.Skipped {
		utils.Success("No workbench job found for %s; skipping\n", result.PullRequestURL)
		return
	}

	fmt.Printf("Created workbench PR follow-up %s for %s\n", result.PromptID, result.PullRequestURL)
	utils.Success("Workbench Job URL: %s\n", result.WorkbenchJobURL)
}
