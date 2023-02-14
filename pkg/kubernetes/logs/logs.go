package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	defaultPodLogsTimeout = 20 * time.Second
)

func Logs(namespace, resource string, tailLines int64) error {
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0))
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	builder := f.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(namespace).DefaultNamespace().
		SingleResourceType()
	builder.ResourceNames("pods", resource)
	infos, err := builder.Do().Infos()
	if err != nil {
		return err
	}
	if len(infos) != 1 {
		return fmt.Errorf("expected a resource")
	}
	object := infos[0].Object

	options, err := logOptions(tailLines)
	if err != nil {
		return err
	}
	requests, err := logsForObject(object, options, defaultPodLogsTimeout, false)
	if err != nil {
		return err
	}
	if len(requests) > 1 {
		return parallelConsumeRequest(requests)
	}

	return sequentialConsumeRequest(requests)
}

func parallelConsumeRequest(requests map[corev1.ObjectReference]rest.ResponseWrapper) error {
	reader, writer := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(len(requests))
	for objRef, request := range requests {
		go func(objRef corev1.ObjectReference, request rest.ResponseWrapper) {
			defer wg.Done()
			if err := defaultConsumeRequest(request, os.Stdout); err != nil {
				fmt.Fprintf(writer, "error: %v\n", err)
			}

		}(objRef, request)
	}

	go func() {
		wg.Wait()
		writer.Close()
	}()

	_, err := io.Copy(os.Stdout, reader)
	return err
}

func sequentialConsumeRequest(requests map[corev1.ObjectReference]rest.ResponseWrapper) error {
	for _, request := range requests {
		if err := defaultConsumeRequest(request, os.Stdout); err != nil {
			return err
		}
	}

	return nil
}

func logOptions(tailLines int64) (*corev1.PodLogOptions, error) {
	logOptions := &corev1.PodLogOptions{
		Follow:    true,
		TailLines: &tailLines,
	}

	return logOptions, nil
}

func defaultConsumeRequest(request rest.ResponseWrapper, out io.Writer) error {
	readCloser, err := request.Stream(context.TODO())
	if err != nil {
		return err
	}
	defer readCloser.Close()

	r := bufio.NewReader(readCloser)
	for {
		bytes, err := r.ReadBytes('\n')
		if _, err := out.Write(bytes); err != nil {
			return err
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}
