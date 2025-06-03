package server

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/mitchellh/go-homedir"

	prov "github.com/pluralsh/plural/pkg/provider"
)

const azureEnvFile = `
export AZURE_CLIENT_ID={{ .ClientId }}
export AZURE_TENANT_ID={{ .TenantId }}
export AZURE_CLIENT_SECRET={{ .ClientSecret }}
export ARM_CLIENT_ID={{ .ClientId }}
export ARM_TENANT_ID={{ .TenantId }}
export ARM_CLIENT_SECRET={{ .ClientSecret }}
export ARM_USE_MSI=true
export ARM_SUBSCRIPTION_ID={{ .SubscriptionId }}
`

func setupProvider(setup *SetupRequest) error {
	if marked("cloud") {
		return nil
	}

	if setup.Provider == "aws" {
		return setupAws(setup)
	}

	if setup.Provider == "gcp" {
		return setupGcp(setup)
	}

	if setup.Provider == "azure" {
		return setupAzure(setup)
	}

	return nil
}

func setupGcp(setup *SetupRequest) error {
	f, err := homedir.Expand("~/gcp.json")
	if err != nil {
		return fmt.Errorf("error getting the gcp.json path: %w", err)
	}

	if err := os.WriteFile(f, []byte(setup.Credentials.Gcp.ApplicationCredentials), 0644); err != nil {
		return fmt.Errorf("error writing gcp credentials: %w", err)
	}

	if out, err := execCmdWithOutput("gcloud", "auth", "activate-service-account", "--key-file", f, "--project", setup.Workspace.Project); err != nil {
		return fmt.Errorf("error authenticating to gcloud: %s", out)
	}

	return nil
}

func setupAzure(setup *SetupRequest) error {
	az := setup.Credentials.Azure
	setup.Context = map[string]interface{}{
		"TenantId":       az.TenantId,
		"SubscriptionId": az.SubscriptionId,
		"StorageAccount": az.StorageAccount,
	}

	tpl, err := template.New("azure").Parse(azureEnvFile)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	out.Grow(5 * 1024)
	if err := tpl.Execute(&out, az); err != nil {
		return err
	}

	f, err := homedir.Expand("~/.env")
	if err != nil {
		return err
	}

	if err := os.WriteFile(f, out.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing azure env file: %w", err)
	}

	if out, err := execCmdWithOutput("az", "login", "--service-principal", "-u", az.ClientId, "-p", az.ClientSecret, "--tenant", az.TenantId); err != nil {
		return fmt.Errorf("error logging into az cli: %s", out)
	}

	return nil
}

func setupAws(setup *SetupRequest) error {
	aws := setup.Credentials.Aws

	if err := awsConfig("default.region", setup.Workspace.Region); err != nil {
		return fmt.Errorf("error configuring default aws region: %w", err)
	}

	if err := awsConfig("aws_access_key_id", aws.AccessKeyId); err != nil {
		return fmt.Errorf("error configuring aws access key: %w", err)
	}

	if err := awsConfig("aws_secret_access_key", aws.SecretAccessKey); err != nil {
		return fmt.Errorf("error configuring aws secret key: %w", err)
	}

	accountId, err := prov.GetAwsAccount()
	if err != nil {
		return fmt.Errorf("error getting aws account: %w", err)
	}

	setup.Workspace.Project = accountId
	return nil
}

func awsConfig(args ...string) error {
	allArgs := []string{"configure", "set"}
	allArgs = append(allArgs, args...)
	return execCmd("aws", allArgs...)
}
