package server

import (
	"os"
	"syscall"
	"context"
	"time"
	"os/signal"
	"net/http"
	"github.com/gin-gonic/gin"
)

func Run() error {
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.POST("/setup", serverFunc(setupCli))
	}

	term := make(chan os.Signal) // OS termination signal
	fail := make(chan error)     // Teardown failure signal

	go func() {
			signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
			<-term // waits for termination signal
			// context with 30s timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			// all teardown process must complete within 30 seconds
			fail <- teardown(ctx)
	}()
	
	if err := r.Run(":8080"); err != nil && err != http.ErrServerClosed {
		return err
	}

	return <-fail
}

func teardown(ctx context.Context) error {
	return syncGit()
}