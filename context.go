package context

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	ctxConfig "github.com/ahmetson/dev-lib/base/config"
	"github.com/ahmetson/dev-lib/dep_client"
	"github.com/ahmetson/dev-lib/dev"
	"github.com/ahmetson/os-lib/arg"
)

type Interface interface {
	SetConfig(p configClient.Interface)
	Config() configClient.Interface
	SetDepManager(dep_client.Interface) error
	DepManager() dep_client.Interface
	Type() ctxConfig.ContextType
	Start() error
	Running() bool
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
