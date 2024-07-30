package console

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	consoleclient "github.com/pluralsh/console/go/client"
	"k8s.io/cli-runtime/pkg/printers"
)

// Each level has 2 spaces for PrefixWriter
const (
	LEVEL_0 = iota
	LEVEL_1
	LEVEL_2
	LEVEL_3
	LEVEL_4
)

func DescribeServiceContext(sc *consoleclient.ServiceContextFragment) (string, error) {
	return tabbedString(func(out io.Writer) error {
		w := NewPrefixWriter(out)
		w.Write(LEVEL_0, "Id:\t%s\n", sc.ID)
		w.Write(LEVEL_0, "Configuration:\n  \tName\tContext\n")
		w.Write(LEVEL_1, "\t----\t------\n")
		for name, value := range sc.Configuration {
			configurationJson, _ := json.Marshal(value)
			w.Write(LEVEL_1, "\t%v \t%v\n", name, string(configurationJson))
		}
		return nil
	})
}

func DescribeCluster(cluster *consoleclient.ClusterFragment) (string, error) {
	return tabbedString(func(out io.Writer) error {
		w := NewPrefixWriter(out)
		w.Write(LEVEL_0, "Id:\t%s\n", cluster.ID)
		w.Write(LEVEL_0, "Name:\t%s\n", cluster.Name)
		if cluster.Handle != nil {
			w.Write(LEVEL_0, "Handle:\t@%s\n", *cluster.Handle)
		}
		if cluster.Version != nil {
			w.Write(LEVEL_0, "Version:\t%s\n", *cluster.Version)
		}
		if cluster.CurrentVersion != nil {
			w.Write(LEVEL_0, "Current Version:\t%s\n", *cluster.CurrentVersion)
		}
		if cluster.PingedAt != nil {
			w.Write(LEVEL_0, "Pinged At:\t%s\n", *cluster.PingedAt)
		}
		if cluster.Self != nil {
			w.Write(LEVEL_0, "Self:\t%v\n", *cluster.Self)
		}

		if cluster.Provider != nil {
			w.Write(LEVEL_0, "Provider:\n")
			w.Write(LEVEL_1, "Id:\t%s\n", cluster.Provider.ID)
			w.Write(LEVEL_1, "Name:\t%s\n", cluster.Provider.Name)
			w.Write(LEVEL_1, "Namespace:\t%s\n", cluster.Provider.Namespace)
			w.Write(LEVEL_1, "Editable:\t%v\n", *cluster.Provider.Editable)
			w.Write(LEVEL_1, "Cloud:\t%v\n", cluster.Provider.Cloud)
			if cluster.Provider.Repository != nil {
				w.Write(LEVEL_1, "Git:\n")
				w.Write(LEVEL_2, "Id:\t%s\n", cluster.Provider.Repository.ID)
				w.Write(LEVEL_2, "Url:\t%s\n", cluster.Provider.Repository.URL)
				if cluster.Provider.Repository.AuthMethod != nil {
					w.Write(LEVEL_2, "Auth Method:\t%v\n", *cluster.Provider.Repository.AuthMethod)
				}
				if cluster.Provider.Repository.Health != nil {
					w.Write(LEVEL_2, "Health:\t%v\n", *cluster.Provider.Repository.Health)
				}
				if cluster.Provider.Repository.Error != nil {
					w.Write(LEVEL_2, "Error:\t%s\n", *cluster.Provider.Repository.Error)
				}
			}
		}

		return nil
	})
}

