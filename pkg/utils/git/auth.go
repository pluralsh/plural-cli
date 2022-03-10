package git

import (
	"regexp"
	cryptossh "golang.org/x/crypto/ssh"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var (
  scpLikeUrlRegExp = regexp.MustCompile(`^(?:(?P<user>[^@]+)@)?(?P<host>[^:\s]+):(?:(?P<port>[0-9]{1,5})(?:\/|:))?(?P<path>[^\\].*\/[^\\].*)$`)
)

func UrlComponents(url string) (user, host, port, path string) {
	m := scpLikeUrlRegExp.FindStringSubmatch(url)
	return m[1], m[2], m[3], m[4]
}

func BasicAuth(user, password string) (transport.AuthMethod, error) {
	return &http.BasicAuth{Username: user, Password: password}, nil
}

func SSHAuth(user, pem, passphrase string) (transport.AuthMethod, error) {
	hostKeyCallback := cryptossh.InsecureIgnoreHostKey()
	keys, err := ssh.NewPublicKeys(user, []byte(pem), passphrase)
	if err != nil {
		return keys, err
	}

	keys.HostKeyCallback = hostKeyCallback
	return keys, err
}