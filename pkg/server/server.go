package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func SetUpRouter() *gin.Engine {
	r := gin.Default()
	r.Use(ErrorHandler())
	v1 := r.Group("/v1")
	{
		v1.POST("/setup", serverFunc(setupCli))
		v1.GET("/health", healthcheck)
		v1.GET("/configuration", serverFunc(configuration))
		v1.GET("/applications", serverFunc(listApplications))
		v1.GET("/shutdown", serverFunc(shutdown))
		v1.POST("/context/configuration", serverFunc(contextConfiguration))
		v1.POST("/shutdown", serverFunc(shutdown))
	}
	return r
}

func Run() error {
	gin.SetMode(gin.ReleaseMode)
	r := SetUpRouter()

	term := make(chan os.Signal, 1) // OS termination signal
	fail := make(chan error)        // Teardown failure signal

	go func() {
		signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
		<-term // waits for termination signal
		// context with 30s timeout
		_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// all teardown process must complete within 30 seconds
		fail <- teardown()
	}()

	if err := r.Run(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return <-fail
}

func healthcheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func shutdown(c *gin.Context) error {
	if err := syncGit(); err != nil {
		return err
	}

	c.String(http.StatusOK, "OK")
	return nil
}

func teardown() error {
	return syncGit()
}
