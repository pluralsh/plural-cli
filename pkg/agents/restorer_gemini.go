package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	console "github.com/pluralsh/console/go/client"
)

// GeminiRestorer restores Gemini CLI session state and launches gemini resume.
type GeminiRestorer struct {
	// baseRestorer supplies shared archive and filesystem helpers.
	baseRestorer
}

func (r *GeminiRestorer) Provider() console.AgentRuntimeType { return console.AgentRuntimeTypeGemini }

func (r *GeminiRestorer) Prepare(_ context.Context, opts RestoreOptions) (*PreparedSession, error) {
	session, err := r.prepare(opts, "chats", ".")
	if err != nil {
		return nil, err
	}

	geminiHome, err := r.configDir()
	if err != nil {
		return nil, err
	}
	archivePath := "chats"
	if opts.Manifest.Session.ArchivePath != "" {
		archivePath = opts.Manifest.Session.ArchivePath
	}
	chatsDir := filepath.Join(opts.WorkDir, archivePath)
	localChatsDir := filepath.Join(geminiHome, "tmp", r.repoDirBaseName(opts.RepoPath), "chats")
	confirmOverwrite := opts.sessionOverwritePrompt(console.AgentRuntimeTypeGemini)
	shouldImport, err := r.removeExistingSessionFiles(localChatsDir, session.SessionID, confirmOverwrite)
	if err != nil {
		return nil, fmt.Errorf("restore gemini chats: %w", err)
	}
	if !shouldImport {
		return session, nil
	}
	if err := r.copyDir(chatsDir, localChatsDir, console.AgentRuntimeTypeGemini, confirmOverwrite); err != nil {
		return nil, fmt.Errorf("restore gemini chats: %w", err)
	}

	return session, nil
}

func (r *GeminiRestorer) Resume(ctx context.Context, session *PreparedSession) error {
	return r.resume(ctx, session, nil, "gemini", "--resume", session.SessionID)
}

func (r *GeminiRestorer) configDir() (string, error) {
	return r.baseRestorer.configDir("GEMINI_CLI_HOME", ".gemini")
}

func (r *GeminiRestorer) removeExistingSessionFiles(chatsDir, sessionID string, confirmOverwrite OverwritePrompt) (bool, error) {
	if sessionID == "" {
		return true, nil
	}

	var matches []string
	err := filepath.WalkDir(chatsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode().Type() != 0 {
			return nil
		}
		found, err := r.chatSessionID(path)
		if err != nil || found != sessionID {
			return nil
		}
		matches = append(matches, path)
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

type geminiChatMetadata struct {
	// SessionID is the current Gemini chat session identifier field.
	SessionID string `json:"sessionId"`
	// LegacySessionID is the older Gemini chat session identifier field.
	LegacySessionID string `json:"session_id"`
}

func (r *GeminiRestorer) chatSessionID(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var metadata geminiChatMetadata
	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return "", err
	}
	if metadata.SessionID != "" {
		return metadata.SessionID, nil
	}
	return metadata.LegacySessionID, nil
}
