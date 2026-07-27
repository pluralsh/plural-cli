package welcome

import "context"

type AppProfile struct {
	Configured    bool
	Name          string
	Email         string
	Endpoint      string
	SavedProfiles int
}

type ConsoleConnection struct {
	Configured bool
	URL        string
}

type Workspace struct {
	Configured bool
	Path       string
	Name       string
	Project    string
	Provider   string
	Region     string
	Owner      string
}

// Snapshot is the credential-free local state shown by the welcome
// screen.
type Snapshot struct {
	Version     string
	App         AppProfile
	Console     ConsoleConnection
	Workspace   Workspace
	KubeContext string
	Diagnostics []string
}

// Source reads local state without presentation concerns.
type Source interface {
	Read(ctx context.Context) (Snapshot, error)
}

// Loader is the narrow dependency consumed by the TUI.
type Loader interface {
	Load(ctx context.Context) (Snapshot, error)
}

type Service struct{ source Source }

func NewService(source Source) *Service {
	return &Service{source: source}
}

func (s *Service) Load(ctx context.Context) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	return s.source.Read(ctx)
}
