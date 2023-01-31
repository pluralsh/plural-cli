package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/mitchellh/go-homedir"
)

func Init() error {
	ctx := context.Background()
	homeDir, err := homedir.Expand("~/.plural")
	if err != nil {
		return err
	}
	terraformBinPath := filepath.Join(homeDir, "terraform")
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	tf, err := tfexec.NewTerraform(workingDir, terraformBinPath)
	if err != nil {
		return fmt.Errorf("error running NewTerraform: %w", err)
	}
	tf.SetStdout(os.Stdout)

	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		return fmt.Errorf("error running Init: %w", err)
	}

	return nil
}
