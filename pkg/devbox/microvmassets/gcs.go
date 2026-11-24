package microvmassets

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// openBucket opens bucketURL with the default blob mux, except gs:// URLs: those use application default
// credentials when present and otherwise the account of `gcloud auth login`, which does not configure ADC.
func openBucket(ctx context.Context, bucketURL string) (*blob.Bucket, error) {
	if !strings.HasPrefix(bucketURL, "gs://") {
		return blob.OpenBucket(ctx, bucketURL)
	}
	ts, err := gcpTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	client, err := gcp.NewHTTPClient(gcp.DefaultTransport(), ts)
	if err != nil {
		return nil, err
	}
	mux := new(blob.URLMux)
	mux.RegisterBucket(gcsblob.Scheme, &gcsblob.URLOpener{Client: client, TokenSource: ts})
	return mux.OpenBucket(ctx, bucketURL)
}

func gcpTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	creds, adcErr := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if adcErr == nil {
		return creds.TokenSource, nil
	}
	if _, err := exec.LookPath("gcloud"); err != nil {
		return nil, fmt.Errorf("no Google credentials: %w; run `gcloud auth login` or `gcloud auth application-default login`", adcErr)
	}
	src := oauth2.ReuseTokenSource(nil, gcloudTokenSource{})
	if _, err := src.Token(); err != nil {
		return nil, errors.Join(fmt.Errorf("no application default credentials: %w", adcErr), err)
	}
	return src, nil
}

// gcloudTokenSource shells out to the gcloud CLI, so it follows whichever account and impersonation `gcloud config`
// currently has active.
type gcloudTokenSource struct{}

func (gcloudTokenSource) Token() (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	out, err := exec.CommandContext(ctx, "gcloud", "auth", "print-access-token").Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("gcloud auth print-access-token: %w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("gcloud auth print-access-token: %w", err)
	}
	tok := strings.TrimSpace(string(out))
	if tok == "" {
		return nil, errors.New("gcloud auth print-access-token returned no token")
	}
	// gcloud doesn't report the expiry; its tokens last about an hour.
	return &oauth2.Token{AccessToken: tok, TokenType: "Bearer", Expiry: time.Now().Add(45 * time.Minute)}, nil
}
