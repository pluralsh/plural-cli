package cluster

import (
	"fmt"
	"strings"

	tm "github.com/buger/goterm"
	"k8s.io/client-go/kubernetes"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"
	// corev1 "k8s.io/api/core/v1"
)

func Ready(cluster *clusterapi.Cluster) bool {
	cond := findReadiness(cluster)
	tm.Printf("Cluster %s ", cluster.Name)
	if cond == nil {
		warn("WAITING")
		tm.Println("")
		return false
	}

	if cond.Status == "True" {
		success("READY")
		tm.Println("")
		return true
	}

	if cond.Status == "False" {
		warn("WAITING")
	} else if cond.Status == "Unknown" {
		highlight("UNKNOWN")
	}

	tm.Println("")
	return false
}

func Print(client *kubernetes.Clientset, cluster *clusterapi.Cluster) (err error) {
	Ready(cluster)
	if _, err := tm.Println(cluster.ObjectMeta.Name); err != nil {
		return err
	}
	if _, err := tm.Printf("\nCluster ControlPlane Ready: %v\n", cluster.Status.ControlPlaneReady); err != nil {
		return err
	}
	if _, err := tm.Printf("\nCluster Infrastructure Ready: %v\n", cluster.Status.InfrastructureReady); err != nil {
		return err
	}
	// first := true
	// for _, comp := range cluster.Status.ComponentList.Objects {
	// 	if comp.Status != "Ready" {
	// 		if first {
	// 			if _, err := tm.Println("\nUnready Components:"); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		kind := strings.ToLower(comp.Kind)
	// 		if _, err := tm.Printf("- %s/%s :: %s\n", kind, comp.Name, comp.Status); err != nil {
	// 			return err
	// 		}
	// 		additionalDetails(client, kind, comp.Name, app.Namespace)
	// 		if _, err := tm.Printf("\tUse `kubectl describe %s %s -n %s` to investigate\n", kind, comp.Name, app.Namespace); err != nil {
	// 			return err
	// 		}
	// 		first = false
	// 	}
	// }

	// first = true
	// for _, comp := range app.Status.ComponentList.Objects {
	// 	if comp.Status == "Ready" {
	// 		if first {
	// 			if _, err := tm.Println("\nReady Components:"); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		if _, err := tm.Printf("- %s/%s :: %s\n", strings.ToLower(comp.Kind), comp.Name, comp.Status); err != nil {
	// 			return err
	// 		}
	// 		first = false
	// 	}
	// }
	return
}

func Flush() {
	for idx, str := range strings.SplitAfter(tm.Screen.String(), "\n") {
		if idx == tm.Height()-1 {
			_, err := tm.Output.WriteString("...")
			if err != nil {
				return
			}
			break
		}

		_, err := tm.Output.WriteString(str)
		if err != nil {
			return
		}
	}

	if err := tm.Output.Flush(); err != nil {
		return
	}
	tm.Screen.Reset()
}

func findReadiness(cluster *clusterapi.Cluster) (condition *clusterapi.Condition) {
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == clusterapi.ReadyCondition {
			condition = &cond
			return
		}
	}
	return
}

func warn(line string, args ...interface{}) {
	if _, err := tm.Print(tm.Color(fmt.Sprintf(line, args...), tm.YELLOW)); err != nil {
		return
	}
}

func success(line string, args ...interface{}) {
	if _, err := tm.Print(tm.Color(fmt.Sprintf(line, args...), tm.GREEN)); err != nil {
		return
	}
}

func highlight(line string, args ...interface{}) {
	if _, err := tm.Print(tm.Bold(fmt.Sprintf(line, args...))); err != nil {
		return
	}
}
