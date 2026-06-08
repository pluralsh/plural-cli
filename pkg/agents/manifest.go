package agents

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	console "github.com/pluralsh/console/go/client"
)

// SessionManifest is the metadata embedded in an uploaded agent session archive.
type SessionManifest struct {
	// Version is the manifest schema version.
	Version int `json:"version"`
	// AgentRunID is the console agent run that produced the archive.
	AgentRunID string `json:"agentRunId"`
	// Provider identifies the runtime whose local state is stored in the archive.
	Provider console.AgentRuntimeType `json:"provider"`
	// Repository is the git remote URL expected for the local checkout.
	Repository string `json:"repository"`
	// Branch is the expected local branch after repository preparation.
	Branch string `json:"branch,omitempty"`
	// Session describes where provider-specific session state is stored.
	Session SessionMetadata `json:"session"`
}

// SessionMetadata identifies the provider session data inside the archive.
type SessionMetadata struct {
	// ID is the provider-specific session identifier used by resume commands.
	ID string `json:"id,omitempty"`
	// Path is the original provider session path recorded by the runtime.
	Path string `json:"path,omitempty"`
	// ArchivePath is the subtree in the archive containing session files.
	ArchivePath string `json:"archivePath,omitempty"`
}

// UnmarshalJSON accepts provider values from older archives while normalizing
// them to the generated console enum representation.
func (m *SessionManifest) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Version    int             `json:"version"`
		AgentRunID string          `json:"agentRunId"`
		Provider   string          `json:"provider"`
		Repository string          `json:"repository"`
		Branch     string          `json:"branch,omitempty"`
		Session    SessionMetadata `json:"session"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	provider := console.AgentRuntimeType(strings.ToUpper(strings.TrimSpace(decoded.Provider)))
	if provider != "" && !provider.IsValid() {
		return fmt.Errorf("%s is not a valid AgentRuntimeType", decoded.Provider)
	}
	*m = SessionManifest{
		Version:    decoded.Version,
		AgentRunID: decoded.AgentRunID,
		Provider:   provider,
		Repository: decoded.Repository,
		Branch:     decoded.Branch,
		Session:    decoded.Session,
	}
	return nil
}

// Validate checks that the archive manifest is compatible with the selected run
// and safe to extract.
func (m *SessionManifest) Validate(run *console.AgentRunMinimalFragment) error {
	if m == nil {
		return fmt.Errorf("session manifest is empty")
	}
	if m.Version != 1 {
		return fmt.Errorf("unsupported session manifest version %d", m.Version)
	}
	if strings.TrimSpace(m.AgentRunID) == "" {
		return fmt.Errorf("session manifest agentRunId is required")
	}
	if run != nil && m.AgentRunID != run.ID {
		return fmt.Errorf("session manifest run id %q does not match selected run %q", m.AgentRunID, run.ID)
	}
	if strings.TrimSpace(m.Repository) == "" {
		return fmt.Errorf("session manifest repository is required")
	}
	if m.Session.ArchivePath != "" {
		archivePath := strings.TrimSpace(m.Session.ArchivePath)
		cleaned := filepath.ToSlash(filepath.Clean(archivePath))
		if filepath.IsAbs(archivePath) || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
			return fmt.Errorf("unsafe session archivePath %q", m.Session.ArchivePath)
		}
		m.Session.ArchivePath = cleaned
	}
	return nil
}
