package provider

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
	plrlErrors "github.com/pluralsh/plural-cli/pkg/utils/errors"

	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	provUtils "github.com/pluralsh/plural-cli/pkg/provider/utils"
)

type AWSProvider struct {
	Clus          string `survey:"cluster"`
	project       string
	bucket        string
	Reg           string `survey:"region"`
	storageClient *s3.Client
	writer        manifest.Writer
	goContext     *context.Context
	ctx           map[string]interface{}
}

var (
	awsRegions = []string{
		"af-south-1",
		"eu-north-1",
		"ap-south-1",
		"eu-west-3",
		"eu-west-2",
		"eu-south-1",
		"eu-west-1",
		"ap-northeast-3",
		"ap-northeast-2",
		"me-south-1",
		"ap-northeast-1",
		"sa-east-1",
		"ca-central-1",
		"ap-east-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"eu-central-1",
		"ap-southeast-3",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}
)

func mkAWS(conf config.Config) (provider *AWSProvider, err error) {
	ctx := context.Background()

	iamSession, err := GetAWSCallerIdentity(ctx)
	if err != nil {
		return nil, plrlErrors.ErrorWrap(err, "Failed to get AWS caller identity")
	}

	provider = &AWSProvider{
		goContext: &ctx,
		ctx: map[string]any{
			"IAMSession": iamSession,
		},
	}

	var awsSurvey = []*survey.Question{
		{
			Name:     "cluster",
			Prompt:   &survey.Input{Message: "Enter the name of your cluster:", Default: clusterFlag},
			Validate: validCluster,
		},
		{
			Name:     "region",
			Prompt:   &survey.Select{Message: "What region will you deploy to?", Default: "us-east-2", Options: awsRegions},
			Validate: survey.Required,
		},
	}

	if err = survey.Ask(awsSurvey, provider); err != nil {
		return
	}

	client, err := getClient(provider.Reg, *provider.goContext)
	if err != nil {
		return
	}

	account, err := GetAwsAccount(ctx)
	if err != nil {
		err = plrlErrors.ErrorWrap(err, "Failed to get aws account (is your aws cli configured?)")
		return
	}

	if len(account) == 0 {
		err = plrlErrors.ErrorWrap(fmt.Errorf("unable to find AWS account ID, make sure that your AWS CLI is configured"), "AWS cli error:")
		return
	}

	provider.project = account
	provider.storageClient = client

	azones, err := getAvailabilityZones(ctx, provider.Region())
	if err != nil {
		return
	}

	projectManifest := manifest.ProjectManifest{
		Cluster:           provider.Cluster(),
		Project:           provider.Project(),
		Provider:          api.ProviderAWS,
		Region:            provider.Region(),
		Context:           provider.Context(),
		AvailabilityZones: azones,
		Owner:             &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}

	provider.writer = projectManifest.Configure(cloudFlag, provider.Cluster())
	provider.bucket = projectManifest.Bucket
	return
}

func awsFromManifest(man *manifest.ProjectManifest) (*AWSProvider, error) {
	ctx := context.Background()
	client, err := getClient(man.Region, ctx)
	if err != nil {
		return nil, err
	}
	providerCtx := map[string]interface{}{}

	return &AWSProvider{Clus: man.Cluster, project: man.Project, bucket: man.Bucket, Reg: man.Region, storageClient: client, goContext: &ctx, ctx: providerCtx}, nil
}

func getClient(region string, context context.Context) (*s3.Client, error) {
	cfg, err := getAwsConfig(context)
	if err != nil {
		return nil, err
	}

	cfg.Region = region
	return s3.NewFromConfig(cfg), nil
}

func getEC2Client(ctx context.Context, region string) (*ec2.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Region = region
	return ec2.NewFromConfig(cfg), nil
}

func getAvailabilityZones(ctx context.Context, region string) ([]string, error) {
	return fetchAZ(ctx, region, true)
}

