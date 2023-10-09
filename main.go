package main

import (
	"log"
	"os"

	"github.com/fatih/color"

	"github.com/pluralsh/plural/cmd/plural"
)

func main() {
	// init Kube when k8s config exists
	p := &plural.Plural{}
	app := plural.CreateNewApp(p)
	if os.Getenv("ENABLE_COLOR") != "" {
		color.NoColor = false
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
