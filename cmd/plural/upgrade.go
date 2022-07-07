package main

import (
	"fmt"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/urfave/cli"
)

func handleUpgrade(c *cli.Context) error {
	name := c.String("name")
	message := c.String("message")
	client := api.NewClient()

	id, err := client.CreateUpgrade(name, message)
	if err != nil {
		return err
	}

	fmt.Printf("Created upgrade: %s", id)
	return nil
}
