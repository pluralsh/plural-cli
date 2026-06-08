package agents

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	console "github.com/pluralsh/console/go/client"
)

// CodexRestorer restores Codex session state and launches codex resume.
type CodexRestorer struct {
	// baseRestorer supplies shared archive and filesystem helpers.
	baseRestorer
}

func (r *CodexRestorer) Provider() console.AgentRuntimeType { return console.AgentRuntimeTypeCodex }

func (r *CodexRestorer) Prepare(_ context.Context, opts RestoreOptions) (*PreparedSession, error) {
	session, err := r.prepare(opts, "sessions", ".codex")
	if err != nil {
		return nil, err
	}

	codexHome, err := r.configDir()
	if err != nil {
		return nil, err
	}
	archivePath := "sessions"
	if opts.Manifest.Session.ArchivePath != "" {
		archivePath = opts.Manifest.Session.ArchivePath
	}
	stagingDir := filepath.Join(opts.WorkDir, ".codex", archivePath)
	sessionFile, err := r.archivedSessionFile(stagingDir, session.SessionID)
	if err != nil {
		return nil, err
	}

	confirmOverwrite := opts.sessionOverwritePrompt(console.AgentRuntimeTypeCodex)
	shouldImport, err := r.removeExistingSessionFiles(filepath.Join(codexHome, "sessions"), session.SessionID, confirmOverwrite)
	if err != nil {
		return nil, fmt.Errorf("restore codex session: %w", err)
	}
	if !shouldImport {
		return session, nil
	}

	targetRel, err := r.localSessionRel(stagingDir, sessionFile)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(sessionFile)
	if err != nil {
		return nil, err
	}
	if err := r.copyFile(sessionFile, filepath.Join(codexHome, "sessions", targetRel), info.Mode().Perm(), confirmOverwrite); err != nil {
		return nil, fmt.Errorf("restore codex session: %w", err)
	}

	return session, nil
}

func (r *CodexRestorer) Resume(ctx context.Context, prepared *PreparedSession) error {
	return r.resume(ctx, prepared, nil, "codex", "resume", prepared.SessionID, "-C", ".")
}

func (r *CodexRestorer) configDir() (string, error) {
	return r.baseRestorer.configDir("CODEX_HOME", ".codex")
}

func (r *CodexRestorer) archivedSessionFile(sessionsDir, sessionID string) (string, error) {
	var files []string
	err := filepath.WalkDir(sessionsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".jsonl" {
			return nil
		}
		if sessionID == "" {
			files = append(files, path)
			return nil
		}
		found, err := r.sessionID(path)
		if err == nil && found == sessionID {
			files = append(files, path)
			return nil
		}
		if strings.Contains(filepath.Base(path), sessionID) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(files) == 1 {
		return files[0], nil
	}
	if len(files) == 0 {
		if sessionID == "" {
			return "", fmt.Errorf("codex archive does not contain a session file")
		}
		return "", fmt.Errorf("codex archive does not contain session %s", sessionID)
	}
	if sessionID == "" {
		return "", fmt.Errorf("codex session id is required when archive contains multiple session files")
	}
	return "", fmt.Errorf("codex archive contains multiple files for session %s", sessionID)
}

func (r *CodexRestorer) localSessionRel(sessionsDir, sessionFile string) (string, error) {
	rel, err := filepath.Rel(sessionsDir, sessionFile)
	if err != nil {
		return "", err
	}
	rel = filepath.Clean(rel)
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) >= 4 && r.isDatePath(parts[0], parts[1], parts[2]) {
		return filepath.FromSlash(strings.Join(parts, "/")), nil
	}

	dateDir, err := r.sessionDateDir(sessionFile)
	if err != nil {
		return "", err
	}
	return filepath.Join(dateDir, filepath.Base(sessionFile)), nil
}

func (r *CodexRestorer) removeExistingSessionFiles(sessionsDir, sessionID string, confirmOverwrite OverwritePrompt) (bool, error) {
	if sessionID == "" {
		return true, nil
	}

	var matches []string
	err := filepath.WalkDir(sessionsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".jsonl" {
			return nil
		}
		found, err := r.sessionID(path)
		if err == nil && found == sessionID {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if len(matches) == 0 {
		return true, nil
	}

	overwrite, err := confirmOverwrite(matches[0])
	if err != nil {
		return false, err
	}
	if !overwrite {
		return false, nil
	}
	for _, match := range matches {
		if err := os.Remove(match); err != nil && !os.IsNotExist(err) {
			return false, err
		}
	}
	return true, nil
}

type codexSessionLine struct {
	// ID is the session identifier used by newer Codex JSONL entries.
	ID string `json:"id"`
	// SessionID is the legacy session identifier field.
	SessionID string `json:"session_id"`
	// ThreadID is the Codex conversation thread identifier.
	ThreadID string `json:"thread_id"`
	// Timestamp is the entry timestamp used for local session placement.
	Timestamp string `json:"timestamp"`
	// Payload contains nested session metadata for some Codex versions.
	Payload codexSessionPayload `json:"payload"`
}

type codexSessionPayload struct {
	// ID is the nested session identifier used by newer Codex payloads.
	ID string `json:"id"`
	// SessionID is the nested legacy session identifier field.
	SessionID string `json:"session_id"`
	// ThreadID is the nested Codex conversation thread identifier.
	ThreadID string `json:"thread_id"`
	// Timestamp is the nested entry timestamp.
	Timestamp string `json:"timestamp"`
}

func (r *CodexRestorer) sessionID(path string) (string, error) {
	line, err := r.sessionMetadata(path)
	if err != nil {
		return "", err
	}
	switch {
	case line.Payload.ID != "":
		return line.Payload.ID, nil
	case line.Payload.SessionID != "":
		return line.Payload.SessionID, nil
	case line.Payload.ThreadID != "":
		return line.Payload.ThreadID, nil
	case line.ID != "":
		return line.ID, nil
	case line.SessionID != "":
		return line.SessionID, nil
	default:
		return line.ThreadID, nil
	}
}

func (r *CodexRestorer) sessionDateDir(path string) (string, error) {
	line, err := r.sessionMetadata(path)
	if err != nil {
		return "", err
	}
	timestamp := line.Payload.Timestamp
	if timestamp == "" {
		timestamp = line.Timestamp
	}
	if timestamp != "" {
		parsed, err := time.Parse(time.RFC3339Nano, timestamp)
		if err == nil {
			return parsed.Format(filepath.Join("2006", "01", "02")), nil
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return info.ModTime().Format(filepath.Join("2006", "01", "02")), nil
}

func (r *CodexRestorer) sessionMetadata(path string) (*codexSessionLine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		var line codexSessionLine
		if err := json.Unmarshal([]byte(raw), &line); err != nil {
			continue
		}
		if line.ID != "" || line.SessionID != "" || line.ThreadID != "" ||
			line.Payload.ID != "" || line.Payload.SessionID != "" || line.Payload.ThreadID != "" ||
			line.Timestamp != "" || line.Payload.Timestamp != "" {
			return &line, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("codex session metadata not found in %s", path)
}

func (r *CodexRestorer) isDatePath(year, month, day string) bool {
	if len(year) != 4 || len(month) != 2 || len(day) != 2 {
		return false
	}
	_, err := time.Parse("2006/01/02", strings.Join([]string{year, month, day}, "/"))
	return err == nil
}
