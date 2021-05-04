package template

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/utils"
)

func repoRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	res, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(res))
}

func repoName() string {
	root := repoRoot()
	return path.Base(root)
}

func repoUrl() string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	res, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(res))
}

func branchName() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	res, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(res))
}

func createWebhook(domain string) (api.Webhook, error) {
	client := api.NewClient()
	url := fmt.Sprintf("https://%s/v1/webhook", domain)
	return client.CreateWebhook(url)
}

func dumpConfig() (string, error) {
	conf := config.Read()
	io, err := conf.Marshal()
	return string(io), err
}

func dumpAesKey() (string, error) {
	key, err := crypto.Materialize()
	if err != nil {
		return "", err
	}

	io, err := key.Marshal()
	return string(io), err
}

func readFile(path string) string {
	res, err := utils.ReadFile(path)
	if err != nil {
		return ""
	}
	return res
}

func readLine(prompt string) (string, error) {
	return utils.ReadLine(prompt + ": ")
}

func readLineDefault(prompt string, def string) (string, error) {
	result, err := utils.ReadLine(fmt.Sprintf("%s [%s]: ", prompt, def))
	if result == "" {
		return def, nil
	}

	return result, err
}

func homeDir(parts ...string) (string, error) {
	home, err := os.UserHomeDir()
	return path.Join(home, path.Join(parts...)), err
}

func knownHosts() (string, error) {
	known_hosts, err := homeDir(".ssh", "known_hosts")
	if err != nil {
		return "", err
	}

	res, _ := utils.ReadFile(known_hosts)
	return res, nil
}

func probe(obj interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")
	val := obj
	for _, key := range keys {
		typed := val.(map[string]interface{})
		value, ok := typed[key]
		if !ok {
			return nil, fmt.Errorf("Could not find %s", key)
		}
		val = value
	}
	return val, nil
}

func dedupe(obj interface{}, path string, val string) string {
	probed, err := probe(obj, path)
	if err != nil {
		return val
	}

	return fmt.Sprintf("%s", probed)
}

func importValue(tool, path string) string {
	return fmt.Sprintf("'{{ .Import.%s.%s }}'", tool, path)
}