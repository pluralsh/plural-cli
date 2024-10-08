package provider

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"

	"github.com/AlecAivazis/survey/v2"
	v1 "k8s.io/api/core/v1"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	"github.com/pluralsh/plural-cli/pkg/template"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
)

type KINDProvider struct {
	Clust  string `survey:"cluster"`
	Proj   string
	bucket string
	Reg    string
	ctx    map[string]interface{}
	writer manifest.Writer
}

var kindSurvey = []*survey.Question{
	{
		Name:     "cluster",
		Prompt:   &survey.Input{Message: "Enter the name of your cluster:"},
		Validate: validCluster,
	},
}

func mkKind(conf config.Config) (provider *KINDProvider, err error) {
	var resp struct {
		Cluster string
	}
	if err = survey.Ask(kindSurvey, &resp); err != nil {
		return
	}

	provider = &KINDProvider{
		resp.Cluster,
		"",
		"",
		"us-east-1",
		map[string]interface{}{},
		nil,
	}

	projectManifest := manifest.ProjectManifest{
		Cluster:  provider.Cluster(),
		Project:  provider.Project(),
		Provider: api.ProviderKind,
		Region:   provider.Region(),
		Context:  provider.Context(),
		Owner:    &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}

	provider.writer = projectManifest.Configure(cloudFlag, provider.Cluster())
	provider.bucket = projectManifest.Bucket
	return
}

func kindFromManifest(man *manifest.ProjectManifest) (*KINDProvider, error) {
	return &KINDProvider{man.Cluster, man.Project, man.Bucket, man.Region, man.Context, nil}, nil
}

func (kind *KINDProvider) CreateBucket() error { return nil }

func (kind *KINDProvider) CreateBackend(prefix string, version string, ctx map[string]interface{}) (string, error) {

	ctx["Region"] = kind.Region()
	ctx["Bucket"] = kind.Bucket()
	ctx["Prefix"] = prefix
	ctx["ClusterCreated"] = false
	ctx["__CLUSTER__"] = kind.Cluster()
	if cluster, ok := ctx["cluster"]; ok {
		ctx["Cluster"] = cluster
		ctx["ClusterCreated"] = true
	} else {
		ctx["Cluster"] = fmt.Sprintf(`"%s"`, kind.Cluster())
	}

	if err := utils.WriteFile(pathing.SanitizeFilepath(filepath.Join(kind.Bucket(), ".gitignore")), []byte("!/**")); err != nil {
		return "", err
	}
	if err := utils.WriteFile(pathing.SanitizeFilepath(filepath.Join(kind.Bucket(), ".gitattributes")), []byte("/** filter=plural-crypt diff=plural-crypt\n.gitattributes !filter !diff")); err != nil {
		return "", err
	}
	scaffold, err := GetProviderScaffold(api.ToGQLClientProvider(api.ProviderKind), version)
	if err != nil {
		return "", err
	}
	return template.RenderString(scaffold, ctx)
}

func (kind *KINDProvider) KubeConfig() error {
	if kubernetes.InKubernetes() {
		return nil
	}
	cmd := exec.Command(
		"kind", "export", "kubeconfig", "--name", kind.Cluster())
	return utils.Execute(cmd)
}

func (kind *KINDProvider) KubeContext() string {
	return fmt.Sprintf("kind-%s", kind.Cluster())
}

func (kind *KINDProvider) Name() string {
	return api.ProviderKind
}

func (kind *KINDProvider) Cluster() string {
	return kind.Clust
}

func (kind *KINDProvider) Project() string {
	return kind.Proj
}

func (kind *KINDProvider) Bucket() string {
	return kind.bucket
}

func (kind *KINDProvider) Region() string {
	return kind.Reg
}

func (*KINDProvider) Permissions() (permissions.Checker, error) {
	return permissions.NullChecker(), nil
}

func (kind *KINDProvider) Context() map[string]interface{} {
	return kind.ctx
}

func (prov *KINDProvider) Decommision(node *v1.Node) error {
	return nil
}

func (prov *KINDProvider) Preflights() []*Preflight {
	return nil
}

func (kind *KINDProvider) Flush() error {
	if kind.writer == nil {
		return nil
	}

	return kind.writer()
}
