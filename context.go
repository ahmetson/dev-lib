package context

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	ctxConfig "github.com/ahmetson/dev-lib/base/config"
	"github.com/ahmetson/dev-lib/base/dep_manager"
	"github.com/ahmetson/dev-lib/dev"
)

type Interface interface {
	SetConfig(p configClient.Interface)
	Config() configClient.Interface
	SetDepManager(dep_manager.Interface) error
	DepManager() dep_manager.Interface
	Type() ctxConfig.ContextType
}

// A New orchestra
func New(ctxType ctxConfig.ContextType) (Interface, error) {
	if ctxType != ctxConfig.DevContext {
		return nil, fmt.Errorf("only %s supported, not %s", ctxConfig.DevContext, ctxType)
	}

	return dev.New()
}
