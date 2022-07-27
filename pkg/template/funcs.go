package template

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/utils"
)

func fileExists(path string) bool {
	return utils.Exists(path)
}

func pathJoin(parts ...string) string {
	return path.Join(parts...)
}

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

func readPassword(prompt string) (string, error) {
	return utils.ReadPwd(prompt + ": ")
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

func dedupeObj(obj interface{}, path string, val interface{}) interface{} {
	probed, err := probe(obj, path)
	if err != nil {
		return val
	}

	return probed
}

func namespace(name string) string {
	conf := config.Read()
	return conf.Namespace(name)
}

func secret(namespace, name string) map[string]interface{} {
	kube, _ := kubernetes.Kubernetes()
	res := map[string]interface{}{}
	secret, err := kube.Secret(namespace, name)
	if err != nil {
		return res
	}

	for k, v := range secret.Data {
		res[k] = string(v)
	}
	return res
}

func importValue(tool, path string) string {
	return fmt.Sprintf(`"{{ .Import.%s.%s }}"`, tool, path)
}

func toYaml(val interface{}) (string, error) {
	res, err := yaml.Marshal(val)
	return string(res), err
}

func eabCredential(cluster, provider string) (*api.EabCredential, error) {
	client := api.NewClient()
	return client.GetEabCredential(cluster, provider)
}
