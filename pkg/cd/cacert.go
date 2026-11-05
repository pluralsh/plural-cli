package cd

import (
	"crypto/tls"
	"crypto/x509"
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
func FetchTLSChain(rawURL string) (string, error) {
	hostPort, err := tlsDialAddress(rawURL)
	if err != nil {
		return "", err
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		hostPort,
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

// MergeWithSystemRoots appends OS/browser-trusted roots to privatePEM.
func MergeWithSystemRoots(privatePEM string) (string, error) {
	privatePEM = strings.TrimSpace(privatePEM)
	if privatePEM == "" {
		return "", fmt.Errorf("empty private CA PEM")
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
	for _, path := range systemCABundlePaths {
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}
	// last resort: ensure SystemCertPool works even if we cannot export PEMs
	if pool, err := x509.SystemCertPool(); err == nil && pool != nil {
		return nil, fmt.Errorf("found a system cert pool but no PEM bundle file; set SSL_CERT_FILE or provide a full merged bundle via --private-ca-file")
	}
	return nil, fmt.Errorf("could not locate a system CA bundle (tried %s)", strings.Join(systemCABundlePaths, ", "))
}

// IndentPEM indents every line for embedding under a YAML `|` block.
func IndentPEM(pemData string, spaces int) string {
	pad := strings.Repeat(" ", spaces)
	lines := strings.Split(strings.TrimRight(pemData, "\n"), "\n")
	for i, line := range lines {
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n") + "\n"
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

// askPrivateCA prompts whether to configure a private CA and returns an indented
// merged PEM suitable for values templates, or empty if skipped.
func askPrivateCA() (indentedPEM string, err error) {
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

	merged, err := MergeWithSystemRoots(private)
	if err != nil {
		return "", err
	}
	return IndentPEM(merged, 6), nil
}
