package cluster

import (
	"fmt"
	"strings"

	tm "github.com/buger/goterm"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"
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
