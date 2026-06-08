package agents

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGitRepositoryApplyPatchUsesBaseBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	root := t.TempDir()
	source := filepath.Join(root, "source")
	origin := filepath.Join(root, "origin.git")
	clone := filepath.Join(root, "clone")

	mkdir(t, source)
	git(t, source, "init")
	git(t, source, "config", "user.email", "test@example.com")
	git(t, source, "config", "user.name", "Test User")
	git(t, source, "config", "commit.gpgsign", "false")
	git(t, source, "checkout", "-b", "main")
	writeFile(t, filepath.Join(source, "README.md"), "base\n")
	git(t, source, "add", "README.md")
	git(t, source, "commit", "-m", "base")
	git(t, root, "clone", "--bare", source, origin)
	git(t, root, "clone", origin, clone)
	git(t, clone, "checkout", "-b", "unrelated")

	writeFile(t, filepath.Join(source, "README.md"), "base\npatched\n")
	patch := git(t, source, "diff")
	httpClient := &http.Client{Transport: roundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(patch)),
		}, nil
	})}

	repository := NewGitRepository(httpClient, nil)
	err := repository.applyPatch(context.Background(), clone, t.TempDir(), "https://example.com/agent.patch", "main", "plural-agent/run-1")
	if err != nil {
		t.Fatalf("applyPatch returned error: %v", err)
	}

	branch := strings.TrimSpace(git(t, clone, "rev-parse", "--abbrev-ref", "HEAD"))
	if branch != "plural-agent/run-1" {
		t.Fatalf("expected patch branch to be checked out, got %q", branch)
	}
	content, err := os.ReadFile(filepath.Join(clone, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "base\npatched\n" {
		t.Fatalf("unexpected patched file content: %q", string(content))
	}
}

func TestGitRepositoryPatchBranchNameUsesTimestamp(t *testing.T) {
	branch := NewGitRepository(nil, nil).patchBranchName()
	if !regexp.MustCompile(`^plural-agent/\d{8}-\d{6}\.\d{9}$`).MatchString(branch) {
		t.Fatalf("unexpected patch branch name %q", branch)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}
