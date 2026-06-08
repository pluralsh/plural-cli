package agents

import (
	"context"
	"fmt"
	"strings"

	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/urfave/cli"

	pkgagents "github.com/pluralsh/plural-cli/pkg/agents"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/utils"
	gitutils "github.com/pluralsh/plural-cli/pkg/utils/git"
)

const recentRunsLimit int64 = 50
const branchDisplayLimit = 40
const promptDisplayLimit = 80

type Service struct {
	client      *client.Plural
	session     *pkgagents.SessionService
	interaction pkgagents.Interaction
}

func NewService(client *client.Plural) *Service {
	interaction := pkgagents.NewSurveyInteraction()
	service := &Service{
		client:      client,
		interaction: interaction,
		session:     pkgagents.NewSessionService(pkgagents.WithSessionInteraction(interaction)),
	}
	return service
}

func (p *Plural) handleResume(c *cli.Context) error {
	return common.LatestVersion(p.service.Resume)(c)
}

func (s *Service) Resume(c *cli.Context) error {
	if err := s.client.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	consoleClient := s.client.ConsoleClient

	run, err := s.selectRun(consoleClient, c.Args().First())
	if err != nil {
		return err
	}
	if run == nil {
		utils.Success("No agent runs with uploaded sessions found.\n")
		return nil
	}
	if err := s.selectPullRequest(run); err != nil {
		return err
	}

	ctx := context.Background()
	bundle, err := s.session.Download(ctx, run)
	if err != nil {
		return err
	}

	repoPath, err := s.promptRepoPath(bundle.Manifest)
	if err != nil {
		return err
	}

	utils.Highlight("Restoring %s session for run %s...\n", bundle.Manifest.Provider, bundle.Run.ID)
	if err := s.session.RestoreAndResume(ctx, bundle, repoPath); err != nil {
		return err
	}
	return nil
}

func (s *Service) selectRun(consoleClient console.ConsoleClient, runID string) (*consoleclient.AgentRunMinimalFragment, error) {
	if len(runID) > 0 {
		run, err := consoleClient.GetAgentRun(runID)
		if err != nil {
			return nil, err
		}
		if run.GetUpload() == nil || run.GetUpload().GetSession() == nil {
			return nil, fmt.Errorf("agent run %s has no uploaded session", runID)
		}

		return run, nil
	}

	runs, err := s.listResumableRuns(consoleClient)
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, nil
	}

	if err := s.printRuns(runs); err != nil {
		return nil, err
	}

	labels := make([]string, 0, len(runs))
	byLabel := map[string]*consoleclient.AgentRunMinimalFragment{}
	for _, run := range runs {
		label := s.selectorLabel(run)
		labels = append(labels, label)
		byLabel[label] = run
	}

	selected, err := s.interaction.Select("Select an agent run to resume:", labels)
	if err != nil {
		return nil, err
	}
	return byLabel[selected], nil
}

func (s *Service) listResumableRuns(consoleClient console.ConsoleClient) ([]*consoleclient.AgentRunMinimalFragment, error) {
	res, err := consoleClient.ListAgentRuns(recentRunsLimit)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("returned objects list [ListAgentRuns] is nil")
	}
	runs := make([]*consoleclient.AgentRunMinimalFragment, 0, len(res))
	for _, run := range res {
		if run == nil || run.GetUpload() == nil || run.GetUpload().GetSession() == nil {
			continue
		}
		runs = append(runs, run)
	}
	return runs, nil
}

func (s *Service) printRuns(runs []*consoleclient.AgentRunMinimalFragment) error {
	headers := []string{"Repo", "Branch", "PR Ref", "Provider", "Prompt", "Run ID"}
	return utils.PrintTable(runs, headers, func(run *consoleclient.AgentRunMinimalFragment) ([]string, error) {
		return []string{
			s.repoName(run.Repository),
			s.displayRunBranch(run),
			s.displayRunPullRequestRef(run),
			s.display(s.runProvider(run)),
			s.displayPrompt(run.Prompt),
			run.ID,
		}, nil
	})
}

func (s *Service) selectorLabel(run *consoleclient.AgentRunMinimalFragment) string {
	return fmt.Sprintf("%s  %s  %s  %s", s.repoName(run.Repository), s.display(s.runProvider(run)), s.displayPrompt(run.Prompt), run.ID)
}

func (s *Service) promptRepoPath(manifest *pkgagents.SessionManifest) (string, error) {
	def, err := gitutils.Root()
	if err != nil || def == "" {
		def = "."
	}
	return s.interaction.Directory(s.localClonePrompt(manifest.Repository), def)
}

