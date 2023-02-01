package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/helm"
	"github.com/pluralsh/plural/pkg/pluralfile"
	scftmpl "github.com/pluralsh/plural/pkg/scaffold/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/urfave/cli"
	"sigs.k8s.io/yaml"
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
	path := c.String("values")
	f, err := scftmpl.TmpValuesFile(path)
	if err != nil {
		return err
	}

	defer func(name string) {
		_ = os.Remove(name)
	}(f.Name())

	namespace := "default"
	actionConfig, err := helm.GetActionConfig(namespace)
	if err != nil {
		return err
	}
	values, err := getValues(f.Name())
	if err != nil {
		return err
	}
	res, err := helm.Template(actionConfig, c.Args().Get(0), namespace, "./", false, false, values)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}

func handleHelmUpload(c *cli.Context) error {
	conf := config.Read()
	pth, repo := c.Args().Get(0), c.Args().Get(1)

	f, err := buildValuesFromTemplate(pth)
	if err != nil {
		return err
	}
	defer func(name string) {
		_ = os.Remove(name)

	}(f.Name())

	utils.Highlight("linting helm: ")
	values, err := getValues(f.Name())
	if err != nil {
		return err
	}
	if err := helm.Lint(pth, "default", values); err != nil {
		return err
	}

	cmUrl := fmt.Sprintf("%s/cm/%s", conf.BaseUrl(), repo)
	return helm.Push(pth, cmUrl)
}

func buildValuesFromTemplate(pth string) (f *os.File, err error) {
	templatePath := pathing.SanitizeFilepath(filepath.Join(pth, "values.yaml.tpl"))
	_, err = utils.ReadFile(templatePath)
	if os.IsNotExist(err) {
		templatePath = pathing.SanitizeFilepath(filepath.Join(pth, "values.yaml.lua"))
		_, err := utils.ReadFile(templatePath)
		if err != nil {
			return nil, err
		}

	}

	return scftmpl.TmpValuesFile(templatePath)
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

func getValues(path string) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	valsContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(valsContent, &values); err != nil {
		return nil, err
	}
	return values, nil
}
