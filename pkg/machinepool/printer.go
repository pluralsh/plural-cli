package machinepool

import (
	"fmt"
	"strings"

	tm "github.com/buger/goterm"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiExp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
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
