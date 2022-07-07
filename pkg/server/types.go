package server

type Workspace struct {
	Cluster      string `json:"cluster"`
	Project      string `json:"project"`
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	BucketPrefix string `json:"bucket_prefix"`
	Subdomain    string `json:"subdomain"`
}

type Aws struct {
	AccessKeyId     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

type Gcp struct {
	ApplicationCredentials string `json:"application_credentials"`
}

type Credentials struct {
	Aws *Aws `json:"aws"`
	Gcp *Gcp `json:"gcp"`
}

type User struct {
	GitUser     string `json:"gitUser"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
}

type GitInfo struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type SetupRequest struct {
	Workspace     *Workspace   `json:"workspace"`
	Credentials   *Credentials `json:"credentials"`
	User          *User        `json:"user"`
	Provider      string       `json:"provider"`
	AesKey        string       `json:"aes_key"`
	GitUrl        string       `json:"git_url"`
	GitInfo       *GitInfo     `json:"git_info"`
	SshPublicKey  string       `json:"ssh_public_key"`
	SshPrivateKey string       `json:"ssh_private_key"`
	IsDemo        bool         `json:"is_demo"`
}
