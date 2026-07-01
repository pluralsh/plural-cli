package agents

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	console "github.com/pluralsh/console/go/client"
)

// Repository prepares and validates the user's local clone before session resume.
type Repository interface {
	// Prepare checks out or creates the branch that should be used for resume.
	Prepare(ctx context.Context, run *console.AgentRunMinimalFragment, bundle *SessionBundle, repoPath string) (string, error)
	// ValidateRepository verifies the local repository origin matches the session manifest.
	ValidateRepository(repoPath string, manifest *SessionManifest) error
	// Validate verifies the local repository matches the session manifest.
	Validate(repoPath string, manifest *SessionManifest) error
}

// GitRepository implements Repository operations using the local git executable.
type GitRepository struct {
	// httpClient downloads patch files when a PR branch was not created.
	httpClient *http.Client
	// confirmer prompts before changing the selected local checkout.
	confirmer Confirmer
}

// NewGitRepository returns a git-backed repository preparer.
func NewGitRepository(httpClient *http.Client, confirmer Confirmer) *GitRepository {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &GitRepository{httpClient: httpClient, confirmer: confirmer}
}

// Prepare checks out a PR ref when present, or creates a local branch from the
// base branch and applies the uploaded patch when a PR has not been created.
func (p *GitRepository) Prepare(ctx context.Context, run *console.AgentRunMinimalFragment, bundle *SessionBundle, repoPath string) (string, error) {
	ref := agentRunRef(run)
	if ref != "" {
		if err := p.validate(repoPath, bundle.Manifest, false); err != nil {
			return "", err
		}
		branch := p.localBranchName(ref)
		if currentBranch, err := p.currentBranch(repoPath); err == nil && currentBranch == branch {
			return branch, nil
		}
		confirmed, err := p.confirmAction(fmt.Sprintf("Checkout branch %s in the selected local clone before resuming?", branch))
		if err != nil {
			return "", err
		}
		if !confirmed {
			return "", nil
		}
		if err := p.checkoutRef(repoPath, ref, branch); err != nil {
			return "", err
		}
		return branch, nil
	}
	patchURL := agentRunPatchURL(run)
	if patchURL == "" {
		return "", nil
	}
	if err := p.validate(repoPath, bundle.Manifest, false); err != nil {
		return "", err
	}

	branch := p.patchBranchName()
	baseBranch := p.baseBranch(run, bundle.Manifest)
	confirmed, err := p.confirmAction(p.applyPatchPrompt(branch, baseBranch))
	if err != nil {
		return "", err
	}
	if !confirmed {
		return "", nil
	}
	if err := p.applyPatch(ctx, repoPath, bundle.WorkDir, patchURL, baseBranch, branch); err != nil {
		return "", err
	}
	return branch, nil
}

// ValidateRepository ensures the selected directory is the expected repository.
func (p *GitRepository) ValidateRepository(repoPath string, manifest *SessionManifest) error {
	return p.validate(repoPath, manifest, false)
}

// Validate ensures the selected directory is the expected repository and, when
// a branch is specified, that the checkout is on that branch.
func (p *GitRepository) Validate(repoPath string, manifest *SessionManifest) error {
	if repoPath == "" {
		return fmt.Errorf("repository path is required")
	}
	if manifest == nil {
		return fmt.Errorf("session manifest is required")
	}
	if _, err := p.git(repoPath, "rev-parse", "--show-toplevel"); err != nil {
		return fmt.Errorf("selected directory %q is not an existing git clone. Select an already cloned local checkout for %s: %w", repoPath, manifest.Repository, err)
	}
	remote, err := p.git(repoPath, "ls-remote", "--get-url", "origin")
	if err != nil {
		return fmt.Errorf("could not read git origin for selected local clone %q. Select an already cloned local checkout for %s: %w", repoPath, manifest.Repository, err)
	}
	if p.normalizeGitURL(remote) != p.normalizeGitURL(manifest.Repository) {
		return fmt.Errorf("selected local clone is not the expected repository. Selected origin is %q, expected %q", remote, manifest.Repository)
	}
	if manifest.Branch != "" {
		branch, err := p.git(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return fmt.Errorf("could not determine current branch for selected repository %q: %w", repoPath, err)
		}
		if branch != manifest.Branch {
			return fmt.Errorf("current branch %q does not match expected branch %q. Check out %q and rerun plural agents resume", branch, manifest.Branch, manifest.Branch)
		}
	}
	return nil
}

func (p *GitRepository) normalizeGitURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, "/")
	raw = strings.TrimSuffix(raw, ".git")
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err == nil {
			host := strings.ToLower(parsed.Host)
			path := strings.Trim(strings.TrimSuffix(parsed.Path, ".git"), "/")
			return host + "/" + path
		}
	}
	if strings.HasPrefix(raw, "git@") {
		raw = strings.TrimPrefix(raw, "git@")
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 {
			return strings.ToLower(parts[0]) + "/" + strings.Trim(strings.TrimSuffix(parts[1], ".git"), "/")
		}
	}
	raw = strings.TrimPrefix(raw, "ssh://git@")
	return strings.ToLower(filepath.ToSlash(strings.Trim(raw, "/")))
}

