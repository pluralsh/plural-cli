package access

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/pluralsh/plural-cli/pkg/bridge"
)

type Profile = bridge.Profile
type Identity = bridge.Identity
type ConsoleProfile = bridge.ConsoleProfile
type AuthContext = bridge.AuthContext
type DeviceAuthorization = bridge.DeviceAuthorization

// ServiceAccount is a selectable acting identity. Its credential is never
// persisted as profile metadata.
type ServiceAccount struct {
	ID    string `yaml:"id"`
	Email string `yaml:"email"`
}

// State is the non-secret, independently switchable access registry.
type State struct {
	Profiles        []Profile        `yaml:"profiles"`
	ConsoleProfiles []ConsoleProfile `yaml:"consoleProfiles"`
	ActiveProfileID string           `yaml:"activeProfile"`
	ActiveConsoleID string           `yaml:"activeConsole"`
}

type Snapshot struct {
	State           State
	Context         AuthContext
	ServiceAccounts []ServiceAccount
}

type Repository interface {
	Load(context.Context) (State, error)
	Save(context.Context, State) error
}
type ServiceAccountSource interface {
	ListServiceAccounts(context.Context, Profile, string) ([]ServiceAccount, error)
}

// Manager is the narrow contract consumed by the Access screen.
type Manager interface {
	Load(context.Context) (Snapshot, error)
	BeginDeviceLogin(context.Context, string) (DeviceAuthorization, error)
	CompleteDeviceLogin(context.Context, string, DeviceAuthorization, string) (Profile, error)
	AddConsoleProfile(context.Context, string, string, string) (ConsoleProfile, error)
	ActivateProfile(context.Context, string) error
	ActivateConsole(context.Context, string) error
	SearchServiceAccounts(context.Context, string) ([]ServiceAccount, error)
	Impersonate(context.Context, string) error
	StopImpersonating()
}

// Service coordinates registries, secure credentials, and ephemeral sessions.
type Service struct {
	repository      Repository
	credentials     bridge.CredentialStore
	auth            *bridge.AuthService
	serviceAccounts ServiceAccountSource
	mu              sync.RWMutex
	acting          *Identity
}

func NewService(repository Repository, credentials bridge.CredentialStore, auth *bridge.AuthService, serviceAccounts ServiceAccountSource) *Service {
	return &Service{repository: repository, credentials: credentials, auth: auth, serviceAccounts: serviceAccounts}
}

