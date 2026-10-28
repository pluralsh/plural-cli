package cd

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestContextIngressClass(t *testing.T) {
	assert.Equal(t, "", contextIngressClass(nil))
	assert.Equal(t, "", contextIngressClass(map[string]interface{}{}))
	assert.Equal(t, "haproxy", contextIngressClass(map[string]interface{}{"IngressClass": "haproxy"}))
}

func TestValidateResourceName(t *testing.T) {
	assert.NoError(t, validateResourceName(""))
	assert.NoError(t, validateResourceName("letsencrypt-prod"))
	assert.NoError(t, validateResourceName("issuer.example.com"))
	assert.Error(t, validateResourceName("letsencrypt: prod"))
	assert.Error(t, validateResourceName("nginx\n  bar: baz"))
	assert.Error(t, validateResourceName("Traefik"))
}

// hangingServer accepts connections but never responds.
func hangingServer(t *testing.T) *httptest.Server {
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-done
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(done) }) // runs first, so srv.Close doesn't wait on blocked handlers
	return srv
}

func withDiscoveryTimeout(t *testing.T, d time.Duration) {
	old := discoveryTimeout
	discoveryTimeout = d
	t.Cleanup(func() { discoveryTimeout = old })
}

func TestListIngressOptionsTimesOut(t *testing.T) {
	withDiscoveryTimeout(t, 100*time.Millisecond)

	start := time.Now()
	issuers, classes := listIngressOptions(&rest.Config{Host: hangingServer(t).URL})

	assert.Empty(t, issuers)
	assert.Empty(t, classes)
	assert.Less(t, time.Since(start), time.Second)
}

func TestListIngressOptionsHangingExecPlugin(t *testing.T) {
	withDiscoveryTimeout(t, 100*time.Millisecond)
	exec := &clientcmdapi.ExecConfig{
		APIVersion:      "client.authentication.k8s.io/v1",
		Command:         "sleep",
		Args:            []string{"2"}, // well past the deadline; kept short since go test waits for the orphaned plugin's stderr
		InteractiveMode: clientcmdapi.IfAvailableExecInteractiveMode,
	}

	start := time.Now()
	issuers, classes := listIngressOptions(&rest.Config{Host: hangingServer(t).URL, ExecProvider: exec})

	assert.Empty(t, issuers)
	assert.Empty(t, classes)
	assert.Less(t, time.Since(start), time.Second)
	assert.Equal(t, clientcmdapi.IfAvailableExecInteractiveMode, exec.InteractiveMode, "caller's exec config must not be mutated")
}
