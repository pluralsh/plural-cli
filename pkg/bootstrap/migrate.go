package bootstrap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	migratorapi "github.com/pluralsh/cluster-api-migration/pkg/api"
	"github.com/pluralsh/cluster-api-migration/pkg/migrator"
	"github.com/pluralsh/polly/containers"
	delinkeranalyze "github.com/pluralsh/terraform-delinker/api/analyze/v1alpha1"
	delinkerdelink "github.com/pluralsh/terraform-delinker/api/delink/v1alpha1"
	delinkerexec "github.com/pluralsh/terraform-delinker/api/exec/v1alpha1"
	delinkerplan "github.com/pluralsh/terraform-delinker/api/plan/v1alpha1"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"

	"github.com/pluralsh/plural/pkg/api"
	bootstrapaws "github.com/pluralsh/plural/pkg/bootstrap/aws"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func newConfiguration(cliProvider provider.Provider, clusterProvider migratorapi.ClusterProvider) (*migratorapi.Configuration, error) {
	switch clusterProvider {
	case migratorapi.ClusterProviderGCP:
		kubeconfigPath, err := getKubeconfigPath()
		if err != nil {
			log.Fatalln(err)
		}

		return &migratorapi.Configuration{
			GCPConfiguration: &migratorapi.GCPConfiguration{
				Project:        cliProvider.Project(),
				Region:         cliProvider.Region(),
				Name:           cliProvider.Cluster(),
				KubeconfigPath: kubeconfigPath,
			},
		}, nil
	case migratorapi.ClusterProviderAzure:
		context := cliProvider.Context()
		config := migratorapi.Configuration{
			AzureConfiguration: &migratorapi.AzureConfiguration{
				SubscriptionID: utils.ToString(context["SubscriptionId"]),
				ResourceGroup:  cliProvider.Project(),
				Name:           cliProvider.Cluster(),
			},
		}

		if err := config.Validate(); err != nil {
			log.Fatalln(err)
		}

		return &config, nil
	case migratorapi.ClusterProviderAWS:
		err := os.Setenv("AWS_REGION", cliProvider.Region())
		if err != nil {
			return nil, err
		}

		config := &migratorapi.Configuration{
			AWSConfiguration: &migratorapi.AWSConfiguration{
				ClusterName: cliProvider.Cluster(),
				Region:      cliProvider.Region(),
			},
		}
		return config, nil
	case migratorapi.ClusterProviderKind:
		return &migratorapi.Configuration{
			KindConfiguration: &migratorapi.KindConfiguration{
				ClusterName: cliProvider.Cluster(),
			},
		}, nil

	}

	return nil, fmt.Errorf("unknown provider, no configuration found")
}

// getMigrator returns configured migrator for current provider.
func getMigrator() (migratorapi.Migrator, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	clusterProvider := migratorapi.ClusterProvider(prov.Name())

	configuration, err := newConfiguration(prov, clusterProvider)
	if err != nil {
		return nil, err
	}

	return migrator.NewMigrator(clusterProvider, configuration)
}

func isDesiredKubernetesVersion(key string, value, diffValue any) bool {
	if key != "kubernetesVersion" {
		return false
	}

	defaultKubernetesVersion, _ := diffValue.(string)
	currentKubernetesVersion, _ := value.(string)

	defaultKubernetesVersion = strings.TrimPrefix(defaultKubernetesVersion, "v")
	currentKubernetesVersion = strings.TrimPrefix(currentKubernetesVersion, "v")

	return len(defaultKubernetesVersion) > 0 && strings.HasPrefix(currentKubernetesVersion, defaultKubernetesVersion)
}

// generateValuesFile generates values.yaml file based on current cluster configuration that will be used by Cluster API.
func generateValuesFile() error {
	utils.Highlight("Generating values.yaml file based on current cluster configuration...\n")

	gitRootDir, err := git.Root()
	if err != nil {
		return err
	}

	bootstrapHelmDir := pathing.SanitizeFilepath(filepath.Join(gitRootDir, "bootstrap", "helm", "bootstrap"))
	valuesFile := pathing.SanitizeFilepath(filepath.Join(bootstrapHelmDir, "values.yaml"))
	defaultValuesFile := pathing.SanitizeFilepath(filepath.Join(bootstrapHelmDir, "default-values.yaml"))

	m, err := getMigrator()
	if err != nil {
		return err
	}

	migratorValues, err := m.Convert()
	if err != nil {
		return err
	}

	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	if prov.Name() == api.ProviderAWS {
		availabilityZoneSet := containers.NewSet[string]()
		for _, subnet := range migratorValues.Cluster.AWSCloudSpec.NetworkSpec.Subnets {
			availabilityZoneSet.Add(subnet.AvailabilityZone)
		}
		man, err := manifest.FetchProject()
		if err != nil {
			return err
		}
		man.AvailabilityZones = availabilityZoneSet.List()
		if err := man.Flush(); err != nil {
			return err
		}
	}

	chart, err := loader.Load(bootstrapHelmDir)
	if err != nil {
		return err
	}

	defaultValues, err := chartutil.ReadValuesFile(defaultValuesFile)
	if err != nil {
		return err
	}

	// Nullify main values.yaml as we only use it to generate migration values
	chart.Values = nil
	chartValues, err := chartutil.CoalesceValues(chart, defaultValues)
	if err != nil {
		return err
	}

	migrationYamlData, err := yaml.Marshal(Bootstrap{ClusterAPICluster: migratorValues})
	if err != nil {
		return err
	}

	migrationValues, err := chartutil.ReadValues(migrationYamlData)
	if err != nil {
		return err
	}

	values := utils.DiffMap(migrationValues, chartValues, isDesiredKubernetesVersion)

	valuesYamlData, err := yaml.Marshal(values)
	if err != nil {
		return err
	}

	if utils.Exists(valuesFile) {
		if err := os.WriteFile(valuesFile, valuesYamlData, 0644); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("can't save %s file", valuesFile)
	}

	utils.Success("values.yaml saved successfully!\n")

	return nil
}

