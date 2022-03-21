package provider

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
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
	storageClient *s3.S3
}

var awsSurvey = []*survey.Question{
	{
		Name:     "cluster",
		Prompt:   &survey.Input{Message: "Enter the name of your cluster:"},
		Validate: validCluster,
	},
	{
		Name:     "region",
		Prompt:   &survey.Input{Message: "What region will you deploy to?", Default: "us-east-2"},
		Validate: survey.Required,
	},
}

func mkAWS(conf config.Config) (*AWSProvider, error) {
	provider := &AWSProvider{}
	if err := survey.Ask(awsSurvey, provider); err != nil {
		return nil, err
	}

	client, err := getClient(provider.Reg)
	if err != nil {
		return nil, err
	}

	account, err := GetAwsAccount()
	if err != nil {
		return nil, errors.ErrorWrap(err, "Failed to get aws account (is your aws cli configured?)")
	}

	if len(account) <= 0 {
		return nil, errors.ErrorWrap(fmt.Errorf("Unable to find aws account id, is your aws cli configured?"), "AWS cli error:")
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

	if err := projectManifest.Configure(); err != nil {
		return nil, err
	}

	provider.bucket = projectManifest.Bucket
	return provider, nil
}

func awsFromManifest(man *manifest.ProjectManifest) (*AWSProvider, error) {
	client, err := getClient(man.Region)
	if err != nil {
		return nil, err
	}

	return &AWSProvider{man.Cluster, man.Project, man.Bucket, man.Region, client}, nil
}

func getClient(region string) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		return nil, errors.ErrorWrap(err, "Failed to initialize aws client: ")
	}

	return s3.New(sess), nil
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
	_, err := client.HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(name)})

	if err != nil {
		_, err = client.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(name)})
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

func (prov *AWSProvider) Decommision(node *v1.Node) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(prov.Region()),
	})

	if err != nil {
		return errors.ErrorWrap(err, "Failed to establish aws session")
	}

	svc := ec2.New(sess)
	instances, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{Name: aws.String("private-dns-name"), Values: []*string{aws.String(node.ObjectMeta.Name)}},
		},
	})

	if err != nil {
		return errors.ErrorWrap(err, "failed to find node in ec2")
	}

	instance := instances.Reservations[0].Instances[0]

	_, err = svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
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

	json.Unmarshal(out, &res)
	return res.Account, nil
}
