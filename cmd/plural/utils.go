package plural

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli"
	"sigs.k8s.io/yaml"
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
			Action: latestVersion(handleImageBump),
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

	return os.WriteFile(path, []byte(content), 0644)
}

func readHelmYaml(path string, result *map[string]interface{}) error {
	content, err := os.ReadFile(path)
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

	return os.WriteFile(path, io, 0644)
}
