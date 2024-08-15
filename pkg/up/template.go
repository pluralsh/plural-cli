package up

import (
	"bytes"
	"text/template"
	"time"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/retry"
)

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
		"Subdomain":      ctx.Manifest.Network.Subdomain,
		"Network":        ctx.Manifest.Network,
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
