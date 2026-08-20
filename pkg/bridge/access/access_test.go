package access

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pluralsh/plural-cli/pkg/bridge"
)

type fakeAuthFactory struct{ client *fakeAuthClient }

func (f fakeAuthFactory) New(context.Context, string, string) bridge.AuthClient { return f.client }

type fakeAuthClient struct{}

func (*fakeAuthClient) DeviceLogin(context.Context) (bridge.DeviceAuthorization, error) {
	return bridge.DeviceAuthorization{LoginURL: "https://example.com/login", DeviceToken: "device"}, nil
}
func (*fakeAuthClient) PollLoginToken(context.Context, string) (string, error) {
	return "jwt", nil
}
func (*fakeAuthClient) CurrentIdentity(context.Context) (string, error) {
	return "dev@example.com", nil
}
func (*fakeAuthClient) ImpersonateServiceAccount(context.Context, string) (string, string, error) {
	return "session-jwt", "deploy@example.com", nil
}
func (*fakeAuthClient) GrabAccessToken(context.Context) (string, error) {
	return "access-token", nil
}

type memoryAccessRepository struct{ state State }

func (r *memoryAccessRepository) Load(context.Context) (State, error) { return r.state, nil }
func (r *memoryAccessRepository) Save(_ context.Context, state State) error {
	r.state = state
	return nil
}

type memoryCredentials struct {
	values      map[string]string
	unavailable bool
}

func (s *memoryCredentials) Get(_ context.Context, id string) (string, error) {
	if s.unavailable {
		return "", errors.New("unavailable")
	}
	value, ok := s.values[id]
	if !ok {
		return "", os.ErrNotExist
	}
	return value, nil
}
func (s *memoryCredentials) Set(_ context.Context, id, value string) error {
	if s.unavailable {
		return errors.New("unavailable")
	}
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[id] = value
	return nil
}
func (s *memoryCredentials) Delete(_ context.Context, id string) error {
	if s.unavailable {
		return errors.New("unavailable")
	}
	delete(s.values, id)
	return nil
}

func TestAccessProfilesSwitchIndependentlyAndClearActingIdentity(t *testing.T) {
	repository := &memoryAccessRepository{state: State{
		Profiles: []Profile{{ID: "app-a", Email: "a@example.com"}, {ID: "app-b", Email: "b@example.com"}}, ActiveProfileID: "app-a",
		ConsoleProfiles: []ConsoleProfile{{ID: "console-a"}, {ID: "console-b"}}, ActiveConsoleID: "console-a",
	}}
	credentials := &memoryCredentials{values: map[string]string{"app-a": "base-token"}}
	service := NewService(repository, credentials, bridge.NewAuthService(fakeAuthFactory{&fakeAuthClient{}}, time.Millisecond), nil)
	if err := service.Impersonate(t.Context(), "deploy@example.com"); err != nil {
		t.Fatalf("Impersonate() error = %v", err)
	}
	if err := service.ActivateConsole(t.Context(), "console-b"); err != nil {
		t.Fatalf("ActivateConsole() error = %v", err)
	}
	snapshot, _ := service.Load(t.Context())
	if snapshot.State.ActiveProfileID != "app-a" || snapshot.State.ActiveConsoleID != "console-b" || snapshot.Context.Acting == nil {
		t.Fatalf("independent Console switch lost state: %#v", snapshot)
	}
	if err := service.ActivateProfile(t.Context(), "app-b"); err != nil {
		t.Fatalf("ActivateProfile() error = %v", err)
	}
	snapshot, _ = service.Load(t.Context())
	if snapshot.State.ActiveConsoleID != "console-b" || snapshot.Context.Acting != nil {
		t.Fatalf("base switch did not preserve Console/clear acting: %#v", snapshot)
	}
	if credentials.values["app-a"] != "base-token" {
		t.Fatal("impersonation overwrote the base credential")
	}
}

