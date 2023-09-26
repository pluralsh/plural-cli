package console

type AuthMethod string

const (
	AuthMethodBasic AuthMethod = "BASIC"
	AuthMethodSSH   AuthMethod = "SSH"
)

type GitHealth string

const (
	GitHealthPullable GitHealth = "PULLABLE"
	GitHealthFailed   GitHealth = "FAILED"
)

type ComponentState string

const (
	ComponentStateRunning ComponentState = "RUNNING"
	ComponentStatePending ComponentState = "PENDING"
	ComponentStateFailed  ComponentState = "FAILED"
	ComponentStateUnknown ComponentState = "UNKNOWN"
)

type Cluster struct {
	Id             string
	Name           string
	Version        string
	CurrentVersion string
	Provider       *ClusterProvider
	NodePools      []NodePool
}

type ClusterProvider struct {
	Id         string
	Name       string
	Namespace  string
	Cloud      string
	Editable   bool
	Repository *GitRepository
	Service    *ServiceDeployment
}

type NodePool struct {
	Id           string
	Name         string
	MinSize      int64
	MaxSize      int64
	InstanceType string
}

type GitRepository struct {
	Id         string
	Editable   bool
	Health     GitHealth
	AuthMethod AuthMethod
	URL        string
}

type ServiceDeployment struct {
	Id         string
	Name       string
	Namespace  string
	Version    string
	Editable   bool
	DeletedAt  *string
	Components []Component
	Git        GitRef
	Repository *GitRepository
	Sha        string
	Tarball    string
}

type Component struct {
	Id        string
	Name      string
	Group     string
	Kind      string
	Namespace string
	State     ComponentState
	Synced    bool
	Version   string
}

type GitRef struct {
	Folder string
	Ref    string
}
