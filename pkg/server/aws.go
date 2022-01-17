package server

import (
	"github.com/pluralsh/plural/pkg/provider"	
)

func setupAws(setup *SetupRequest) error {
	aws := setup.Credentials.Aws
	

	if err := awsConfig("default.region", setup.Workspace.Region); err != nil {
		return err
	}

	if err := awsConfig("aws_access_key_id", aws.AccessKeyId); err != nil {
		return err
	}

	if err := awsConfig("aws_secret_access_key", aws.SecretAccessKey); err != nil {
		return err
	}

	accountId, err := provider.GetAwsAccount()
	if err != nil {
		return err
	}

	setup.Workspace.Project = accountId
	return nil
}

func awsConfig(args ...string) error {
	allArgs := []string{"configure", "set"}
	allArgs = append(allArgs, args...)
	return execCmd("aws", allArgs...)
}