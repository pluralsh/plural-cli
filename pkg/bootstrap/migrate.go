package bootstrap

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pluralsh/cluster-api-migration/pkg/api"
	"github.com/pluralsh/cluster-api-migration/pkg/migrator"
	"sigs.k8s.io/yaml"

	delinkeranalyze "github.com/pluralsh/terraform-delinker/api/analyze/v1alpha1"
	delinkerdelink "github.com/pluralsh/terraform-delinker/api/delink/v1alpha1"
	delinkerexec "github.com/pluralsh/terraform-delinker/api/exec/v1alpha1"
	delinkerplan "github.com/pluralsh/terraform-delinker/api/plan/v1alpha1"

	api2 "github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func newConfiguration(cliProvider provider.Provider, clusterProvider api.ClusterProvider) (*api.Configuration, error) {
	switch clusterProvider {
	case api.ClusterProviderGoogle:
		kubeconfigPath, err := getKubeconfigPath()
		if err != nil {
			log.Fatalln(err)
		}

		context := cliProvider.Context()
		credentials, err := base64.StdEncoding.DecodeString(utils.ToString(context["Credentials"]))
		if err != nil {
			log.Fatalln(err)
		}

		return &api.Configuration{
			GCPConfiguration: &api.GCPConfiguration{
				Credentials:    string(credentials),
				Project:        cliProvider.Project(),
				Region:         cliProvider.Region(),
				Name:           cliProvider.Cluster(),
				KubeconfigPath: kubeconfigPath,
			},
		}, nil
	case api.ClusterProviderAzure:
		context := cliProvider.Context()

		config := api.Configuration{
			AzureConfiguration: &api.AzureConfiguration{
				SubscriptionID: utils.ToString(context["SubscriptionId"]),
				ResourceGroup:  cliProvider.Project(),
				Name:           cliProvider.Cluster(),
			},
		}

		if err := config.Validate(); err != nil {
			log.Fatalln(err)
		}

		return &config, nil
	case api.ClusterProviderAWS:
		err := os.Setenv("AWS_REGION", cliProvider.Region())
		if err != nil {
			return nil, err
		}

		config := &api.Configuration{
			AWSConfiguration: &api.AWSConfiguration{
				ClusterName: cliProvider.Cluster(),
				Region:      cliProvider.Region(),
			},
		}
		return config, nil
	}

	return nil, fmt.Errorf("unknown provider, no configuration found")
}

// getMigrator returns configured migrator for current provider.
func getMigrator() (api.Migrator, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	clusterProvider := api.ClusterProvider(prov.Name())

	configuration, err := newConfiguration(prov, clusterProvider)
	if err != nil {
		return nil, err
	}

	return migrator.NewMigrator(clusterProvider, configuration)
}

// generateValuesFile generates values.yaml file based on current cluster configuration that will be used by Cluster API.
func generateValuesFile() error {
	utils.Highlight("Generating values.yaml file based on current cluster configuration...\n")

	m, err := getMigrator()
	if err != nil {
		return err
	}

	values, err := m.Convert()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(Bootstrap{ClusterAPICluster: values})
	if err != nil {
		return err
	}

	gitRootDir, err := git.Root()
	if err != nil {
		return err
	}

	bootstrapRepo := filepath.Join(gitRootDir, "bootstrap")
	valuesFile := pathing.SanitizeFilepath(filepath.Join(bootstrapRepo, "helm", "bootstrap", "values.yaml"))
	if utils.Exists(valuesFile) {
		if err := os.WriteFile(valuesFile, data, 0644); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("can't save %s file", valuesFile)
	}

	utils.Success("values.yaml saved successfully!\n")

	return nil
}

// getProviderTags returns list of tags to set on provider resources during migration.
func getProviderTags(provider, cluster string) []string {
	switch provider {
	case "aws":
		return []string{
			fmt.Sprintf("kubernetes.io/cluster/%s=owned", cluster),
			fmt.Sprintf("sigs.k8s.io/cluster-api-provider-aws/cluster/%s=owned", cluster),
		}
	case "azure":
		return []string{
			fmt.Sprintf("sigs.k8s.io_cluster-api-provider-azure_cluster_%s=owned", cluster),
			"sigs.k8s.io_cluster-api-provider-azure_role=common",
		}
	default:
		return []string{}
	}
}

// tagResources adds Cluster API tags on provider resources.
func tagResources(arguments []string) error {
	m, err := getMigrator()
	if err != nil {
		return err
	}

	tags := map[string]string{}
	for _, arg := range arguments {
		split := strings.Split(arg, "=")
		if len(split) == 2 {
			tags[split[0]] = split[1]
		}
	}

	return m.AddTags(tags)
}

