package welcome

import (
	"context"
	"testing"
)

type welcomeSourceFunc func(context.Context) (Snapshot, error)

func (f welcomeSourceFunc) Read(ctx context.Context) (Snapshot, error) { return f(ctx) }

func TestWelcomeServiceReturnsReadOnlySnapshot(t *testing.T) {
	want := Snapshot{Version: "v1", App: AppProfile{Configured: true, Email: "dev@example.com"}}
	service := NewService(welcomeSourceFunc(func(context.Context) (Snapshot, error) { return want, nil }))
	got, err := service.Load(t.Context())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.App.Email != want.App.Email {
		t.Fatalf("Load() = %#v", got)
	}
}
