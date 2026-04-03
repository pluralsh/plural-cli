package up

import "testing"

func TestGetGitUsername(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "github scp-style",
			url:  "git@github.com:acme-corp/widget-service.git",
			want: "git",
		},
		{
			name: "github ssh transport",
			url:  "ssh://git@github.com/acme-corp/widget-service.git",
			want: "git",
		},
		{
			name: "gitlab nested group",
			url:  "git@gitlab.com:engineering/platform/api-gateway.git",
			want: "git",
		},
		{
			name: "azure devops ssh (v3 path)",
			url:  "my-org@ssh.dev.azure.com:v3/MyOrg/MyProject/MyRepo",
			want: "my-org",
		},
		{
			name: "bitbucket cloud workspace repo",
			url:  "another-org@bitbucket.org:acme-workspace/mobile-app.git",
			want: "another-org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGitUsername(tt.url); got != tt.want {
				t.Errorf("getGitUsername(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}
