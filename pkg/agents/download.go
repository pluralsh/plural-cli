package agents

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	console "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

// DownloadedArchive describes a session archive stored on disk.
type DownloadedArchive struct {
	// Path is the local archive file path.
	Path string
	// WorkDir is the directory containing the archive and temporary restore data.
	WorkDir string
}

// ArchiveStore downloads and finalizes uploaded agent session archives.
type ArchiveStore interface {
	// Download fetches an archive and stores it under a temporary provider path.
	Download(ctx context.Context, url, runID string, provider console.AgentRuntimeType) (*DownloadedArchive, error)
	// Finalize moves the archive under the provider directory recorded by the manifest.
	Finalize(download *DownloadedArchive, runID string, provider console.AgentRuntimeType) (*DownloadedArchive, error)
}

// HTTPArchiveStore downloads session archives over HTTP into the Plural cache.
type HTTPArchiveStore struct {
	// client performs archive download requests.
	client *http.Client
}

// NewHTTPArchiveStore returns an archive store using the supplied HTTP client.
func NewHTTPArchiveStore(client *http.Client) *HTTPArchiveStore {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPArchiveStore{client: client}
}

func (s *HTTPArchiveStore) Download(ctx context.Context, url, runID string, provider console.AgentRuntimeType) (*DownloadedArchive, error) {
	if len(provider) == 0 {
		provider = "_downloads"
	}
	workDir, err := s.sessionWorkDir(provider, runID)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("create session download directory: %w", err)
	}

	finalPath := filepath.Join(workDir, SessionTarName)
	partialPath := finalPath + ".partial"
	if err := s.download(ctx, url, partialPath); err != nil {
		_ = os.Remove(partialPath)
		return nil, err
	}
	if err := os.Rename(partialPath, finalPath); err != nil {
		return nil, fmt.Errorf("store session archive: %w", err)
	}
	return &DownloadedArchive{Path: finalPath, WorkDir: workDir}, nil
}

func (s *HTTPArchiveStore) Finalize(download *DownloadedArchive, runID string, provider console.AgentRuntimeType) (*DownloadedArchive, error) {
	if len(provider) == 0 {
		return nil, fmt.Errorf("provider is required")
	}
	finalDir, err := s.sessionWorkDir(provider, runID)
	if err != nil {
		return nil, err
	}
	finalPath := filepath.Join(finalDir, SessionTarName)
	if download.WorkDir == finalDir && download.Path == finalPath {
		return download, nil
	}
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return nil, fmt.Errorf("create session directory: %w", err)
	}
	if err := os.Rename(download.Path, finalPath); err != nil {
		if copyErr := utils.CopyFile(download.Path, finalPath); copyErr != nil {
			return nil, fmt.Errorf("move session archive: %w", err)
		}
		_ = os.Remove(download.Path)
	}
	return &DownloadedArchive{Path: finalPath, WorkDir: finalDir}, nil
}

func (s *HTTPArchiveStore) download(ctx context.Context, url, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("download session archive: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download session archive: unexpected status %d", resp.StatusCode)
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

func (s *HTTPArchiveStore) sessionWorkDir(provider console.AgentRuntimeType, runID string) (string, error) {
	return config.PluralDir("ai", "agents", "sessions", string(provider), runID)
}
