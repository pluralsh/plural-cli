package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func Run() error {
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.POST("/setup", serverFunc(setupCli))
		v1.GET("/health", healthcheck)
	}

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

	if err := r.Run(":8080"); err != nil && err != http.ErrServerClosed {
		return err
	}

	return <-fail
}

func healthcheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func teardown() error {
	return syncGit()
}
