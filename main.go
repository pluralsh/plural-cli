package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/fatih/color"

	"github.com/pluralsh/plural/cmd/plural"
)

func main() {
	rand.Seed(time.Now().UnixNano())
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
