package provider

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
)

type AzureProvider struct {
	cluster       string
	project       string
	bucket        string
	region        string
}

const azureBackendTemplate = `terraform {
	backend "s3" {
		bucket = {{ .Values.Bucket | quote }}
		key = "{{ .Values.__CLUSTER__ }}/{{ .Values.Prefix }}/terraform.tfstate"
		region = {{ .Values.Region | quote }}
	}

	required_providers {
    azure = {
      source  = "hashicorp/azure"
      version = "~> 3.36.0"
    }
		kubernetes = {
			source  = "hashicorp/kubernetes"
			version = "~> 2.0.3"
		}
  }
}

provider "azure" {
  region = {{ .Values.Region | quote }}
}

data "aws_eks_cluster" "cluster" {
  name = {{ .Values.Cluster }}
}

data "aws_eks_cluster_auth" "cluster" {
  name = {{ .Values.Cluster }}
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
  token                  = data.aws_eks_cluster_auth.cluster.token
}
`

func mkAzure() (*AzureProvider, error) {
	cluster, _ := utils.ReadLine("Enter the name of your cluster: ")
	bucket, _ := utils.ReadLine("Enter the name of a storage bucket to use for state, eg: <yourprojectname>-tf-state: ")
	region, _ := utils.ReadLine("Enter the region you want to deploy to eg us-east-2: ")

	provider := &AzureProvider{
		cluster,
		"",
		bucket,
		region,
	}

	projectManifest := manifest.ProjectManifest{
		Cluster:  cluster,
		Project:  "",
		Bucket:   bucket,
		Provider: AZURE,
		Region:   provider.Region(),
	}
	path := manifest.ProjectManifestPath()
	projectManifest.Write(path)

	return provider, nil
}

func azureFromManifest(man *manifest.Manifest) (*AzureProvider, error) {
	return &AzureProvider{man.Cluster, man.Project, man.Bucket, man.Region}, nil
}

func (azure *AzureProvider) CreateBackend(prefix string, ctx map[string]interface{}) (string, error) {
	ctx["Region"] = azure.Region()
	ctx["Bucket"] = azure.Bucket()
	ctx["Prefix"] = prefix
	ctx["__CLUSTER__"] = azure.Cluster()
	if _, ok := ctx["Cluster"]; !ok {
		ctx["Cluster"] = fmt.Sprintf("\"%s\"", azure.Cluster())
	}

	return template.RenderString(azureBackendTemplate, ctx)
}

func (azure *AzureProvider) KubeConfig() error {
	if utils.InKubernetes() {
		return nil
	}

	cmd := exec.Command(
		"az", "eks", "update-kubeconfig")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (azure *AzureProvider) Install() (err error) {
	if exists, _ := utils.Which("az"); exists {
		utils.Success("azure cli already installed!\n")
		return
	}

	fmt.Println("visit https://docs.microsoft.com/en-us/cli/azure/install-azure-cli to install")
	return
}

func (az *AzureProvider) Name() string {
	return AZURE
}

func (az *AzureProvider) Cluster() string {
	return az.cluster
}

func (az *AzureProvider) Project() string {
	return az.project
}

func (az *AzureProvider) Bucket() string {
	return az.bucket
}

func (az *AzureProvider) Region() string {
	return az.region
}
