package proxy

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	"github.com/pluralsh/plural/pkg/kubernetes/portforward"
	"github.com/pluralsh/plural/pkg/utils"
)

type postgres struct {
	Proxy *v1alpha1.Proxy
	Pwd   string
}

type dbConnection interface {
	Connect(namespace string) error
}

func buildConnection(secret string, proxy *v1alpha1.Proxy) (dbConnection, error) {
	switch proxy.Spec.DbConfig.Engine {
	case "postgres":
		return &postgres{Pwd: secret, Proxy: proxy}, nil
	default:
		return nil, fmt.Errorf("Unsupported engine %s", proxy.Spec.DbConfig.Engine)
	}
}

func (pg *postgres) Connect(namespace string) error {
	err := portForward(namespace, pg.Proxy, pg.Proxy.Spec.DbConfig.Port)
	if err != nil {
		return err
	}

	utils.Highlight("Wait a bit while the port-forward boots up\n")
	time.Sleep(5 * time.Second)
	cmd := exec.Command("psql", fmt.Sprintf("host=127.0.0.1 sslmode=allow user=%s dbname=%s", pg.Proxy.Spec.Credentials.User, pg.Proxy.Spec.DbConfig.Name))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", pg.Pwd))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func portForward(namespace string, proxy *v1alpha1.Proxy, port int32) error {
	errorChan := make(chan error, 1)
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{})
	var err error
	go func() {
		err = portforward.PortForward(namespace, proxy.Spec.Target, []string{fmt.Sprint(port)}, stopChan, readyChan)
		errorChan <- err
	}()

	select {
	case <-readyChan:
		break
	case <-errorChan:
		close(stopChan)
		close(errorChan)
		break
	}
	return err
}
