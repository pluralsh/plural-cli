// Package edge exposes Console-backed edge image build and flash for the TUI.
package edge

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
	pkgedge "github.com/pluralsh/plural-cli/pkg/edge"
)

var errNoConsole = errors.New("connect a Console profile before building an edge image")

type ConsoleResolver interface {
	ActiveConsole(context.Context) (url, token string, err error)
}

type API interface {
	pkgedge.ConsoleAPI
}

type ClientFactory func(token, url string) (API, error)

type Loader interface {
	BuildImage(ctx context.Context, options pkgedge.ImageOptions, log func(string)) error
	Flash(ctx context.Context, options pkgedge.FlashOptions, log func(string)) error
}

type Service struct {
	resolve   ConsoleResolver
	newClient ClientFactory
	flash     func(pkgedge.FlashOptions) error
}

func NewService(resolve ConsoleResolver) *Service {
	return &Service{
		resolve: resolve,
		newClient: func(token, url string) (API, error) {
			return console.NewConsoleClient(token, url)
		},
		flash: pkgedge.Flash,
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
