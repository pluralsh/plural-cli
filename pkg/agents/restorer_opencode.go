package agents

import (
	"context"
	"path/filepath"

	console "github.com/pluralsh/console/go/client"
)

// OpencodeRestorer imports and resumes Opencode session state.
type OpencodeRestorer struct {
	// baseRestorer supplies shared archive and filesystem helpers.
	baseRestorer
}

func (r *OpencodeRestorer) Provider() console.AgentRuntimeType {
	return console.AgentRuntimeTypeOpencode
}

func (r *OpencodeRestorer) Prepare(_ context.Context, opts RestoreOptions) (*PreparedSession, error) {
	if opts.Manifest.Session.ArchivePath != "" {
		if err := r.archive.ExtractSubtree(opts.ArchivePath, opts.Manifest.Session.ArchivePath, filepath.Join(opts.WorkDir, opts.Manifest.Session.ArchivePath)); err != nil {
			return nil, err
		}
	} else if err := r.archive.ExtractSubtree(opts.ArchivePath, "opencode", filepath.Join(opts.WorkDir, "opencode")); err != nil {
		return nil, err
	}
	return &PreparedSession{
		RepoPath:  opts.RepoPath,
		WorkDir:   opts.WorkDir,
		SessionID: opts.Manifest.Session.ID,
	}, nil
}

func (r *OpencodeRestorer) Resume(ctx context.Context, prepared *PreparedSession) error {
	// import session file
	if err := Executable(prepared.WorkDir).Run(ctx, "opencode", "import", "opencode/agent-session.json"); err != nil {
		return err
	}

	// run imported session
	if err := Executable(prepared.RepoPath).Run(ctx, "opencode", "-s", prepared.SessionID); err != nil {
		return err
	}

	return nil
}
