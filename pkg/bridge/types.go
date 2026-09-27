// Package bridge defines the frontend-independent boundary between Plural CLI
// presentation layers and authentication infrastructure.
//
// Authentication starts with an AuthClientFactory, which supplies transport
// implementations to AuthService. AuthService produces an AuthSession while
// reporting optional AuthEvents. Persisted Profile metadata is deliberately
// separated from secrets, which are resolved through CredentialStore.
// AuthContext then combines the selected base profile, an optional ephemeral
// acting identity, and an independently selected Console profile.
//
// View-oriented aggregation lives in the bridge/access and bridge/welcome
// subpackages. This package contains the shared vocabulary and infrastructure
// contracts used by those services and by legacy CLI commands.
package bridge

import (
	"context"
	"time"
)

// Operation identifies an authentication operation that can be attached to an
// Error for diagnostics and recovery decisions.
type Operation string

const (
	// OperationDeviceLogin starts device authorization.
	OperationDeviceLogin Operation = "DeviceLogin"
	// OperationPollLoginToken waits for device authorization to complete.
	OperationPollLoginToken Operation = "PollLoginToken"
	// OperationCurrentIdentity resolves the identity associated with a credential.
	OperationCurrentIdentity Operation = "Me"
	// OperationImpersonateServiceAccount exchanges a base credential for an acting identity.
	OperationImpersonateServiceAccount Operation = "ImpersonateServiceAccount"
	// OperationGrabAccessToken exchanges an authenticated session for a durable access token.
	OperationGrabAccessToken Operation = "GrabAccessToken"
)

// ErrorCode is a stable category that presentation layers can map to recovery
// actions without matching backend error strings.
type ErrorCode string

const (
	// ErrorUnauthenticated indicates that valid authentication is required.
	ErrorUnauthenticated ErrorCode = "unauthenticated"
	// ErrorUnauthorized indicates that the identity lacks permission.
	ErrorUnauthorized ErrorCode = "unauthorized"
	// ErrorInvalid indicates invalid caller input or persisted state.
	ErrorInvalid ErrorCode = "invalid"
	// ErrorUnavailable indicates a dependency or operation is unavailable.
	ErrorUnavailable ErrorCode = "unavailable"
	// ErrorCancelled indicates cancellation or deadline expiry.
	ErrorCancelled ErrorCode = "cancelled"
)

// Error adds operation and recovery semantics while preserving its cause.
type Error struct {
	Code      ErrorCode
	Operation Operation
	Err       error
}

// DeviceAuthorization contains the user-facing URL and opaque token for a
// device-login flow.
type DeviceAuthorization struct {
	LoginURL    string
	DeviceToken string
}

// AuthClient is the context-aware authentication transport required by the
// bridge. Implementations adapt concrete APIs without leaking them to callers.
type AuthClient interface {
	DeviceLogin(ctx context.Context) (DeviceAuthorization, error)
	PollLoginToken(ctx context.Context, deviceToken string) (string, error)
	CurrentIdentity(ctx context.Context) (string, error)
	ImpersonateServiceAccount(ctx context.Context, email string) (credential, effectiveEmail string, err error)
	GrabAccessToken(ctx context.Context) (string, error)
}

// AuthClientFactory creates an authentication client for an endpoint and an
// optional existing credential.
type AuthClientFactory interface {
	New(ctx context.Context, endpoint, credential string) AuthClient
}

// AuthService coordinates authentication clients, device-login polling, token
// exchange, and optional service-account impersonation.
type AuthService struct {
	clients      AuthClientFactory
	pollInterval time.Duration
}

// AuthEventKind identifies a meaningful transition during session creation.
type AuthEventKind string

const (
	// AuthEventIdentified reports the base identity resolved from a credential.
	AuthEventIdentified AuthEventKind = "identified"
	// AuthEventImpersonated reports a successful service-account exchange.
	AuthEventImpersonated AuthEventKind = "impersonated"
)

// AuthEvent reports an identity or credential transition to interested callers.
type AuthEvent struct {
	Kind       AuthEventKind
	Email      string
	Credential string
}

// AuthSession is the result of authentication and optional impersonation.
type AuthSession struct {
	BaseEmail      string
	EffectiveEmail string
	Credential     string
	Impersonated   bool
}

// Profile identifies a Plural App login. Its credential is stored separately
// and resolved through CredentialStore by profile ID.
type Profile struct {
	ID       string
	Name     string
	Email    string
	Endpoint string
}

// Identity is the effective actor for the current session.
type Identity struct {
	Email          string
	ServiceAccount bool
}

// ConsoleProfile identifies a Console connection selected independently from
// the Plural App profile.
type ConsoleProfile struct {
	ID    string
	Name  string
	URL   string
	Actor string
}

// AuthContext keeps persisted base identity separate from the effective
// in-memory actor and the independently selected Console connection.
type AuthContext struct {
	Base    *Profile
	Acting  *Identity
	Console *ConsoleProfile
}

// ProfileRepository stores non-secret identity metadata.
type ProfileRepository interface {
	List(ctx context.Context) ([]Profile, error)
	Get(ctx context.Context, id string) (Profile, error)
	Save(ctx context.Context, profile Profile) error
}

// CredentialStore stores secret material independently from profile metadata.
type CredentialStore interface {
	Get(ctx context.Context, profileID string) (string, error)
	Set(ctx context.Context, profileID, credential string) error
	Delete(ctx context.Context, profileID string) error
}

// SessionExchanger creates an ephemeral acting identity without replacing the
// persisted credential of its base profile.
type SessionExchanger interface {
	Impersonate(ctx context.Context, base Profile, serviceAccount string) (Identity, string, error)
}

// AuthContextLoader resolves active identity state for a caller-owned context.
type AuthContextLoader interface {
	Load(ctx context.Context) (AuthContext, error)
}
