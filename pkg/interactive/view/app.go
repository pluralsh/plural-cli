package view

import (
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/interactive/ui"
)

// App represents an application view.
type App struct {
	version string
	*ui.App
	Content *PageStack
	// command       *Command
	// factory       *watch.Factory
	// cancelFn      context.CancelFunc
	// clusterModel  *model.ClusterInfo
	// cmdHistory    *model.History
	// filterHistory *model.History
	// conRetry      int32
	// showHeader    bool
	// showLogo      bool
	// showCrumbs    bool
}

func NewApp(cfg *config.Config) *App {
	a := App{
		App:           ui.NewApp(cfg, cfg.K9s.CurrentContext),
		cmdHistory:    model.NewHistory(model.MaxHistory),
		filterHistory: model.NewHistory(model.MaxHistory),
		Content:       NewPageStack(),
	}

	a.Views()["statusIndicator"] = ui.NewStatusIndicator(a.App, a.Styles)
	a.Views()["clusterInfo"] = NewClusterInfo(&a)

	return &a
}
