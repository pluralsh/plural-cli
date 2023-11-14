package portforward

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/kubernetes/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func PortForward(namespace, resource string, ports []string, stopChan, readyChan chan struct{}) error {
	obj, forwardablePod, err := utils.GetPodWithObject(namespace, resource)
	if err != nil {
		return err
	}
	podName := forwardablePod.Name
	if len(podName) == 0 {
		return fmt.Errorf("pod name or resource type/name must be specified")
	}

	var podPorts []string
	// handle service port mapping to target port if needed
	switch t := obj.(type) {
	case *corev1.Service:
		err = checkUDPPortInService(ports, t)
		if err != nil {
			return err
		}
		podPorts, err = translateServicePortToTargetPort(ports, *t, *forwardablePod)
		if err != nil {
			return err
		}
	default:
		err = checkUDPPortInPod(ports, forwardablePod)
		if err != nil {
			return err
		}
		podPorts, err = convertPodNamedPortToNumber(ports, *forwardablePod)
		if err != nil {
			return err
		}
	}
	if len(podPorts) < 1 {
		return fmt.Errorf("at least 1 PORT is required for port-forward")
	}
	kube, err := kubernetes.Kubernetes()
	if err != nil {
		return err
	}

	pod, err := kube.GetClient().CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	req := kube.GetRestClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod.Name).
		SubResource("portforward")

	return forwardPorts(http.MethodPost, req.URL(), podPorts, stopChan, readyChan)
}

func forwardPorts(method string, url *url.URL, ports []string, stopChan, readyChan chan struct{}) error {

	clientConfig, err := kubernetes.KubeConfig()
	if err != nil {
		return err
	}

	transport, upgrader, err := spdy.RoundTripperFor(clientConfig)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforward.New(dialer, ports, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}
