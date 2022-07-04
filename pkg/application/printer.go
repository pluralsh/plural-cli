package application

import (
	"fmt"
	"strings"

	tm "github.com/buger/goterm"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/application/api/v1beta1"
	// corev1 "k8s.io/api/core/v1"
)

func Ready(app *v1beta1.Application) bool {
	cond := findReadiness(app)
	if _, err := tm.Printf("Application %s (%s) ", app.Name, app.Spec.Descriptor.Version); err != nil {
		return false
	}
	if cond == nil {
		warn("WAITING")
		_, err := tm.Println("")
		return err == nil
	}

	if cond.Status == "True" {
		success("READY")
		_, err := tm.Println("")
		return err == nil
	}

	if cond.Status == "False" {
		warn("WAITING")
	} else if cond.Status == "Unknown" {
		highlight("UNKNOWN")
	}

	if _, err := tm.Println(""); err != nil {
		return false
	}
	return false
}

func Print(client *kubernetes.Clientset, app *v1beta1.Application) (err error) {
	Ready(app)
	if _, err := tm.Println(app.Spec.Descriptor.Description); err != nil {
		return err
	}
	if _, err := tm.Printf("\nComponents Ready: %s\n", app.Status.ComponentsReady); err != nil {
		return err
	}
	first := true
	for _, comp := range app.Status.ComponentList.Objects {
		if comp.Status != "Ready" {
			if first {
				if _, err := tm.Println("\nUnready Components:"); err != nil {
					return err
				}
			}
			kind := strings.ToLower(comp.Kind)
			if _, err := tm.Printf("- %s/%s :: %s\n", kind, comp.Name, comp.Status); err != nil {
				return err
			}
			additionalDetails(client, kind, comp.Name, app.Namespace)
			if _, err := tm.Printf("\tUse `kubectl describe %s %s -n %s` to investigate\n", kind, comp.Name, app.Namespace); err != nil {
				return err
			}
			first = false
		}
	}

	first = true
	for _, comp := range app.Status.ComponentList.Objects {
		if comp.Status == "Ready" {
			if first {
				if _, err := tm.Println("\nReady Components:"); err != nil {
					return err
				}
			}
			if _, err := tm.Printf("- %s/%s :: %s\n", strings.ToLower(comp.Kind), comp.Name, comp.Status); err != nil {
				return err
			}
			first = false
		}
	}
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

	err := tm.Output.Flush()
	if err != nil {
		return
	}
	tm.Screen.Reset()
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

func warn(line string, args ...interface{}) {
	_, err := tm.Print(tm.Color(fmt.Sprintf(line, args...), tm.YELLOW))
	if err != nil {
		return
	}
}

func success(line string, args ...interface{}) {
	_, err := tm.Print(tm.Color(fmt.Sprintf(line, args...), tm.GREEN))
	if err != nil {
		return
	}
}

func highlight(line string, args ...interface{}) {
	_, err := tm.Print(tm.Bold(fmt.Sprintf(line, args...)))
	if err != nil {
		return
	}
}
