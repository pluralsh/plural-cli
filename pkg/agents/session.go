package agents

import (
	"context"
	"fmt"

	console "github.com/pluralsh/console/go/client"
)

// SessionBundle contains a downloaded agent session archive and the console run
// metadata needed to restore it into a local checkout.
type SessionBundle struct {
	// Run is the minimal console run metadata needed for local resume.
	Run *console.AgentRunMinimalFragment
	// Manifest is the validated manifest read from the session archive.
	Manifest *SessionManifest
	// ArchivePath is the local path to the finalized session archive.
	ArchivePath string
	// WorkDir is the local working directory used while restoring the session.
	WorkDir string
}

// SessionService coordinates download, repository preparation, validation, and
// provider-specific session restoration.
type SessionService struct {
	// store downloads uploaded session archives.
	store ArchiveStore
	// archive reads and extracts session archives.
	archive ArchiveReader
	// repository prepares and validates the local git checkout.
	repository Repository
	// registry resolves provider-specific restorers.
	registry RestorerRegistry
	// interaction handles confirmations needed during restore.
	interaction Confirmer
}

// SessionServiceOption customizes SessionService dependencies for tests and
// alternate UI/storage implementations.
type SessionServiceOption func(*SessionService)

// WithSessionInteraction sets the prompt implementation used by session restore.
func WithSessionInteraction(interaction Confirmer) SessionServiceOption {
	return func(s *SessionService) {
		s.interaction = interaction
	}
}

// WithSessionRepository sets the repository preparer used before restoration.
func WithSessionRepository(repository Repository) SessionServiceOption {
	return func(s *SessionService) {
		s.repository = repository
	}
}

// NewSessionService builds a service with production defaults unless overridden.
func NewSessionService(options ...SessionServiceOption) *SessionService {
	archive := TarGzipArchiveReader{}
	interaction := NewSurveyInteraction()
	service := &SessionService{
		store:       NewHTTPArchiveStore(nil),
		archive:     archive,
		registry:    NewDefaultRestorerRegistry(archive),
		interaction: interaction,
	}
	for _, option := range options {
		option(service)
	}
	if service.repository == nil {
		service.repository = NewGitRepository(nil, service.interaction)
	}
	return service
}

// Download fetches the run's uploaded session archive and validates its manifest.
func (s *SessionService) Download(ctx context.Context, run *console.AgentRunMinimalFragment) (*SessionBundle, error) {
	if run == nil {
		return nil, fmt.Errorf("agent run is required")
	}
	sessionURL := agentRunSessionURL(run)
	if sessionURL == "" {
		return nil, fmt.Errorf("agent run %s has no uploaded session archive", run.ID)
	}

	download, err := s.store.Download(ctx, sessionURL, run.ID, agentRunProvider(run))
	if err != nil {
		return nil, err
	}
	manifest, err := s.archive.ReadManifest(download.Path)
	if err != nil {
		return nil, err
	}
	if err := manifest.Validate(run); err != nil {
		return nil, err
	}
	if manifest.Session.ArchivePath != "" {
		ok, err := s.archive.Contains(download.Path, manifest.Session.ArchivePath)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("session archive does not contain manifest archivePath %q", manifest.Session.ArchivePath)
		}
	}
	finalDownload, err := s.store.Finalize(download, run.ID, manifest.Provider)
	if err != nil {
		return nil, err
	}
	return &SessionBundle{Run: run, Manifest: manifest, ArchivePath: finalDownload.Path, WorkDir: finalDownload.WorkDir}, nil
}

// RestoreAndResume prepares the local repository, restores provider state, and
// launches the provider-specific resume command.
func (s *SessionService) RestoreAndResume(ctx context.Context, bundle *SessionBundle, repoPath string) error {
	if bundle == nil || bundle.Manifest == nil {
		return fmt.Errorf("session bundle is required")
	}
	if bundle.Run == nil {
		return fmt.Errorf("session bundle run is required")
	}
	resumeBranch, err := s.repository.Prepare(ctx, bundle.Run, bundle, repoPath)
	if err != nil {
		return err
	}
	if resumeBranch != "" {
		bundle.Manifest.Branch = resumeBranch
	}
	if err := s.repository.Validate(repoPath, bundle.Manifest); err != nil {
		return err
	}
	restorer, err := s.registry.ForProvider(bundle.Manifest.Provider)
	if err != nil {
		return err
	}
	prepared, err := restorer.Prepare(ctx, RestoreOptions{
		RepoPath:    repoPath,
		ArchivePath: bundle.ArchivePath,
		WorkDir:     bundle.WorkDir,
		Manifest:    bundle.Manifest,
		Interaction: s.interaction,
	})
	if err != nil {
		return err
	}
	return restorer.Resume(ctx, prepared)
}
