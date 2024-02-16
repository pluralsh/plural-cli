package plural

import (
	"bytes"
	"io"
	"os"

	"github.com/pluralsh/gqlclient"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	lua "github.com/pluralsh/plural-cli/pkg/scaffold/template"
	"github.com/pluralsh/plural-cli/pkg/template"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

func testTemplate(c *cli.Context) error {
	conf := config.Read()
	client := api.NewClient()
	installations, _ := client.GetInstallations()
	repoName := c.Args().Get(0)
	templateTypeFlag := c.String("templateType")
	templateType := gqlclient.TemplateTypeGotemplate
	testTemplate, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	if templateTypeFlag != "" {
		templateType = gqlclient.TemplateType(templateTypeFlag)
	}

	for _, installation := range installations {
		if installation.Repository.Name != repoName {
			continue
		}

		var output []byte
		vals := genDefaultValues(conf, installation)

		if templateType == gqlclient.TemplateTypeLua {
			output, err = luaTmpValues(string(testTemplate), vals)
			if err != nil {
				return err
			}
		} else {
			output, err = goTmpValues(string(testTemplate), vals)
			if err != nil {
				return err
			}
		}
		if _, err := os.Stdout.Write(output); err != nil {
			return err
		}
	}

	return nil
}

func genDefaultValues(conf config.Config, installation *api.Installation) map[string]interface{} {
	return map[string]interface{}{
		"Values":   installation.Context,
		"License":  installation.LicenseKey,
		"Region":   "region",
		"Project":  "example",
		"Cluster":  "cluster",
		"Provider": "provider",
		"Config":   conf,
		"Context":  map[string]interface{}{},
	}
}

func goTmpValues(valuesTmpl string, defaultValues map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(5 * 1024)
	tmpl, err := template.MakeTemplate(valuesTmpl)
	if err != nil {
		return nil, err
	}
	if err = tmpl.Execute(&buf, defaultValues); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func luaTmpValues(valuesTmpl string, defaultValues map[string]interface{}) ([]byte, error) {
	output, err := lua.ExecuteLua(defaultValues, valuesTmpl)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(output)
}

type GrafanaDashboard struct {
	Title  string
	Panels []struct {
		Title   string
		Targets []struct {
			Expr         string
			LegendFormat string
		}
	}
}
