package up

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

func ping(url string) error {
	done := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	go func() {
		for {
			fmt.Printf("Pinging %s...\n", url)
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == 200 {
				utils.Success("Found status code 200, console up!\n")
				done <- true
				return
			}
			time.Sleep(10 * time.Second)
		}
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("Console failed to become ready after 5 minutes, you might want to inspect the resources in the plrl-console namespace")
	}
}
