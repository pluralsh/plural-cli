package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/helm"
	"github.com/pluralsh/plural/pkg/output"
	"github.com/pluralsh/plural/pkg/pluralfile"
	scftmpl "github.com/pluralsh/plural/pkg/scaffold/template"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

func (p *Plural) pushCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "terraform",
			Usage:     "pushes a terraform module",
			ArgsUsage: "path/to/module REPO",
			Action:    latestVersion(p.handleTerraformUpload),
		},
		{
			Name:      "helm",
			Usage:     "pushes a helm chart",
			ArgsUsage: "path/to/chart REPO",
			Action:    latestVersion(handleHelmUpload),
		},
		{
			Name:      "recipe",
			Usage:     "pushes a recipe",
			ArgsUsage: "path/to/recipe.yaml REPO",
			Action:    latestVersion(p.handleRecipeUpload),
		},
		{
			Name:      "artifact",
			Usage:     "creates an artifact for the repo",
			ArgsUsage: "path/to/def.yaml REPO",
			Action:    latestVersion(p.handleArtifact),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "platform",
					Value: "mac",
					Usage: "name of the OS this binary is built for",
				},
				cli.StringFlag{
					Name:  "arch",
					Value: "amd64",
					Usage: "machine architecture the binary is compatible with",
				},
			},
		},
		{
			Name:      "crd",
			Usage:     "registers a new crd for a chart",
			ArgsUsage: "path/to/def.yaml REPO CHART",
			Action:    latestVersion(p.createCrd),
		},
	}
}

func apply(c *cli.Context) error {
	path, _ := os.Getwd()
	var file = pathing.SanitizeFilepath(filepath.Join(path, "Pluralfile"))
	if c.IsSet("file") {
		file, _ = filepath.Abs(c.String("file"))
	}

	if err := os.Chdir(filepath.Dir(file)); err != nil {
		return err
	}

	plrl, err := pluralfile.Parse(file)
	if err != nil {
		return err
	}

	lock, err := plrl.Lock(file)
	if err != nil {
		return err
	}
	return plrl.Execute(file, lock)
}

func (p *Plural) handleTerraformUpload(c *cli.Context) error {
	p.InitPluralClient()
	_, err := p.UploadTerraform(c.Args().Get(0), c.Args().Get(1))
	return api.GetErrorResponse(err, "UploadTerraform")
}

func handleHelmTemplate(c *cli.Context) error {
	conf := config.Read()
	f, err := tmpValuesFile(c.String("values"), &conf)
	if err != nil {
		return fmt.Errorf("Failed to template vals: %w", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(f.Name())

	cmd := exec.Command("helm", "template", c.Args().Get(0), "-f", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func handleHelmUpload(c *cli.Context) error {
	conf := config.Read()
	pth, repo := c.Args().Get(0), c.Args().Get(1)

	f, err := buildValuesFromTemplate(pth, &conf)
	if err != nil {
		return err
	}
	defer func(name string) {
		_ = os.Remove(name)

	}(f.Name())

	utils.Highlight("linting helm: ")
	cmd, outputWriter := executor.SuppressedCommand("helm", "lint", pth, "-f", f.Name())
	if err := executor.RunCommand(cmd, outputWriter); err != nil {
		return err
	}

	cmUrl := fmt.Sprintf("%s/cm/%s", conf.BaseUrl(), repo)
	return helm.Push(pth, cmUrl)
}

func buildValuesFromTemplate(pth string, conf *config.Config) (f *os.File, err error) {
	isLuaTemplate := false
	templatePath := pathing.SanitizeFilepath(filepath.Join(pth, "values.yaml.tpl"))
	valuesTmpl, err := utils.ReadFile(templatePath)
	if os.IsNotExist(err) {
		templatePath = pathing.SanitizeFilepath(filepath.Join(pth, "values.yaml.lua"))
		valuesTmpl, err = utils.ReadFile(templatePath)
		if err != nil {
			return nil, err
		}
		isLuaTemplate = true
	}

	if !isLuaTemplate {
		return tmpValuesFile(templatePath, conf)
	}
	f, err = os.CreateTemp("", "values.yaml")
	if err != nil {
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	vals := genDefaultValues(conf)

	output, err := scftmpl.ExecuteLua(vals, valuesTmpl)
	if err != nil {
		return nil, err
	}

	io, err := yaml.Marshal(output)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(io))
	_, err = f.Write(io)
	if err != nil {
		return nil, err
	}
	return
}

func tmpValuesFile(path string, conf *config.Config) (f *os.File, err error) {
	valuesTmpl, err := utils.ReadFile(path)
	if err != nil {
		return
	}
	tmpl, err := template.MakeTemplate(valuesTmpl)
	if err != nil {
		return
	}

	vals := genDefaultValues(conf)
	var buf bytes.Buffer

	if err = tmpl.Execute(&buf, vals); err != nil {
		return
	}

	f, err = os.CreateTemp("", "values.yaml")
	if err != nil {
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	fmt.Println(buf.String())
	err = wkspace.FormatValues(f, buf.String(), output.New())
	return
}

func (p *Plural) handleRecipeUpload(c *cli.Context) error {
	p.InitPluralClient()
	fullPath, _ := filepath.Abs(c.Args().Get(0))
	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	recipeInput, err := api.ConstructRecipe(contents)
	if err != nil {
		return err
	}

	_, err = p.CreateRecipe(c.Args().Get(1), recipeInput)
	return api.GetErrorResponse(err, "CreateRecipe")
}

func (p *Plural) handleArtifact(c *cli.Context) error {
	p.InitPluralClient()
	fullPath, _ := filepath.Abs(c.Args().Get(0))
	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	input, err := api.ConstructArtifactAttributes(contents)
	if err != nil {
		return err
	}
	input.Platform = c.String("platform")
	input.Arch = c.String("arch")
	_, err = p.CreateArtifact(c.Args().Get(1), input)
	return api.GetErrorResponse(err, "CreateArtifact")
}

func (p *Plural) createCrd(c *cli.Context) error {
	p.InitPluralClient()
	fullPath, _ := filepath.Abs(c.Args().Get(0))
	repo := c.Args().Get(1)
	chart := c.Args().Get(2)
	err := p.CreateCrd(repo, chart, fullPath)
	return api.GetErrorResponse(err, "CreateCrd")
}

func genDefaultValues(conf *config.Config) map[string]interface{} {
	return map[string]interface{}{
		"Values":   map[string]interface{}{},
		"License":  "example-license",
		"Region":   "region",
		"Project":  "example",
		"Cluster":  "cluster",
		"Provider": "provider",
		"Config":   conf,
		"Context":  map[string]interface{}{},
	}
}
