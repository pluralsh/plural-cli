package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	v1 "k8s.io/api/core/v1"

	"github.com/joho/godotenv"
	"github.com/linode/linodego"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
)

type LinodeProvider struct {
	cluster string `survey:"cluster"`
	bucket  string
	region  string `survey:"region"`
	client  *linodego.Client
	ctx     map[string]interface{}
	writer  manifest.Writer
}

//nolint:all
func getLinodeSurvey() (surveys []*survey.Question, err error) {
	client, err := linodeClient()
	if err != nil {
		return
	}
	regions, err := client.ListRegions(context.Background(), nil)
	if err != nil {
		return
	}
	linodeRegions := algorithms.Map(regions, func(r linodego.Region) string { return r.ID })
	surveys = []*survey.Question{
		{
			Name:     "cluster",
			Prompt:   &survey.Input{Message: "Enter the name of your cluster"},
			Validate: validCluster,
		},
		{
			Name:     "region",
			Prompt:   &survey.Select{Message: "What region will you deploy to?", Default: "us-east", Options: linodeRegions},
			Validate: survey.Required,
		},
	}
	return
}

//nolint:all
func mkLinode(conf config.Config) (provider *LinodeProvider, err error) {
	provider = &LinodeProvider{}
	s, err := getLinodeSurvey()
	if err != nil {
		return
	}

	if err = survey.Ask(s, provider); err != nil {
		return
	}

	client, err := linodeClient()
	if err != nil {
		return
	}

	provider.client = client

	projectManifest := manifest.ProjectManifest{
		Cluster:  provider.Cluster(),
		Project:  provider.Project(),
		Provider: api.ProviderGCP,
		Region:   provider.Region(),
		Context:  provider.Context(),
		Owner:    &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}

	provider.writer = projectManifest.Configure(cloudFlag, provider.Cluster())
	provider.bucket = projectManifest.Bucket
	return
}

//nolint:all
func linodeFromManifest(man *manifest.ProjectManifest) (*LinodeProvider, error) {
	client, err := linodeClient()
	if err != nil {
		return nil, err
	}

	return &LinodeProvider{
		client:  client,
		cluster: man.Cluster,
		region:  man.Region,
		bucket:  man.Bucket,
	}, nil
}

//nolint:all
func linodeClient() (*linodego.Client, error) {
	apiKey, ok := os.LookupEnv("LINODE_TOKEN")
	if !ok {
		return nil, fmt.Errorf("env var LINODE_TOKEN must be set")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	return lo.ToPtr(linodego.NewClient(oauth2Client)), nil
}

func linodeEnvFile() (string, error) {
	root, err := git.Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".env"), nil
}

func (l *LinodeProvider) CreateBucket() error {
	envPath, err := linodeEnvFile()
	if err != nil {
		return err
	}

	if utils.Exists(envPath) {
		return godotenv.Load(envPath)
	}

	ctx := context.Background()
	if _, err := l.client.CreateObjectStorageBucket(ctx, linodego.ObjectStorageBucketCreateOptions{
		Cluster: "us-east-1",
		Label:   l.bucket,
		ACL:     linodego.ACLPrivate,
	}); err != nil {
		return err
	}

	key, err := l.client.CreateObjectStorageKey(ctx, linodego.ObjectStorageKeyCreateOptions{
		Label: l.bucket,
		BucketAccess: lo.ToPtr([]linodego.ObjectStorageKeyBucketAccess{
			{
				Cluster:     "us-east-1",
				BucketName:  l.bucket,
				Permissions: "read_write",
			},
		}),
	})
	if err != nil {
		return err
	}

	env := fmt.Sprintf("AWS_ACCESS_KEY_ID=%s\nAWS_SECRET_ACCESS_KEY=%s\n", key.AccessKey, key.SecretKey)
	if err := utils.WriteFile(envPath, []byte(env)); err != nil {
		return err
	}

	return godotenv.Load(envPath)
}

func (l *LinodeProvider) CreateBackend(prefix string, version string, ctx map[string]interface{}) (string, error) {
	return "", nil
}

func (l *LinodeProvider) KubeConfig() error {
	if kubernetes.InKubernetes() {
		return nil
	}

	// TODO: implement
	return nil
}

func (l *LinodeProvider) KubeContext() string {
	return fmt.Sprintf("linode-%s", l.Cluster())
}

func (l *LinodeProvider) Name() string {
	return "linode"
}

func (l *LinodeProvider) Cluster() string {
	return l.cluster
}

func (l *LinodeProvider) Project() string {
	return ""
}

func (l *LinodeProvider) Bucket() string {
	return l.bucket
}

func (l *LinodeProvider) Region() string {
	return l.region
}

func (*LinodeProvider) Permissions() (permissions.Checker, error) {
	return permissions.NullChecker(), nil
}

func (l *LinodeProvider) Context() map[string]interface{} {
	return l.ctx
}

func (l *LinodeProvider) Decommision(node *v1.Node) error {
	return nil
}

func (l *LinodeProvider) Preflights() []*Preflight {
	return nil
}

func (l *LinodeProvider) Flush() error {
	if l.writer == nil {
		return nil
	}

	return l.writer()
}
