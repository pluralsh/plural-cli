package up

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
	"github.com/urfave/cli"

	cdpkg "github.com/pluralsh/plural-cli/cmd/command/cd"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/provider/gcp"
	"github.com/pluralsh/plural-cli/pkg/up"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

const (
	defaultBootstrapBranch = "main"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:  "up",
		Usage: "sets up your repository and an initial management cluster",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "endpoint",
				Usage: "the endpoint for the plural installation you're working with",
			},
			cli.StringFlag{
				Name:  "service-account",
				Usage: "email for the service account you'd like to use for this workspace",
			},
			cli.BoolFlag{
				Name:  "ignore-preflights",
				Usage: "whether to ignore preflight check failures prior to init",
			},
			cli.BoolFlag{
				Name:  "dry-run",
				Usage: "whether to simply generate the up repo, but not deploy anything",
			},
			cli.BoolFlag{
				Name:  "cloud",
				Usage: "Whether you're provisioning against a cloud-hosted Plural Console",
			},
			cli.StringFlag{
				Name:  "commit",
				Usage: "commits your changes with this message",
			},
			cli.StringFlag{
				Name:  "git-ref",
				Usage: "branch or tag name to use for cloning the bootstrap repository",
				Value: defaultBootstrapBranch,
			},
		},
		Action: common.LatestVersion(p.handleUp),
	}
}

func (p *Plural) handleUp(c *cli.Context) error {
	// provider.IgnoreProviders([]string{"GENERIC", "KIND"})
	if err := common.HandleLogin(c); err != nil {
		return err
	}
	p.InitPluralClient()
	dryRun := c.Bool("dry-run")

	cd := &cdpkg.Plural{Plural: p.Plural}

	var name, url string
	var err error

	if c.Bool("cloud") {
		name, url, err = p.choseCluster()
		if err != nil {
			return err
		}

		cdpkg.SetConsoleURL(url)
		provider.SetClusterFlag(name)
		if err := cd.HandleCdLogin(c); err != nil {
			return err
		}

		if err := p.backfillEncryption(); err != nil {
			return err
		}
	}

	if err := p.HandleInit(c); err != nil {
		return err
	}

	if err := askAppDomain(); err != nil {
		return err
	}

	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	ctx, err := up.Build(c.Bool("cloud"))
	if err != nil {
		return err
	}

	byok := ctx.Provider.Name() == api.BYOK

	if c.Bool("cloud") {
		id, err := getCluster(cd)
		if err != nil {
			return err
		}

		ctx.ImportCluster = lo.ToPtr(id)
		ctx.CloudCluster = name
		if byok {
			if err := p.InitConsoleClient("", ""); err != nil {
				return err
			}
			if err := p.ReinstallOperator(c, lo.ToPtr(id), nil, ""); err != nil {
				return err
			}
		}
	}

	if err := ctx.Backfill(); err != nil {
		return err
	}

	gitRef := lo.Ternary(len(c.String("git-ref")) > 0, c.String("git-ref"), defaultBootstrapBranch)
	dir, err := ctx.Generate(gitRef)
	defer func() { os.RemoveAll(dir) }()
	if err != nil {
		return err
	}

	if dryRun {
		utils.Success("Finished generating the repo, no deployment will occur due to the --dry-run flag\n")
		return nil
	}

	if !byok {
		if !common.Affirm(common.AffirmUp, "PLURAL_UP_AFFIRM_DEPLOY") {
			return fmt.Errorf("cancelled deploy")
		}
	}

	if err := ctx.Deploy(func() error {
		utils.Highlight("\n==> Enter a commit message to push your configuration\n\n")
		if commit := common.CommitMsg(c); commit != "" {
			utils.Highlight("Pushing upstream...\n")
			return git.Sync(repoRoot, commit, c.Bool("force"))
		}
		return nil
	}); err != nil {
		return err
	}

	utils.Success("Finished setting up your management cluster!\n")
	if byok {
		utils.Highlight("Since you're using BYOK, be sure to complete setup of your management cluster\n")
		utils.Highlight("IMPORTANT: You'll need to configure IAM permissions for the plrl-deploy-operator/stacks service account.\n")
		utils.Highlight("This is no longer handled automatically. See the terraform example in the docs for the required IAM policy.\n")
		return nil
	}
	utils.Highlight("Feel free to use terraform as you normally would, and leverage the gitops setup we've generated in the bootstrap/ subfolder\n")
	return nil
}

