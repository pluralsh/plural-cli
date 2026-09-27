package edge

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

// DownloadOptions are the CLI flags for plural edge download.
type DownloadOptions struct {
	OCIURL string
	To     string
	Fetch  func(ociURL string) (io.ReadCloser, error)
	Untar  func(dst string, r io.Reader) error
}

// Download pulls an OCI image and unpacks it into To (cwd when empty).
func Download(opts DownloadOptions, log func(string)) error {
	ociURL := strings.TrimSpace(opts.OCIURL)
	if ociURL == "" {
		return fmt.Errorf("oci-url is required")
	}
	dest := strings.TrimSpace(opts.To)
	if dest == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dest = wd
	}
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}

	logf := bootstrapLog(log)
	logf("unpacking image contents to %s", dest)

	fetch := opts.Fetch
	if fetch == nil {
		fetch = fetchOCIImage
	}
	reader, err := fetch(ociURL)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := reader.Close(); cerr != nil && log != nil {
			log(cerr.Error())
		}
	}()

	untar := opts.Untar
	if untar == nil {
		untar = utils.Untar
	}
	return untar(dest, reader)
}

func fetchOCIImage(ociURL string) (io.ReadCloser, error) {
	ref, err := name.ParseReference(ociURL)
	if err != nil {
		return nil, err
	}
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, err
	}
	return mutate.Extract(img), nil
}
