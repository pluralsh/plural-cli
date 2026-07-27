package bridge

import "testing"

func TestAuthContextKeepsBaseAndActingIdentitySeparate(t *testing.T) {
	ctx := AuthContext{
		Base:   &Profile{ID: "personal", Email: "dev@example.com"},
		Acting: &Identity{Email: "deploy@example.com", ServiceAccount: true},
	}

	if err := ctx.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got := ctx.EffectiveEmail(); got != "deploy@example.com" {
		t.Fatalf("EffectiveEmail() = %q", got)
	}
	if got := ctx.Base.Email; got != "dev@example.com" {
		t.Fatalf("base profile was changed: %q", got)
	}
}

func TestAuthContextRejectsActingIdentityWithoutBase(t *testing.T) {
	ctx := AuthContext{Acting: &Identity{Email: "deploy@example.com"}}
	if err := ctx.Validate(); err == nil {
		t.Fatal("Validate() expected an error")
	}
}
