package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

// CreateInput is the credential-free create payload.
type CreateInput struct {
	ClusterID string
	Name      string
	Namespace string
	RepoID    string
	GitRef    string
	GitFolder string
	Kustomize string
	Version   string
	DryRun    bool
}

// UpdateInput is the credential-free update payload.
type UpdateInput struct {
	ID        string
	GitRef    string
	GitFolder string
	Kustomize string
	Version   string
	DryRun    *bool
}

// CloneInput is the credential-free clone payload.
type CloneInput struct {
	SourceID      string
	DestClusterID string
	Name          string
	Namespace     string
}

// Loader is the narrow contract consumed by the Services screen.
type Loader interface {
	ListClusters(ctx context.Context, query string) ([]Cluster, error)
	List(ctx context.Context, clusterID string, after *string, query string) (Page, error)
	Get(ctx context.Context, id string) (Detail, error)
	Kick(ctx context.Context, id string) (Detail, error)
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, input CreateInput) (Detail, error)
	Update(ctx context.Context, input UpdateInput) (Detail, error)
	Clone(ctx context.Context, input CloneInput) (Detail, error)
	DownloadTarball(ctx context.Context, id, dir string) (string, error)
}

// API is the Console surface required by this package.
type API interface {
	ListClusters() (*gqlclient.ListClusters, error)
	ListClusterServices(clusterId, handle *string) ([]*gqlclient.ServiceDeploymentEdgeFragment, error)
	GetClusterService(serviceId, serviceName, clusterName *string) (*gqlclient.ServiceDeploymentExtended, error)
	KickClusterService(serviceId, serviceName, clusterName *string) (*gqlclient.ServiceDeploymentExtended, error)
	DeleteClusterService(serviceId string) (*gqlclient.DeleteServiceDeployment, error)
	CreateClusterService(clusterId, clusterName *string, attr gqlclient.ServiceDeploymentAttributes) (*gqlclient.ServiceDeploymentExtended, error)
	UpdateClusterService(serviceId, serviceName, clusterName *string, attributes gqlclient.ServiceUpdateAttributes) (*gqlclient.ServiceDeploymentExtended, error)
	CloneService(clusterId string, serviceId, serviceName, clusterName *string, attributes gqlclient.ServiceCloneAttributes) (*gqlclient.ServiceDeploymentFragment, error)
	GetDeployToken(clusterId, clusterName *string) (string, error)
}

