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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseCertificatesRejectsInvalid(t *testing.T) {
	_, err := parseCertificates("not a cert")
	require.Error(t, err)

	_, err = parseCertificates("-----BEGIN CERTIFICATE-----\nbm90LWEtY2VydA==\n-----END CERTIFICATE-----\n")
	require.Error(t, err)
}

func TestMergeWithSystemRoots(t *testing.T) {
	dir := t.TempDir()
	bundle := filepath.Join(dir, "ca-certificates.crt")

	old := systemCABundlePaths
	systemCABundlePaths = []string{bundle}
	t.Cleanup(func() { systemCABundlePaths = old })

	caCert, _ := mustCreateCA(t)
	var private strings.Builder
	require.NoError(t, pem.Encode(&private, &pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw}))

	_, err := MergeWithSystemRoots("not-a-pem")
	require.Error(t, err)

	require.NoError(t, os.WriteFile(bundle, []byte(private.String()), 0o644))
	merged, err := MergeWithSystemRoots(private.String())
	require.NoError(t, err)
	require.Contains(t, merged, "BEGIN CERTIFICATE")
}

func TestFetchTLSChain(t *testing.T) {
	caCert, caKey := mustCreateCA(t)
	leafCert, leafKey := mustCreateLeaf(t, caCert, caKey)

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
	go func() {
		_ = srv.Serve(tlsLn)
	}()
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

func TestReadSystemCABundleSSLCertFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.crt")
	require.NoError(t, os.WriteFile(path, []byte("custom-roots\n"), 0o644))
	t.Setenv("SSL_CERT_FILE", path)

	data, err := readSystemCABundle()
	require.NoError(t, err)
	require.Equal(t, "custom-roots\n", string(data))
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