// delinkTerraformState delinks resources managed by Cluster API from Terraform state.
func delinkTerraformState(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("path argument is missing")
	}

	path := args[0]
	planner := delinkerplan.NewPlanner(delinkerplan.WithTerraform(delinkerexec.WithDir(path)))

	plan, err := planner.Plan()
	if err != nil {
		return err
	}

	report := delinkeranalyze.NewAnalyzer(plan).Analyze(tfjson.ActionDelete)
	delinker := delinkerdelink.NewDelinker(delinkerdelink.WithTerraform(delinkerexec.WithDir(path)))
	return delinker.Run(report)
}

// getMigrationFlags returns list of provider-specific flags used during cluster migration.
func getMigrationFlags(provider string) []string {
	switch provider {
	case "aws":
		return []string{
			"--set", "cluster-api-provider-aws.cluster-api-provider-aws.bootstrapMode=false",
		}
	default:
		return []string{}
	}
}

// getMigrationSteps returns list of steps to run during cluster migration.
func getMigrationSteps(runPlural ActionFunc) ([]*Step, error) {
	projectManifest, err := manifest.FetchProject()
	if err != nil {
		return nil, err
	}

	gitRootDir, err := git.Root()
	if err != nil {
		return nil, err
	}

	bootstrapPath := pathing.SanitizeFilepath(filepath.Join(gitRootDir, "bootstrap"))
	terraformPath := filepath.Join(bootstrapPath, "terraform")
	tags := getProviderTags(projectManifest.Provider, projectManifest.Cluster)
	flags := getMigrationFlags(projectManifest.Provider)

	var steps []*Step

	if projectManifest.Provider == "azure" {
		// Setting PLURAL_PACKAGES_UNINSTALL variable to avoid confirmation prompt on package uninstall.
		err := os.Setenv("PLURAL_PACKAGES_UNINSTALL", "true")
		if err != nil {
			return nil, err
		}

		steps = append(steps, []*Step{
			{
				Name:       "Uninstall azure-identity package",
				Args:       append([]string{"plural", "packages", "uninstall", "helm", "bootstrap", "azure-identity"}),
				TargetPath: gitRootDir,
				Execute:    runPlural,
			},
			{
				Name:       "Clear package cache",
				TargetPath: gitRootDir,
				Execute: func(_ []string) error {
					api2.ClearPackageCache()

					return nil
				},
			},
		}...)
	}

	return append(steps, []*Step{
		{
			Name:       "Build values",
			Args:       []string{"plural", "build", "--only", "bootstrap", "--force"},
			TargetPath: gitRootDir,
			Execute:    runPlural,
		},
		{
			Name:    "Bootstrap CRDs",
			Args:    []string{"plural", "wkspace", "crds", bootstrapPath},
			Execute: runPlural,
		},
		{
			Name:    "Install Cluster API operators",
			Args:    append([]string{"plural", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, flags...),
			Execute: runPlural,
		},
		{
			Name:    "Add Cluster API tags for provider resources",
			Args:    tags,
			Execute: tagResources,
		},
		{
			Name:    "Deploy cluster",
			Args:    append([]string{"plural", "wkspace", "helm", "bootstrap"}, flags...),
			Execute: runPlural,
		},
		{
			Name:    "Wait for cluster",
			Args:    []string{"plural", "clusters", "wait", "bootstrap", projectManifest.Cluster},
			Execute: runPlural,
		},
		{
			Name:    "Wait for machine pools",
			Args:    []string{"plural", "clusters", "mpwait", "bootstrap", projectManifest.Cluster},
			Execute: runPlural,
		},
		{
			Name:       "Mark cluster as migrated to Cluster API",
			TargetPath: gitRootDir,
			Execute: func(_ []string) error {
				path := manifest.ProjectManifestPath()
				project, err := manifest.ReadProject(path)
				if err != nil {
					return err
				}

				project.ClusterAPI = true
				return project.Write(path)
			},
		},
		{
			Name:       "Build values",
			Args:       []string{"plural", "build", "--only", "bootstrap", "--force"},
			TargetPath: gitRootDir,
			Execute:    runPlural,
		},
		{
			Name:    "Delink resources managed by Cluster API from Terraform state",
			Args:    []string{terraformPath},
			Execute: delinkTerraformState,
		},
		{
			Name:       "Run Terraform init",
			Args:       []string{"init", "-upgrade"},
			TargetPath: terraformPath,
			Execute:    runTerraform,
		},
		{
			Name:       "Run Terraform apply",
			Args:       []string{"apply", "-auto-approve"},
			TargetPath: terraformPath,
			Execute:    runTerraform,
		},
	}...), nil
}

// MigrateCluster migrates existing clusters to Cluster API.
func MigrateCluster(runPlural ActionFunc) error {
	utils.Highlight("Migrating cluster to Cluster API...\n")

	err := generateValuesFile()
	if err != nil {
		return err
	}

	steps, err := getMigrationSteps(runPlural)
	if err != nil {
		return err
	}

	err = executeSteps(steps)
	if err != nil {
		return err
	}

	utils.Success("Cluster migrated successfully!\n")
	return nil
}
