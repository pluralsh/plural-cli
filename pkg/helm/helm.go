package helm

import (
	"fmt"
	"log"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

const enableDebug = false

func debug(format string, v ...interface{}) {
	if enableDebug {
		format = fmt.Sprintf("[debug] %s\n", format)
		err := log.Output(2, fmt.Sprintf(format, v...))
		if err != nil {
			log.Panic(err)
		}
	}
}

func GetActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	settings := cli.New()
	settings.SetNamespace(namespace)
	log.Prefix()
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "", debug); err != nil {
		return nil, err
	}
	return actionConfig, nil
}
