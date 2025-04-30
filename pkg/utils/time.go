package utils

import (
	"fmt"
	"time"
)

func WaitFor(timeout, interval time.Duration, f func() (bool, error)) error {
	var lastErr string
	timeup := time.After(timeout)
	for {
		select {
		case <-timeup:
			return fmt.Errorf("time limit exceeded, last error: %s", lastErr)
		default:
		}

		stop, err := f()
		if stop {
			return nil
		}
		if err != nil {
			return err
		}

		time.Sleep(interval)
	}
}
