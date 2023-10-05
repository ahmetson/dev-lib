package context

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	ctxConfig "github.com/ahmetson/dev-lib/config"
	"github.com/ahmetson/dev-lib/dev"
	"github.com/ahmetson/dev-lib/proxy_client"
	"github.com/ahmetson/os-lib/arg"
)

type Interface interface {
	SetConfig(p configClient.Interface)
	Config() configClient.Interface
	SetProxyClient(p proxy_client.Interface) error
	ProxyClient() proxy_client.Interface
	Type() ctxConfig.ContextType
	StartConfig() error
	StartDepManager() error
	StartProxyHandler() error
	Close() error // Close the dep handler and config handler. The dep manager client is not closed
	Running() bool
	SetService(string, string) // SetService sets the service parameters
}

// A New orchestra. Optionally pass the type of the context.
// Or the context type could be retrieved from the config.ContextFlag.
func New(ctxTypes ...ctxConfig.ContextType) (Interface, error) {
	ctxType := ctxConfig.DevContext // default is used a dev context

	if len(ctxTypes) > 0 {
		ctxType = ctxTypes[0]
	} else if arg.FlagExist(ctxConfig.ContextFlag) {
		ctxType = arg.FlagValue(ctxConfig.ContextFlag)
	}

	if ctxType == ctxConfig.DevContext {
		return dev.New()
	}

	return nil, fmt.Errorf("only %s supported, not %s", ctxConfig.DevContext, ctxType)
}