func DescribeService(service *consoleclient.ServiceDeploymentExtended) (string, error) {
	return tabbedString(func(out io.Writer) error {
		w := NewPrefixWriter(out)
		w.Write(LEVEL_0, "Id:\t%s\n", service.ID)
		w.Write(LEVEL_0, "Name:\t%s\n", service.Name)
		w.Write(LEVEL_0, "Namespace:\t%s\n", service.Namespace)
		w.Write(LEVEL_0, "Version:\t%s\n", service.Version)
		if service.Tarball != nil {
			w.Write(LEVEL_0, "Tarball:\t%s\n", *service.Tarball)
		} else {
			w.Write(LEVEL_0, "Tarball:\t%s\n", "<none>")
		}
		if service.DeletedAt != nil {
			w.Write(LEVEL_0, "Status:\tTerminating (lasts %s)\n", *service.DeletedAt)
		}
		dryRun := false
		if service.DryRun != nil {
			dryRun = *service.DryRun
		}
		templated := true
		if service.Templated != nil {
			templated = *service.Templated
		}
		w.Write(LEVEL_0, "Dry run:\t%v\n", dryRun)
		w.Write(LEVEL_0, "Templated:\t%v\n", templated)
		w.Write(LEVEL_0, "Git:\t\n")
		w.Write(LEVEL_1, "Ref:\t%s\n", service.Git.Ref)
		w.Write(LEVEL_1, "Folder:\t%s\n", service.Git.Folder)
		if service.Revision != nil {
			w.Write(LEVEL_1, "Revision:\t\n")
			w.Write(LEVEL_2, "Id:\t%s\n", service.Revision.ID)

		}
		if service.Kustomize != nil {
			w.Write(LEVEL_0, "Kustomize:\t\n")
			w.Write(LEVEL_1, "Path:\t%s\n", service.Kustomize.Path)
		}
		if service.Repository != nil {
			w.Write(LEVEL_0, "Repository:\t\n")
			w.Write(LEVEL_1, "Id:\t%s\n", service.Repository.ID)
			w.Write(LEVEL_1, "Url:\t%s\n", service.Repository.URL)
			if service.Repository.AuthMethod != nil {
				w.Write(LEVEL_1, "AuthMethod:\t%s\n", *service.Repository.AuthMethod)
			}
			w.Write(LEVEL_1, "Status:\t\n")
			if service.Repository.Health != nil {
				w.Write(LEVEL_2, "Health:\t%s\n", *service.Repository.Health)
			}
			if service.Repository.Error != nil {
				w.Write(LEVEL_2, "Error:\t%s\n", *service.Repository.Error)
			}
			configMap := map[string]string{}
			for _, conf := range service.Configuration {
				configMap[conf.Name] = conf.Value
			}
			printConfigMultiline(w, "Configuration", configMap)
			if len(service.Components) > 0 {
				w.Write(LEVEL_0, "Components:\n  Id\tName\tNamespace\tKind\tVersion\tState\tSynced\n")
				w.Write(LEVEL_1, "----\t------\t------\t------\t------\t------\t------\n")
				for _, c := range service.Components {
					namespace := "-"
					version := "-"
					state := "-"
					if c.Namespace != nil {
						namespace = *c.Namespace
					}
					if c.Version != nil {
						version = *c.Version
					}
					if c.State != nil {
						state = string(*c.State)
					}

					w.Write(LEVEL_1, "%v \t%v\t%v\t%v\t%v\t%v\t%v\n", c.ID, c.Name, namespace, c.Kind, version, state, c.Synced)
				}
			} else {
				w.Write(LEVEL_0, "Components: %s\n", "<none>")
			}
			if len(service.Errors) > 0 {
				w.Write(LEVEL_0, "Errors:\n  Source\tMessage\n")
				w.Write(LEVEL_1, "----\t------\n")
				for _, c := range service.Errors {

					w.Write(LEVEL_1, "%v \t%v\n", c.Source, c.Message)
				}
			} else {
				w.Write(LEVEL_0, "Errors: %s\n", "<none>")
			}

		}
		return nil
	})
}

var maxConfigLen = 140

func printConfigMultiline(w PrefixWriter, title string, configurations map[string]string) {
	w.Write(LEVEL_0, "%s:\t", title)

	// to print labels in the sorted order
	keys := make([]string, 0, len(configurations))
	for key := range configurations {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		w.WriteLine("<none>")
		return
	}
	sort.Strings(keys)
	indent := "\t"
	for i, key := range keys {
		if i != 0 {
			w.Write(LEVEL_0, indent)
		}
		value := strings.TrimSuffix(configurations[key], "\n")
		if (len(value)+len(key)+2) > maxConfigLen || strings.Contains(value, "\n") {
			w.Write(LEVEL_0, "%s:\n", key)
			for _, s := range strings.Split(value, "\n") {
				w.Write(LEVEL_0, "%s  %s\n", indent, shorten(s, maxConfigLen-2))
			}
		} else {
			w.Write(LEVEL_0, "%s: %s\n", key, value)
		}
	}
}

func shorten(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength] + "..."
	}
	return s
}

func tabbedString(f func(io.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 2, ' ', 0)

	err := f(out)
	if err != nil {
		return "", err
	}

	err = out.Flush()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

type flusher interface {
	Flush()
}

// PrefixWriter can write text at various indentation levels.
type PrefixWriter interface {
	// Write writes text with the specified indentation level.
	Write(level int, format string, a ...interface{})
	// WriteLine writes an entire line with no indentation level.
	WriteLine(a ...interface{})
	// Flush forces indentation to be reset.
	Flush()
}

// prefixWriter implements PrefixWriter
type prefixWriter struct {
	out io.Writer
}

var _ PrefixWriter = &prefixWriter{}

// NewPrefixWriter creates a new PrefixWriter.
func NewPrefixWriter(out io.Writer) PrefixWriter {
	return &prefixWriter{out: out}
}

func (pw *prefixWriter) Write(level int, format string, a ...interface{}) {
	levelSpace := "  "
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += levelSpace
	}
	output := fmt.Sprintf(prefix+format, a...)
	err := printers.WriteEscaped(pw.out, output)
	if err != nil {
		return
	}
}

func (pw *prefixWriter) WriteLine(a ...interface{}) {
	output := fmt.Sprintln(a...)
	err := printers.WriteEscaped(pw.out, output)
	if err != nil {
		return
	}
}

func (pw *prefixWriter) Flush() {
	if f, ok := pw.out.(flusher); ok {
		f.Flush()
	}
}
