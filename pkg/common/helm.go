package common

import (
	"fmt"
	"os"

	"github.com/pluralsh/plural-cli/pkg/helm"
	scftmpl "github.com/pluralsh/plural-cli/pkg/scaffold/template"
	"github.com/urfave/cli"
	"sigs.k8s.io/yaml"
)

func HandleHelmTemplate(c *cli.Context) error {
	path := c.String("values")
	f, err := scftmpl.TmpValuesFile(path)
	if err != nil {
		return err
	}

	defer func(name string) {
		_ = os.Remove(name)
	}(f.Name())

	name := "default"
	namespace := "default"
	actionConfig, err := helm.GetActionConfig(namespace)
	if err != nil {
		return err
	}
	values, err := getValues(f.Name())
	if err != nil {
		return err
	}
	res, err := helm.Template(actionConfig, name, namespace, c.Args().Get(0), false, false, values)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
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