func (s *Service) Kick(ctx context.Context, id string) (Detail, error) {
	if err := ctx.Err(); err != nil {
		return Detail{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingID}
	}
	client, err := s.client(ctx)
	if err != nil {
		return Detail{}, err
	}
	service, err := client.KickClusterService(&id, nil, nil)
	if err != nil {
		return Detail{}, err
	}
	if service == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingService}
	}
	return detailFromExtended(service), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingID}
	}
	client, err := s.client(ctx)
	if err != nil {
		return err
	}
	_, err = client.DeleteClusterService(id)
	return err
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Detail, error) {
	if err := ctx.Err(); err != nil {
		return Detail{}, err
	}
	input.ClusterID = strings.TrimSpace(input.ClusterID)
	input.Name = strings.TrimSpace(input.Name)
	input.RepoID = strings.TrimSpace(input.RepoID)
	input.GitRef = strings.TrimSpace(input.GitRef)
	input.GitFolder = strings.TrimSpace(input.GitFolder)
	if input.ClusterID == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingCluster}
	}
	if input.Name == "" || input.RepoID == "" || input.GitRef == "" || input.GitFolder == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingCreateFields}
	}
	if input.Namespace == "" {
		input.Namespace = "default"
	}
	if input.Version == "" {
		input.Version = "0.0.1"
	}
	client, err := s.client(ctx)
	if err != nil {
		return Detail{}, err
	}
	attrs := gqlclient.ServiceDeploymentAttributes{
		Name:         input.Name,
		Namespace:    input.Namespace,
		Version:      lo.ToPtr(input.Version),
		RepositoryID: lo.ToPtr(input.RepoID),
		Git:          &gqlclient.GitRefAttributes{Ref: input.GitRef, Folder: input.GitFolder},
		DryRun:       lo.ToPtr(input.DryRun),
	}
	if input.Kustomize != "" {
		attrs.Kustomize = &gqlclient.KustomizeAttributes{Path: input.Kustomize}
	}
	service, err := client.CreateClusterService(&input.ClusterID, nil, attrs)
	if err != nil {
		return Detail{}, err
	}
	if service == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingService}
	}
	return detailFromExtended(service), nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (Detail, error) {
	if err := ctx.Err(); err != nil {
		return Detail{}, err
	}
	input.ID = strings.TrimSpace(input.ID)
	if input.ID == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingID}
	}
	client, err := s.client(ctx)
	if err != nil {
		return Detail{}, err
	}
	attrs := gqlclient.ServiceUpdateAttributes{}
	if input.GitRef != "" || input.GitFolder != "" {
		attrs.Git = &gqlclient.GitRefAttributes{Ref: input.GitRef, Folder: input.GitFolder}
	}
	if input.Version != "" {
		attrs.Version = lo.ToPtr(input.Version)
	}
	if input.DryRun != nil {
		attrs.DryRun = input.DryRun
	}
	if input.Kustomize != "" {
		attrs.Kustomize = &gqlclient.KustomizeAttributes{Path: input.Kustomize}
	}
	service, err := client.UpdateClusterService(&input.ID, nil, nil, attrs)
	if err != nil {
		return Detail{}, err
	}
	if service == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingService}
	}
	return detailFromExtended(service), nil
}

func (s *Service) Clone(ctx context.Context, input CloneInput) (Detail, error) {
	if err := ctx.Err(); err != nil {
		return Detail{}, err
	}
	input.SourceID = strings.TrimSpace(input.SourceID)
	input.DestClusterID = strings.TrimSpace(input.DestClusterID)
	input.Name = strings.TrimSpace(input.Name)
	if input.SourceID == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingID}
	}
	if input.DestClusterID == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingCluster}
	}
	if input.Name == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingCreateFields}
	}
	if input.Namespace == "" {
		input.Namespace = "default"
	}
	client, err := s.client(ctx)
	if err != nil {
		return Detail{}, err
	}
	attrs := gqlclient.ServiceCloneAttributes{Name: input.Name, Namespace: lo.ToPtr(input.Namespace)}
	frag, err := client.CloneService(input.DestClusterID, &input.SourceID, nil, nil, attrs)
	if err != nil {
		return Detail{}, err
	}
	if frag == nil {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingService}
	}
	return s.Get(ctx, frag.ID)
}

func (s *Service) DownloadTarball(ctx context.Context, id, dir string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "", &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingID}
	}
	client, err := s.client(ctx)
	if err != nil {
		return "", err
	}
	service, err := client.GetClusterService(&id, nil, nil)
	if err != nil {
		return "", err
	}
	if service == nil {
		return "", &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingService}
	}
	if service.Tarball == nil || strings.TrimSpace(*service.Tarball) == "" {
		return "", &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingTarball}
	}
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = filepath.Join(".", service.Name+"-tarball")
	}
	if err := utils.EnsureEmptyDir(dir); err != nil {
		return "", err
	}
	if service.Cluster == nil {
		return "", &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingCluster}
	}
	token, err := client.GetDeployToken(&service.Cluster.ID, nil)
	if err != nil {
		return "", err
	}
	resp, err := utils.ReadRemoteFileWithRetries(*service.Tarball, token, 3)
	if err != nil {
		return "", err
	}
	defer func(c io.Closer) { _ = c.Close() }(resp)
	if err := utils.Untar(dir, resp); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir, nil
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("tarball directory missing after unpack: %w", err)
	}
	return abs, nil
}
