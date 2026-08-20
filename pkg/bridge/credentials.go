package bridge

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const credentialService = "plural-cli"

// KeyringCredentialStore stores secrets in the operating-system credential
// store. It contains no filesystem policy and is independently replaceable.
type KeyringCredentialStore struct{ Service string }

func (s KeyringCredentialStore) service() string {
	if s.Service != "" {
		return s.Service
	}
	return credentialService
}
func (s KeyringCredentialStore) Get(_ context.Context, id string) (string, error) {
	return keyring.Get(s.service(), id)
}
func (s KeyringCredentialStore) Set(_ context.Context, id, secret string) error {
	return keyring.Set(s.service(), id, secret)
}
func (s KeyringCredentialStore) Delete(_ context.Context, id string) error {
	return keyring.Delete(s.service(), id)
}

// FileCredentialStore is the owner-only fallback for hosts without a usable
// keyring. IDs are hashed so untrusted profile names cannot escape its root.
type FileCredentialStore struct{ Dir string }

func (s FileCredentialStore) path(id string) string {
	hash := sha256.Sum256([]byte(id))
	return filepath.Join(s.Dir, fmt.Sprintf("%x", hash[:])+".credential")
}
func (s FileCredentialStore) Get(ctx context.Context, id string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	value, err := os.ReadFile(s.path(id))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(value)), nil
}
func (s FileCredentialStore) Set(ctx context.Context, id, secret string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(s.Dir, 0700); err != nil {
		return err
	}
	if err := os.Chmod(s.Dir, 0700); err != nil {
		return err
	}
	target := s.path(id)
	if err := os.WriteFile(target, []byte(secret+"\n"), 0600); err != nil {
		return err
	}
	return os.Chmod(target, 0600)
}
func (s FileCredentialStore) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := os.Remove(s.path(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// ResilientCredentialStore prefers a keyring, falls back to owner-only files,
// and lazily migrates fallback secrets when the keyring becomes available.
type ResilientCredentialStore struct{ Primary, Fallback CredentialStore }

func (s ResilientCredentialStore) Get(ctx context.Context, id string) (string, error) {
	if s.Primary != nil {
		if value, err := s.Primary.Get(ctx, id); err == nil {
			return value, nil
		}
	}
	if s.Fallback == nil {
		return "", os.ErrNotExist
	}
	value, err := s.Fallback.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if s.Primary != nil && s.Primary.Set(ctx, id, value) == nil {
		_ = s.Fallback.Delete(ctx, id)
	}
	return value, nil
}
func (s ResilientCredentialStore) Set(ctx context.Context, id, secret string) error {
	if s.Primary != nil && s.Primary.Set(ctx, id, secret) == nil {
		if s.Fallback != nil {
			_ = s.Fallback.Delete(ctx, id)
		}
		return nil
	}
	if s.Fallback == nil {
		return errors.New("no credential store is available")
	}
	return s.Fallback.Set(ctx, id, secret)
}
func (s ResilientCredentialStore) Delete(ctx context.Context, id string) error {
	var primaryErr error
	if s.Primary != nil {
		primaryErr = s.Primary.Delete(ctx, id)
	}
	if s.Fallback != nil {
		if err := s.Fallback.Delete(ctx, id); err != nil {
			return err
		}
	}
	if errors.Is(primaryErr, keyring.ErrNotFound) {
		return nil
	}
	return primaryErr
}
