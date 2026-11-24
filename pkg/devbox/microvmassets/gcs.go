package microvmassets

import (
	"context"
	"encoding/json"
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
	src := oauth2.ReuseTokenSource(nil, &gcloudTokenSource{ctx: ctx, run: runGcloud})
	if _, err := src.Token(); err != nil {
		return nil, errors.Join(fmt.Errorf("no application default credentials: %w", adcErr), err)
	}
	return src, nil
}

// fallbackTokenLifetime is how long a token is cached when gcloud does not report its expiry (it omits it for
// impersonated and some service account credentials). Short, so a token that is nearly expired is refreshed.
const fallbackTokenLifetime = 5 * time.Minute

// gcloudTokenSource shells out to the gcloud CLI, so it follows whichever account and impersonation `gcloud config`
// currently has active. Its subprocess is bound to ctx, so deadlines and Ctrl-C apply.
type gcloudTokenSource struct {
	ctx context.Context
	run func(ctx context.Context) ([]byte, error)
}

func runGcloud(ctx context.Context) ([]byte, error) {
	out, err := exec.CommandContext(ctx, "gcloud", "auth", "print-access-token", "--format=json(token,token_expiry)").Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("gcloud auth print-access-token: %w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("gcloud auth print-access-token: %w", err)
	}
	return out, nil
}

func (g *gcloudTokenSource) Token() (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(g.ctx, time.Minute)
	defer cancel()
	out, err := g.run(ctx)
	if err != nil {
		return nil, err
	}
	var res struct {
		Token       string `json:"token"`
		TokenExpiry string `json:"token_expiry"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, fmt.Errorf("parsing gcloud auth print-access-token output: %w", err)
	}
	if res.Token == "" {
		return nil, errors.New("gcloud auth print-access-token returned no token")
	}
	expiry := time.Now().Add(fallbackTokenLifetime)
	if t, err := time.Parse(time.RFC3339, res.TokenExpiry); err == nil {
		expiry = t
	}
	return &oauth2.Token{AccessToken: res.Token, TokenType: "Bearer", Expiry: expiry}, nil
}
