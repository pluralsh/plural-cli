package validation

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/pluralsh/plural/pkg/api"
)

var (
	tfRequirements = map[string]string{
		"aws-bootstrap":   ">= 0.1.53",
		"gcp-bootstrap":   ">= 0.2.24",
		"azure-bootstrap": ">= 0.2.0",
	}
	helmRequirements = map[string]string{
		"bootstrap": ">= 0.8.72",
	}
)

func ValidateMigration(client api.Client) error {
	repo, err := client.GetRepository("bootstrap")
	if err != nil {
		return err
	}

	charts, tfs, err := client.GetPackageInstallations(repo.Id)
	if err != nil {
		return err
	}
	chartsByName, tfsByName := map[string]*api.ChartInstallation{}, map[string]*api.TerraformInstallation{}
	for _, chart := range charts {
		chartsByName[chart.Chart.Name] = chart
	}
	for _, tf := range tfs {
		tfsByName[tf.Terraform.Name] = tf
	}

	for name, req := range tfRequirements {
		if tf, ok := tfsByName[name]; ok {
			if !testSemver(req, tf.Version.Version) {
				return fmt.Errorf("You must have installed the %s terraform module at least at version %s to run cluster migration, your version is %s", name, req, tf.Version.Version)
			}
		}
	}

	for name, req := range helmRequirements {
		if chart, ok := chartsByName[name]; ok {
			if !testSemver(req, chart.Version.Version) {
				return fmt.Errorf("You must have installed the %s helm chart at least at version %s to run cluster migration, your version is %s", name, req, chart.Version.Version)
			}
		}
	}

	return nil
}

func testSemver(constraint, vsn string) bool {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false
	}
	v, err := semver.NewVersion(vsn)
	if err != nil {
		return false
	}

	return c.Check(v)
}
