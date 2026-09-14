package stacks

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/stacks"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

var (
	errMissingStackID = errors.New("stack id is required")
	errMissingActor   = errors.New("plural app email is required to generate a terraform backend (run plural login)")
)

// GenBackendInput generates '_override.tf' for a stack (CLI: plural stacks gen-backend).
type GenBackendInput struct {
	StackID       string
	Dir           string
	Address       string // optional override; otherwise fetched from Console runs
	LockAddress   string
	UnlockAddress string
}

// GenBackendResult is the credential-free outcome of gen-backend.
type GenBackendResult struct {
	FilePath string
	Dir      string
}

// Loader is the narrow contract consumed by the Stacks screen.
type Loader interface {
	List(ctx context.Context, after *string, query string) (Page, error)
	Get(ctx context.Context, id string) (Detail, error)
	GenBackend(ctx context.Context, input GenBackendInput) (GenBackendResult, error)
}

// API is the Console surface required by this package.
type API interface {
	ListStacks() (*gqlclient.ListInfrastructureStacks, error)
	GetStack(id string) (*gqlclient.InfrastructureStackFragment, error)
	ListStackRuns(stackID string) (*gqlclient.ListStackRuns, error)
}

func (s *Service) GenBackend(ctx context.Context, input GenBackendInput) (GenBackendResult, error) {
	if err := ctx.Err(); err != nil {
		return GenBackendResult{}, err
	}
	id := strings.TrimSpace(input.StackID)
	if id == "" {
		return GenBackendResult{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingStackID}
	}
	dir := strings.TrimSpace(input.Dir)
	if dir == "" {
		dir = "."
	}

	if s.resolve == nil {
		return GenBackendResult{}, &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoConsole}
	}
	_, token, err := s.resolve.ActiveConsole(ctx)
	if err != nil {
		return GenBackendResult{}, err
	}
	client, err := s.client(ctx)
	if err != nil {
		return GenBackendResult{}, err
	}

	address, lock, unlock := strings.TrimSpace(input.Address), strings.TrimSpace(input.LockAddress), strings.TrimSpace(input.UnlockAddress)
	if address == "" || lock == "" || unlock == "" {
		stateUrls, err := stacks.GetTerraformStateUrls(client, id)
		if err != nil {
			return GenBackendResult{}, err
		}
		if address == "" {
			address = lo.FromPtr(stateUrls.Address)
		}
		if lock == "" {
			lock = lo.FromPtr(stateUrls.Lock)
		}
		if unlock == "" {
			unlock = lo.FromPtr(stateUrls.Unlock)
		}
	}

	actor, err := s.actorEmail()
	if err != nil {
		return GenBackendResult{}, err
	}

	fileName, err := stacks.GenerateOverrideTemplate(&stacks.OverrideTemplateInput{
		Address:       address,
		LockAddress:   lock,
		UnlockAddress: unlock,
		Actor:         actor,
		DeployToken:   token,
	}, dir)
	if err != nil {
		return GenBackendResult{}, err
	}
	if err := git.AppendGitIgnore(dir, []string{fileName}); err != nil {
		return GenBackendResult{}, err
	}
	return GenBackendResult{FilePath: filepath.Join(dir, fileName), Dir: dir}, nil
}

func (s *Service) actorEmail() (string, error) {
	if s.actor != nil {
		return s.actor()
	}
	if !config.Exists() {
		return "", &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errMissingActor}
	}
	cfg := config.Read()
	if strings.TrimSpace(cfg.Email) == "" {
		return "", &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errMissingActor}
	}
	return cfg.Email, nil
}
