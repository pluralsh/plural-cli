package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/samber/lo"
)

type ConfigurationUpdate struct {
	Configuration map[string]map[string]interface{} `json:"configuration,omitempty"`
	Buckets       []string                          `json:"buckets"`
	Domains       []string                          `json:"domains"`
	Bundles       []*manifest.Bundle                `json:"bundles"`
}

func contextConfiguration(c *gin.Context) error {
	var update ConfigurationUpdate
	if err := c.BindJSON(&update); err != nil {
		return err
	}

	context, err := manifest.ReadContext(manifest.ContextPath())
	if err != nil {
		return err
	}

	for k, v := range update.Configuration {
		context.Configuration[k] = v
	}

	context.Buckets = lo.Uniq(append(context.Buckets, update.Buckets...))
	context.Domains = lo.Uniq(append(context.Domains, update.Domains...))
	context.Bundles = lo.UniqBy(append(context.Bundles, update.Bundles...), func(b *manifest.Bundle) string {
		return fmt.Sprintf("%s:%s", b.Repository, b.Name)
	})

	if err := context.Write(manifest.ContextPath()); err != nil {
		return err
	}

	c.JSON(http.StatusOK, context)
	return nil
}
