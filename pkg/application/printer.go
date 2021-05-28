package application

import (
	"fmt"
	"sigs.k8s.io/application/api/v1beta1"
	"github.com/fatih/color"
	// corev1 "k8s.io/api/core/v1"
)

func Print(app *v1beta1.Application) (err error) {
	fmt.Printf("Application: %s (%s)  ", app.Name, app.Spec.Descriptor.Version)
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

	fmt.Println("")

	fmt.Println(app.Spec.Descriptor.Description)
	fmt.Printf("\nComponents Ready: %s\n", app.Status.ComponentsReady)
	for _, comp := range app.Status.ComponentList.Objects {
		if comp.Status != "Ready" {
			fmt.Println("%s/%s :: %s", comp.Kind, comp.Name, comp.Status)
		}
	}
	return
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
	color.New(color.FgYellow, color.Bold).Printf(line, args...)
}

func success(line string, args... interface{}) {
	color.New(color.FgGreen, color.Bold).Printf(line, args...)
}

func highlight(line string, args... interface{}) {
	color.New(color.Bold).Printf(line, args...)
}