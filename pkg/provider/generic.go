package provider

import (
	"fmt"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/kubernetes"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	MINIO = "minio"
)

type GENERICProvider struct {
	Clust        string `survey:"cluster"`
	Proj         string
	bucket       string
	Reg          string
	ctx          map[string]interface{}
	ObjectStr    *manifest.ObjectStoreConfig
	writer       manifest.Writer
	configAccess clientcmd.ConfigAccess
}

func (gen *GENERICProvider) getContexts() (string, []string) {
	output := []string{}
	config, err := gen.configAccess.GetStartingConfig()
	if err != nil {
		utils.HighlightError(err)
		return "", output
	}

	for name := range config.Contexts {
		output = append(output, name)
	}
	return config.CurrentContext, output
}

func getContexts() []string {
	options := GENERICProvider{
		configAccess: clientcmd.NewDefaultPathOptions(),
	}
	_, contexts := options.getContexts()
	return contexts
}

func getCurrentContext() string {
	options := GENERICProvider{
		configAccess: clientcmd.NewDefaultPathOptions(),
	}

	currentContext, _ := options.getContexts()
	return currentContext
}

var genericSurvey = []*survey.Question{
	{
		Name: "cluster",
		Prompt: &survey.Select{
			Message: "Select the kubeconfig context to use for this cluster:",
			Options: getContexts(),
			Default: getCurrentContext(),
		},
	},
	{
		Name: "osProvider",
		Prompt: &survey.Select{
			Message: "Select an object store provider:",
			Options: []string{MINIO},
		},
	},
	{
		Name: "osEndpoint",
		Prompt: &survey.Input{
			Message: "Enter the endpoint for your object store:",
		},
	},
	{
		Name: "osUsername",
		Prompt: &survey.Input{
			Message: "Enter the username for your object store:",
		},
	},
	{
		Name: "osPassword",
		Prompt: &survey.Password{
			Message: "Enter the password for your object store:",
		},
	},
	{
		Name: "osSSL",
		Prompt: &survey.Confirm{
			Message: "Does your object store use SSL?",
			Default: true,
		},
	},
	{
		Name: "osInsecure",
		Prompt: &survey.Confirm{
			Message: "Should we validate the SSL certificate for your object store?",
			Default: true,
		},
	},
}

func mkGeneric(conf config.Config) (provider *GENERICProvider, err error) {
	var resp struct {
		Cluster    string
		OsProvider string
		OsEndpoint string
		OsUsername string
		OsPassword string
		OsSSL      bool
		OsInsecure bool
	}
	if err = survey.Ask(genericSurvey, &resp); err != nil {
		return
	}

	provider = &GENERICProvider{
		resp.Cluster,
		"",
		"",
		"us-east-1",
		map[string]interface{}{},
		&manifest.ObjectStoreConfig{
			Provider: resp.OsProvider,
			Endpoint: resp.OsEndpoint,
			Username: resp.OsUsername,
			Password: resp.OsPassword,
			Ssl:      resp.OsSSL,
			Insecure: resp.OsInsecure,
		},
		nil,
		clientcmd.NewDefaultPathOptions(),
	}

	projectManifest := manifest.ProjectManifest{
		Cluster:     provider.Cluster(),
		Project:     provider.Project(),
		Provider:    GENERIC,
		Region:      provider.Region(),
		Context:     provider.Context(),
		ObjectStore: provider.ObjectStore(),
		Owner:       &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}

	provider.writer = projectManifest.Configure()
	provider.bucket = projectManifest.Bucket
	return
}

func genericFromManifest(man *manifest.ProjectManifest) (*GENERICProvider, error) {
	return &GENERICProvider{man.Cluster, man.Project, man.Bucket, man.Region, man.Context, man.ObjectStore, nil, clientcmd.NewDefaultPathOptions()}, nil
}

func (gen *GENERICProvider) CreateBackend(prefix string, version string, ctx map[string]interface{}) (string, error) {

	ctx["Region"] = gen.Region()
	ctx["Bucket"] = gen.Bucket()
	ctx["Prefix"] = prefix
	ctx["ClusterCreated"] = false
	ctx["__CLUSTER__"] = gen.Cluster()
	if cluster, ok := ctx["cluster"]; ok {
		ctx["Cluster"] = cluster
		ctx["ClusterCreated"] = true
	} else {
		ctx["Cluster"] = fmt.Sprintf(`"%s"`, gen.Cluster())
	}

	if err := utils.WriteFile(pathing.SanitizeFilepath(filepath.Join(gen.Bucket(), ".gitignore")), []byte("!/**")); err != nil {
		return "", err
	}
	if err := utils.WriteFile(pathing.SanitizeFilepath(filepath.Join(gen.Bucket(), ".gitattributes")), []byte("/** filter=plural-crypt diff=plural-crypt\n.gitattributes !filter !diff")); err != nil {
		return "", err
	}
	scaffold, err := GetProviderScaffold("GENERIC", version)
	if err != nil {
		return "", err
	}
	return template.RenderString(scaffold, ctx)
}

func (gen *GENERICProvider) KubeConfig() error {

	if kubernetes.InKubernetes() {
		return nil
	}

	config, err := gen.configAccess.GetStartingConfig()
	if err != nil {
		return err
	}
	config.CurrentContext = gen.Cluster()
	return clientcmd.ModifyConfig(gen.configAccess, *config, true)
}

func (gen *GENERICProvider) Name() string {
	return GENERIC
}

func (gen *GENERICProvider) Cluster() string {
	return gen.Clust
}

func (gen *GENERICProvider) Project() string {
	return gen.Proj
}

func (gen *GENERICProvider) Bucket() string {
	return gen.bucket
}

func (gen *GENERICProvider) Region() string {
	return gen.Reg
}

func (gen *GENERICProvider) Context() map[string]interface{} {
	return gen.ctx
}

func (gen *GENERICProvider) ObjectStore() *manifest.ObjectStoreConfig {
	return gen.ObjectStr
}

func (prov *GENERICProvider) Decommision(node *v1.Node) error {
	return nil
}

func (prov *GENERICProvider) Preflights() []*Preflight {
	return nil
}

func (gen *GENERICProvider) Flush() error {
	if gen.writer == nil {
		return nil
	}

	return gen.writer()
}
