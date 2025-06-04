package up

import (
	"bytes"
	"regexp"
	"text/template"
	"time"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/retry"
)

func (ctx *Context) redact(file string) error {
	buf, err := utils.ReadFile(file)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`(?msU)# BEGIN REMOVE(.*)# END REMOVE`)
	return utils.WriteFile(file, []byte(re.ReplaceAllString(buf, "")))
}

func (ctx *Context) uncomment(file string) error {
	buf, err := utils.ReadFile(file)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`# UNCOMMENT\s+`)
	return utils.WriteFile(file, []byte(re.ReplaceAllString(buf, "")))
}

func (ctx *Context) templateFrom(file, to string) error {
	buf, err := utils.ReadFile(file)
	if err != nil {
		return err
	}

	res, err := ctx.template(buf)
	if err != nil {
		return err
	}

	return utils.WriteFile(to, []byte(res))
}

func (ctx *Context) template(tmplate string) (string, error) {
	cluster, provider := ctx.Provider.Cluster(), ctx.Provider.Name()
	client := api.NewClient()
	retrier := retry.NewConstant(15*time.Millisecond, 3)
	eabCredential, err := retry.Retry(retrier, func() (*api.EabCredential, error) {
		return client.GetEabCredential(cluster, provider)
	})
	if err != nil {
		return "", err
	}

	values := map[string]interface{}{
		"Cluster":        cluster,
		"Provider":       provider,
		"Bucket":         ctx.Provider.Bucket(),
		"Project":        ctx.Provider.Project(),
		"Region":         ctx.Provider.Region(),
		"Context":        ctx.Provider.Context(),
		"Config":         ctx.Config,
		"RepoUrl":        ctx.RepoUrl,
		"Identifier":     ctx.identifier(),
		"Acme":           eabCredential,
		"StacksIdentity": ctx.StacksIdentity,
		"RequireDB":      !ctx.Cloud,
		"CloudCluster":   ctx.CloudCluster,
		"Cloud":          ctx.Cloud,
		"ClusterName":    cluster,
		"ProjectID":      ctx.Provider.Project(),
	}
	if ctx.Manifest.Network != nil {
		values["Subdomain"] = ctx.Manifest.Network.Subdomain
		values["Network"] = ctx.Manifest.Network
	}
	if ctx.Manifest.AppDomain != "" {
		values["AppDomain"] = ctx.Manifest.AppDomain
	}

	tpl := template.New("tpl")
	if ctx.Delims != nil {
		tpl.Delims(ctx.Delims.left, ctx.Delims.right)
	}

	readyTpl, err := tpl.Parse(tmplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = readyTpl.Execute(&buf, values)
	return buf.String(), err
}
