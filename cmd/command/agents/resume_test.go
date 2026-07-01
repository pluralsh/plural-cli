package agents

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	consoleclient "github.com/pluralsh/console/go/client"

	pkgagents "github.com/pluralsh/plural-cli/pkg/agents"
)

func TestAgentRunSelectorLabelOmitsBranch(t *testing.T) {
	service := &Service{}
	branch := "feature/do-not-show-this-in-selector"
	provider := consoleclient.AgentRuntimeTypeGemini
	run := &consoleclient.AgentRunMinimalFragment{
		ID:         "run-1",
		Repository: "git@github.com:pluralsh/plural.git",
		Branch:     &branch,
		Prompt:     "restore this session",
		Runtime:    &consoleclient.AgentRunMinimalFragment_Runtime{Type: provider},
	}

	label := service.selectorLabel(run)
	if strings.Contains(label, branch) {
		t.Fatalf("expected selector label to omit branch, got %q", label)
	}
	if !strings.Contains(label, "plural") || !strings.Contains(label, string(provider)) || !strings.Contains(label, run.Prompt) || !strings.Contains(label, run.ID) {
		t.Fatalf("selector label is missing expected fields: %q", label)
	}
}

func TestDisplayBranchTruncatesLongBranch(t *testing.T) {
	service := &Service{}
	branch := "feature/" + strings.Repeat("very-long-branch-name-", 4)
	displayed := service.displayBranch(branch)

	if len([]rune(displayed)) != branchDisplayLimit {
		t.Fatalf("expected branch display length %d, got %d: %q", branchDisplayLimit, len([]rune(displayed)), displayed)
	}
	if !strings.HasSuffix(displayed, "...") {
		t.Fatalf("expected truncated branch to end with ellipsis, got %q", displayed)
	}
}

func TestDisplayPromptTruncatesLongPrompt(t *testing.T) {
	service := &Service{}
	prompt := strings.Repeat("write a detailed fix ", 10)
	displayed := service.displayPrompt(prompt)

	if len([]rune(displayed)) != promptDisplayLimit {
		t.Fatalf("expected prompt display length %d, got %d: %q", promptDisplayLimit, len([]rune(displayed)), displayed)
	}
	if !strings.HasSuffix(displayed, "...") {
		t.Fatalf("expected truncated prompt to end with ellipsis, got %q", displayed)
	}
}

func TestTruncateDisplayHandlesUnicode(t *testing.T) {
	service := &Service{}
	displayed := service.truncateDisplay("abcąęłóxyz", 8)
	if displayed != "abcąę..." {
		t.Fatalf("unexpected unicode truncation: %q", displayed)
	}
}

func TestLocalClonePromptClarifiesExistingClone(t *testing.T) {
	service := &Service{}
	prompt := service.localClonePrompt("git@github.com:pluralsh/plural.git")
	if !strings.Contains(prompt, "Existing local clone directory") {
		t.Fatalf("expected prompt to clarify existing local clone, got %q", prompt)
	}
}

func TestDisplayRunBranchAndPullRequestRef(t *testing.T) {
	service := &Service{}
	branch := "main"
	ref := "plrl/run-1"
	run := &consoleclient.AgentRunMinimalFragment{
		Branch: &branch,
		PullRequests: []*consoleclient.AgentRunMinimalFragment_PullRequests{
			{ID: "pr-1", URL: "https://github.com/pluralsh/plural/pull/123", Ref: &ref},
		},
	}

	if got := service.displayRunBranch(run); got != branch {
		t.Fatalf("expected base branch %q, got %q", branch, got)
	}
	if got := service.displayRunPullRequestRef(run); got != "plrl/run-1 (#123)" {
		t.Fatalf("unexpected PR ref display: %q", got)
	}
}

func TestDisplayRunPullRequestRefFallsBackToID(t *testing.T) {
	service := &Service{}
	ref := "plrl/run-1"
	run := &consoleclient.AgentRunMinimalFragment{
		PullRequests: []*consoleclient.AgentRunMinimalFragment_PullRequests{
			{ID: "pr-id", URL: "https://example.com/pull/not-a-number", Ref: &ref},
		},
	}

	if got := service.displayRunPullRequestRef(run); got != "plrl/run-1 (pr-id)" {
		t.Fatalf("unexpected PR ref display: %q", got)
	}
}

