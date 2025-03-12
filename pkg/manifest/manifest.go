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

func (pMan *ProjectManifest) Write(path string) error {
	versioned := &VersionedProjectManifest{
		ApiVersion: "plural.sh/v1alpha1",
		Kind:       "ProjectManifest",
		Metadata:   &Metadata{Name: pMan.Cluster},
		Spec:       pMan,
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

func (man *ProjectManifest) Flush() error {
	return man.Write(ProjectManifestPath())
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

func (pMan *ProjectManifest) Configure(cloud bool, cluster string) Writer {
	pMan.BucketPrefix = cluster
	pMan.Bucket = fmt.Sprintf("plrl-cloud-%s", cluster)

	if !cloud {
		answer := ""
		input := &survey.Input{Message: fmt.Sprintf("Enter a unique, memorable string to use for bucket naming, e.g. an abbreviation for your company:")}
		if err := survey.AskOne(input, &answer, survey.WithValidator(func(val interface{}) error {
			res, _ := val.(string)
			return utils.ValidateRegex(res, "[a-z][0-9\\-a-z]+", "String can only contain alphanumeric characters or hyphens")
		})); err != nil {
			return nil
		}

		pMan.BucketPrefix = answer
		pMan.Bucket = fmt.Sprintf("%s-tf-state", answer)
		if err := pMan.ConfigureNetwork(); err != nil {
			return nil
		}
	}
	return func() error { return pMan.Write(ProjectManifestPath()) }
}

func (pMan *ProjectManifest) ConfigureNetwork() error {
	if pMan.Network != nil {
		return nil
	}

	subdomain := ""
	input := &survey.Input{Message: fmt.Sprintf("Enter subdomain of %s domain that you want to use:", pluralDomain)}
	if err := survey.AskOne(input, &subdomain, survey.WithValidator(func(val interface{}) error {
		res, _ := val.(string)

		if !strings.HasSuffix(res, pluralDomain) {
			res += "." + pluralDomain
		}

		if err := utils.ValidateDns(res); err != nil {
			return err
		}

		client := api.NewClient()
		if err := client.CreateDomain(res); err != nil {
			return fmt.Errorf("Domain %s is taken or your user doesn't have sufficient permissions to create domains", val)
		}

		return nil
	})); err != nil {
		return err
	}

	pMan.Network = &NetworkConfig{Subdomain: subdomain, PluralDns: true}

	return nil
}
