package main

import (
	"path/filepath"
	"io/ioutil"
	"strings"
	"fmt"
	"os"
	"bufio"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v2"
	"github.com/urfave/cli"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/coreos/go-semver/semver"
	"regexp"
)

func utilsCommands() []cli.Command {
	return []cli.Command{
		{
			Name: "image-bump",
			ArgsUsage: "CHART",
			Usage: "Bumps a chart's image tag",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "path",
					Usage: "path to tag in helm values file",
				},
				cli.StringFlag{
					Name: "tag",
					Usage: "the image tag to set to",
				},
			},
			Action: handleImageBump,
		},
	}
}

func handleImageBump(c *cli.Context) error {
	chartPath := c.Args().Get(0)
	path, tag := c.String("path"), c.String("tag")

	chartPath, _ = filepath.Abs(chartPath)

	chart := make(map[string]interface{})
	vals  := make(map[string]interface{})
	
	if err := readHelmYaml(filepath.Join(chartPath, "Chart.yaml"), &chart); err != nil {
		return err
	}

	if err := readHelmYaml(filepath.Join(chartPath, "values.yaml"), &vals); err != nil {
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

	if err := replaceVals(filepath.Join(chartPath, "values.yaml"), tag); err != nil {
		return err
	}

	sv.BumpPatch()
	chart["version"] = sv.String()

	return writeHelmYaml(filepath.Join(chartPath, "Chart.yaml"), chart)
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
			line = fmt.Sprintf(matches[1] + " ## PLRL-REPLACE[%s]", val, matches[1])
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