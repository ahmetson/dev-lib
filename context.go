package context

import (
	"fmt"
	"github.com/ahmetson/config-lib"
	ctxConfig "github.com/ahmetson/dev-lib/config"
	"github.com/ahmetson/dev-lib/dep"
	"github.com/ahmetson/dev-lib/orchestra/dev"
	"github.com/ahmetson/log-lib"
)

type Interface interface {
	SetConfig(config.Interface)
	Config() config.Interface
	SetDepManager(dep.Interface) error
	DepManager() dep.Interface
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

	depManager, err := dep.New(engine)
	if err != nil {
		return nil, fmt.Errorf("dep.new: %w", err)
	}

	if err := ctx.SetDepManager(depManager); err != nil {
		return nil, fmt.Errorf("ctx.SetDepManager: %w", err)
	}

	return ctx, nil
}
