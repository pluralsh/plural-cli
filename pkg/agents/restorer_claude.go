package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	console "github.com/pluralsh/console/go/client"
)

// ClaudeRestorer restores Claude CLI session state and launches Claude resume.
type ClaudeRestorer struct {
	// baseRestorer supplies shared archive and filesystem helpers.
	baseRestorer
}

func (r *ClaudeRestorer) Provider() console.AgentRuntimeType { return console.AgentRuntimeTypeClaude }

func (r *ClaudeRestorer) Prepare(_ context.Context, opts RestoreOptions) (*PreparedSession, error) {
	archivePath := "projects"
	if opts.Manifest.Session.ArchivePath != "" {
		archivePath = opts.Manifest.Session.ArchivePath
	}

	configDir, err := r.configDir()
	if err != nil {
		return nil, err
	}
	stagingDir := filepath.Join(opts.WorkDir, ".claude", "projects")
	if err := r.archive.ExtractSubtree(opts.ArchivePath, archivePath, stagingDir); err != nil {
		return nil, err
	}
	projectDir, err := r.archivedProjectDir(stagingDir, opts.Manifest.Session.ID)
	if err != nil {
		return nil, err
	}
	localProjectDir := filepath.Join(configDir, "projects", r.toProjectDirName(opts.RepoPath))
	if err := r.copyDir(projectDir, localProjectDir, console.AgentRuntimeTypeClaude, opts.sessionOverwritePrompt(console.AgentRuntimeTypeClaude)); err != nil {
		return nil, fmt.Errorf("restore claude project: %w", err)
	}

	return &PreparedSession{
		RepoPath:  opts.RepoPath,
		WorkDir:   opts.WorkDir,
		SessionID: opts.Manifest.Session.ID,
	}, nil
}

func (r *ClaudeRestorer) Resume(ctx context.Context, prepared *PreparedSession) error {
	if prepared.SessionID == "" {
		return fmt.Errorf("claude session id is required to resume")
	}
	return Executable(prepared.RepoPath).Run(ctx, "claude", "--resume", prepared.SessionID)
}

func (r *ClaudeRestorer) configDir() (string, error) {
	return r.baseRestorer.configDir("CLAUDE_CONFIG_DIR", ".claude")
}

func (r *ClaudeRestorer) archivedProjectDir(projectsDir, sessionID string) (string, error) {
	if sessionID != "" {
		sessionFile := sessionID + ".jsonl"
		if _, err := os.Stat(filepath.Join(projectsDir, sessionFile)); err == nil {
			return projectsDir, nil
		}
		entries, err := os.ReadDir(projectsDir)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			projectDir := filepath.Join(projectsDir, entry.Name())
			if _, err := os.Stat(filepath.Join(projectDir, sessionFile)); err == nil {
				return projectDir, nil
			}
		}
		return "", fmt.Errorf("claude archive does not contain session %s", sessionID)
	}

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return "", err
	}
	var projectDirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			projectDirs = append(projectDirs, filepath.Join(projectsDir, entry.Name()))
		}
	}
	if len(projectDirs) == 1 {
		return projectDirs[0], nil
	}
	if len(projectDirs) == 0 {
		return projectsDir, nil
	}
	return "", fmt.Errorf("claude session id is required when archive contains multiple project directories")
}

func (r *ClaudeRestorer) toProjectDirName(repoPath string) string {
	if abs, err := filepath.Abs(repoPath); err == nil {
		repoPath = abs
	}
	return strings.ReplaceAll(filepath.Clean(repoPath), string(filepath.Separator), "-")
}
