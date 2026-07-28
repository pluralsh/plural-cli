package workbenches

import (
	"errors"
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"

	pluralclient "github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
)

func TestCommandShape(t *testing.T) {
	command := Command(pluralclient.Plural{})

	assert.Equal(t, "workbenches", command.Name)
	assert.Contains(t, command.Aliases, "wb")
	require.Len(t, command.Subcommands, 1)
	assert.Equal(t, "pr-followup", command.Subcommands[0].Name)
	assert.Equal(t, map[string]bool{
		"console-url": true,
		"token":       true,
	}, flagNames(command.Flags))
	assert.Equal(t, map[string]bool{
		"base-url":     true,
		"commit":       true,
		"defer":        true,
		"prompt":       true,
		"provider":     true,
		"skip-missing": true,
		"url":          true,
	}, flagNames(command.Subcommands[0].Flags))
}

func TestHandlePRFollowupUsesConfiguredConsoleClient(t *testing.T) {
	url := "https://github.com/pluralsh/plural-cli/pull/5078"
	prompt := " verify the fix "
	consoleMock := mocks.NewConsoleClient(t)
	consoleMock.On("EnqueueWorkbenchPRFollowup", url, prompt, 2*time.Minute).Return("prompt-1", nil).Once()
	workbenches := NewWorkbenches(pluralclient.Plural{ConsoleClient: consoleMock})
	ctx := prFollowupContext(t, "--url", url, "--prompt", prompt, "--defer", "2m")

	err := workbenches.handlePRFollowup(ctx)

	require.NoError(t, err)
	consoleMock.AssertExpectations(t)
}

func TestHandlePRFollowupPropagatesConsoleError(t *testing.T) {
	url := "https://github.com/pluralsh/plural-cli/pull/5078"
	consoleMock := mocks.NewConsoleClient(t)
	consoleMock.On("EnqueueWorkbenchPRFollowup", url, mock.Anything, time.Duration(0)).Return("", errors.New("pull request not found")).Once()
	workbenches := NewWorkbenches(pluralclient.Plural{ConsoleClient: consoleMock})
	ctx := prFollowupContext(t, "--url", url, "--prompt", "verify")

	err := workbenches.handlePRFollowup(ctx)

	require.EqualError(t, err, "pull request not found")
}

func TestHandlePRFollowupSkipsMissingPullRequest(t *testing.T) {
	url := "https://github.com/pluralsh/plural-cli/pull/5078"
	consoleMock := mocks.NewConsoleClient(t)
	consoleMock.On("EnqueueWorkbenchPRFollowup", url, mock.Anything, time.Duration(0)).Return("", errors.New("GraphQL error: pull request not found: EnqueueWorkbenchPrFollowup")).Once()
	workbenches := NewWorkbenches(pluralclient.Plural{ConsoleClient: consoleMock})
	ctx := prFollowupContext(t, "--url", url, "--prompt", "verify", "--skip-missing")

	err := workbenches.handlePRFollowup(ctx)

	require.NoError(t, err)
	consoleMock.AssertExpectations(t)
}

func flagNames(flags []cli.Flag) map[string]bool {
	names := make(map[string]bool, len(flags))
	for _, commandFlag := range flags {
		names[commandFlag.GetName()] = true
	}

	return names
}

func prFollowupContext(t *testing.T, args ...string) *cli.Context {
	t.Helper()

	flags := flag.NewFlagSet("pr-followup", flag.ContinueOnError)
	flags.String("url", "", "")
	flags.String("commit", "", "")
	flags.String("base-url", "", "")
	flags.String("prompt", "", "")
	flags.String("provider", string(ProviderAuto), "")
	flags.String("defer", "0s", "")
	flags.Bool("skip-missing", false, "")
	require.NoError(t, flags.Parse(args))

	return cli.NewContext(nil, flags, nil)
}
