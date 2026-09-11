package bridge

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeAuthFactory struct{ client *fakeAuthClient }

func (f fakeAuthFactory) New(context.Context, string, string) AuthClient { return f.client }

type fakeAuthClient struct {
	polls       int
	accessCalls int
}

func (*fakeAuthClient) DeviceLogin(context.Context) (DeviceAuthorization, error) {
	return DeviceAuthorization{LoginURL: "https://example.com/login", DeviceToken: "device"}, nil
}

func (f *fakeAuthClient) PollLoginToken(context.Context, string) (string, error) {
	f.polls++
	if f.polls < 2 {
		return "", errors.New("pending")
	}
	return "jwt", nil
}

func (*fakeAuthClient) CurrentIdentity(context.Context) (string, error) {
	return "dev@example.com", nil
}

func (*fakeAuthClient) ImpersonateServiceAccount(context.Context, string) (string, string, error) {
	return "session-jwt", "deploy@example.com", nil
}

func (f *fakeAuthClient) GrabAccessToken(context.Context) (string, error) {
	f.accessCalls++
	return "access-token", nil
}

func TestAwaitDeviceLoginRetriesAndCanComplete(t *testing.T) {
	client := &fakeAuthClient{}
	service := NewAuthService(fakeAuthFactory{client}, time.Millisecond)
	credential, err := service.AwaitDeviceLogin(t.Context(), "", "device")
	if err != nil {
		t.Fatalf("AwaitDeviceLogin() error = %v", err)
	}
	if credential != "jwt" || client.polls != 2 {
		t.Fatalf("credential = %q, polls = %d", credential, client.polls)
	}
}

func TestAwaitDeviceLoginHonorsCancellation(t *testing.T) {
	client := &fakeAuthClient{polls: -100}
	service := NewAuthService(fakeAuthFactory{client}, time.Hour)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := service.AwaitDeviceLogin(ctx, "", "device")
	if !IsCode(err, ErrorCancelled) {
		t.Fatalf("AwaitDeviceLogin() error = %v", err)
	}
}

func TestSessionOnlyImpersonationDoesNotExchangeOrPersistBaseIdentity(t *testing.T) {
	client := &fakeAuthClient{}
	service := NewAuthService(fakeAuthFactory{client}, time.Millisecond)
	var events []AuthEvent
	session, err := service.EstablishSession(t.Context(), "", "base-jwt", "deploy@example.com", false, func(event AuthEvent) {
		events = append(events, event)
	})
	if err != nil {
		t.Fatalf("EstablishSession() error = %v", err)
	}
	if session.BaseEmail != "dev@example.com" || session.EffectiveEmail != "deploy@example.com" {
		t.Fatalf("session = %#v", session)
	}
	if session.Credential != "session-jwt" || client.accessCalls != 0 {
		t.Fatalf("credential = %q, access calls = %d", session.Credential, client.accessCalls)
	}
	if len(events) != 2 || events[0].Kind != AuthEventIdentified || events[1].Kind != AuthEventImpersonated {
		t.Fatalf("events = %#v", events)
	}
}

func TestPersistedImpersonationExchangesAccessToken(t *testing.T) {
	client := &fakeAuthClient{}
	service := NewAuthService(fakeAuthFactory{client}, time.Millisecond)
	session, err := service.EstablishSession(t.Context(), "", "base-jwt", "deploy@example.com", true, nil)
	if err != nil {
		t.Fatalf("EstablishSession() error = %v", err)
	}
	if session.Credential != "access-token" || client.accessCalls != 1 {
		t.Fatalf("session = %#v, access calls = %d", session, client.accessCalls)
	}
}
