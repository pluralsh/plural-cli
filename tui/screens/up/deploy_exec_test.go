package up

import (
	"strings"
	"testing"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
)

func TestShouldStreamLiveRunner(t *testing.T) {
	if !shouldStreamLiveRunner(nil) {
		t.Fatal("nil runner should stream")
	}
	if !shouldStreamLiveRunner(upbridge.LiveRunner{}) {
		t.Fatal("live runner should stream")
	}
	if shouldStreamLiveRunner(&stubRunner{}) {
		t.Fatal("stub runner should not stream")
	}
	if !shouldExecDeploy(upbridge.LiveRunner{}) {
		t.Fatal("shouldExecDeploy alias broken")
	}
}

func TestLineWriterSplitsLines(t *testing.T) {
	ch := make(chan string, 8)
	w := &lineWriter{ch: ch}
	if _, err := w.Write([]byte("hello\nwor")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("ld\n")); err != nil {
		t.Fatal(err)
	}
	w.Close()
	var got []string
	for len(ch) > 0 {
		got = append(got, <-ch)
	}
	if strings.Join(got, "|") != "hello|world" {
		t.Fatalf("got %#v", got)
	}
}

func TestAppendOpLogCaps(t *testing.T) {
	var lines []string
	for i := 0; i < maxOpLogLines+10; i++ {
		lines = appendOpLog(lines, "x")
	}
	if len(lines) != maxOpLogLines {
		t.Fatalf("len=%d", len(lines))
	}
}
