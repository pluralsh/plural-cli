//go:build generate

package ui

import (
	"log"
)

// Used to generate frontend bindings used to call backend.
// Only needed when 'generate' tag is provided during the build.
func init() {
	err := Run(nil, nil)
	if err != nil {
		log.Fatal(err)
	}
}
