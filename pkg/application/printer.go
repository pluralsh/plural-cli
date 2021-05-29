package application

import (
	"fmt"
	"sigs.k8s.io/application/api/v1beta1"
	"strings"
	tm "github.com/buger/goterm"
	// corev1 "k8s.io/api/core/v1"
)

func Print(app *v1beta1.Application) (err error) {
	tm.Printf("Application: %s (%s)  ", app.Name, app.Spec.Descriptor.Version)
	cond := findReadiness(app)
	if cond != nil {
		if cond.Status == "True" {
			success("READY")
		} else if cond.Status == "False" {
			warn("WAITING")
		} else if cond.Status == "Unknown" {
			highlight("UNKNOWN")
		}
	}

	tm.Println("")

	tm.Println(app.Spec.Descriptor.Description)
	tm.Printf("\nComponents Ready: %s\n", app.Status.ComponentsReady)
	first := true
	for _, comp := range app.Status.ComponentList.Objects {
		if comp.Status != "Ready" {
			if first {
				tm.Println("\nUnready Components:")
			}
			kind := strings.ToLower(comp.Kind)
			tm.Printf("- %s/%s :: %s\n", kind, comp.Name, comp.Status)
			tm.Printf("\tUse `kubectl describe %s %s -n %s` to investigate\n", kind, comp.Name, app.Namespace)
			first = false
		}
	}

	first = true
	for _, comp := range app.Status.ComponentList.Objects {
		if comp.Status == "Ready" {
			if first {
				tm.Println("\nReady Components:")
			}
			tm.Printf("- %s/%s :: %s\n", strings.ToLower(comp.Kind), comp.Name, comp.Status)
			first = false
		}
	}
	return
}

func Flush() {
	for idx, str := range strings.SplitAfter(tm.Screen.String(), "\n") {
		if idx == tm.Height() - 1 {
			tm.Output.WriteString("...")
			break
		}

		tm.Output.WriteString(str)
	}

	tm.Output.Flush()
	tm.Screen.Reset()
}

func Ready(app *v1beta1.Application) bool {
	cond := findReadiness(app)
	if cond == nil {
		return false
	}
	return cond.Status == "True"
}

func findReadiness(app *v1beta1.Application) (condition *v1beta1.Condition) {
	for _, cond := range app.Status.Conditions {
		if cond.Type == v1beta1.Ready {
			condition = &cond
			return
		}
	}
	return
}

func warn(line string, args... interface{}) {
	tm.Print(tm.Color(fmt.Sprintf(line, args...), tm.YELLOW))
}

func success(line string, args... interface{}) {
	tm.Print(tm.Color(fmt.Sprintf(line, args...), tm.GREEN))
}

func highlight(line string, args... interface{}) {
	tm.Print(tm.Bold(fmt.Sprintf(line, args...)))
}