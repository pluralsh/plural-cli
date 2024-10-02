package push

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/client"

	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/helm"
	scftmpl "github.com/pluralsh/plural-cli/pkg/scaffold/template"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/urfave/cli"
	"sigs.k8s.io/yaml"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "push",
		Usage:       "utilities for pushing tf or helm packages",
		Subcommands: p.pushCommands(),
		Category:    "Publishing",
	}
}

func (p *Plural) pushCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "terraform",
			Usage:     "pushes a terraform module",
			ArgsUsage: "{path-to-module} {repo}",
			Action:    common.LatestVersion(p.handleTerraformUpload),
		},
		{
			Name:      "helm",
			Usage:     "pushes a helm chart",
			ArgsUsage: "{path-to-chart} {repo}",
			Action:    common.LatestVersion(handleHelmUpload),
		},
		{
			Name:      "recipe",
			Usage:     "pushes a recipe",
			ArgsUsage: "{path-to-recipe} {repo}",
			Action:    common.LatestVersion(p.handleRecipeUpload),
		},
		{
			Name:      "artifact",
			Usage:     "creates an artifact for the repo",
			ArgsUsage: "{path-to-def} {repo}",
			Action:    common.LatestVersion(p.handleArtifact),
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
			ArgsUsage: "{path-to-def} {repo} {chart}",
			Action:    common.LatestVersion(p.createCrd),
		},
	}
}

func (p *Plural) handleTerraformUpload(c *cli.Context) error {
	p.InitPluralClient()
	_, err := p.UploadTerraform(c.Args().Get(0), c.Args().Get(1))
	return api.GetErrorResponse(err, "UploadTerraform")
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
