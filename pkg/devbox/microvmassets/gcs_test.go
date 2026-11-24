package microvmassets

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGcloudTokenSourceUsesReportedExpiry(t *testing.T) {
	exp := time.Now().Add(20 * time.Minute).UTC().Truncate(time.Second)
	g := &gcloudTokenSource{ctx: context.Background(), run: func(context.Context) ([]byte, error) {
		return []byte(`{"token":"tok","token_expiry":"` + exp.Format(time.RFC3339) + `"}`), nil
	}}
	tok, err := g.Token()
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "tok" || !tok.Expiry.Equal(exp) {
		t.Errorf("got %q expiring %v, want tok expiring %v", tok.AccessToken, tok.Expiry, exp)
	}
}

func TestGcloudTokenSourceFallsBackToShortLifetime(t *testing.T) {
	g := &gcloudTokenSource{ctx: context.Background(), run: func(context.Context) ([]byte, error) {
		return []byte(`{"token":"tok"}`), nil
	}}
	tok, err := g.Token()
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Until(tok.Expiry); d <= 0 || d > fallbackTokenLifetime {
		t.Errorf("fallback expiry %v away, want within %v", d, fallbackTokenLifetime)
	}
}

func TestGcloudTokenSourceHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := &gcloudTokenSource{ctx: ctx, run: func(ctx context.Context) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	if _, err := g.Token(); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestGcloudTokenSourceRejectsEmptyToken(t *testing.T) {
	g := &gcloudTokenSource{ctx: context.Background(), run: func(context.Context) ([]byte, error) { return []byte(`{}`), nil }}
	if _, err := g.Token(); err == nil {
		t.Error("expected an error for an empty token")
	}
}