func (p *Plural) choseCluster() (name, url string, err error) {
	prior := console.ReadConfig()
	instances, err := p.GetConsoleInstances()
	if err != nil {
		return
	}

	clusterNames := []string{}
	clusterMap := map[string]string{}

	for _, cluster := range instances {
		if prior.Url != "" && strings.EqualFold(common.GetHostnameFromURL(prior.Url), common.GetHostnameFromURL(cluster.URL)) {
			name = cluster.Name
			url = cluster.URL
			return
		}
		clusterNames = append(clusterNames, cluster.Name)
		clusterMap[cluster.Name] = cluster.URL
	}

	prompt := &survey.Select{
		Message: "Select one of the following clusters:",
		Options: clusterNames,
	}
	if err = survey.AskOne(prompt, &name, survey.WithValidator(survey.Required)); err != nil {
		return
	}
	url = clusterMap[name]
	return
}

func askAppDomain() error {
	skip, ok := utils.GetEnvBoolValue("PLURAL_UP_SKIP_APP_DOMAIN")
	if ok && skip {
		return nil
	}

	project, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	var domain string
	message := "Enter the domain for your application. It's expected that the root domain already exist in your clouds DNS provider. Leave empty to ignore:"
	if project.Provider == api.ProviderGCP {
		message = "Enter the DNS zone name for your application. This should be the DNS zone name already configured in your cloud's DNS provider. Leave empty to ignore:"
	}
	prompt := &survey.Input{
		Message: message,
	}
	if err := survey.AskOne(prompt, &domain); err != nil {
		return err
	}

	return processAppDomain(domain, project)
}

func processAppDomain(domain string, project *manifest.ProjectManifest) error {
	if lo.IsEmpty(domain) {
		// No domain was provided, domain checks and setup can be skipped.
		return nil
	}

	switch project.Provider {
	case api.ProviderAWS:
		// For AWS, we need to validate that the domain is set up in Route 53.
		if err := provider.ValidateAWSDomainRegistration(context.Background(), domain, project.Region); err != nil {
			return err
		}
	case api.ProviderAzure:
		// For Azure, we need to validate that the domain is set up in Azure DNS.
		if err := provider.ValidateAzureDomainRegistration(context.Background(), domain, project.Project); err != nil {
			return err
		}
	case api.ProviderGCP:
		// For GCP, besides just validating that the domain is set up,
		// we also need to determine the managed DNS zone to use.
		// If there is one it will be automatically selected, if there are multiple,
		// the user will be prompted to select one.

		d := strings.TrimSuffix(domain, ".") + "." // GCP stores zone names with a trailing dot.

		managedZones, err := gcp.ManagedZones(project.Project)
		if err != nil {
			return err
		}

		if len(managedZones) == 0 {
			return fmt.Errorf("no DNS managed zones found for domain %s in project %s", d, project.Project)
		}

		filteredZones := algorithms.Filter(managedZones, func(dnsName string) bool {
			return dnsName == d
		})

		candidateZones := managedZones
		if len(filteredZones) > 0 {
			candidateZones = filteredZones
		}

		var managedZone string
		if len(candidateZones) == 1 {
			managedZone = candidateZones[0]
		} else {
			if err := survey.AskOne(&survey.Select{Message: "Select managed DNS zone:", Options: candidateZones},
				&managedZone, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
		}

		if project.Context == nil {
			project.Context = map[string]interface{}{}
		}
		project.Context["ManagedZone"] = managedZone
	}

	// Save the domain and other changes to the project manifest.
	project.AppDomain = domain
	return project.Flush()
}

func getCluster(cd *cdpkg.Plural) (id string, err error) {
	if cd == nil {
		err = fmt.Errorf("your CLI is not logged into Plural, try running `plural login` to generate local credentials")
		return
	}

	clusters, err := cd.ListClusters()
	if err != nil {
		return
	}

	for _, cluster := range clusters {
		if lo.FromPtr(cluster.Node.Handle) == "mgmt" {
			return cluster.Node.ID, nil
		}
	}

	err = fmt.Errorf("could not find the management cluster in your Plural cloud instance, contact support for assistance")
	return
}
