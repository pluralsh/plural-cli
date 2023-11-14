package utils

import (
	"fmt"
	"sort"
	"time"

	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/podutils"
)

const (
	// Amount of time to wait until at least one pod is running
	defaultPodWaitTimeout = 60 * time.Second
)

func GetPodWithObject(namespace, resource string) (runtime.Object, *corev1.Pod, error) {
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0))
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	builder := f.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(namespace).DefaultNamespace().
		SingleResourceType()
	builder.ResourceNames("pods", resource)
	obj, err := builder.Do().Object()
	if err != nil {
		return nil, nil, err
	}

	return attachablePodForObject(obj, defaultPodWaitTimeout)

}

func attachablePodForObject(object runtime.Object, timeout time.Duration) (runtime.Object, *corev1.Pod, error) {
	if t, ok := object.(*corev1.Pod); ok {
		return object, t, nil
	}

	clientConfig, err := kubernetes.KubeConfig()
	if err != nil {
		return nil, nil, err
	}
	clientset, err := corev1client.NewForConfig(clientConfig)
	if err != nil {
		return nil, nil, err
	}

	namespace, selector, err := polymorphichelpers.SelectorsForObject(object)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot attach to %T: %w", object, err)
	}
	sortBy := func(pods []*corev1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(clientset, namespace, selector.String(), timeout, sortBy)
	return object, pod, err
}
