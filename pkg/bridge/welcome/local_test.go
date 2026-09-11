package welcome

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestLocalContextSourceReadsStateWithoutCredentials(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	config.SetConfig(nil)
	t.Cleanup(func() { config.SetConfig(nil) })

	appConfig := &config.Config{Email: "dev@example.com", Token: "app-secret"}
	if err := appConfig.Flush(); err != nil {
		t.Fatalf("save app config: %v", err)
	}
	consoleConfig := &console.Config{Url: "https://console.example.com", Token: "console-secret"}
	if err := consoleConfig.Save(); err != nil {
		t.Fatalf("save console config: %v", err)
	}

	root := filepath.Join(t.TempDir(), "workspace")
	nested := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	project := &manifest.ProjectManifest{
		Cluster: "platform-prod", Project: "acme", Provider: "aws", Region: "eu-west-1",
		Owner: &manifest.Owner{Email: "deploy@example.com"},
	}
	if err := project.Write(filepath.Join(root, "workspace.yaml")); err != nil {
		t.Fatalf("write workspace: %v", err)
	}
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(nested); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDirectory) })

	kubeconfig := filepath.Join(home, "kubeconfig")
	kube := clientcmdapi.NewConfig()
	kube.CurrentContext = "plural-platform-prod"
	if err := clientcmd.WriteToFile(*kube, kubeconfig); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}
	t.Setenv("KUBECONFIG", kubeconfig)

	snapshot, err := NewLocalSource("v1").Read(t.Context())
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if snapshot.Version != "v1" || snapshot.App.Email != "dev@example.com" || snapshot.App.Endpoint != "https://app.plural.sh" {
		t.Fatalf("app snapshot = %#v", snapshot.App)
	}
	if snapshot.Console.URL != "https://console.example.com" {
		t.Fatalf("console snapshot = %#v", snapshot.Console)
	}
	if snapshot.Workspace.Name != "platform-prod" || snapshot.Workspace.Owner != "deploy@example.com" {
		t.Fatalf("workspace snapshot = %#v", snapshot.Workspace)
	}
	if snapshot.KubeContext != "plural-platform-prod" {
		t.Fatalf("kube context = %q", snapshot.KubeContext)
	}
}
