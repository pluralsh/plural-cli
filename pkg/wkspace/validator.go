package wkspace

import (
	"fmt"
	"os/exec"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
)

func Preflight() error {
	requirements := []string{"helm", "kubectl", "terraform", "git"}
	for _, req := range requirements {
		if ok, _ := utils.Which(req); !ok {
			return utils.HighlightError(fmt.Errorf("%s not installed", req))
		}
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if _, err := cmd.CombinedOutput(); err != nil {
		return utils.HighlightError(fmt.Errorf("not in a git repository, or repository has no initial commit"))
	}

	return nil
}

func (wk *Workspace) Validate() error {
	for _, tf := range wk.Terraform {
		if err := wk.providersValid(tf.Terraform.Dependencies.Providers); err != nil {
			return err
		}
	}

	for _, chart := range wk.Charts {
		if err := wk.providersValid(chart.Chart.Dependencies.Providers); err != nil {
			return err
		}
	}

	return nil
}

func (wk *Workspace) providersValid(providers []string) error {
	if len(providers) == 0 {
		return nil
	}

	pass := false
	for _, provider := range providers {
		if wk.match(provider) {
			pass = true
		}
	}

	if !pass {
		return fmt.Errorf("Provider %s is not supported for any of %v", wk.Provider.Name(), providers)
	}

	return nil
}

func (wk *Workspace) match(prov string) bool {
	switch wk.Provider.Name() {
	case provider.GCP:
		return prov == "GCP"
	case provider.AWS:
		return prov == "AWS"
	case provider.AZURE:
		return prov == "AZURE"
	default:
		return false
	}
}
