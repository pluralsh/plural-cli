package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"

	"github.com/pluralsh/plural-cli/cmd/command/plural"
)

func main() {
	// init Kube when k8s config exists
	p := &plural.Plural{}
	fmt.Println("HELLO AND WELCOME TO CUSTOM PLURAL CLI")
	app := plural.CreateNewApp(p)
	if os.Getenv("ENABLE_COLOR") != "" {
		color.NoColor = false
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
