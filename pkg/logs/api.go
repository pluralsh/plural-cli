package logs

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func List(namespace string) (*v1alpha1.LogTailList, error) {
	kube, err := utils.Kubernetes()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return kube.Plural.PlatformV1alpha1().LogTails(namespace).List(ctx, metav1.ListOptions{})
}

func Tail(namespace string, name string) error {
	kube, err := utils.Kubernetes()
	if err != nil {
		return err
	}

	ctx := context.Background()
	tail, err := kube.Plural.PlatformV1alpha1().LogTails(namespace).Get(ctx, name, metav1.GetOptions{})
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
