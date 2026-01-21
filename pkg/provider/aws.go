package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/preflights"
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
	provider = &AWSProvider{}
	iamSession, callerIdentity, err := GetAWSCallerIdentity(ctx)
	if err != nil {
		return provider, plrlErrors.ErrorWrap(err, "Failed to get AWS caller identity")
	}
	provider.goContext = &ctx
	provider.ctx = map[string]any{
		"IAMSession": iamSession,
	}
	fmt.Printf("\nUsing %s AWS profile\n", getAWSProfileName())
	fmt.Printf("Caller identity ARN: %s\n", lo.FromPtr(callerIdentity.Arn))
	fmt.Printf("Caller identity account: %s\n", lo.FromPtr(callerIdentity.Account))
	fmt.Printf("Caller identity user ID: %s\n\n", lo.FromPtr(callerIdentity.UserId))

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

	provider.project = lo.FromPtr(callerIdentity.Account)
	if len(provider.project) == 0 {
		err = plrlErrors.ErrorWrap(fmt.Errorf("unable to find AWS account ID, make sure that your AWS CLI is configured"), "AWS cli error:")
		return
	}

	provider.storageClient = client

	projectManifest := manifest.ProjectManifest{
		Cluster:           provider.Cluster(),
		Project:           provider.Project(),
		Provider:          api.ProviderAWS,
		Region:            provider.Region(),
		Context:           provider.Context(),
		AvailabilityZones: []string{},
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

	return &AWSProvider{Clus: man.Cluster, project: man.Project, bucket: man.Bucket, Reg: man.Region, storageClient: client, goContext: &ctx, ctx: man.Context}, nil
}

func getClient(region string, context context.Context) (*s3.Client, error) {
	cfg, err := getAwsConfig(context)
	if err != nil {
		return nil, err
	}

	cfg.Region = region
	return s3.NewFromConfig(cfg), nil
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

func (aws *AWSProvider) mkBucket(name string) error {
	client := aws.storageClient
	_, err := client.HeadBucket(*aws.goContext, &s3.HeadBucketInput{Bucket: &name})

	if err != nil {
		bucket := &s3.CreateBucketInput{
			Bucket: &name,
		}

		if aws.Region() != "us-east-1" {
			bucket.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
				LocationConstraint: s3Types.BucketLocationConstraint(aws.Region()),
			}
		}

		_, err = client.CreateBucket(*aws.goContext, bucket)
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

func (aws *AWSProvider) Preflights() []*preflights.Preflight {
	return []*preflights.Preflight{
		{Name: "Test IAM Permissions", Callback: aws.testIamPermissions},
	}
}

func (aws *AWSProvider) Flush() error {
	if aws == nil || aws.writer == nil {
		return nil
	}
	return aws.writer()
}

func (aws *AWSProvider) Permissions() (permissions.Checker, error) {
	return permissions.NewAwsChecker(*aws.goContext)
}

func (aws *AWSProvider) Decommision(node *v1.Node) error {
	cfg, err := awsConfig.LoadDefaultConfig(*aws.goContext)

	if err != nil {
		return plrlErrors.ErrorWrap(err, "Failed to establish aws session")
	}

	cfg.Region = aws.Region()

	name := "private-dns-name"

	svc := ec2.NewFromConfig(cfg)
	instances, err := svc.DescribeInstances(*aws.goContext, &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{Name: &name, Values: []string{node.Name}},
		},
	})

	if err != nil {
		return plrlErrors.ErrorWrap(err, "failed to find node in ec2")
	}

	instance := instances.Reservations[0].Instances[0]

	_, err = svc.TerminateInstances(*aws.goContext, &ec2.TerminateInstancesInput{
		InstanceIds: []string{*instance.InstanceId},
	})

	return plrlErrors.ErrorWrap(err, "failed to terminate instance")
}

func ValidateAWSDomainRegistration(ctx context.Context, domain, region string) error {
	cfg, err := getAwsConfig(ctx)
	if err != nil {
		return err
	}

	d := strings.TrimSuffix(domain, ".") + "." // Route53 stores zone names with trailing dot.

	cfg.Region = region // Route53 is a global service, but AWS SDK requires a region to be set.
	svc := route53.NewFromConfig(cfg)

	var marker *string
	for {
		input := &route53.ListHostedZonesInput{}
		if marker != nil {
			input.Marker = marker
		}

		output, err := svc.ListHostedZones(ctx, input)
		if err != nil {
			return plrlErrors.ErrorWrap(err, "Failed to list hosted zones: ")
		}

		for _, hz := range output.HostedZones {
			if lo.FromPtr(hz.Name) == d {
				return nil // Domain is registered, return without error.
			}
		}

		if output.IsTruncated && output.NextMarker != nil {
			marker = output.NextMarker
		} else {
			break
		}
	}

	return fmt.Errorf("domain %s not found", domain)
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

func getAWSProfileName() string {
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		return profile
	}

	if profile := os.Getenv("AWS_DEFAULT_PROFILE"); profile != "" {
		return profile
	}

	return "default"
}

// GetAWSCallerIdentity returns the IAM role ARN of the current caller identity.
func GetAWSCallerIdentity(ctx context.Context) (string, *sts.GetCallerIdentityOutput, error) {
	cfg, err := getAwsConfig(ctx)
	if err != nil {
		return "", nil, err
	}

	svc := sts.NewFromConfig(cfg)
	callerIdentity, err := svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", callerIdentity, plrlErrors.ErrorWrap(err, "Error getting caller identity: ")
	}

	callerIdentityArn := lo.FromPtr(callerIdentity.Arn)
	roleName, _ := RoleNameSessionFromARN(callerIdentityArn)
	if !lo.IsEmpty(roleName) {
		role, err := iam.NewFromConfig(cfg).GetRole(ctx, &iam.GetRoleInput{RoleName: &roleName})
		if err != nil {
			return "", callerIdentity, plrlErrors.ErrorWrap(err, "Error getting IAM role: ")
		}

		return lo.FromPtr(role.Role.Arn), callerIdentity, nil
	}

	return callerIdentityArn, callerIdentity, nil
}
