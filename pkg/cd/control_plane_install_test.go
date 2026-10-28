package cd

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestContextIngressClass(t *testing.T) {
	assert.Equal(t, "", contextIngressClass(nil))
	assert.Equal(t, "", contextIngressClass(map[string]interface{}{}))
	assert.Equal(t, "haproxy", contextIngressClass(map[string]interface{}{"IngressClass": "haproxy"}))
}

func TestListIngressOptionsTimesOut(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-done // accept the connection but never respond
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(done) }) // runs first, so srv.Close doesn't wait on blocked handlers

	old := discoveryTimeout
	discoveryTimeout = 100 * time.Millisecond
	t.Cleanup(func() { discoveryTimeout = old })

	start := time.Now()
	issuers, classes := listIngressOptions(&rest.Config{Host: srv.URL})

	assert.Empty(t, issuers)
	assert.Empty(t, classes)
	assert.Less(t, time.Since(start), 2*time.Second)
}
