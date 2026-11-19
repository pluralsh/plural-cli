package cd

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
)

var systemCABundlePaths = []string{
	"/etc/ssl/certs/ca-certificates.crt",
	"/etc/pki/tls/certs/ca-bundle.crt",
	"/etc/ssl/cert.pem",
}

// FetchTLSChain dials host (OpenSSL s_client -showcerts style) and returns PEM
// for intermediates/roots presented by the peer. The leaf certificate is skipped.
//
// TLS verification is intentionally skipped so we can extract the private PKI
// chain from a host that is not yet trusted by this machine.
func FetchTLSChain(rawURL string) (string, error) {
	hostPort, err := tlsDialAddress(rawURL)
	if err != nil {
		return "", err
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		hostPort,
		// #nosec G402 -- intentional: discover private CA chain from a host not yet trusted locally
		&tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12},
	)
	if err != nil {
		return "", fmt.Errorf("tls dial %s: %w", hostPort, err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return "", fmt.Errorf("no peer certificates from %s", hostPort)
	}

	var b strings.Builder
	certs := state.PeerCertificates
	if len(certs) == 1 {
		return "", fmt.Errorf("server at %s only presented a leaf certificate; provide the CA PEM with --private-ca-file or paste it when prompted", hostPort)
	}
	for _, cert := range certs[1:] {
		if err := pem.Encode(&b, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

func tlsDialAddress(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty tls address")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("missing host in %q", raw)
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}
	return net.JoinHostPort(host, port), nil
}

// parseCertificates decodes and parses all CERTIFICATE PEM blocks.
func parseCertificates(pemData string) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	rest := []byte(pemData)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse certificate PEM: %w", err)
		}
		certs = append(certs, cert)
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("no valid CERTIFICATE PEM blocks found")
	}
	return certs, nil
}

// certificateFingerprints returns SHA-256 fingerprints for display/confirmation.
func certificateFingerprints(certs []*x509.Certificate) []string {
	out := make([]string, 0, len(certs))
	for _, cert := range certs {
		sum := sha256.Sum256(cert.Raw)
		out = append(out, strings.ToUpper(hex.EncodeToString(sum[:])))
	}
	return out
}

// MergeWithSystemRoots validates privatePEM and appends OS/browser-trusted roots.
func MergeWithSystemRoots(privatePEM string) (string, error) {
	privatePEM = strings.TrimSpace(privatePEM)
	if privatePEM == "" {
		return "", fmt.Errorf("empty private CA PEM")
	}
	if _, err := parseCertificates(privatePEM); err != nil {
		return "", fmt.Errorf("invalid private CA PEM: %w", err)
	}
	roots, err := readSystemCABundle()
	if err != nil {
		return "", err
	}
	var out strings.Builder
	out.WriteString(privatePEM)
	if !strings.HasSuffix(privatePEM, "\n") {
		out.WriteByte('\n')
	}
	out.WriteString(strings.TrimSpace(string(roots)))
	out.WriteByte('\n')
	return out.String(), nil
}

func readSystemCABundle() ([]byte, error) {
	if path := os.Getenv("SSL_CERT_FILE"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read SSL_CERT_FILE %q: %w", path, err)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("SSL_CERT_FILE %q is empty", path)
		}
		return data, nil
	}
	for _, path := range systemCABundlePaths {
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}
	return nil, fmt.Errorf("could not locate a system CA bundle (tried SSL_CERT_FILE and %s)", strings.Join(systemCABundlePaths, ", "))
}

// ResolvePrivateCABundle builds a merged CA PEM from optional URL and/or file.
func ResolvePrivateCABundle(caURL, caFile string) (string, error) {
	var private string
	var err error
	switch {
	case caFile != "":
		data, err := os.ReadFile(caFile)
		if err != nil {
			return "", fmt.Errorf("read private ca file: %w", err)
		}
		private = string(data)
	case caURL != "":
		private, err = FetchTLSChain(caURL)
		if err != nil {
			return "", err
		}
	default:
		return "", nil
	}
	return MergeWithSystemRoots(private)
}

// askPrivateCA prompts whether to configure a private CA and returns a PEM
// suitable for values templates (templates indent under data: |).
func askPrivateCA() (pemData string, err error) {
	var usePrivate bool
	if err = survey.AskOne(&survey.Confirm{
		Message: "Are you using a private / internal CA that console and agents must trust?",
		Default: false,
	}, &usePrivate); err != nil {
		return "", err
	}
	if !usePrivate {
		return "", nil
	}

	var mode string
	if err = survey.AskOne(&survey.Select{
		Message: "How do you want to provide the private CA?",
		Options: []string{
			"Download cert chain from an HTTPS host",
			"Paste / load a PEM file path",
		},
	}, &mode); err != nil {
		return "", err
	}

	var private string
	switch mode {
	case "Download cert chain from an HTTPS host":
		var host string
		if err = survey.AskOne(&survey.Input{
			Message: "Example HTTPS URL or host using that PKI (eg https://git.corp.example):",
		}, &host, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}
		private, err = FetchTLSChain(host)
		if err != nil {
			return "", err
		}
		certs, err := parseCertificates(private)
		if err != nil {
			return "", err
		}
		fps := certificateFingerprints(certs)
		fmt.Println("Downloaded CA certificate fingerprints (SHA-256):")
		for i, fp := range fps {
			subject := certs[i].Subject.String()
			fmt.Printf("  %d. %s\n     %s\n", i+1, subject, fp)
		}
		var confirm bool
		if err = survey.AskOne(&survey.Confirm{
			Message: "Trust these CA certificates end-to-end for console and agents?",
			Default: false,
		}, &confirm); err != nil {
			return "", err
		}
		if !confirm {
			return "", fmt.Errorf("private CA download was not confirmed")
		}
	default:
		var path string
		if err = survey.AskOne(&survey.Input{
			Message: "Path to a PEM file containing your private CA / chain:",
		}, &path, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read ca file: %w", err)
		}
		private = string(data)
	}

	return MergeWithSystemRoots(private)
}
