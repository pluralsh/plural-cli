package workbenches

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/samber/lo"
)

type PullRequestOptions struct {
	URL      string
	Commit   string
	BaseURL  string
	Provider string
}

type PullRequestResolver struct {
	repository PullRequestRepository
	providers  []PullRequestProvider
}

type repositoryAddress struct {
	url  string
	host string
}

func NewPullRequestResolver(repository PullRequestRepository) *PullRequestResolver {
	if repository == nil {
		repository = GitPullRequestRepository{}
	}

	return &PullRequestResolver{
		repository: repository,
		providers:  defaultPullRequestProviders(),
	}
}

func (r *PullRequestResolver) Resolve(options PullRequestOptions) (string, error) {
	if options.URL != "" && options.Commit != "" {
		return "", fmt.Errorf("url and commit cannot be used together")
	}

	if options.URL != "" {
		if err := r.validateHTTPURL(options.URL); err != nil {
			return "", fmt.Errorf("invalid pull request URL: %w", err)
		}

		return options.URL, nil
	}

	if r.repository == nil {
		return "", fmt.Errorf("git repository is not configured")
	}

	var provider PullRequestProvider
	var err error
	if options.Provider != "" && options.Provider != string(ProviderAuto) {
		provider, err = r.provider(options.Provider, "")

		if err != nil {
			return "", err
		}
	}

	ref := options.Commit
	if ref == "" {
		ref = "HEAD"
	}

	subject, err := r.repository.CommitSubject(ref)
	if err != nil {
		return "", fmt.Errorf("could not read commit %q: %w", ref, err)
	}

	if options.BaseURL != "" {
		if err := r.validateHTTPURL(options.BaseURL); err != nil {
			return "", fmt.Errorf("invalid repository base URL: %w", err)
		}
	}

	address := repositoryAddress{url: options.BaseURL}
	if options.Provider == "" || options.Provider == string(ProviderAuto) || address.url == "" {
		address, err = r.repositoryAddress(options.BaseURL)
		if err != nil {
			return "", err
		}
	}

	if provider == nil {
		provider, err = r.provider(options.Provider, address.host)
		if err != nil {
			return "", err
		}
	}

	number, found := provider.PullRequestNumber(subject)
	if !found {
		return "", fmt.Errorf("commit %q does not identify a %s pull request", ref, provider.Name())
	}

	return provider.PullRequestURL(strings.TrimRight(address.url, "/"), number), nil
}

func (r *PullRequestResolver) provider(name, host string) (PullRequestProvider, error) {
	if name == "" || name == string(ProviderAuto) {
		provider, found := lo.Find(r.providers, func(provider PullRequestProvider) bool {
			return provider.Supports(host)
		})

		if !found {
			return nil, fmt.Errorf("cannot infer source control provider from host %q; provide --provider", host)
		}

		return provider, nil
	}

	provider, found := lo.Find(r.providers, func(provider PullRequestProvider) bool {
		return string(provider.Name()) == name
	})

	if !found {
		return nil, fmt.Errorf("unsupported source control provider %q", name)
	}

	return provider, nil
}

func (r *PullRequestResolver) repositoryAddress(baseURL string) (repositoryAddress, error) {
	remote, err := r.repository.RemoteURL()
	if err != nil {
		return repositoryAddress{}, fmt.Errorf("could not read origin URL: %w", err)
	}

	address, err := r.parseRepositoryAddress(remote)
	if err != nil {
		return repositoryAddress{}, err
	}

	if baseURL != "" {
		address.url = baseURL
	}

	return address, nil
}

func (*PullRequestResolver) parseRepositoryAddress(raw string) (repositoryAddress, error) {
	if raw == "" {
		return repositoryAddress{}, fmt.Errorf("origin URL is empty")
	}

	if !strings.Contains(raw, "://") {
		if at := strings.LastIndex(raw, "@"); at >= 0 {
			raw = raw[at+1:]
		}

		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return repositoryAddress{}, fmt.Errorf("cannot parse origin URL %q", raw)
		}

		host := parts[0]
		path := strings.TrimSuffix(strings.Trim(parts[1], "/"), ".git")
		return repositoryAddress{url: "https://" + host + "/" + path, host: strings.ToLower(host)}, nil
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return repositoryAddress{}, fmt.Errorf("cannot parse origin URL %q", raw)
	}

	path := strings.TrimSuffix(strings.Trim(parsed.Path, "/"), ".git")
	if path == "" {
		return repositoryAddress{}, fmt.Errorf("origin URL %q does not contain a repository path", raw)
	}

	return repositoryAddress{
		url:  "https://" + parsed.Host + "/" + path,
		host: strings.ToLower(parsed.Hostname()),
	}, nil
}

func (*PullRequestResolver) validateHTTPURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("scheme must be http or https")
	}
	if parsed.Hostname() == "" {
		return fmt.Errorf("host is required")
	}

	return nil
}
