package plural

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pluralsh/cluster-api-migration/pkg/migrator"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"sigs.k8s.io/yaml"

	"github.com/pluralsh/cluster-api-migration/pkg/api"
)

func newConfiguration(cliProvider provider.Provider) (api.ClusterProvider, *api.Configuration) {
	context := cliProvider.Context()
	clusterProvider := api.ClusterProvider(cliProvider.Name())
	switch clusterProvider {
	case api.ClusterProviderGoogle:
		kubeconfigPath := os.Getenv("KUBECONFIG")
		credentials, err := base64.StdEncoding.DecodeString(os.Getenv("GCP_B64ENCODED_CREDENTIALS"))
		if err != nil {
			panic(err)
		}

		return clusterProvider, &api.Configuration{
			GCPConfiguration: &api.GCPConfiguration{
				Credentials:    string(credentials),
				Project:        cliProvider.Project(),
				Region:         cliProvider.Region(),
				Name:           cliProvider.Cluster(),
				KubeconfigPath: kubeconfigPath,
			},
		}
	case api.ClusterProviderAzure:
		config := api.Configuration{
			AzureConfiguration: &api.AzureConfiguration{
				SubscriptionID: utils.ToString(context["SubscriptionId"]),
				ResourceGroup:  cliProvider.Project(),
				Name:           cliProvider.Cluster(),
				SSHPublicKey:   "c3NoLXJzYSBBQUFBQjNOemFDMXljMkVBQUFBREFRQUJBQUFCQVFEVTRhRmJ3anNSSG5nU1Jsdmh2TUVoeW5DN29zTDBTRWEvQ0c3ZWtFOXlFS2Npa0pTanVNcGZrTm80NnFPRld3ZWFhV1QyRFlDdUpEQXRuZ3dyT1A3dmNKbEFnbmcvQW9qMDJ1VHJRZVoySW5qekhnQlFvWmgrcE9kYm8wdDVwTXVNckxmUFdlY2M3aGN6TlVDSDRxYjNnNEl1VG9SbjhkVUFjb1UxZFNkVWpTTDk2U3BmNTdKVHBaTTVwRGFPT1ZXL0llUHdDSVlJODdqdGZUQ3ZGSi9JR2tNK0hPbGlzcE1FQ05UY25saEE2a0QvRHdvVTR0eVVSKzFlS3ExU1hpb3ZZcHcrYkFlaTFReGptK1VMUzNXSzdubnZCbE5seHB3MjRPUzBleGphRGJ4TzhjYkRSYXhzVWMwZW9JbUROU1pWUjhMbWc3UzJnRCsvOGErbTA3VnI=", // Mocked as it is not used but required in forked version of CAPZ.
			},
		}

		if err := config.Validate(); err != nil {
			log.Fatalln(err)
		}

		return clusterProvider, &config
	case api.ClusterProviderAWS:
		os.Setenv("AWS_REGION", cliProvider.Region())
		config := &api.Configuration{
			AWSConfiguration: &api.AWSConfiguration{
				ClusterName: cliProvider.Cluster(),
				Region:      cliProvider.Region(),
			},
		}
		return clusterProvider, config
	}

	return "", nil
}

type Bootstrap struct {
	ClusterAPICluster *api.Values `json:"cluster-api-cluster"`
}

func ExecuteMigration() error {
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
	root, err := git.Root()
	if err != nil {
		return err
	}
	bootstrapRepo := filepath.Join(root, "bootstrap")
	bootstrapRepoPath := pathing.SanitizeFilepath(bootstrapRepo)
	valuesFile := pathing.SanitizeFilepath(filepath.Join(bootstrapRepo, "helm", "bootstrap", "values.yaml"))
	if utils.Exists(valuesFile) {
		if err := os.WriteFile(valuesFile, data, 0644); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("can't save %s file", valuesFile)
	}

	for _, step := range clusterAPIMigrateSteps(bootstrapRepoPath) {
		utils.Highlight("%s \n", step.Name)
		err := os.Chdir(step.TargetPath)
		if err != nil {
			return err
		}
		err = step.Execute(step.Args)
		if err != nil {
			return err
		}
	}

	return nil
}

func clusterAPIMigrateSteps(path string) []*Step {
	pm, _ := manifest.FetchProject()

	sanitizedPath := pathing.SanitizeFilepath(path)
	providerBootstrapFlags := []string{}
	providerTags := []string{}

	root, _ := git.Root()
	switch pm.Provider {
	case "aws":
		providerBootstrapFlags = []string{
			"--set", "cluster-api-provider-aws.cluster-api-provider-aws.bootstrapMode=false",
		}
		providerTags = []string{
			fmt.Sprintf("kubernetes.io/cluster/%s=owned", pm.Cluster),
		}
	case "azure":
		providerTags = []string{
			fmt.Sprintf("sigs.k8s.io_cluster-api-provider-azure_cluster_%s=owned", pm.Cluster),
			"sigs.k8s.io_cluster-api-provider-azure_role=common",
		}
	}

	steps := []*Step{
		{
			Name:       "build values",
			Args:       []string{"plural", "build", "--only", "bootstrap", "--force"},
			TargetPath: root,
			Execute:    RunPlural,
		},
		{
			Name:       "terraform init",
			Args:       []string{"init", "-upgrade"},
			TargetPath: filepath.Join(path, "terraform"),
			Execute:    RunTerraform,
		},
		{
			Name:       "terraform apply",
			Args:       []string{"apply", "-auto-approve"},
			TargetPath: filepath.Join(path, "terraform"),
			Execute:    RunTerraform,
		},
		{
			Name:       "bootstrap crds",
			Args:       []string{"plural", "wkspace", "crds", sanitizedPath},
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "install capi operators",
			Args:       append([]string{"plural", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "add tags",
			Args:       providerTags,
			TargetPath: sanitizedPath,
			Execute:    RunAddTags,
		},
	}

	if pm.Provider == "azure" {
		os.Setenv("PLURAL_PACKAGES_UNINSTALL", "true")
		steps = append(steps, []*Step{{
			Name:       "uninstall azure-identity",
			Args:       append([]string{"plural", "packages", "uninstall", "helm", "bootstrap", "azure-identity"}),
			TargetPath: root,
			Execute:    RunPlural,
		}}...)
	}

	return append(steps, []*Step{
		{
			Name:       "deploy cluster",
			Args:       append([]string{"plural", "wkspace", "helm", "bootstrap"}, providerBootstrapFlags...),
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "wait-for-cluster",
			Args:       []string{"plural", "clusters", "wait", "bootstrap", pm.Cluster},
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "wait-for-machines-running",
			Args:       []string{"plural", "clusters", "mpwait", "bootstrap", pm.Cluster},
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
	}...)
}

func getMigrator() (api.Migrator, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}
	return migrator.NewMigrator(newConfiguration(prov))
}

func RunAddTags(arguments []string) error {
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

func RunPlural(arguments []string) error {
	return CreateNewApp(&Plural{}).Run(arguments)
}

func RunTerraform(arguments []string) error {
	return execCommand("terraform", arguments...)
}

func execCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
