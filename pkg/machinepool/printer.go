package machinepool

import (
	"fmt"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"
	"strings"

	tm "github.com/buger/goterm"
	"k8s.io/client-go/kubernetes"
	clusterapiExp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	// corev1 "k8s.io/api/core/v1"
)

func Ready(mp *clusterapiExp.MachinePool) bool {
	phase := findReadiness(mp)
	tm.Printf("MachinePool %s ", mp.Name)

	switch phase {
	case clusterapiExp.MachinePoolPhasePending:
		warn("PENDING")
		tm.Println("")
		return false
	case clusterapiExp.MachinePoolPhaseProvisioning:
		warn("PROVISIONING")
		tm.Println("")
		return false
	case clusterapiExp.MachinePoolPhaseProvisioned:
		warn("PROVISIONED")
		tm.Println("")
		return false
	case clusterapiExp.MachinePoolPhaseRunning:
		success("RUNNING")
		tm.Println("")
		return true
	case clusterapiExp.MachinePoolPhaseDeleting:
		warn("DELETING")
		tm.Println("")
		return false
	case clusterapiExp.MachinePoolPhaseFailed:
		warn("FAILED")
		tm.Println("")
		return false
	case clusterapiExp.MachinePoolPhaseUnknown:
		highlight("UNKNOWN")
		tm.Println("")
		return false
	case clusterapiExp.MachinePoolPhaseScalingUp:
		warn("SCALING UP")
		tm.Println("")
		return false
	case clusterapiExp.MachinePoolPhaseScalingDown:
		warn("SCALING DOWN")
		tm.Println("")
		return false
	case clusterapiExp.MachinePoolPhaseScaling:
		warn("SCALING")
		tm.Println("")
		return false
	}

	tm.Println("")
	return false
}

func Print(client *kubernetes.Clientset, mp *clusterapiExp.MachinePool) (err error) {
	Ready(mp)
	if _, err := tm.Println(mp.ObjectMeta.Name); err != nil {
		return err
	}
	// if _, err := tm.Printf("\nCluster ControlPlane Ready: %v\n", cluster.Status.ControlPlaneReady); err != nil {
	// 	return err
	// }
	// if _, err := tm.Printf("\nCluster Infrastructure Ready: %v\n", cluster.Status.InfrastructureReady); err != nil {
	// 	return err
	// }
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

func findReadiness(mp *clusterapiExp.MachinePool) clusterapiExp.MachinePoolPhase {
	return clusterapiExp.MachinePoolPhase(mp.Status.Phase)
}

func findCondition(mp *clusterapiExp.MachinePool) (condition clusterapi.Condition) {
	for _, cond := range mp.Status.Conditions {
		if cond.Type == clusterapi.ReadyCondition {
			condition = cond
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
