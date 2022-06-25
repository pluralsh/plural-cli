package git

import (
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	cryptossh "golang.org/x/crypto/ssh"
)

var (
	scpLikeUrlRegExp = regexp.MustCompile(`^(?:(?P<user>[^@]+)@)?(?P<host>[^:\s]+):(?:(?P<port>[0-9]{1,5})(?:\/|:))?(?P<path>[^\\].*\/[^\\].*)$`)
)

func UrlComponents(url string) (user, host, port, path string, err error) {
	m := scpLikeUrlRegExp.FindStringSubmatch(url)
	if len(m) < 5 {
		err = fmt.Errorf("%s is not a valid git ssh url", url)
		return
	}

	return m[1], m[2], m[3], m[4], nil
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
