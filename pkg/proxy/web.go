package proxy

import (
	"fmt"
	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural/pkg/utils"
	"time"
)

func execWeb(namespace string, proxy *v1alpha1.Proxy, kube *utils.Kube) error {
	config := proxy.Spec.WebConfig
	fwd, err := portForward(namespace, proxy, config.Port)
	if err != nil {
		return err
	}
	defer fwd.Process.Kill()

	utils.Highlight("Wait a bit while the port-forward boots up\n\n")
	time.Sleep(5 * time.Second)

	printCredentials(proxy, namespace, kube)
	fmt.Printf("\nVisit http://localhost:%d%s\n", config.Port, config.Path)
	utils.ReadLine("Press enter to close the proxy")
	return nil
}

func printCredentials(proxy *v1alpha1.Proxy, namespace string, kube *utils.Kube) error {
	creds := proxy.Spec.Credentials
	if creds == nil { 
		return nil 
	}

	secret, err := kube.Secret(namespace, creds.Secret)
	if err != nil {
		return err
	}
	user, err := fetchUserPassword(secret, creds)
	if err != nil {
		return err
	}

	highlightedEntry("Username", user.User)
	highlightedEntry("Password", user.Password)

	return nil
}

func highlightedEntry(label, value string) {
	utils.Highlight(label + ": ")
	fmt.Println(value)
}

