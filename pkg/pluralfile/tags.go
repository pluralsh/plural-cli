package pluralfile

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils"
	"gopkg.in/yaml.v3"
)

type VersionSpec struct {
	Chart     *string
	Terraform *string
	Version   string
}

type VersionTags struct {
	Spec *VersionSpec
	Tags []string
}

type Tags struct {
	File string
}

func (a *Tags) Type() ComponentName {
	return TAG
}

func (a *Tags) Key() string {
	return a.File
}

func (t *Tags) Push(repo string, sha string) (string, error) {
	newsha, err := executor.MkHash(t.File, []string{})
	if err != nil || newsha == sha {
		if err == nil {
			utils.Highlight("No change for %s\n", t.File)
		}
		return sha, err
	}

	f, err := os.Open(t.File)
	if err != nil {
		return sha, err
	}

	utils.Highlight("updating tags for %s", t.File)
	client := api.NewClient()
	d := yaml.NewDecoder(f)
	for {
		tagSpec := &VersionTags{}
		err = d.Decode(tagSpec)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("")
			return sha, err
		}

		vspec := &api.VersionSpec{
			Repository: repo,
			Chart:      tagSpec.Spec.Chart,
			Terraform:  tagSpec.Spec.Terraform,
			Version:    tagSpec.Spec.Version,
		}
		if err := client.UpdateVersion(vspec, tagSpec.Tags); err != nil {
			fmt.Println("")
			return sha, err
		}

		utils.Highlight(".")
	}

	utils.Success("\u2713\n")
	return newsha, nil
}
