package provider

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
)

type TestProvider struct {
	Clust  string `survey:"cluster"`
	Proj   string
	bucket string
	Reg    string
	ctx    map[string]interface{}
}

func (t TestProvider) Name() string {
	return api.TEST
}

func (t TestProvider) Cluster() string {
	return t.Clust
}

func (t TestProvider) Project() string {
	return t.Proj
}

func (t TestProvider) Region() string {
	return t.Reg
}

func (t TestProvider) Bucket() string {
	return t.bucket
}

func (t TestProvider) KubeConfig() error {
	return nil
}

func (t TestProvider) KubeContext() string {
	return t.Clust
}

func (t TestProvider) CreateBackend(prefix string, version string, ctx map[string]interface{}) (string, error) {
	return "test", nil
}

func (t TestProvider) Context() map[string]interface{} {
	return map[string]interface{}{}
}

func (t TestProvider) Decommision(_ *corev1.Node) error {
	return nil
}

func (t TestProvider) Preflights() []*Preflight {
	return nil
}

func (t TestProvider) Permissions() (permissions.Checker, error) {
	return permissions.NullChecker(), nil
}

func (t TestProvider) Flush() error {
	return nil
}

func testFromManifest(man *manifest.ProjectManifest) (*TestProvider, error) {
	return &TestProvider{man.Cluster, man.Project, man.Bucket, man.Region, man.Context}, nil
}