func (s *Service) Load(ctx context.Context) (Snapshot, error) {
	state, err := s.repository.Load(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	result := Snapshot{State: state}
	if profile, ok := findProfile(state.Profiles, state.ActiveProfileID); ok {
		result.Context.Base = &profile
	}
	if profile, ok := findConsole(state.ConsoleProfiles, state.ActiveConsoleID); ok {
		result.Context.Console = &profile
	}
	s.mu.RLock()
	if s.acting != nil {
		copy := *s.acting
		result.Context.Acting = &copy
	}
	s.mu.RUnlock()
	return result, result.Context.Validate()
}

func (s *Service) BeginDeviceLogin(ctx context.Context, endpoint string) (DeviceAuthorization, error) {
	if s.auth == nil {
		return DeviceAuthorization{}, errors.New("device login is unavailable")
	}
	return s.auth.BeginDeviceLogin(ctx, normalizeAppEndpoint(endpoint))
}

func (s *Service) CompleteDeviceLogin(ctx context.Context, name string, authorization DeviceAuthorization, endpoint string) (Profile, error) {
	if s.auth == nil {
		return Profile{}, errors.New("device login is unavailable")
	}
	endpoint = normalizeAppEndpoint(endpoint)
	credential, err := s.auth.AwaitDeviceLogin(ctx, endpoint, authorization.DeviceToken)
	if err != nil {
		return Profile{}, err
	}
	session, err := s.auth.EstablishSession(ctx, endpoint, credential, "", true, nil)
	if err != nil {
		return Profile{}, err
	}
	profile := Profile{ID: stableID("app", name, session.BaseEmail, endpoint), Name: strings.TrimSpace(name), Email: session.BaseEmail, Endpoint: endpoint}
	if profile.Name == "" {
		profile.Name = "default"
	}
	if err := s.credentials.Set(ctx, profile.ID, session.Credential); err != nil {
		return Profile{}, err
	}
	state, err := s.repository.Load(ctx)
	if err != nil {
		return Profile{}, err
	}
	state.Profiles = upsertProfile(state.Profiles, profile)
	state.ActiveProfileID = profile.ID
	if err := s.repository.Save(ctx, state); err != nil {
		_ = s.credentials.Delete(ctx, profile.ID)
		return Profile{}, err
	}
	s.StopImpersonating()
	return profile, nil
}

func (s *Service) AddConsoleProfile(ctx context.Context, name, rawURL, token string) (ConsoleProfile, error) {
	normalized, err := normalizeConsoleURL(rawURL)
	if err != nil {
		return ConsoleProfile{}, err
	}
	profile := ConsoleProfile{ID: stableID("console", name, normalized), Name: strings.TrimSpace(name), URL: normalized}
	if profile.Name == "" {
		profile.Name = "default"
	}
	if strings.TrimSpace(token) == "" {
		return ConsoleProfile{}, errors.New("Console token is required")
	}
	if err := s.credentials.Set(ctx, profile.ID, token); err != nil {
		return ConsoleProfile{}, err
	}
	state, err := s.repository.Load(ctx)
	if err != nil {
		return ConsoleProfile{}, err
	}
	state.ConsoleProfiles = upsertConsole(state.ConsoleProfiles, profile)
	state.ActiveConsoleID = profile.ID
	if err := s.repository.Save(ctx, state); err != nil {
		_ = s.credentials.Delete(ctx, profile.ID)
		return ConsoleProfile{}, err
	}
	return profile, nil
}

func (s *Service) ActivateProfile(ctx context.Context, id string) error {
	state, err := s.repository.Load(ctx)
	if err != nil {
		return err
	}
	if _, ok := findProfile(state.Profiles, id); !ok {
		return fmt.Errorf("Plural App profile %q not found", id)
	}
	state.ActiveProfileID = id
	s.StopImpersonating()
	return s.repository.Save(ctx, state)
}

func (s *Service) ActivateConsole(ctx context.Context, id string) error {
	state, err := s.repository.Load(ctx)
	if err != nil {
		return err
	}
	if _, ok := findConsole(state.ConsoleProfiles, id); !ok {
		return fmt.Errorf("Console profile %q not found", id)
	}
	state.ActiveConsoleID = id
	return s.repository.Save(ctx, state)
}

func (s *Service) SearchServiceAccounts(ctx context.Context, query string) ([]ServiceAccount, error) {
	snapshot, err := s.Load(ctx)
	if err != nil || snapshot.Context.Base == nil {
		return nil, err
	}
	if s.serviceAccounts == nil {
		return nil, nil
	}
	return s.serviceAccounts.ListServiceAccounts(ctx, *snapshot.Context.Base, query)
}

func (s *Service) Impersonate(ctx context.Context, email string) error {
	if s.auth == nil {
		return errors.New("impersonation is unavailable")
	}
	snapshot, err := s.Load(ctx)
	if err != nil {
		return err
	}
	if snapshot.Context.Base == nil {
		return errors.New("connect a Plural App profile before impersonating")
	}
	credential, err := s.credentials.Get(ctx, snapshot.Context.Base.ID)
	if err != nil {
		return err
	}
	session, err := s.auth.EstablishSession(ctx, snapshot.Context.Base.Endpoint, credential, email, false, nil)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.acting = &Identity{Email: session.EffectiveEmail, ServiceAccount: true}
	s.mu.Unlock()
	return nil
}

func (s *Service) StopImpersonating() { s.mu.Lock(); s.acting = nil; s.mu.Unlock() }

func findProfile(values []Profile, id string) (Profile, bool) {
	for _, value := range values {
		if value.ID == id {
			return value, true
		}
	}
	return Profile{}, false
}
func findConsole(values []ConsoleProfile, id string) (ConsoleProfile, bool) {
	for _, value := range values {
		if value.ID == id {
			return value, true
		}
	}
	return ConsoleProfile{}, false
}
func upsertProfile(values []Profile, value Profile) []Profile {
	for i := range values {
		if values[i].ID == value.ID {
			values[i] = value
			return values
		}
	}
	values = append(values, value)
	sort.Slice(values, func(i, j int) bool { return values[i].Name < values[j].Name })
	return values
}
func upsertConsole(values []ConsoleProfile, value ConsoleProfile) []ConsoleProfile {
	for i := range values {
		if values[i].ID == value.ID {
			values[i] = value
			return values
		}
	}
	values = append(values, value)
	sort.Slice(values, func(i, j int) bool { return values[i].Name < values[j].Name })
	return values
}
func normalizeAppEndpoint(endpoint string) string {
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(endpoint), "https://"), "/")
}
func normalizeConsoleURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", errors.New("Console URL must be an absolute https URL")
	}
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	return parsed.String(), nil
}
func stableID(parts ...string) string {
	joined := strings.ToLower(strings.Join(parts, "\x00"))
	var hash uint64 = 1469598103934665603
	for i := range joined {
		hash ^= uint64(joined[i])
		hash *= 1099511628211
	}
	return fmt.Sprintf("%s-%x", parts[0], hash)
}
