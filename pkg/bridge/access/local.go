package access

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/pluralsh/plural-cli/pkg/bridge"
)

const accessRegistryName = "access.yml"

// LocalRepository persists only non-secret registry metadata and imports
// legacy config.yml/console.yml records the first time it is loaded.
type LocalRepository struct {
	Dir         string
	Credentials bridge.CredentialStore
	mu          sync.Mutex
}

func NewLocalRepository(home string, credentials bridge.CredentialStore) *LocalRepository {
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	return &LocalRepository{Dir: filepath.Join(home, ".plural"), Credentials: credentials}
}

func (r *LocalRepository) Load(ctx context.Context) (State, error) {
	if err := ctx.Err(); err != nil {
		return State{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	contents, err := os.ReadFile(filepath.Join(r.Dir, accessRegistryName))
	if err == nil {
		var state State
		if err := yaml.Unmarshal(contents, &state); err != nil {
			return State{}, err
		}
		state, changed, err := r.ensureLegacyConsole(ctx, state)
		if err != nil {
			return State{}, err
		}
		if changed {
			if err := r.save(state); err != nil {
				return State{}, err
			}
		}
		return state, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return State{}, err
	}
	state, err := r.importLegacy(ctx)
	if err != nil {
		return State{}, err
	}
	if err := r.save(state); err != nil {
		return State{}, err
	}
	return state, nil
}

func (r *LocalRepository) Save(ctx context.Context, state State) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.save(state)
}

func (r *LocalRepository) save(state State) error {
	contents, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(r.Dir, 0700); err != nil {
		return err
	}
	if err := os.Chmod(r.Dir, 0700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(r.Dir, ".access-*.tmp")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if err := temporary.Chmod(0600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(contents); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(r.Dir, accessRegistryName))
}

type legacyAppConfig struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		Email    string `yaml:"email"`
		Token    string `yaml:"token"`
		Endpoint string `yaml:"endpoint"`
	} `yaml:"spec"`
}
type legacyConsoleConfig struct {
	Kind string `yaml:"kind"`
	Spec struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
	} `yaml:"spec"`
}

func (r *LocalRepository) importLegacy(ctx context.Context) (State, error) {
	state := State{}
	entries, err := os.ReadDir(r.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") || entry.Name() == accessRegistryName {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(r.Dir, entry.Name()))
		if err != nil {
			return state, err
		}
		if entry.Name() == "console.yml" {
			var legacy legacyConsoleConfig
			if yaml.Unmarshal(contents, &legacy) != nil || legacy.Spec.URL == "" {
				continue
			}
			profile := ConsoleProfile{ID: stableID("console", "default", legacy.Spec.URL), Name: "default", URL: legacy.Spec.URL}
			state.ConsoleProfiles = upsertConsole(state.ConsoleProfiles, profile)
			state.ActiveConsoleID = profile.ID
			if legacy.Spec.Token != "" && r.Credentials != nil {
				if err := r.Credentials.Set(ctx, profile.ID, legacy.Spec.Token); err != nil {
					return state, err
				}
			}
			continue
		}
		var legacy legacyAppConfig
		if yaml.Unmarshal(contents, &legacy) != nil || legacy.Kind != "Config" || legacy.Spec.Email == "" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yml")
		if entry.Name() == "config.yml" {
			name = "default"
		}
		if legacy.Metadata.Name != "" {
			name = legacy.Metadata.Name
		}
		profile := Profile{ID: stableID("app", name, legacy.Spec.Email, legacy.Spec.Endpoint), Name: name, Email: legacy.Spec.Email, Endpoint: legacy.Spec.Endpoint}
		state.Profiles = upsertProfile(state.Profiles, profile)
		if entry.Name() == "config.yml" {
			state.ActiveProfileID = profile.ID
		}
		if legacy.Spec.Token != "" && r.Credentials != nil {
			if err := r.Credentials.Set(ctx, profile.ID, legacy.Spec.Token); err != nil {
				return state, err
			}
		}
	}
	if state.ActiveProfileID == "" && len(state.Profiles) > 0 {
		state.ActiveProfileID = state.Profiles[0].ID
	}
	return state, nil
}

// ensureLegacyConsole imports ~/.plural/console.yml when the Access registry
// has no usable Console profile yet (common after plural cd login while
// access.yml already existed from App login).
func (r *LocalRepository) ensureLegacyConsole(ctx context.Context, state State) (State, bool, error) {
	if _, ok := findConsole(state.ConsoleProfiles, state.ActiveConsoleID); ok {
		return state, false, nil
	}
	contents, err := os.ReadFile(filepath.Join(r.Dir, "console.yml"))
	if errors.Is(err, os.ErrNotExist) {
		if state.ActiveConsoleID != "" && len(state.ConsoleProfiles) == 0 {
			state.ActiveConsoleID = ""
			return state, true, nil
		}
		return state, false, nil
	}
	if err != nil {
		return state, false, err
	}
	var legacy legacyConsoleConfig
	if yaml.Unmarshal(contents, &legacy) != nil || legacy.Spec.URL == "" {
		return state, false, nil
	}
	profile := ConsoleProfile{ID: stableID("console", "default", legacy.Spec.URL), Name: "default", URL: legacy.Spec.URL}
	state.ConsoleProfiles = upsertConsole(state.ConsoleProfiles, profile)
	state.ActiveConsoleID = profile.ID
	if legacy.Spec.Token != "" && r.Credentials != nil {
		if err := r.Credentials.Set(ctx, profile.ID, legacy.Spec.Token); err != nil {
			return state, false, err
		}
	}
	return state, true, nil
}

// NewLocalManager builds the production persistence stack. Callers can
// still inject every boundary separately in tests or alternate frontends.
func NewLocalManager(home string, auth *bridge.AuthService, serviceAccounts ServiceAccountSource) *Service {
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	credentials := bridge.ResilientCredentialStore{
		Primary:  bridge.KeyringCredentialStore{},
		Fallback: bridge.FileCredentialStore{Dir: filepath.Join(home, ".plural", "credentials")},
	}
	if serviceAccounts == nil {
		serviceAccounts = PluralServiceAccountSource{Credentials: credentials}
	}
	repository := NewLocalRepository(home, credentials)
	return NewService(repository, credentials, auth, serviceAccounts)
}
