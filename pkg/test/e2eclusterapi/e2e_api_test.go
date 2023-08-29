//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/machinepool"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/polly/containers"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiListInstallations(t *testing.T) {

	cmd := exec.Command("plural", "api", "list", "installations")
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	installations := make([]string, 0)
	rows := strings.Split(string(cmdOutput[:]), "\n")
	for _, row := range rows[1:] { // Skip the heading row and iterate through the remaining rows
		row = strings.ReplaceAll(row, "|", "")
		cols := strings.Fields(row) // Extract each column from the row.
		if len(cols) == 3 {
			installations = append(installations, cols[0])
		}
	}
	expected := []string{"bootstrap"}
	expects := containers.ToSet(expected)
	assert.True(t, expects.Equal(containers.ToSet(installations)), fmt.Sprintf("the expected %s is different then %s", expected, installations))
}

func TestPackagesList(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)

	testDir := path.Join(homeDir, "test")

	err = os.Chdir(testDir)
	assert.NoError(t, err)
	cmd := exec.Command("plural", "packages", "list", "bootstrap")
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	packages := make([]string, 0)
	rows := strings.Split(string(cmdOutput[:]), "\n")
	for _, row := range rows[3:] { // Skip the heading row and iterate through the remaining rows
		row = strings.ReplaceAll(row, "|", "")
		cols := strings.Fields(row) // Extract each column from the row.
		if len(cols) == 3 {
			packages = append(packages, cols[1])
		}
	}
	expected := []string{"bootstrap", "kind-bootstrap-cluster-api", "cluster-api-control-plane", "cluster-api-bootstrap", "plural-certmanager-webhook", "cluster-api-cluster", "cluster-api-core", "cluster-api-provider-docker"}
	expects := containers.ToSet(expected)
	assert.True(t, expects.Equal(containers.ToSet(packages)), fmt.Sprintf("the expected %s is different then %s", expected, packages))
}

func TestUpdateNodePools(t *testing.T) {
	cmd := exec.Command("plural", "ops", "cluster")
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	nodes := make([]string, 0)
	rows := strings.Split(string(cmdOutput[:]), "\n")
	for _, row := range rows[3:] { // Skip the heading row and iterate through the remaining rows
		row = strings.ReplaceAll(row, "|", "")
		cols := strings.Fields(row) // Extract each column from the row.
		if len(cols) == 3 {
			nodes = append(nodes, cols[0])
		}
	}

	assert.Equal(t, 3, len(nodes), fmt.Sprintf("expected %d nodes got %d", 3, len(nodes)))
	kubeConf, err := kubernetes.KubeConfig()
	if err != nil {
		t.Fatal(err)
	}

	mpools, err := machinepool.ListAll(kubeConf)
	if err != nil {
		t.Fatal(err)
	}
	if len(mpools) != 1 {
		t.Fatal("expected one machine pool")
	}
	mp := mpools[0]

	mps, err := machinepool.NewForConfig(kubeConf)
	if err != nil {
		t.Fatal(err)
	}

	client := mps.MachinePools("bootstrap")
	machinePool, err := client.Get(context.Background(), mp.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	replicas := *machinePool.Spec.Replicas

	if replicas != 2 {
		t.Fatal("expected 2 replicas")
	}
	replicas = 3
	machinePool.Spec.Replicas = &replicas

	machinePool, err = client.Update(context.Background(), machinePool)
	if err != nil {
		t.Fatal(err)
	}

	kube, err := kubernetes.Kubernetes()
	if err != nil {
		t.Fatal(err)
	}

	if err := utils.WaitFor(5*time.Minute, 5*time.Second, func() (bool, error) {
		nodeList, err := kube.Nodes()
		if err != nil {
			t.Fatal(err)
		}
		if len(nodeList.Items) == 4 {
			return true, nil
		}
		return false, nil
	}); err != nil {
		t.Fatal(err)
	}

	machinePool, err = client.Get(context.Background(), mp.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	replicas = 2
	machinePool.Spec.Replicas = &replicas
	machinePool, err = client.Update(context.Background(), machinePool)
	if err != nil {
		t.Fatal(err)
	}
	if err := utils.WaitFor(5*time.Minute, 5*time.Second, func() (bool, error) {
		nodeList, err := kube.Nodes()
		if err != nil {
			t.Fatal(err)
		}
		if len(nodeList.Items) == 3 {
			return true, nil
		}
		return false, nil
	}); err != nil {
		t.Fatal(err)
	}

	machinePool, err = client.Get(context.Background(), mp.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	version := "v1.23.17"
	machinePool.Spec.Template.Spec.Version = &version
	machinePool, err = client.Update(context.Background(), machinePool)
	if err != nil {
		t.Fatal(err)
	}
	if err := utils.WaitFor(5*time.Minute, 5*time.Second, func() (bool, error) {
		nodeList, err := kube.Nodes()
		if err != nil {
			t.Fatal(err)
		}
		for _, node := range nodeList.Items {
			if node.Status.NodeInfo.KubeletVersion == version {
				return true, nil
			}
		}

		return false, nil
	}); err != nil {
		t.Fatal(err)
	}
}
