package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v2"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/exp"
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

func ManifestPath(repo string) (string, error) {
	root, found := utils.ProjectRoot()
	if !found {
		return "", fmt.Errorf("You're not within an installation repo")
	}

	return pathing.SanitizeFilepath(filepath.Join(root, repo, "manifest.yaml")), nil
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

	// Override Cluster API flag silently
	if !exp.IsFeatureEnabled(exp.EXP_PLURAL_CAPI) {
		man.ClusterAPI = false
	}

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

func Read(path string) (man *Manifest, err error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return
	}

	versioned := &VersionedManifest{}
	err = yaml.Unmarshal(contents, versioned)
	if err != nil || versioned.Spec == nil {
		man = &Manifest{}
		err = yaml.Unmarshal(contents, man)
		return
	}

	man = versioned.Spec
	return
}

func (pMan *ProjectManifest) Configure() Writer {
	utils.Highlight("\nLet's get some final information about your workspace set up\n\n")

	res, _ := utils.ReadAlphaNum("Give us a unique, memorable string to use for bucket naming, eg an abbreviation for your company: ")
	pMan.BucketPrefix = res
	pMan.Bucket = fmt.Sprintf("%s-tf-state", res)

	if err := pMan.ConfigureNetwork(); err != nil {
		return nil
	}

	if exp.IsFeatureEnabled(exp.EXP_PLURAL_CAPI) {
		pMan.ClusterAPI = true
	}

	return func() error { return pMan.Write(ProjectManifestPath()) }
}

func (pMan *ProjectManifest) ConfigureNetwork() error {
	if pMan.Network != nil {
		return nil
	}

	utils.Highlight("\nOk, let's get your network configuration set up now...\n")
	pluralDns := utils.Confirm("Do you want to use plural's dns provider?")
	modifier := " (eg something.mydomain.com)"
	if pluralDns {
		modifier = ", must be a subdomain under onplural.sh"
	}

	subdomain := ""
	input := &survey.Input{Message: fmt.Sprintf("\nWhat do you want to use as your domain%s: ", modifier)}
	if err := survey.AskOne(input, &subdomain, survey.WithValidator(func(val interface{}) error {
		res, _ := val.(string)
		if err := utils.ValidateDns(res); err != nil {
			return err
		}

		if pluralDns && !strings.HasSuffix(res, pluralDomain) {
			return fmt.Errorf("Not an onplural.sh domain")
		}

		if pluralDns {
			client := api.NewClient()
			if err := client.CreateDomain(res); err != nil {
				return fmt.Errorf("Domain %s is taken or your user doesn't have sufficient permissions to create domains", val)
			}
		}

		return nil
	})); err != nil {
		return err
	}

	pMan.Network = &NetworkConfig{Subdomain: subdomain, PluralDns: pluralDns}

	return nil
}
