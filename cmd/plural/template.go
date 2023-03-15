package plural

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/pluralsh/gqlclient"

	"github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	lua "github.com/pluralsh/plural/pkg/scaffold/template"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
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

func formatDashboard(c *cli.Context) error {
	if err := v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}
	s := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, scheme.Scheme,
		scheme.Scheme)

	dashboard := v1alpha1.Dashboard{}
	grafana := GrafanaDashboard{}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &grafana); err != nil {
		return err
	}

	graphs := make([]*v1alpha1.DashboardGraph, 0)
	for _, panel := range grafana.Panels {
		graph := &v1alpha1.DashboardGraph{}
		graph.Name = panel.Title
		graph.Queries = make([]*v1alpha1.GraphQuery, 0)
		for _, target := range panel.Targets {
			query := &v1alpha1.GraphQuery{
				Query:  target.Expr,
				Legend: target.LegendFormat,
			}
			graph.Queries = append(graph.Queries, query)
		}
		graphs = append(graphs, graph)
	}

	dashboard.Spec.Graphs = graphs
	dashboard.Spec.Timeslices = []string{"1h", "2h", "6h", "1d", "7d"}
	dashboard.Spec.DefaultTime = "1h"
	dashboard.Spec.Name = grafana.Title

	return s.Encode(&dashboard, os.Stdout)
}
