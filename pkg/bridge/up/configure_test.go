package up

import "testing"

func TestValidateBucketPrefix(t *testing.T) {
	if err := ValidateBucketPrefix("acme"); err != nil {
		t.Fatalf("acme: %v", err)
	}
	if err := ValidateBucketPrefix("Acme"); err == nil {
		t.Fatal("expected uppercase rejection")
	}
	if err := ValidateBucketPrefix("1bad"); err == nil {
		t.Fatal("expected leading digit rejection")
	}
}

func TestPluralDomain(t *testing.T) {
	if got := PluralDomain("demo"); got != "demo.onplural.sh" {
		t.Fatalf("got %q", got)
	}
	if got := PluralDomain("demo.onplural.sh"); got != "demo.onplural.sh" {
		t.Fatalf("full got %q", got)
	}
}

func TestValidatePluralSubdomain(t *testing.T) {
	if err := ValidatePluralSubdomain("demo"); err != nil {
		t.Fatalf("demo: %v", err)
	}
}
