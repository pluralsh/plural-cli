package up

import (
	"fmt"
	"os"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

// ExistingWorkspace is the subset of workspace.yaml the TUI needs when skipping init.
type ExistingWorkspace struct {
	ProviderID   string
	Cluster      string
	Region       string
	Project      string
	BucketPrefix string
	PluralDNS    string
	AppDomain    string
}

// HasWorkspace reports whether ./workspace.yaml (or project-root workspace.yaml) exists.
// Mirrors HandleInitWithProject's early return.
func HasWorkspace() bool {
	return utils.Exists("./workspace.yaml") || utils.Exists(manifest.ProjectManifestPath())
}

// LoadExistingWorkspace reads the project manifest for the skip-init path.
func LoadExistingWorkspace() (ExistingWorkspace, error) {
	pm, err := manifest.FetchProject()
	if err != nil {
		return ExistingWorkspace{}, err
	}
	ws := ExistingWorkspace{
		ProviderID:   api.NormalizeProvider(pm.Provider),
		Cluster:      strings.TrimSpace(pm.Cluster),
		Region:       strings.TrimSpace(pm.Region),
		Project:      strings.TrimSpace(pm.Project),
		BucketPrefix: strings.TrimSpace(pm.BucketPrefix),
		AppDomain:    strings.TrimSpace(pm.AppDomain),
	}
	if pm.Network != nil {
		ws.PluralDNS = strings.TrimSpace(pm.Network.Subdomain)
	}
	if ws.ProviderID == "" {
		return ExistingWorkspace{}, fmt.Errorf("workspace.yaml is missing provider")
	}
	if ws.Cluster == "" {
		return ExistingWorkspace{}, fmt.Errorf("workspace.yaml is missing cluster")
	}
	return ws, nil
}

// EnsureExistingWorkspace mirrors Plural.ensureWorkspace: Plural DNS check, branch context, gitignore.
// Does not print CLI Highlight lines — the TUI shows its own status.
func EnsureExistingWorkspace() error {
	proj, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	if proj.Network != nil && proj.Network.PluralDns {
		client := api.NewClient()
		if err := client.CreateDomain(proj.Network.Subdomain); err != nil {
			return err
		}
	}

	branch, err := git.CurrentBranch()
	if err != nil {
		return err
	}
	if proj.Context == nil {
		proj.Context = map[string]interface{}{}
	}
	proj.Context["Branch"] = branch
	if err := proj.Flush(); err != nil {
		return err
	}
	if err := common.EnsureGitIgnore(); err != nil {
		return err
	}
	return nil
}

// WorkspacePathForTest returns the path Write would use in the current directory.
func WorkspacePathForTest() string {
	if _, err := os.Stat("./workspace.yaml"); err == nil {
		return "./workspace.yaml"
	}
	return manifest.ProjectManifestPath()
}
