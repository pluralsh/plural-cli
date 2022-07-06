package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/errors"
	v1 "k8s.io/api/core/v1"
)

type AWSProvider struct {
	Clus          string `survey:"cluster"`
	project       string
	bucket        string
	Reg           string `survey:"region"`
	storageClient *s3.Client
	writer        manifest.Writer
	goContext     *context.Context
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

var awsSurvey = []*survey.Question{
	{
		Name:     "cluster",
		Prompt:   &survey.Input{Message: "Enter the name of your cluster:"},
		Validate: validCluster,
	},
	{
		Name:     "region",
		Prompt:   &survey.Select{Message: "What region will you deploy to?", Default: "us-east-2", Options: awsRegions},
		Validate: survey.Required,
	},
}

func mkAWS(conf config.Config) (provider *AWSProvider, err error) {
	provider = &AWSProvider{}
	if err = survey.Ask(awsSurvey, provider); err != nil {
		return
	}

	ctx := context.Background()

	provider.goContext = &ctx

	client, err := getClient(provider.Reg, *provider.goContext)
	if err != nil {
		return
	}

	account, err := GetAwsAccount()
	if err != nil {
		err = errors.ErrorWrap(err, "Failed to get aws account (is your aws cli configured?)")
		return
	}

	if len(account) == 0 {
		err = errors.ErrorWrap(fmt.Errorf("Unable to find aws account id, is your aws cli configured?"), "AWS cli error:")
		return
	}

	provider.project = account
	provider.storageClient = client

	projectManifest := manifest.ProjectManifest{
		Cluster:  provider.Cluster(),
		Project:  provider.Project(),
		Provider: AWS,
		Region:   provider.Region(),
		Owner:    &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}

	provider.writer = projectManifest.Configure()
	provider.bucket = projectManifest.Bucket
	return
}

func awsFromManifest(man *manifest.ProjectManifest) (*AWSProvider, error) {
	ctx := context.Background()
	client, err := getClient(man.Region, ctx)
	if err != nil {
		return nil, err
	}

	return &AWSProvider{Clus: man.Cluster, project: man.Project, bucket: man.Bucket, Reg: man.Region, storageClient: client, goContext: &ctx}, nil
}

func getClient(region string, context context.Context) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context)

	if err != nil {
		return nil, errors.ErrorWrap(err, "Failed to initialize aws client: ")
	}

	cfg.Region = region

	return s3.NewFromConfig(cfg), nil
}

func (aws *AWSProvider) CreateBackend(prefix string, ctx map[string]interface{}) (string, error) {
	if err := aws.mkBucket(aws.bucket); err != nil {
		return "", errors.ErrorWrap(err, fmt.Sprintf("Failed to create terraform state bucket %s", aws.bucket))
	}

	ctx["Region"] = aws.Region()
	ctx["Bucket"] = aws.Bucket()
	ctx["Prefix"] = prefix
	ctx["__CLUSTER__"] = aws.Cluster()
	if _, ok := ctx["Cluster"]; !ok {
		ctx["Cluster"] = fmt.Sprintf("\"%s\"", aws.Cluster())
	}
	scaffold, err := GetProviderScaffold("AWS")
	if err != nil {
		return "", err
	}
	return template.RenderString(scaffold, ctx)
}

func (aws *AWSProvider) KubeConfig() error {
	if utils.InKubernetes() {
		return nil
	}

	cmd := exec.Command(
		"aws", "eks", "update-kubeconfig", "--name", aws.Cluster(), "--region", aws.Region())
	return utils.Execute(cmd)
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
	return AWS
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
	return map[string]interface{}{}
}

func (aws *AWSProvider) Preflights() []*Preflight {
	return nil
}

func (aws *AWSProvider) Flush() error {
	if aws.writer == nil {
		return nil
	}
	return aws.writer()
}

func (prov *AWSProvider) Decommision(node *v1.Node) error {
	cfg, err := awsConfig.LoadDefaultConfig(*prov.goContext)

	if err != nil {
		return errors.ErrorWrap(err, "Failed to establish aws session")
	}

	cfg.Region = prov.Region()

	name := "private-dns-name"

	svc := ec2.NewFromConfig(cfg)
	instances, err := svc.DescribeInstances(*prov.goContext, &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{Name: &name, Values: []string{node.ObjectMeta.Name}},
		},
	})

	if err != nil {
		return errors.ErrorWrap(err, "failed to find node in ec2")
	}

	instance := instances.Reservations[0].Instances[0]

	_, err = svc.TerminateInstances(*prov.goContext, &ec2.TerminateInstancesInput{
		InstanceIds: []string{*instance.InstanceId},
	})

	return errors.ErrorWrap(err, "failed to terminate instance")
}

func GetAwsAccount() (string, error) {
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(out)
		return "", err
	}

	var res struct {
		Account string
	}

	if err := json.Unmarshal(out, &res); err != nil {
		return "", err
	}
	return res.Account, nil
}
