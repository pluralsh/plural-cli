package template

import (
	"fmt"
	"os"
	"strings"

	console "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/polly/template"
)

func RenderYaml(path string, bindings map[string]interface{}) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(path, ".tpl") {
		return template.RenderTpl(content, bindings)
	}

	if strings.HasSuffix(path, ".liquid") {
		return template.RenderLiquid(content, bindings)
	}

	return content, fmt.Errorf("not a .liquid or .tpl file")
}

func RenderService(path string, svc *console.ServiceDeploymentExtended) ([]byte, error) {
	bindings := map[string]interface{}{
		"Configuration": configMap(svc),
		"Cluster":       clusterConfiguration(svc.Cluster),
		"Contexts":      contexts(svc),
	}

	for k, v := range bindings {
		bindings[strings.ToLower(k)] = v
	}

	return RenderYaml(path, bindings)
}
