package edge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	gqlclient "github.com/pluralsh/console/go/client"
)

type fakeConsole struct {
	userID    string
	projectID string
	token     string
	userErr   error
	projErr   error
	tokenErr  error
	userEmail string
	project   string
}

func (f *fakeConsole) GetUser(email string) (*gqlclient.UserFragment, error) {
	f.userEmail = email
	if f.userErr != nil {
		return nil, f.userErr
	}
	return &gqlclient.UserFragment{ID: f.userID}, nil
}

func (f *fakeConsole) GetProject(name string) (*gqlclient.ProjectFragment, error) {
	f.project = name
	if f.projErr != nil {
		return nil, f.projErr
	}
	return &gqlclient.ProjectFragment{ID: f.projectID}, nil
}

func (f *fakeConsole) CreateBootstrapToken(gqlclient.BootstrapTokenAttributes) (string, error) {
	if f.tokenErr != nil {
		return "", f.tokenErr
	}
	return f.token, nil
}

func TestBuildUsesCloudConfigOverrideWithoutConsole(t *testing.T) {
	dir := t.TempDir()
	cloud := filepath.Join(dir, "cloud.yaml")
	if err := os.WriteFile(cloud, []byte("#cloud-config\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var calls []string
	svc := &Service{
		LoadConfig: func(string) (*Configuration, error) {
			return &Configuration{
				Image:           "kairos:img",
				AurorabootImage: "auroraboot:img",
				CraneImage:      "crane:img",
				Bundles:         map[string]string{"k3s": "k3s:img"},
			}, nil
		},
		Exec: func(name string, args ...string) error {
			calls = append(calls, name+" "+strings.Join(args, " "))
			if len(args) >= 2 && args[0] == "run" && strings.Contains(strings.Join(args, " "), "/tmp/build/kairos.img") {
				buildDir := filepath.Join(dir, "image", "build")
				return os.WriteFile(filepath.Join(buildDir, "kairos.img"), []byte("built"), 0644)
			}
			return nil
		},
		Log: func(string) {},
	}

	if err := svc.Build(ImageOptions{
		OutputDir:   "image",
		WorkingDir:  dir,
		CloudConfig: cloud,
		Model:       "rpi5",
	}); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "docker volume create edge-rootfs") {
		t.Fatalf("missing volume create:\n%s", joined)
	}
	if !strings.Contains(joined, "util unpack kairos:img") {
		t.Fatalf("missing unpack:\n%s", joined)
	}
	if !strings.Contains(joined, "--model rpi5") {
		t.Fatalf("missing model:\n%s", joined)
	}
	if !strings.Contains(joined, "docker volume rm edge-rootfs") {
		t.Fatalf("missing volume cleanup:\n%s", joined)
	}
	got, err := os.ReadFile(filepath.Join(dir, "image", "kairos.img"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "built" {
		t.Fatalf("copied image = %q", got)
	}
}

func TestBuildTemplatesCloudConfigFromConsole(t *testing.T) {
	dir := t.TempDir()
	console := &fakeConsole{userID: "user-1", projectID: "proj-1", token: "boot-token"}
	svc := &Service{
		Client: console,
		LoadConfig: func(string) (*Configuration, error) {
			return &Configuration{Image: "kairos:img", AurorabootImage: "auroraboot:img", CraneImage: "crane:img"}, nil
		},
		FetchCloudConfig: func() (string, error) {
			return "url=@URL@ token=@TOKEN@ user=@USERNAME@ pass=@PASSWORD@", nil
		},
		Exec: func(name string, args ...string) error {
			if len(args) >= 2 && args[0] == "run" && strings.Contains(strings.Join(args, " "), "/tmp/build/kairos.img") {
				return os.WriteFile(filepath.Join(dir, "image", "build", "kairos.img"), []byte("built"), 0644)
			}
			return nil
		},
		Log: func(string) {},
	}

	if err := svc.Build(ImageOptions{
		OutputDir:  "image",
		WorkingDir: dir,
		Project:    "default",
		User:       "ops@example.com",
		Username:   "plural",
		Password:   "secret",
		ConsoleURL: "https://console.example.com",
	}); err != nil {
		t.Fatal(err)
	}
	if console.userEmail != "ops@example.com" || console.project != "default" {
		t.Fatalf("console lookup user=%q project=%q", console.userEmail, console.project)
	}
	got, err := os.ReadFile(filepath.Join(dir, "image", "cloud-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	want := "url=https://console.example.com token=boot-token user=plural pass=secret"
	if string(got) != want {
		t.Fatalf("cloud-config = %q", got)
	}
}

func TestBuildRequiresPasswordWithoutCloudConfig(t *testing.T) {
	svc := &Service{
		Client:     &fakeConsole{projectID: "proj-1", token: "t"},
		LoadConfig: func(string) (*Configuration, error) { return &Configuration{Image: "kairos:img"}, nil },
		Log:        func(string) {},
	}
	err := svc.Build(ImageOptions{WorkingDir: t.TempDir(), ConsoleURL: "https://console.example.com"})
	if err == nil || !strings.Contains(err.Error(), "password cannot be empty") {
		t.Fatalf("error = %v", err)
	}
}