func TestPromptRepoPathUsesCurrentMatchingRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	repo := initGitRepo(t, "git@github.com:pluralsh/plural.git")
	chdir(t, repo)

	interaction := &testInteraction{}
	service := &Service{
		interaction: interaction,
		repository:  pkgagents.NewGitRepository(nil, nil),
	}
	path, err := service.promptRepoPath(&pkgagents.SessionManifest{
		Repository: "https://github.com/pluralsh/plural.git",
	})
	if err != nil {
		t.Fatalf("promptRepoPath returned error: %v", err)
	}
	if path != repo {
		t.Fatalf("expected current repository path %q, got %q", repo, path)
	}
	if interaction.directoryMessage != "" {
		t.Fatalf("expected directory prompt to be skipped, got %q", interaction.directoryMessage)
	}
}

func TestPromptRepoPathPromptsWhenCurrentRepositoryDoesNotMatch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	repo := initGitRepo(t, "git@github.com:pluralsh/other.git")
	chdir(t, repo)

	interaction := &testInteraction{}
	service := &Service{
		interaction: interaction,
		repository:  pkgagents.NewGitRepository(nil, nil),
	}
	path, err := service.promptRepoPath(&pkgagents.SessionManifest{
		Repository: "git@github.com:pluralsh/plural.git",
	})
	if err != nil {
		t.Fatalf("promptRepoPath returned error: %v", err)
	}
	if path != repo {
		t.Fatalf("expected prompted default path %q, got %q", repo, path)
	}
	if interaction.directoryMessage == "" {
		t.Fatalf("expected directory prompt when current repository does not match")
	}
	if interaction.directoryDefault != repo {
		t.Fatalf("expected prompt default %q, got %q", repo, interaction.directoryDefault)
	}
}

func TestSelectPullRequestUsesSingleRefWithoutPrompt(t *testing.T) {
	ref := "plrl/run-1"
	service := &Service{interaction: &testInteraction{}}
	run := &consoleclient.AgentRunMinimalFragment{
		PullRequests: []*consoleclient.AgentRunMinimalFragment_PullRequests{
			{ID: "pr-1", Ref: &ref},
		},
	}

	if err := service.selectPullRequest(run); err != nil {
		t.Fatalf("selectPullRequest returned error: %v", err)
	}
	if len(run.PullRequests) != 1 || run.PullRequests[0].GetRef() == nil || *run.PullRequests[0].GetRef() != ref {
		t.Fatalf("expected single pull request ref to be selected, got %#v", run.PullRequests)
	}
}

func TestSelectPullRequestPromptsForMultipleRefs(t *testing.T) {
	ref1 := "plrl/run-1"
	ref2 := "plrl/run-2"
	title := "second branch"
	interaction := &testInteraction{}
	service := &Service{interaction: interaction}
	run := &consoleclient.AgentRunMinimalFragment{
		PullRequests: []*consoleclient.AgentRunMinimalFragment_PullRequests{
			{ID: "pr-1", Ref: &ref1},
			{ID: "pr-2", Ref: &ref2, Title: &title},
		},
	}
	interaction.selectResult = service.pullRequestLabel(run.PullRequests[1])

	if err := service.selectPullRequest(run); err != nil {
		t.Fatalf("selectPullRequest returned error: %v", err)
	}
	if interaction.selectMessage != "Select a pull request branch to resume:" {
		t.Fatalf("expected pull request prompt, got %q", interaction.selectMessage)
	}
	if len(interaction.selectOptions) != 2 {
		t.Fatalf("expected two pull request options, got %v", interaction.selectOptions)
	}
	if len(run.PullRequests) != 1 || run.PullRequests[0].ID != "pr-2" {
		t.Fatalf("expected selected pull request to be retained, got %#v", run.PullRequests)
	}
}

type testInteraction struct {
	selectMessage    string
	selectOptions    []string
	selectResult     string
	directoryMessage string
	directoryDefault string
}

func (i *testInteraction) Confirm(string, bool) (bool, error) {
	return false, nil
}

func (i *testInteraction) Select(message string, options []string) (string, error) {
	i.selectMessage = message
	i.selectOptions = append([]string(nil), options...)
	if i.selectResult == "" {
		return "", fmt.Errorf("unexpected select prompt")
	}
	return i.selectResult, nil
}

func (i *testInteraction) Directory(message, def string) (string, error) {
	i.directoryMessage = message
	i.directoryDefault = def
	return def, nil
}

func initGitRepo(t *testing.T, origin string) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "init")
	runGit(t, repo, "remote", "add", "origin", origin)
	return repo
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatal(err)
		}
	})
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}
