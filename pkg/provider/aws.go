package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
)

type AWSProvider struct {
	cluster       string
	project       string
	bucket        string
	region        string
	storageClient *s3.S3
}

func mkAWS() (*AWSProvider, error) {
	cluster, _ := utils.ReadLine("Enter the name of your cluster: ")
	bucket, _ := utils.ReadLine("Enter the name of a s3 bucket to use for state, eg: <yourprojectname>-tf-state: ")
	region, _ := utils.ReadLine("Enter the region you want to deploy to eg us-east-2: ")

	client, err := getClient(region)
	if err != nil {
		return nil, err
	}

	account, err := getAwsAccount()
	if err != nil {
		return nil, err
	}

	provider := &AWSProvider{
		cluster,
		account,
		bucket,
		region,
		client,
	}

	projectManifest := manifest.ProjectManifest{
		Cluster:  cluster,
		Project:  account,
		Bucket:   bucket,
		Provider: AWS,
		Region:   provider.Region(),
	}
	path := manifest.ProjectManifestPath()
	projectManifest.Write(path)

	return provider, nil
}

func awsFromManifest(man *manifest.Manifest) (*AWSProvider, error) {
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
		return nil, err
	}

	return s3.New(sess), nil
}

func (aws *AWSProvider) CreateBackend(prefix string, ctx map[string]interface{}) (string, error) {
	if err := aws.mkBucket(aws.bucket); err != nil {
		return "", err
	}

	ctx["Region"] = aws.Region()
	ctx["Bucket"] = aws.Bucket()
	ctx["Prefix"] = prefix
	ctx["__CLUSTER__"] = aws.Cluster()
	if _, ok := ctx["Cluster"]; !ok {
		ctx["Cluster"] = fmt.Sprintf("\"%s\"", aws.Cluster())
	}
	return template.RenderString(awsBackendTemplate, ctx)
}

func (aws *AWSProvider) KubeConfig() error {
	if utils.InKubernetes() {
		return nil
	}

	cmd := exec.Command(
		"aws", "eks", "update-kubeconfig", "--name", aws.cluster, "--region", aws.region)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

func (aws *AWSProvider) Install() (err error) {
	if exists, _ := utils.Which("aws"); exists {
		utils.Success("aws cli already installed!\n")
		return
	}

	fmt.Println("AWS requires you to manually pkg install the aws cli")
	osName := runtime.GOOS
	if osName == "darwin" {
		osName = "mac"
	}

	fmt.Printf("Visit https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-%s.html to install\n", osName)
	return
}

func (aws *AWSProvider) Name() string {
	return AWS
}

func (aws *AWSProvider) Cluster() string {
	return aws.cluster
}

func (aws *AWSProvider) Project() string {
	return aws.project
}

func (aws *AWSProvider) Bucket() string {
	return aws.bucket
}

func (aws *AWSProvider) Region() string {
	return aws.region
}

func (aws *AWSProvider) Context() map[string]interface{} {
	return map[string]interface{}{}
}

func getAwsAccount() (string, error) {
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
