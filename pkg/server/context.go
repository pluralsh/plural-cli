package server

import (
	"encoding/json"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
	"github.com/pluralsh/plural/pkg/manifest"
)

func contextConfiguration(c *gin.Context) error {
	var configuration map[string]map[string]interface{}
	if err := c.BindJSON(&configuration); err != nil {
		return err
	}

	configurationBytes, err := json.Marshal(configuration)
	if err != nil {
		return err
	}

	context, err := manifest.ReadContext(manifest.ContextPath())
	if err != nil {
		return err
	}

	contextBytes, err := json.Marshal(context)
	if err != nil {
		return err
	}

	patchedJSON, err := jsonpatch.MergePatch(contextBytes, configurationBytes)
	if err != nil {
		return err
	}
	var patchContext manifest.Context
	err = json.Unmarshal(patchedJSON, &patchContext)
	if err != nil {
		return err
	}

	if err := patchContext.Write(manifest.ContextPath()); err != nil {
		return err
	}

	c.JSON(http.StatusOK, patchContext)
	return nil
}
