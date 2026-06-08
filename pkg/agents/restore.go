package agents

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	console "github.com/pluralsh/console/go/client"
)

// RestoreOptions are the inputs a provider restorer needs to recreate local
// session state from an uploaded archive.
type RestoreOptions struct {
	// RepoPath is the selected local repository checkout.
	RepoPath string
	// ArchivePath is the local path to the downloaded session archive.
	ArchivePath string
	// WorkDir is the local directory used for extracted restore data.
	WorkDir string
	// Manifest is the validated session manifest from the archive.
	Manifest *SessionManifest
	// Interaction prompts before overwriting existing provider session data.
	Interaction Confirmer
	// ConfirmOverwrite overrides overwrite prompting for tests and custom flows.
	ConfirmOverwrite OverwritePrompt
}

// PreparedSession is the local provider session after archive extraction and
// before invoking the provider's resume command.
type PreparedSession struct {
	// RepoPath is the local repository where the provider resume command runs.
	RepoPath string
	// WorkDir is the local directory containing restored provider files.
	WorkDir string
	// SessionID is the provider-specific session identifier to resume.
	SessionID string
}

// SessionRestorer restores and resumes sessions for one agent runtime provider.
type SessionRestorer interface {
	// Provider returns the console runtime type this restorer supports.
	Provider() console.AgentRuntimeType
	// Prepare restores provider state from the archive into local files.
	Prepare(ctx context.Context, opts RestoreOptions) (*PreparedSession, error)
	// Resume launches the provider-specific resume command.
	Resume(ctx context.Context, prepared *PreparedSession) error
}

// RestorerRegistry resolves the provider-specific restorer for a session.
type RestorerRegistry interface {
	// ForProvider returns the restorer registered for provider.
	ForProvider(provider console.AgentRuntimeType) (SessionRestorer, error)
}

// MapRestorerRegistry is an in-memory provider-to-restorer registry.
type MapRestorerRegistry struct {
	// restorers maps console runtime type to provider implementation.
	restorers map[console.AgentRuntimeType]SessionRestorer
}

// NewRestorerRegistry registers the supplied provider restorers.
func NewRestorerRegistry(restorers ...SessionRestorer) *MapRestorerRegistry {
	registry := &MapRestorerRegistry{restorers: map[console.AgentRuntimeType]SessionRestorer{}}
	for _, restorer := range restorers {
		registry.restorers[restorer.Provider()] = restorer
	}
	return registry
}

func (r *MapRestorerRegistry) ForProvider(provider console.AgentRuntimeType) (SessionRestorer, error) {
	restorer, ok := r.restorers[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported session provider %q", provider)
	}
	return restorer, nil
}

// NewDefaultRestorerRegistry returns the built-in restorers supported by the CLI.
func NewDefaultRestorerRegistry(archive ArchiveReader) *MapRestorerRegistry {
	if archive == nil {
		archive = TarGzipArchiveReader{}
	}

	return NewRestorerRegistry(
		&ClaudeRestorer{baseRestorer: baseRestorer{archive: archive}},
		&CodexRestorer{baseRestorer: baseRestorer{archive: archive}},
		&GeminiRestorer{baseRestorer: baseRestorer{archive: archive}},
		&OpencodeRestorer{baseRestorer: baseRestorer{archive: archive}},
	)
}

type baseRestorer struct {
	archive ArchiveReader
}

func (r *baseRestorer) prepare(opts RestoreOptions, archivePath, providerDir string) (*PreparedSession, error) {
	if opts.Manifest.Session.ArchivePath != "" {
		archivePath = opts.Manifest.Session.ArchivePath
	}

	if err := r.archive.ExtractSubtree(opts.ArchivePath, archivePath, filepath.Join(opts.WorkDir, providerDir, archivePath)); err != nil {
		return nil, err
	}
	return &PreparedSession{
		RepoPath:  opts.RepoPath,
		WorkDir:   opts.WorkDir,
		SessionID: opts.Manifest.Session.ID,
	}, nil
}

func (r *baseRestorer) resume(ctx context.Context, prepared *PreparedSession, env []string, command string, args ...string) error {
	return Executable(prepared.RepoPath, env...).Run(ctx, command, args...)
}

func (r *baseRestorer) configDir(envName, defaultDir string) (string, error) {
	if configDir := strings.TrimSpace(os.Getenv(envName)); configDir != "" {
		return configDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, defaultDir), nil
}

func (r *baseRestorer) repoDirBaseName(repoPath string) string {
	if abs, err := filepath.Abs(repoPath); err == nil {
		repoPath = abs
	}
	return filepath.Base(filepath.Clean(repoPath))
}

func (r *baseRestorer) copyDir(src, dst string, provider console.AgentRuntimeType, confirmOverwrite OverwritePrompt) error {
	if confirmOverwrite == nil {
		confirmOverwrite = overwriteExisting
	}
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0755)
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch mode := info.Mode(); {
		case mode.IsDir():
			return os.MkdirAll(target, mode.Perm())
		case mode.Type() == 0:
			return r.copyFile(path, target, mode.Perm(), confirmOverwrite)
		default:
			return fmt.Errorf("unsupported %s session file %q", provider, path)
		}
	})
}

func (r *baseRestorer) copyFile(src, dst string, mode os.FileMode, confirmOverwrite OverwritePrompt) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	if _, err := os.Lstat(dst); err == nil {
		overwrite, err := confirmOverwrite(dst)
		if err != nil {
			return err
		}
		if !overwrite {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
