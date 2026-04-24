package up

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
	"time"

	"github.com/pluralsh/console/go/polly/retry"

	"github.com/pluralsh/plural-cli/pkg/api"
	plrltpl "github.com/pluralsh/plural-cli/pkg/template"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func (c *Context) redact(file string) error {
	buf, err := utils.ReadFile(file)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`(?msU)# BEGIN REMOVE(.*)# END REMOVE`)
	return utils.WriteFile(file, []byte(re.ReplaceAllString(buf, "")))
}

func (c *Context) uncomment(file string) error {
	buf, err := utils.ReadFile(file)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`# UNCOMMENT\s+`)
	return utils.WriteFile(file, []byte(re.ReplaceAllString(buf, "")))
}

func (c *Context) templateFrom(file, to string) error {
	buf, err := utils.ReadFile(file)
	if err != nil {
		return err
	}

	res, err := c.template(buf)
	if err != nil {
		return err
	}

	return utils.WriteFile(to, []byte(res))
}

func (c *Context) template(tmplate string) (string, error) {
	cluster, provider := c.Provider.Cluster(), c.Provider.Name()

	client := api.NewClient()

	me, err := client.Me()
	if err != nil {
		return "", fmt.Errorf("you must run `plural login` before installing")
	}
	eabCredential := &api.EabCredential{}
	if c.Provider.Name() != api.BYOK && !c.ignorePreflights {
		retrier := retry.NewConstant(15*time.Millisecond, 3)
		eabCredential, err = retry.Retry(retrier, func() (*api.EabCredential, error) { return client.GetEabCredential(cluster, provider) })
		if err != nil {
			return "", err
		}
	}

	values := map[string]interface{}{
		"UserEmail":      me.Email,
		"Cluster":        cluster,
		"Provider":       provider,
		"Bucket":         c.Provider.Bucket(),
		"Project":        c.Provider.Project(),
		"Region":         c.Provider.Region(),
		"Context":        c.Provider.Context(),
		"Config":         c.Config,
		"RepoUrl":        c.RepoUrl,
		"Identifier":     c.identifier(),
		"Acme":           eabCredential,
		"StacksIdentity": c.StacksIdentity,
		"RequireDB":      !c.Cloud,
		"CloudCluster":   c.CloudCluster,
		"Cloud":          c.Cloud,
		"ClusterName":    cluster,
		"ProjectID":      c.Provider.Project(),
		"GitUsername":    c.GitUsername,
		"GitPassword":    c.GitPassword,
	}
	if c.Manifest.Network != nil {
		values["Subdomain"] = c.Manifest.Network.Subdomain
		values["Network"] = c.Manifest.Network
	}
	if c.Manifest.AppDomain != "" {
		values["AppDomain"] = c.Manifest.AppDomain
	}

	tpl := template.New("tpl").Funcs(plrltpl.GetFuncMap())
	if c.Delims != nil {
		tpl.Delims(c.Delims.left, c.Delims.right)
	}

	readyTpl, err := tpl.Parse(tmplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = readyTpl.Execute(&buf, values)
	return buf.String(), err
}