func TestActiveConsolePrefersRegistryThenLegacy(t *testing.T) {
	repository := &memoryAccessRepository{state: State{
		ConsoleProfiles: []ConsoleProfile{{ID: "console-a", Name: "production", URL: "https://console.example.com"}},
		ActiveConsoleID: "console-a",
	}}
	credentials := &memoryCredentials{values: map[string]string{"console-a": "registry-token"}}
	service := NewService(repository, credentials, nil, nil)

	url, token, err := service.ActiveConsole(t.Context())
	if err != nil {
		t.Fatalf("ActiveConsole() error = %v", err)
	}
	if url != "https://console.example.com" || token != "registry-token" {
		t.Fatalf("ActiveConsole() = %q, %q", url, token)
	}

	empty := NewService(&memoryAccessRepository{}, &memoryCredentials{}, nil, nil)
	original := readLegacyConsole
	t.Cleanup(func() { readLegacyConsole = original })
	readLegacyConsole = func() (string, string, bool) { return "https://legacy.example.com", "legacy-token", true }
	url, token, err = empty.ActiveConsole(t.Context())
	if err != nil || url != "https://legacy.example.com" || token != "legacy-token" {
		t.Fatalf("legacy ActiveConsole() = %q, %q, %v", url, token, err)
	}

	readLegacyConsole = func() (string, string, bool) { return "", "", false }
	_, _, err = empty.ActiveConsole(t.Context())
	if !bridge.IsCode(err, bridge.ErrorUnauthenticated) {
		t.Fatalf("missing console error = %v", err)
	}
}

func TestCompleteDeviceLoginStoresBaseCredentialOutsideMetadata(t *testing.T) {
	repository := &memoryAccessRepository{}
	credentials := &memoryCredentials{}
	service := NewService(repository, credentials, bridge.NewAuthService(fakeAuthFactory{&fakeAuthClient{}}, time.Millisecond), nil)
	profile, err := service.CompleteDeviceLogin(t.Context(), "personal", DeviceAuthorization{DeviceToken: "device"}, "app.plural.sh")
	if err != nil {
		t.Fatalf("CompleteDeviceLogin() error = %v", err)
	}
	if repository.state.ActiveProfileID != profile.ID || profile.Email != "dev@example.com" {
		t.Fatalf("profile = %#v state = %#v", profile, repository.state)
	}
	if credentials.values[profile.ID] != "access-token" {
		t.Fatalf("stored credential = %q", credentials.values[profile.ID])
	}
}

func TestFileCredentialStoreUsesOwnerOnlyPermissions(t *testing.T) {
	store := bridge.FileCredentialStore{Dir: filepath.Join(t.TempDir(), "credentials")}
	if err := store.Set(t.Context(), "../../unsafe", "secret"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	entries, err := os.ReadDir(store.Dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries = %v, error = %v", entries, err)
	}
	info, _ := entries[0].Info()
	if info.Mode().Perm() != 0600 {
		t.Fatalf("credential mode = %o", info.Mode().Perm())
	}
	dirInfo, _ := os.Stat(store.Dir)
	if dirInfo.Mode().Perm() != 0700 {
		t.Fatalf("directory mode = %o", dirInfo.Mode().Perm())
	}
}

func TestResilientCredentialStoreMigratesFallback(t *testing.T) {
	primary := &memoryCredentials{unavailable: true}
	fallback := &memoryCredentials{values: map[string]string{"profile": "secret"}}
	store := bridge.ResilientCredentialStore{Primary: primary, Fallback: fallback}
	if value, err := store.Get(t.Context(), "profile"); err != nil || value != "secret" {
		t.Fatalf("fallback Get() = %q, %v", value, err)
	}
	primary.unavailable = false
	if value, err := store.Get(t.Context(), "profile"); err != nil || value != "secret" {
		t.Fatalf("migrating Get() = %q, %v", value, err)
	}
	if primary.values["profile"] != "secret" {
		t.Fatal("fallback credential was not migrated")
	}
	if _, ok := fallback.values["profile"]; ok {
		t.Fatal("fallback credential remained after migration")
	}
}
