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
		// HTTPS URLs – the embedded user from the URL should be returned
		{
			name: "github https with oauth2 user",
			url:  "https://oauth2@github.com/acme-corp/widget-service.git",
			want: "oauth2",
		},
		{
			name: "github https without embedded user falls back to empty string",
			url:  "https://github.com/acme-corp/widget-service.git",
			want: "",
		},
		{
			name: "gitlab https with x-token-auth user",
			url:  "https://x-token-auth@gitlab.com/engineering/api-gateway.git",
			want: "x-token-auth",
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

func TestIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		repoUrl string
		want    string
	}{
		{
			name:    "empty url",
			repoUrl: "",
			want:    "",
		},
		// HTTPS URLs
		{
			name:    "github https",
			repoUrl: "https://github.com/acme-corp/widget-service.git",
			want:    "acme-corp/widget-service",
		},
		{
			name:    "github https without .git suffix",
			repoUrl: "https://github.com/acme-corp/widget-service",
			want:    "acme-corp/widget-service",
		},
		{
			name:    "gitlab https nested group",
			repoUrl: "https://gitlab.com/engineering/platform/api-gateway.git",
			want:    "engineering/platform/api-gateway",
		},
		{
			name:    "http (non-TLS) url",
			repoUrl: "http://git.internal.example.com/myorg/myrepo.git",
			want:    "myorg/myrepo",
		},
		{
			name:    "github https with embedded credentials",
			repoUrl: "https://oauth2:token@github.com/acme-corp/widget-service.git",
			want:    "acme-corp/widget-service",
		},
		// SSH / SCP-style URLs
		{
			name:    "github scp-style",
			repoUrl: "git@github.com:acme-corp/widget-service.git",
			want:    "acme-corp/widget-service",
		},
		{
			name:    "gitlab nested group scp-style",
			repoUrl: "git@gitlab.com:engineering/platform/api-gateway.git",
			want:    "engineering/platform/api-gateway",
		},
		{
			name:    "scp without .git suffix",
			repoUrl: "git@github.com:acme-corp/widget-service",
			want:    "acme-corp/widget-service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{RepoUrl: tt.repoUrl}
			if got := ctx.identifier(); got != tt.want {
				t.Errorf("identifier() for url %q = %q, want %q", tt.repoUrl, got, tt.want)
			}
		})
	}
}

