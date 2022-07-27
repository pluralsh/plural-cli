package logs

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pluralsh/plural/pkg/kubernetes"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

func List(kube kubernetes.Kube, namespace string) (*v1alpha1.LogTailList, error) {
	return kube.LogTailList(namespace)
}

func Tail(kube kubernetes.Kube, namespace string, name string) error {
	tail, err := kube.LogTail(namespace, name)
	if err != nil {
		return err
	}

	args := []string{"logs", fmt.Sprintf("--tail=%d", tail.Spec.Limit)}
	if tail.Spec.Follow {
		args = append(args, "-f")
	}
	args = append(args, tail.Spec.Target, "-n", namespace)

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
