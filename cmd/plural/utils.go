package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func utilsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "image-bump",
			ArgsUsage: "CHART",
			Usage:     "Bumps a chart's image tag",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path",
					Usage: "path to tag in helm values file",
				},
				cli.StringFlag{
					Name:  "tag",
					Usage: "the image tag to set to",
				},
			},
			Action: handleImageBump,
		},
		{
			Name:      "helm-process",
			ArgsUsage: "PATH",
			Usage:     "Process a helm chart for use in a Plural artifact",
			// Flags: []cli.Flag{
			// 	cli.StringFlag{
			// 		Name:  "path",
			// 		Usage: "path to tag in helm values file",
			// 	},
			// 	cli.StringFlag{
			// 		Name:  "tag",
			// 		Usage: "the image tag to set to",
			// 	},
			// },
			Action: handleHelmProcess,
		},
	}
}

func handleImageBump(c *cli.Context) error {
	chartPath := c.Args().Get(0)
	path, tag := c.String("path"), c.String("tag")

	chartPath, _ = filepath.Abs(chartPath)

	chart := make(map[string]interface{})
	vals := make(map[string]interface{})

	if err := readHelmYaml(pathing.SanitizeFilepath(filepath.Join(chartPath, "Chart.yaml")), &chart); err != nil {
		return err
	}

	if err := readHelmYaml(pathing.SanitizeFilepath(filepath.Join(chartPath, "values.yaml")), &vals); err != nil {
		return err
	}

	currentTag := funk.Get(vals, path)
	if currentTag == tag {
		utils.Highlight("No change in version tag\n")
		return nil
	}

	currentVersion := (chart["version"]).(string)
	sv, err := semver.NewVersion(currentVersion)
	if err != nil {
		return err
	}

	if err := replaceVals(pathing.SanitizeFilepath(filepath.Join(chartPath, "values.yaml")), tag); err != nil {
		return err
	}

	sv.BumpPatch()
	chart["version"] = sv.String()

	return writeHelmYaml(pathing.SanitizeFilepath(filepath.Join(chartPath, "Chart.yaml")), chart)
}

const replPattern = `(?s).*## PLRL-REPLACE\[(.*)\].*`

func replaceVals(path string, val string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	result := make([]string, 0)
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	re := regexp.MustCompile(replPattern)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) > 0 {
			line = fmt.Sprintf(matches[1]+" ## PLRL-REPLACE[%s]", val, matches[1])
		}

		result = append(result, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	content := strings.Join(result, "\n")

	return ioutil.WriteFile(path, []byte(content), 0644)
}

func readHelmYaml(path string, result *map[string]interface{}) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(content, result)
}

func writeHelmYaml(path string, vals map[string]interface{}) error {
	io, err := yaml.Marshal(&vals)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, io, 0644)
}

func handleHelmProcess(c *cli.Context) error {
	// Download chart dependencies
	// if err := updateDeps(c); err != nil {
	// 	utils.HighlightError(err)
	// }

	// Load chart and get CRDs from them and write them to our chart
	loadedChart, err := loadHelmChart(c)
	if err != nil {
		utils.HighlightError(err)
		return err
	}

	path := c.Args().Get(0)
	if path == "" {
		path = "."
	}

	//TODO: cleanup this path stuff
	absPath, _ := filepath.Abs(path)

	chartPath := pathing.SanitizeFilepath(absPath)

	// Get all CRDs from dependent charts and put them in the CRDs folder of the artifact
	chartDeps := loadedChart.Dependencies()

	var joinedErrors []string

	regKind := regexp.MustCompile("kind: (Deployment|StatefulSet)")
	// regApiVersion := regexp.MustCompile("apiVersion: .*")
	regName := regexp.MustCompile("name: .*")
	regDot := regexp.MustCompile(" . ")

	for _, dep := range chartDeps {
		for _, crd := range dep.CRDObjects() {
			splitFileName := strings.Split(crd.File.Name, "/")
			splitFileName[len(splitFileName)-1] = dep.Name() + "_" + splitFileName[len(splitFileName)-1]
			newFileName := strings.Join(splitFileName, "/")
			if err := ioutil.WriteFile(chartPath+"/"+newFileName, crd.File.Data, 0644); err != nil {
				utils.HighlightError(err)
				joinedErrors = append(joinedErrors, err.Error())
			}
		}

		for _, template := range dep.Templates {
			// utils.Highlight(fmt.Sprintf("Template Name: %s\n", template.Name))
			// utils.Highlight(fmt.Sprintf("File is of wanted type: %v\n\n", reg.Match(template.Data)))
			if regKind.Match(template.Data) {
				// name := strings.Replace(string(regName.Find(template.Data)), "name: ", "", 1)
				name := regName.Find(template.Data)

				// utils.Highlight(name)

				replaced := regDot.ReplaceAll(name, []byte(" .Subcharts.test "))
				utils.Highlight(string(replaced))
			}
		}
	}

	for _, template := range loadedChart.Templates {
		// utils.Highlight(fmt.Sprintf("Template Name: %s\n", template.Name))
		// utils.Highlight(fmt.Sprintf("File is of wanted type: %v\n\n", reg.Match(template.Data)))
		if regKind.Match(template.Data) {
			utils.Highlight(string(regName.Find(template.Data)))
		}
	}

	var outputError error
	if len(joinedErrors) > 0 {
		outputError = fmt.Errorf(strings.Join(joinedErrors, " - "))
	}

	return outputError
}

// Load a helm chart from a path
func loadHelmChart(c *cli.Context) (*chart.Chart, error) {
	path := c.Args().Get(0)
	if path == "" {
		path = "."
	}

	loadedChart, err := loader.Load(path)
	if err != nil {
		utils.HighlightError(err)
		return nil, err
	}
	return loadedChart, nil
}
