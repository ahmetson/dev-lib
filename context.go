package context

import (
	"fmt"
	"github.com/ahmetson/config-lib"
	ctxConfig "github.com/ahmetson/dev-lib/config"
	"github.com/ahmetson/dev-lib/dep_manager"
	"github.com/ahmetson/dev-lib/dev"
)

type Interface interface {
	SetConfig(config.Interface)
	Config() config.Interface
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
