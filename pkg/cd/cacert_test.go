package cd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIndentPEM(t *testing.T) {
	in := "-----BEGIN CERTIFICATE-----\nABC\n-----END CERTIFICATE-----\n"
	out := IndentPEM(in, 6)
	require.Equal(t, "      -----BEGIN CERTIFICATE-----\n      ABC\n      -----END CERTIFICATE-----\n", out)
}

func TestFetchTLSChain(t *testing.T) {
	caCert, caKey := mustCreateCA(t)
	leafCert, leafKey := mustCreateLeaf(t, caCert, caKey)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	server.TLS.Certificates = []tls.Certificate{{
		Certificate: [][]byte{leafCert.Raw, caCert.Raw},
		PrivateKey:  leafKey,
	}}
	// rebuild listener certs
	server.Close()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	tlsLn := tls.NewListener(ln, &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{leafCert.Raw, caCert.Raw},
			PrivateKey:  leafKey,
		}},
	})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}
	go srv.Serve(tlsLn)
	defer srv.Close()

	pemChain, err := FetchTLSChain("https://" + ln.Addr().String())
	require.NoError(t, err)
	require.Contains(t, pemChain, "BEGIN CERTIFICATE")
	block, _ := pem.Decode([]byte(pemChain))
	require.NotNil(t, block)
	parsed, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	require.True(t, parsed.Equal(caCert))
}

func TestMergeWithSystemRoots(t *testing.T) {
	dir := t.TempDir()
	bundle := filepath.Join(dir, "ca-certificates.crt")
	require.NoError(t, os.WriteFile(bundle, []byte("-----BEGIN CERTIFICATE-----\nPUBLIC\n-----END CERTIFICATE-----\n"), 0o644))

	old := systemCABundlePaths
	systemCABundlePaths = []string{bundle}
	t.Cleanup(func() { systemCABundlePaths = old })

	merged, err := MergeWithSystemRoots("-----BEGIN CERTIFICATE-----\nPRIVATE\n-----END CERTIFICATE-----\n")
	require.NoError(t, err)
	require.Contains(t, merged, "PRIVATE")
	require.Contains(t, merged, "PUBLIC")
}

func mustCreateCA(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert, key
}

func mustCreateLeaf(t *testing.T, ca *x509.Certificate, caKey *ecdsa.PrivateKey) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &key.PublicKey, caKey)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert, key
}
