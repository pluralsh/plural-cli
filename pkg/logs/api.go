package logs

import (
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/kubernetes/logs"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
)

func List(kube kubernetes.Kube, namespace string) (*v1alpha1.LogTailList, error) {
	return kube.LogTailList(namespace)
}

func Print(tails *v1alpha1.LogTailList) error {
	headers := []string{"Name", "Follow", "Target"}
	return utils.PrintTable[v1alpha1.LogTail](tails.Items, headers, func(log v1alpha1.LogTail) ([]string, error) {
		follow := "False"
		if log.Spec.Follow {
			follow = "True"
		}

		return []string{log.Name, follow, log.Spec.Target}, nil
	})
}

func Tail(kube kubernetes.Kube, namespace string, name string) error {
	tail, err := kube.LogTail(namespace, name)
	if err != nil {
		return err
	}

	return logs.Logs(namespace, tail.Spec.Target, int64(tail.Spec.Limit))
}
