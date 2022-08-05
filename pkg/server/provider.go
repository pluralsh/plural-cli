package server

import (
	"fmt"
	"io/ioutil"

	"github.com/mitchellh/go-homedir"

	prov "github.com/pluralsh/plural/pkg/provider"
)

func setupProvider(setup *SetupRequest) error {
	if setup.Provider == "aws" {
		return setupAws(setup)
	}

	if setup.Provider == "gcp" {
		return setupGcp(setup)
	}

	return nil
}

func setupGcp(setup *SetupRequest) error {
	f, err := homedir.Expand("~/gcp.json")
	if err != nil {
		return fmt.Errorf("error getting the gcp.json path: %w", err)
	}

	if err := ioutil.WriteFile(f, []byte(setup.Credentials.Gcp.ApplicationCredentials), 0644); err != nil {
		return fmt.Errorf("error writing gcp credentials: %w", err)
	}

	if err := execCmd("gcloud", "auth", "activate-service-account", "--key-file", f, "--project", setup.Workspace.Project); err != nil {
		return fmt.Errorf("error authenticating to gcloud: %w", err)
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
