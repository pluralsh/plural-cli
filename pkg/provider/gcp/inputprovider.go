package gcp

// InputProvider partially implements the Provider interface.
// It only contains methods where values need to be read from the user.
// This is to allow easily swapping out the survey library.
type InputProvider interface {
	Cluster() string
	Project() string
	Region() string
}

type readonlyInputProvider struct {
	cluster string
	project string
	region  string
}

func (in *readonlyInputProvider) Cluster() string {
	return in.cluster
}

func (in *readonlyInputProvider) Project() string {
	return in.project
}

func (in *readonlyInputProvider) Region() string {
	return in.region
}

func NewReadonlyInputProvider(cluster, project, region string) InputProvider {
	return &readonlyInputProvider{cluster, project, region}
}
