package proxy

import (
	"os"
	"os/exec"

	"github.com/michaeljguarino/forge/types/v1alpha1"
)

func execShell(namespace string, proxy *v1alpha1.Proxy) error {
	shell := proxy.Spec.ShConfig
	var rest []string
	if len(shell.Command) > 0 {
		rest = append([]string{"-e", shell.Command}, shell.Args...)
	}
	args := []string{"exec", "-it", "-n", namespace, proxy.Spec.Target, "--", "/bin/sh"}
	args = append(args, rest...)
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
