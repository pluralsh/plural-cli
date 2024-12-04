package wkspace

import (
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
)

type Workspace struct {
	Provider provider.Provider
	Config   *config.Config
	Context  *manifest.Context
}
