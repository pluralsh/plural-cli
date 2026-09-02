package up

import (
	"testing"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
)

func TestShouldExecLiveRunner(t *testing.T) {
	if !shouldExecLiveRunner(nil) {
		t.Fatal("nil runner should exec")
	}
	if !shouldExecLiveRunner(upbridge.LiveRunner{}) {
		t.Fatal("live runner should exec")
	}
	if shouldExecLiveRunner(&stubRunner{}) {
		t.Fatal("stub runner should not exec")
	}
	if !shouldExecDeploy(upbridge.LiveRunner{}) {
		t.Fatal("shouldExecDeploy alias broken")
	}
}