// GetProviderTags returns list of tags to set on provider resources during migration.
func GetProviderTags(prov, cluster string) []string {
	switch prov {
	case api.ProviderAWS:
		return []string{
			fmt.Sprintf("kubernetes.io/cluster/%s=owned", cluster),
			fmt.Sprintf("sigs.k8s.io/cluster-api-provider-aws/cluster/%s=owned", cluster),
		}
	case api.ProviderAzure:
		return []string{
			fmt.Sprintf("sigs.k8s.io_cluster-api-provider-azure_cluster_%s=owned", cluster),
			"sigs.k8s.io_cluster-api-provider-azure_role=common",
		}
	default:
		return []string{}
	}
}

// GetProviderTagsMap returns map of tags to set on provider resources during migration.
func GetProviderTagsMap(arguments []string) (map[string]string, error) {
	tags := map[string]string{}
	for _, arg := range arguments {
		split := strings.Split(arg, "=")
		if len(split) == 2 {
			tags[split[0]] = split[1]
		} else {
			return nil, fmt.Errorf("invalid tag format")
		}
	}

	return tags, nil
}

// tagResources adds Cluster API tags on provider resources.
func tagResources(arguments []string) error {
	m, err := getMigrator()
	if err != nil {
		return err
	}

	tags, err := GetProviderTagsMap(arguments)
	if err != nil {
		return err
	}

	return m.AddTags(tags)
}

// delinkTerraformState delinks resources managed by Cluster API from Terraform state.
func delinkTerraformState(path string) error {
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
func getMigrationFlags(prov string) []string {
	switch prov {
	case api.ProviderAWS:
		return []string{
			"--set", "cluster-api-provider-aws.cluster-api-provider-aws.bootstrapMode=false",
		}
	case api.ProviderGCP:
		return []string{
			"--set", "cluster-api-provider-gcp.cluster-api-provider-gcp.bootstrapMode=false",
		}
	default:
		return []string{}
	}
}

// getMigrationSteps returns list of steps to run during cluster migration.
func getMigrationSteps(runPlural ActionFunc) ([]*Step, error) {
	man, err := manifest.FetchProject()
	if err != nil {
		return nil, err
	}

	gitRootDir, err := git.Root()
	if err != nil {
		return nil, err
	}

	bootstrapPath := pathing.SanitizeFilepath(filepath.Join(gitRootDir, "bootstrap"))
	terraformPath := filepath.Join(bootstrapPath, "terraform")
	flags := getMigrationFlags(man.Provider)

	if man.Provider == api.ProviderAzure {
		// Setting PLURAL_PACKAGES_UNINSTALL variable to avoid confirmation prompt on package uninstall.
		err := os.Setenv("PLURAL_PACKAGES_UNINSTALL", "true")
		if err != nil {
			return nil, err
		}
	}

	return []*Step{
		{
			Name: "Ensure Cluster API IAM role has access",
			Execute: func(_ []string) error {
				roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s-capa-controller", man.Project, man.Cluster)
				return bootstrapaws.AddRole(roleArn)
			},
			Skip: man.Provider != api.ProviderAWS,
		},
		{
			Name:       "Uninstall azure-identity package",
			Args:       []string{"plural", "packages", "uninstall", "helm", "bootstrap", "azure-identity"},
			TargetPath: gitRootDir,
			Execute:    runPlural,
			Skip:       man.Provider != api.ProviderAzure,
			Retries:    2,
		},
		{
			Name:       "Clear package cache",
			TargetPath: gitRootDir,
			Execute: func(_ []string) error {
				api.ClearPackageCache()

				return nil
			},
			Skip: man.Provider != api.ProviderAzure,
		},
		{
			Name: "Normalize GCP provider value",
			Execute: func(_ []string) error {
				path := manifest.ProjectManifestPath()
				project, err := manifest.ReadProject(path)
				if err != nil {
					return err
				}

				project.Provider = api.ProviderGCP
				return project.Write(path)
			},
			Skip: man.Provider != api.ProviderGCP,
		},
		{
			Name:       "Set Cluster API flag",
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
			Name: "Add Cluster API tags for provider resources",
			Execute: func(_ []string) error {
				return tagResources(GetProviderTags(man.Provider, man.Cluster))
			},
		},
		{
			Name:    "Deploy cluster",
			Args:    append([]string{"plural", "wkspace", "helm", "bootstrap"}, flags...),
			Execute: runPlural,
		},
		{
			Name:    "Wait for cluster",
			Args:    []string{"plural", "clusters", "wait", "bootstrap", man.Cluster},
			Execute: runPlural,
		},
		{
			Name:    "Wait for machine pools",
			Args:    []string{"plural", "clusters", "mpwait", "bootstrap", man.Cluster},
			Execute: runPlural,
		},
		{
			Name: "Delink resources managed by Cluster API from Terraform state",
			Execute: func(_ []string) error {
				return delinkTerraformState(terraformPath)
			},
			Retries: 2,
		},
		{
			Name:       "Run deploy",
			Args:       []string{"plural", "deploy", "--from", "bootstrap", "--silence", "--commit", "migrate to cluster api"},
			TargetPath: gitRootDir,
			Execute:    runPlural,
		},
	}, nil
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

	err = ExecuteSteps(steps)
	if err != nil {
		return err
	}

	utils.Success("Cluster migrated successfully!\n")
	return nil
}
