package orchestra

import (
	"fmt"
	"github.com/ahmetson/config-lib"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/service-lib/dep"
	"github.com/ahmetson/service-lib/orchestra/dev"
)

type Type = string

const (
	// DevContext indicates that all dependency proxies are in the local machine
	DevContext Type = "development"
	// DefaultContext indicates that all dependencies are in any machine.
	// It's unspecified.
	DefaultContext Type = "default"
)

type Interface interface {
	SetConfig(config.Interface)
	Config() config.Interface
	SetDepManager(dep.Interface) error
	DepManager() dep.Interface
	Type() Type
}

// A New orchestra
func New(ctxType Type) (Interface, error) {
	if ctxType != DevContext {
		return nil, fmt.Errorf("only %s supported, not %s", DevContext, ctxType)
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
