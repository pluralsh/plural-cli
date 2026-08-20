package welcome

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/samber/lo"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

// LocalSource reads existing CLI state without returning credentials or
// contacting remote APIs.
type LocalSource struct{ version string }

func NewLocalSource(version string) LocalSource {
	return LocalSource{version: version}
}

func (s LocalSource) Read(ctx context.Context) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}

	snapshot := Snapshot{Version: s.version}
	snapshot.App = s.readAppProfile()
	snapshot.Console = s.readConsole()
	s.readWorkspace(&snapshot)
	s.readKubeContext(&snapshot)
	return snapshot, nil
}

func (s LocalSource) readAppProfile() AppProfile {
	if !config.Exists() {
		return AppProfile{}
	}

	conf := config.Read()
	profile := AppProfile{
		Configured: conf.Email != "" || conf.Token != "",
		Name:       lo.CoalesceOrEmpty(conf.ProfileName(), "active"),
		Email:      conf.Email,
		Endpoint:   conf.BaseUrl(),
	}
	if profiles, err := config.Profiles(); err == nil {
		profile.SavedProfiles = len(profiles)
	}
	return profile
}

func (s LocalSource) readConsole() ConsoleConnection {
	conf := console.ReadConfig()
	return ConsoleConnection{
		Configured: conf.Url != "" || conf.Token != "",
		URL:        conf.Url,
	}
}

func (s LocalSource) readWorkspace(snapshot *Snapshot) {
	root, found := utils.ProjectRoot()
	if !found {
		return
	}

	project, err := manifest.ReadProject(filepath.Join(root, "workspace.yaml"))
	if err != nil {
		snapshot.Diagnostics = append(snapshot.Diagnostics, fmt.Sprintf("workspace: %v", err))
		return
	}
	snapshot.Workspace = Workspace{
		Configured: true,
		Path:       root,
		Name:       project.Cluster,
		Project:    project.Project,
		Provider:   project.Provider,
		Region:     project.Region,
	}
	if project.Owner != nil {
		snapshot.Workspace.Owner = project.Owner.Email
	}
}

func (s LocalSource) readKubeContext(snapshot *Snapshot) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	raw, err := rules.Load()
	if err != nil {
		snapshot.Diagnostics = append(snapshot.Diagnostics, fmt.Sprintf("kubeconfig: %v", err))
		return
	}
	snapshot.KubeContext = raw.CurrentContext
}
