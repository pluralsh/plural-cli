// Package edge exposes Console-backed edge image build, flash, bootstrap, and download for the TUI.
package edge

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	pluralclient "github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/console"
	pkgedge "github.com/pluralsh/plural-cli/pkg/edge"
)

var errNoConsole = errors.New("connect a Console profile before building an edge image")

type ConsoleResolver interface {
	ActiveConsole(context.Context) (url, token string, err error)
}

type API interface {
	pkgedge.ConsoleAPI
	pkgedge.BootstrapAPI
}

type ClientFactory func(token, url string) (API, error)

type Loader interface {
	BuildImage(ctx context.Context, options pkgedge.ImageOptions, log func(string)) error
	Flash(ctx context.Context, options pkgedge.FlashOptions, log func(string)) error
	Bootstrap(ctx context.Context, options pkgedge.BootstrapOptions, log func(string)) error
	Download(ctx context.Context, options pkgedge.DownloadOptions, log func(string)) error
}

type Service struct {
	resolve   ConsoleResolver
	newClient ClientFactory
	flash     func(pkgedge.FlashOptions) error
	install   pkgedge.InstallOperatorFunc
	download  func(pkgedge.DownloadOptions, func(string)) error
}

func NewService(resolve ConsoleResolver) *Service {
	return &Service{
		resolve: resolve,
		newClient: func(token, url string) (API, error) {
			return console.NewConsoleClient(token, url)
		},
		flash:    pkgedge.Flash,
		download: pkgedge.Download,
	}
}

func (s *Service) client(ctx context.Context) (API, error) {
	if s.resolve == nil {
		return nil, &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoConsole}
	}
	url, token, err := s.resolve.ActiveConsole(ctx)
	if err != nil {
		return nil, err
	}
	return s.newClient(token, url)
}

func (s *Service) BuildImage(ctx context.Context, options pkgedge.ImageOptions, log func(string)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var client pkgedge.ConsoleAPI
	if strings.TrimSpace(options.CloudConfig) == "" {
		api, err := s.client(ctx)
		if err != nil {
			return err
		}
		client = api
		if options.ConsoleURL == "" && s.resolve != nil {
			if url, _, err := s.resolve.ActiveConsole(ctx); err == nil {
				options.ConsoleURL = url
			}
		}
	}
	svc := pkgedge.NewService(client)
	if log != nil {
		svc.Log = log
		reader, writer := io.Pipe()
		svc.Output = writer
		done := make(chan struct{})
		go func() {
			defer close(done)
			scanLines(reader, log)
		}()
		defer func() {
			_ = writer.Close()
			<-done
		}()
	}
	return svc.Build(options)
}

func (s *Service) Flash(ctx context.Context, options pkgedge.FlashOptions, log func(string)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if log != nil {
		log("flashing " + options.Image + " onto " + options.Device)
		reader, writer := io.Pipe()
		options.Log = writer
		done := make(chan struct{})
		go func() {
			defer close(done)
			scanLines(reader, log)
		}()
		defer func() {
			_ = writer.Close()
			<-done
		}()
	}
	flash := s.flash
	if flash == nil {
		flash = pkgedge.Flash
	}
	return flash(options)
}

func (s *Service) Bootstrap(ctx context.Context, options pkgedge.BootstrapOptions, log func(string)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	api, err := s.client(ctx)
	if err != nil {
		return err
	}
	install := s.install
	if install == nil {
		install = operatorInstaller(api)
	}
	return pkgedge.Bootstrap(ctx, api, install, options, log)
}

func (s *Service) Download(ctx context.Context, options pkgedge.DownloadOptions, log func(string)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	download := s.download
	if download == nil {
		download = pkgedge.Download
	}
	return download(options, log)
}

func operatorInstaller(api API) pkgedge.InstallOperatorFunc {
	return func(url, token, chartLoc, clusterID string) error {
		prev, had := os.LookupEnv("PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS")
		_ = os.Setenv("PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS", "true")
		defer func() {
			if had {
				_ = os.Setenv("PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS", prev)
			} else {
				_ = os.Unsetenv("PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS")
			}
		}()
		p := &pluralclient.Plural{}
		if cc, ok := api.(console.ConsoleClient); ok {
			p.ConsoleClient = cc
		}
		return p.DoInstallOperator(url, token, "", chartLoc, clusterID)
	}
}

func scanLines(r io.Reader, log func(string)) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line != "" {
			log(line)
		}
	}
}
