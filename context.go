package context

import (
	"fmt"
	"github.com/ahmetson/config-lib"
	ctxConfig "github.com/ahmetson/dev-lib/config"
	"github.com/ahmetson/dev-lib/dep_manager"
	"github.com/ahmetson/dev-lib/dev"
	"github.com/ahmetson/log-lib"
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

	ctx := dev.New()

	logger, err := log.New(ctxType, true)
	if err != nil {
		return nil, fmt.Errorf("log.New(%s)", ctxType)
	}

	engine, err := config.New(logger)
	if err != nil {
		return nil, fmt.Errorf("config.New: %w", err)
	}

	ctx.SetConfig(engine)

	depManager, err := dep_manager.New(engine)
	if err != nil {
		return nil, fmt.Errorf("dep_manager.new: %w", err)
	}

	if err := ctx.SetDepManager(depManager); err != nil {
		return nil, fmt.Errorf("ctx.SetDepManager: %w", err)
	}

	return ctx, nil
}
