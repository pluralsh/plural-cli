package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pluralsh/plural-cli/pkg/application"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
)

func listApplications(c *gin.Context) error {
	kubeConf, err := kubernetes.KubeConfig()
	if err != nil {
		return err
	}

	apps, err := application.ListAll(kubeConf)
	if err != nil {
		return err
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, apps)
	return nil
}
