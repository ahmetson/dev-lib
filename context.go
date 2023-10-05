package context

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	"github.com/ahmetson/dev-lib/proxy_client"
	"github.com/ahmetson/os-lib/arg"
)

type Interface interface {
	SetConfig(p configClient.Interface)
	Config() configClient.Interface
	SetProxyClient(p proxy_client.Interface) error
	ProxyClient() proxy_client.Interface
	Type() ContextType
	StartConfig() error
	StartDepManager() error
	StartProxyHandler() error
	Close() error // Close the dep handler and config handler. The dep manager client is not closed
	Running() bool
	SetService(string, string) // SetService sets the service parameters
}

// A New orchestra. Optionally pass the type of the context.
// Or the context type could be retrieved from the config.ContextFlag.
func New(ctxTypes ...ContextType) (Interface, error) {
	ctxType := DevContext // default is used a dev context

	if len(ctxTypes) > 0 {
		ctxType = ctxTypes[0]
	} else if arg.FlagExist(ContextFlag) {
		ctxType = arg.FlagValue(ContextFlag)
	}

	if ctxType == DevContext {
		return NewDev()
	}

	return nil, fmt.Errorf("only %s supported, not %s", DevContext, ctxType)
}
