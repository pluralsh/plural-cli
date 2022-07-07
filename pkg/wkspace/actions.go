package wkspace

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

type checker func(s string) bool

var alwaysErr checker = func(s string) bool { return false }

func (c checker) execSuppressed(command string, args ...string) (err error) {
	for retry := 2; retry >= 0; retry-- {
		utils.Highlight("%s %s ~> ", command, strings.Join(args, " "))
		cmd, out := executor.SuppressedCommand(command, args...)
		err = executor.RunCommand(cmd, out)
		if err == nil || c(out.Format()) {
			break
		}
		fmt.Printf("retrying command, number of retries remaining: %d\n", retry)
	}

	return
}

func (w *Workspace) DestroyHelm() error {
	// ensure current kubeconfig is correct before destroying stuff
	if err := w.Provider.KubeConfig(); err != nil {
		return err
	}

	name := w.Installation.Repository.Name

	ns := w.Config.Namespace(name)
	if err := alwaysErr.execSuppressed("helm", "get", "values", name, "-n", ns); err != nil {
		fmt.Println("Helm already uninstalled, continuing...")
		return nil
	}

	r := regexp.MustCompile("release.*not found")
	var ignoreNotFound checker = func(s string) bool { return r.MatchString(s) }
	return ignoreNotFound.execSuppressed("helm", "del", name, "-n", ns)
}

func (w *Workspace) Bounce() error {
	return w.ToMinimal().BounceHelm()
}

func (w *Workspace) HelmDiff() error {
	return w.ToMinimal().DiffHelm()
}

func (w *Workspace) Destroy() error {
	if err := w.DestroyHelm(); err != nil {
		return err
	}

	if err := w.DestroyTerraform(); err != nil {
		return err
	}

	return w.Reset()
}

func (w *Workspace) Reset() error {
	repo := w.Installation.Repository
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	deployFile := pathing.SanitizeFilepath(filepath.Join(repoRoot, repo.Name, "deploy.hcl"))
	if err := os.Remove(deployFile); err != nil {
		return err
	}
	return nil
}

func (w *Workspace) DestroyTerraform() error {
	repo := w.Installation.Repository
	path, err := filepath.Abs(pathing.SanitizeFilepath(filepath.Join(repo.Name, "terraform")))
	if err != nil {
		return err
	}

	time.AfterFunc(1*time.Minute, func() {
		kube, err := utils.Kubernetes()
		if err != nil {
			fmt.Printf("Could not set up k8s client due to %s\n", err)
			return
		}

		ns := w.Config.Namespace(repo.Name)
		if err := kube.FinalizeNamespace(ns); err != nil {
			return
		}
	})

	if err := os.Chdir(path); err != nil {
		return err
	}
	if err := alwaysErr.execSuppressed("terraform", "init", "-upgrade"); err != nil {
		return err
	}

	return alwaysErr.execSuppressed("terraform", "destroy", "-auto-approve")
}