func (s *Service) localClonePrompt(repository string) string {
	return fmt.Sprintf("Existing local clone directory for %s:", repository)
}

func (s *Service) repoName(repository string) string {
	name := gitutils.RepoName(repository)
	if name == "" {
		return repository
	}
	return name
}

func (s *Service) display(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func (s *Service) displayBranch(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(default)"
	}
	return s.truncateDisplay(value, branchDisplayLimit)
}

func (s *Service) displayRunBranch(run *consoleclient.AgentRunMinimalFragment) string {
	if run != nil && run.GetBranch() != nil {
		return s.displayBranch(*run.GetBranch())
	}
	return s.displayBranch("")
}

func (s *Service) displayRunPullRequestRef(run *consoleclient.AgentRunMinimalFragment) string {
	prs := s.pullRequestsWithRef(run)
	switch len(prs) {
	case 0:
		return "-"
	case 1:
		return s.displayPullRequestRef(prs[0])
	default:
		return fmt.Sprintf("%d pull requests", len(prs))
	}
}

func (s *Service) displayPrompt(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return "-"
	}
	return s.truncateDisplay(value, promptDisplayLimit)
}

func (s *Service) selectPullRequest(run *consoleclient.AgentRunMinimalFragment) error {
	prs := s.pullRequestsWithRef(run)
	switch len(prs) {
	case 0:
		return nil
	case 1:
		run.PullRequests = []*consoleclient.AgentRunMinimalFragment_PullRequests{prs[0]}
		return nil
	}

	labels := make([]string, 0, len(prs))
	byLabel := map[string]*consoleclient.AgentRunMinimalFragment_PullRequests{}
	for _, pr := range prs {
		label := s.pullRequestLabel(pr)
		labels = append(labels, label)
		byLabel[label] = pr
	}
	selected, err := s.interaction.Select("Select a pull request branch to resume:", labels)
	if err != nil {
		return err
	}
	run.PullRequests = []*consoleclient.AgentRunMinimalFragment_PullRequests{byLabel[selected]}
	return nil
}

func (s *Service) pullRequestsWithRef(run *consoleclient.AgentRunMinimalFragment) []*consoleclient.AgentRunMinimalFragment_PullRequests {
	if run == nil {
		return nil
	}
	prs := make([]*consoleclient.AgentRunMinimalFragment_PullRequests, 0, len(run.GetPullRequests()))
	for _, pr := range run.GetPullRequests() {
		if pr == nil || pr.GetRef() == nil || strings.TrimSpace(*pr.GetRef()) == "" {
			continue
		}
		prs = append(prs, pr)
	}
	return prs
}

func (s *Service) pullRequestLabel(pr *consoleclient.AgentRunMinimalFragment_PullRequests) string {
	title := "-"
	if pr != nil && pr.GetTitle() != nil && strings.TrimSpace(*pr.GetTitle()) != "" {
		title = s.displayPrompt(*pr.GetTitle())
	}
	status := "-"
	if pr != nil && pr.GetStatus() != nil {
		status = string(*pr.GetStatus())
	}
	id := ""
	if pr != nil {
		id = pr.ID
	}
	return fmt.Sprintf("%s  %s  %s  %s", s.displayPullRequestRef(pr), status, title, id)
}

func (s *Service) displayPullRequestRef(pr *consoleclient.AgentRunMinimalFragment_PullRequests) string {
	if pr == nil || pr.GetRef() == nil {
		return "-"
	}
	ref := s.displayBranch(*pr.GetRef())
	number := s.pullRequestNumber(pr)
	if number == "" {
		return ref
	}
	return fmt.Sprintf("%s (%s)", ref, number)
}

func (s *Service) pullRequestNumber(pr *consoleclient.AgentRunMinimalFragment_PullRequests) string {
	if pr == nil {
		return ""
	}
	if number := numericURLSuffix(pr.GetURL()); number != "" {
		return "#" + number
	}
	if strings.TrimSpace(pr.ID) != "" {
		return pr.ID
	}
	return ""
}

func numericURLSuffix(raw string) string {
	raw = strings.Trim(strings.TrimSpace(raw), "/")
	if raw == "" {
		return ""
	}
	lastSlash := strings.LastIndex(raw, "/")
	if lastSlash >= 0 {
		raw = raw[lastSlash+1:]
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return raw
}

func (s *Service) runProvider(run *consoleclient.AgentRunMinimalFragment) string {
	if run == nil || run.GetRuntime() == nil {
		return ""
	}
	return string(run.GetRuntime().Type)
}

func (s *Service) truncateDisplay(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}
