package logs

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/pluralsh/plural/pkg/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/cmd/util/podcmd"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/podutils"
)

func logsForObject(object, options runtime.Object, timeout time.Duration, allContainers bool) (map[corev1.ObjectReference]rest.ResponseWrapper, error) {
	config, err := kubernetes.KubeConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := corev1client.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return logsForObjectWithClient(clientset, object, options, timeout, allContainers)
}

// this is split for easy test-ability
func logsForObjectWithClient(clientset corev1client.CoreV1Interface, object, options runtime.Object, timeout time.Duration, allContainers bool) (map[corev1.ObjectReference]rest.ResponseWrapper, error) {
	opts, ok := options.(*corev1.PodLogOptions)
	if !ok {
		return nil, errors.New("provided options object is not a PodLogOptions")
	}

	switch t := object.(type) {
	case *corev1.PodList:
		ret := make(map[corev1.ObjectReference]rest.ResponseWrapper)
		for i := range t.Items {
			currRet, err := logsForObjectWithClient(clientset, &t.Items[i], options, timeout, allContainers)
			if err != nil {
				return nil, err
			}
			for k, v := range currRet {
				ret[k] = v
			}
		}
		return ret, nil

	case *corev1.Pod:
		return logsForPod(t, opts, clientset, timeout, allContainers)
	}

	namespace, selector, err := polymorphichelpers.SelectorsForObject(object)
	if err != nil {
		return nil, fmt.Errorf("cannot get the logs from %T: %w", object, err)
	}

	sortBy := func(pods []*corev1.Pod) sort.Interface { return podutils.ByLogging(pods) }
	pod, numPods, err := polymorphichelpers.GetFirstPod(clientset, namespace, selector.String(), timeout, sortBy)
	if err != nil {
		return nil, err
	}
	if numPods > 1 {
		fmt.Fprintf(os.Stderr, "Found %v pods, using pod/%v\n", numPods, pod.Name)
	}

	return logsForObjectWithClient(clientset, pod, options, timeout, allContainers)
}

func logsForPod(t *corev1.Pod, opts *corev1.PodLogOptions, clientset corev1client.CoreV1Interface, timeout time.Duration, allContainers bool) (map[corev1.ObjectReference]rest.ResponseWrapper, error) {
	// if allContainers is true, then we're going to locate all containers and then iterate through them. At that point, "allContainers" is false
	if !allContainers {
		currOpts := new(corev1.PodLogOptions)
		if opts != nil {
			opts.DeepCopyInto(currOpts)
		}
		// in case the "kubectl.kubernetes.io/default-container" annotation is present, we preset the opts.Containers to default to selected
		// container. This gives users ability to preselect the most interesting container in pod.
		if annotations := t.GetAnnotations(); annotations != nil && currOpts.Container == "" {
			var defaultContainer string
			if len(annotations[podcmd.DefaultContainerAnnotationName]) > 0 {
				defaultContainer = annotations[podcmd.DefaultContainerAnnotationName]
			}
			if len(defaultContainer) > 0 {
				if exists, _ := podcmd.FindContainerByName(t, defaultContainer); exists == nil {
					fmt.Fprintf(os.Stderr, "Default container name %q not found in pod %s\n", defaultContainer, t.Name)
				} else {
					currOpts.Container = defaultContainer
				}
			}
		}

		if currOpts.Container == "" {
			// Default to the first container name(aligning behavior with `kubectl exec').
			currOpts.Container = t.Spec.Containers[0].Name
			if len(t.Spec.Containers) > 1 || len(t.Spec.InitContainers) > 0 || len(t.Spec.EphemeralContainers) > 0 {
				fmt.Fprintf(os.Stderr, "Defaulted container %q out of: %s\n", currOpts.Container, podcmd.AllContainerNames(t))
			}
		}

		container, fieldPath := podcmd.FindContainerByName(t, currOpts.Container)
		if container == nil {
			return nil, fmt.Errorf("container %s is not valid for pod %s", currOpts.Container, t.Name)
		}
		ref, err := reference.GetPartialReference(scheme.Scheme, t, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("Unable to construct reference to '%#v': %w", t, err)
		}

		ret := make(map[corev1.ObjectReference]rest.ResponseWrapper, 1)
		ret[*ref] = clientset.Pods(t.Namespace).GetLogs(t.Name, currOpts)
		return ret, nil
	}

	ret := make(map[corev1.ObjectReference]rest.ResponseWrapper)
	for _, c := range t.Spec.InitContainers {
		currOpts := opts.DeepCopy()
		currOpts.Container = c.Name
		currRet, err := logsForObjectWithClient(clientset, t, currOpts, timeout, false)
		if err != nil {
			return nil, err
		}
		for k, v := range currRet {
			ret[k] = v
		}
	}
	for _, c := range t.Spec.Containers {
		currOpts := opts.DeepCopy()
		currOpts.Container = c.Name
		currRet, err := logsForObjectWithClient(clientset, t, currOpts, timeout, false)
		if err != nil {
			return nil, err
		}
		for k, v := range currRet {
			ret[k] = v
		}
	}
	for _, c := range t.Spec.EphemeralContainers {
		currOpts := opts.DeepCopy()
		currOpts.Container = c.Name
		currRet, err := logsForObjectWithClient(clientset, t, currOpts, timeout, false)
		if err != nil {
			return nil, err
		}
		for k, v := range currRet {
			ret[k] = v
		}
	}

	return ret, nil
}
