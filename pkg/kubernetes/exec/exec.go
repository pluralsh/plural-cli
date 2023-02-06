package exec

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/kubernetes/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/cmd/util/podcmd"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/term"
)

func Exec(namespace, resource string, commands []string) error {
	obj, pod, err := utils.GetPodWithObject(namespace, resource)
	if err != nil {
		return err
	}
	if meta.IsListType(obj) {
		return fmt.Errorf("cannot exec into multiple objects at a time")
	}
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
	}

	container, err := podcmd.FindOrDefaultContainerByName(pod, "", true, os.Stderr)
	if err != nil {
		return err
	}

	t := setupTTY()
	sizeQueue := t.MonitorSize(t.GetSize())

	kube, err := kubernetes.Kubernetes()
	if err != nil {
		return err
	}

	fn := func() error {
		req := kube.GetRestClient().Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
			SubResource("exec")
		req.VersionedParams(&corev1.PodExecOptions{
			Container: container.Name,
			Command:   commands,
			Stdin:     true,
			Stdout:    true,
			TTY:       t.Raw,
		}, scheme.ParameterCodec)

		return execute("POST", req.URL(), os.Stdin, os.Stdout, os.Stderr, t.Raw, sizeQueue)
	}

	return t.Safe(fn)
}

func execute(method string, url *url.URL, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	config, err := kubernetes.KubeConfig()
	if err != nil {
		return err
	}
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: terminalSizeQueue,
	})
}

func setupTTY() term.TTY {
	t := term.TTY{
		Out: os.Stdout,
		In:  os.Stdin,
		Raw: true,
	}

	return t
}
