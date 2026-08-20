package bridge

import (
	"context"

	"github.com/pluralsh/plural-cli/pkg/config"
)

// LegacyProfileStore preserves config.yml persistence while keeping it out of
// presentation handlers.
type LegacyProfileStore struct{}

func (LegacyProfileStore) Persist(ctx context.Context, conf *config.Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return conf.Flush()
}

func (LegacyProfileStore) Activate(ctx context.Context, conf *config.Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	config.SetConfig(conf)
	return nil
}
