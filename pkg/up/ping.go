package up

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/likexian/doh"
	"github.com/likexian/doh/dns"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func testDns(domain string) error {
	ping := fmt.Sprintf("Querying %s...\n", domain)
	return retrier(ping, "DNS fully resolved, testing if console is functional...\n", func() error {
		return doTestDns(domain)
	})
}

func doTestDns(domain string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	callDns := func(t dns.Type) error {
		c := doh.Use(doh.CloudflareProvider, doh.GoogleProvider)
		rsp, err := c.Query(ctx, dns.Domain(domain), t)
		if err != nil {
			return err
		}

		// close the client
		c.Close()

		// doh dns answer
		answer := rsp.Answer
		if len(answer) > 0 {
			return nil
		}

		return fmt.Errorf("dns answer was empty")
	}

	if err := callDns(dns.TypeA); err == nil {
		return nil
	}

	if err := callDns(dns.TypeCNAME); err == nil {
		return nil
	}

	if err := callDns(dns.TypeAAAA); err == nil {
		return nil
	}

	return fmt.Errorf("could not resolve %s dns domain, you likely need to wait for this to propagate", domain)
}

func ping(url string) error {
	ping := fmt.Sprintf("Pinging %s...\n", url)
	return retrier(ping, "Found status code 200, console up!\n", func() error {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}

		return fmt.Errorf("console failed to become ready after 5 minutes, you might want to inspect the resources in the plrl-console namespace")
	})
}

// waitForConsole polls until the console deployment in plrl-console has at least
// one ready replica. This works on local/BYOK clusters where TLS or public DNS
// is not available.
func waitForConsole() error {
	return retrier(
		"Waiting for console deployment to become ready...\n",
		"Console deployment is ready!\n",
		func() error {
			kube, err := kubernetes.Kubernetes()
			if err != nil {
				return err
			}

			deploy, err := kube.GetClient().AppsV1().Deployments("plrl-console").
				Get(context.Background(), "console", metav1.GetOptions{})
			if err != nil {
				return err
			}

			if deploy.Status.ReadyReplicas > 0 {
				return nil
			}

			return fmt.Errorf("console deployment not ready yet (%d/%d replicas ready)",
				deploy.Status.ReadyReplicas, deploy.Status.Replicas)
		},
	)
}

func retrier(retryMsg, successMsg string, f func() error) error {
	done := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	var resErr error
	defer cancel()

	go func() {
		for {
			fmt.Print(retryMsg)
			err := f()
			if err == nil {
				utils.Success("%s", successMsg)
				done <- true
				return
			}
			resErr = err
			time.Sleep(10 * time.Second)
		}
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return resErr
	}
}