func fetchAZ(context context.Context, region string, sorted bool) ([]string, error) {
	ec2Client, err := getEC2Client(context, region)
	if err != nil {
		return nil, err
	}
	allAvailabilityZones := true
	dryRun := false
	regionName := "region-name"
	azones, err := ec2Client.DescribeAvailabilityZones(context, &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: &allAvailabilityZones,
		DryRun:               &dryRun,
		Filters: []ec2Types.Filter{
			{
				Name:   &regionName,
				Values: []string{region},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, az := range azones.AvailabilityZones {
		if az.ParentZoneId == nil {
			result = append(result, *az.ZoneName)
		}
	}
	// append when there are fewer zones than 3
	for i := 0; (3 - len(result)) > 0; i++ {
		result = append(result, result[i])
	}
	if sorted {
		sort.Strings(result)
	}
	return result, nil
}

func getAwsConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	return cfg, plrlErrors.ErrorWrap(err, "Failed to initialize aws client: ")
}

func (aws *AWSProvider) CreateBucket() error {
	return aws.mkBucket(aws.bucket)
}

func (aws *AWSProvider) KubeConfig() error {
	if kubernetes.InKubernetes() {
		return nil
	}

	cmd := exec.Command(
		"aws", "eks", "update-kubeconfig", "--name", aws.Cluster(), "--region", aws.Region())
	return utils.Execute(cmd)
}

func (aws *AWSProvider) KubeContext() string {
	return fmt.Sprintf("arn:aws:eks:%s:%s:cluster/%s", aws.Region(), aws.project, aws.Cluster())
}

func (p *AWSProvider) mkBucket(name string) error {
	client := p.storageClient
	_, err := client.HeadBucket(*p.goContext, &s3.HeadBucketInput{Bucket: &name})

	if err != nil {
		bucket := &s3.CreateBucketInput{
			Bucket: &name,
		}

		if p.Region() != "us-east-1" {
			bucket.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
				LocationConstraint: s3Types.BucketLocationConstraint(p.Region()),
			}
		}

		_, err = client.CreateBucket(*p.goContext, bucket)
		return err
	}

	return nil
}

func (aws *AWSProvider) Name() string {
	return api.ProviderAWS
}

func (aws *AWSProvider) Cluster() string {
	return aws.Clus
}

func (aws *AWSProvider) Project() string {
	return aws.project
}

func (aws *AWSProvider) Bucket() string {
	return aws.bucket
}

func (aws *AWSProvider) Region() string {
	return aws.Reg
}

func (aws *AWSProvider) Context() map[string]interface{} {
	return aws.ctx
}

func (aws *AWSProvider) Preflights() []*Preflight {
	return []*Preflight{
		{Name: "Test IAM Permissions", Callback: aws.testIamPermissions},
	}
}

func (aws *AWSProvider) Flush() error {
	if aws.writer == nil {
		return nil
	}
	return aws.writer()
}

func (prov *AWSProvider) Permissions() (permissions.Checker, error) {
	return permissions.NewAwsChecker(*prov.goContext)
}

func (prov *AWSProvider) Decommision(node *v1.Node) error {
	cfg, err := awsConfig.LoadDefaultConfig(*prov.goContext)

	if err != nil {
		return plrlErrors.ErrorWrap(err, "Failed to establish aws session")
	}

	cfg.Region = prov.Region()

	name := "private-dns-name"

	svc := ec2.NewFromConfig(cfg)
	instances, err := svc.DescribeInstances(*prov.goContext, &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{Name: &name, Values: []string{node.Name}},
		},
	})

	if err != nil {
		return plrlErrors.ErrorWrap(err, "failed to find node in ec2")
	}

	instance := instances.Reservations[0].Instances[0]

	_, err = svc.TerminateInstances(*prov.goContext, &ec2.TerminateInstancesInput{
		InstanceIds: []string{*instance.InstanceId},
	})

	return plrlErrors.ErrorWrap(err, "failed to terminate instance")
}

func GetAwsAccount(ctx context.Context) (string, error) {
	cfg, err := getAwsConfig(ctx)
	if err != nil {
		return "", err
	}
	svc := sts.NewFromConfig(cfg)
	result, err := svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", plrlErrors.ErrorWrap(err, "Error finding iam identity: ")
	}

	return *result.Account, nil
}

func (aws *AWSProvider) testIamPermissions() error {
	checker, err := aws.Permissions()
	if err != nil {
		return err
	}

	missing, err := checker.MissingPermissions()
	if err != nil {
		return err
	}

	if len(missing) == 0 {
		return nil
	}

	for _, missed := range missing {
		provUtils.FailedPermission(missed)
	}

	return fmt.Errorf("you do not meet all required iam permissions to deploy an eks cluster: %s, this is not necessarily a full list, we recommend using as close to AdministratorAccess as possible to run plural", strings.Join(missing, ","))
}

// GetAWSCallerIdentity returns the IAM role ARN of the current caller identity.
func GetAWSCallerIdentity(ctx context.Context) (string, error) {
	cfg, err := getAwsConfig(ctx)
	if err != nil {
		return "", err
	}

	svc := sts.NewFromConfig(cfg)
	callerIdentity, err := svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", plrlErrors.ErrorWrap(err, "Error getting caller identity: ")
	}

	callerIdentityArn := lo.FromPtr(callerIdentity.Arn)
	parsedArn, err := arn.Parse(callerIdentityArn)
	if err != nil {
		return "", plrlErrors.ErrorWrap(err, "Error parsing caller identity ARN: ")
	}

	matcher := regexp.MustCompile(`^assumed-role/.+/.{2,}`)
	if strings.EqualFold(parsedArn.Service, sts.ServiceID) && matcher.MatchString(parsedArn.Resource) {
		split := strings.Split(parsedArn.Resource, "/")
		if len(split) < 3 {
			return "", fmt.Errorf("invalid assumed role ARN format: %s", parsedArn.Resource)
		}

		roleName := split[len(split)-2]
		role, err := iam.NewFromConfig(cfg).GetRole(ctx, &iam.GetRoleInput{RoleName: &roleName})
		if err != nil {
			return "", plrlErrors.ErrorWrap(err, "Error getting IAM role: ")
		}

		return lo.FromPtr(role.Role.Arn), nil
	}

	return callerIdentityArn, nil
}
