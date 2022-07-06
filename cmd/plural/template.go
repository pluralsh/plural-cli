package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/urfave/cli"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

func testTemplate(c *cli.Context) error {
	conf := config.Read()
	client := api.NewClient()
	installations, _ := client.GetInstallations()
	repoName := c.Args().Get(0)
	testTemplate, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if installation.Repository.Name != repoName {
			continue
		}

		ctx := installation.Context
		tmpl, err := template.MakeTemplate(string(testTemplate))
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		buf.Grow(5 * 1024)
		vals := map[string]interface{}{
			"Values":   ctx,
			"License":  installation.LicenseKey,
			"Region":   "region",
			"Project":  "example",
			"Cluster":  "cluster",
			"Provider": "provider",
			"Config":   conf,
			"Context":  map[string]interface{}{},
		}
		if err := tmpl.Execute(&buf, vals); err != nil {
			return err
		}

		if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
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
	data, err := ioutil.ReadAll(os.Stdin)
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
