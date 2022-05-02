package application

import (
	"context"
	"time"

	tm "github.com/buger/goterm"
	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func additionalDetails(client *kubernetes.Clientset, kind, name, namespace string) {
	ctx := context.Background()
	if kind == "statefulset" {
		ss, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return
		}

		podDetails(ctx, client, ss.Spec.Selector, namespace)
	}

	if kind == "deployment" {
		dep, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return
		}

		podDetails(ctx, client, dep.Spec.Selector, namespace)
	}
}

func podDetails(ctx context.Context, client *kubernetes.Clientset, selector *metav1.LabelSelector, namespace string) {
	ls, _ := metav1.LabelSelectorAsSelector(selector)
	listOptions := metav1.ListOptions{LabelSelector: ls.String()}
	pods, err := client.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return
	}

	tm.Println("\nPod Health:")
	table := tablewriter.NewWriter(tm.Screen)
	table.SetHeader([]string{"Pod", "Status", "Created"})
	for _, pod := range pods.Items {
		table.Append([]string{pod.Name, string(pod.Status.Phase), pod.CreationTimestamp.Format(time.UnixDate)})
	}

	table.Render()
}