func (p *GitRepository) validate(repoPath string, manifest *SessionManifest, checkBranch bool) error {
	if manifest == nil {
		return p.Validate(repoPath, nil)
	}

	expected := *manifest
	if !checkBranch {
		expected.Branch = ""
	}

	return p.Validate(repoPath, &expected)
}

func (p *GitRepository) checkoutRef(repoPath, ref, branch string) error {
	if err := p.ensureClean(repoPath); err != nil {
		return err
	}
	if _, err := p.git(repoPath, "check-ref-format", "--branch", branch); err != nil {
		return fmt.Errorf("invalid local branch name %q: %w", branch, err)
	}
	if _, err := p.git(repoPath, "fetch", "origin", ref); err != nil {
		return fmt.Errorf("fetch branch %q: %w", ref, err)
	}
	if p.localBranchExists(repoPath, branch) {
		if _, err := p.git(repoPath, "checkout", branch); err != nil {
			return fmt.Errorf("checkout branch %q: %w", branch, err)
		}
		if _, err := p.git(repoPath, "merge", "--ff-only", "FETCH_HEAD"); err != nil {
			return fmt.Errorf("fast-forward branch %q to %q: %w", branch, ref, err)
		}
		return nil
	}
	if _, err := p.git(repoPath, "checkout", "-b", branch, "FETCH_HEAD"); err != nil {
		return fmt.Errorf("checkout branch %q from %q: %w", branch, ref, err)
	}
	return nil
}

func (p *GitRepository) applyPatch(ctx context.Context, repoPath, workDir, patchURL, baseBranch, branch string) error {
	if err := p.ensureClean(repoPath); err != nil {
		return err
	}
	if _, err := p.git(repoPath, "check-ref-format", "--branch", branch); err != nil {
		return fmt.Errorf("invalid local branch name %q: %w", branch, err)
	}
	if strings.TrimSpace(baseBranch) != "" {
		if _, err := p.git(repoPath, "fetch", "origin", baseBranch); err != nil {
			return fmt.Errorf("fetch base branch %q: %w", baseBranch, err)
		}
		if _, err := p.git(repoPath, "checkout", "-b", branch, "FETCH_HEAD"); err != nil {
			return fmt.Errorf("create branch %q from base branch %q: %w", branch, baseBranch, err)
		}
	} else if _, err := p.git(repoPath, "checkout", "-b", branch); err != nil {
		return fmt.Errorf("create branch %q: %w", branch, err)
	}
	patchPath := filepath.Join(workDir, "agent-run.patch")
	if err := p.downloadPatch(ctx, patchURL, patchPath); err != nil {
		return err
	}
	if _, err := p.git(repoPath, "apply", patchPath); err != nil {
		return fmt.Errorf("apply agent run patch: %w", err)
	}
	return nil
}

func (p *GitRepository) ensureClean(repoPath string) error {
	status, err := p.git(repoPath, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("check git status: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		return fmt.Errorf("selected local clone has uncommitted changes; commit or stash them before resuming")
	}
	return nil
}

func (p *GitRepository) localBranchExists(repoPath, branch string) bool {
	_, err := p.git(repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}

func (p *GitRepository) currentBranch(repoPath string) (string, error) {
	return p.git(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}

func (p *GitRepository) downloadPatch(ctx context.Context, url, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download agent run patch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download agent run patch: unexpected status %d", resp.StatusCode)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func (p *GitRepository) git(repoPath string, args ...string) (string, error) {
	return Executable(repoPath).Output(context.Background(), "git", args...)
}

func (p *GitRepository) confirmAction(message string) (bool, error) {
	if p.confirmer == nil {
		return false, nil
	}
	return p.confirmer.Confirm(message, false)
}

func (p *GitRepository) baseBranch(run *console.AgentRunMinimalFragment, manifest *SessionManifest) string {
	if branch := agentRunBranch(run); branch != "" {
		return branch
	}
	if manifest == nil {
		return ""
	}
	return strings.TrimSpace(manifest.Branch)
}

func (p *GitRepository) applyPatchPrompt(branch, baseBranch string) string {
	if strings.TrimSpace(baseBranch) == "" {
		return fmt.Sprintf("Agent run has a patch but no PR branch. Create local branch %s from the current checkout and apply the patch before resuming?", branch)
	}
	return fmt.Sprintf("Agent run has a patch but no PR branch. Create local branch %s from base branch %s and apply the patch before resuming?", branch, baseBranch)
}

func (p *GitRepository) localBranchName(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "refs/heads/")
	ref = strings.TrimPrefix(ref, "origin/")
	return ref
}

func (p *GitRepository) patchBranchName() string {
	return "plural-agent/" + time.Now().UTC().Format("20060102-150405.000000000")
}
