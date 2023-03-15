//go:build generate

package ui

import (
	"log"
)

func init() {
	err := Run()
	if err != nil {
		log.Fatal(err)
	}
}
