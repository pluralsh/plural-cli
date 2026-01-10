package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v2"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/pluralsh/polly/algorithms"
)

const pluralDomain = "onplural.sh"

type Writer func() error

func ProjectManifestPath() string {
	root, found := utils.ProjectRoot()
	if !found {
		path, _ := filepath.Abs("workspace.yaml")
		return path
	}

	return pathing.SanitizeFilepath(filepath.Join(root, "workspace.yaml"))
}

func (pm *ProjectManifest) Write(path string) error {
	versioned := &VersionedProjectManifest{
		ApiVersion: "plural.sh/v1alpha1",
		Kind:       "ProjectManifest",
		Metadata:   &Metadata{Name: pm.Cluster},
		Spec:       pm,
	}

	io, err := yaml.Marshal(&versioned)
	if err != nil {
		return err
	}

	return os.WriteFile(path, io, 0644)
}

func FetchProject() (*ProjectManifest, error) {
	path := ProjectManifestPath()
	return ReadProject(path)
}

func ReadProject(path string) (man *ProjectManifest, err error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("could not find workspace.yaml file, you might need to run `plural init`")
		return
	}

	versioned := &VersionedProjectManifest{}
	err = yaml.Unmarshal(contents, versioned)
	if err != nil || versioned.Spec == nil {
		man = &ProjectManifest{}
		err = yaml.Unmarshal(contents, man)
		return
	}

	man = versioned.Spec

	man.Provider = api.NormalizeProvider(man.Provider)

	return
}

func (pm *ProjectManifest) Flush() error {
	return pm.Write(ProjectManifestPath())
}

func (man *Manifest) Write(path string) error {
	versioned := &VersionedManifest{
		ApiVersion: "plural.sh/v1alpha1",
		Kind:       "Manifest",
		Metadata:   &Metadata{Name: man.Name},
		Spec:       man,
	}

	io, err := yaml.Marshal(&versioned)
	if err != nil {
		return err
	}

	return os.WriteFile(path, io, 0644)
}

func (pm *ProjectManifest) Configure(cloud bool, cluster string) Writer {
	pm.BucketPrefix = cluster
	pm.Bucket = fmt.Sprintf("plrl-cloud-%s-%s", cluster, algorithms.String(4))

	if !cloud {
		answer := ""
		input := &survey.Input{Message: "Enter a unique, memorable string to use for bucket naming, e.g. an abbreviation for your company:"}
		if err := survey.AskOne(input, &answer, survey.WithValidator(func(val interface{}) error {
			res, _ := val.(string)
			return utils.ValidateRegex(res, "[a-z][0-9\\-a-z]+", "bucket name can only contain alphanumeric characters or hyphens")
		})); err != nil {
			return nil
		}

		pm.BucketPrefix = answer
		pm.Bucket = fmt.Sprintf("%s-tf-state", answer)
		if err := pm.ConfigureNetwork(); err != nil {
			return nil
		}
	}
	return func() error { return pm.Write(ProjectManifestPath()) }
}

func (pm *ProjectManifest) ConfigureNetwork() error {
	if pm.Network != nil {
		return nil
	}

	subdomain := ""
	input := &survey.Input{Message: fmt.Sprintf("Enter subdomain of %s domain that you want to use:", pluralDomain)}
	if err := survey.AskOne(input, &subdomain, survey.WithValidator(func(val any) error {
		d := domain(val.(string))
		if err := utils.ValidateDns(d); err != nil {
			return err
		}

		client := api.NewClient()
		if err := client.CreateDomain(d); err != nil {
			return fmt.Errorf("domain %s is taken or your user doesn't have sufficient permissions to create domains", val)
		}

		return nil
	})); err != nil {
		return err
	}

	pm.Network = &NetworkConfig{Subdomain: domain(subdomain), PluralDns: true}

	return nil
}

func domain(subdomain string) string {
	if strings.HasSuffix(subdomain, pluralDomain) {
		return subdomain
	}

	return subdomain + "." + pluralDomain
}
