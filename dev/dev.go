// Package dev sets up the developer context.
package dev

import (
	"fmt"
	"github.com/ahmetson/config-lib"
	config2 "github.com/ahmetson/dev-lib/base/config"
	dep_manager2 "github.com/ahmetson/dev-lib/base/dep_manager"
	devConfig "github.com/ahmetson/dev-lib/config"
	"github.com/ahmetson/dev-lib/dep_manager"
	"github.com/ahmetson/handler-lib/base"
)

// A Context handles the config of the contexts
type Context struct {
	engine       config.Interface
	depManager   dep_manager2.Interface
	controller   base.Interface
	serviceReady bool
}

// New creates Developer context.
// Loads it with the Dev Configuration and Dev DepManager Manager.
func New() (*Context, error) {
	ctx := &Context{}

	engine, err := config.NewDev()
	if err != nil {
		return nil, fmt.Errorf("config.New: %w", err)
	}

	ctx.SetConfig(engine)
	if err := devConfig.SetDevDefaults(engine); err != nil {
		return nil, fmt.Errorf("config.SetDevDefaults: %w", err)
	}

	binPath := engine.GetString(devConfig.BinKey)
	srcPath := engine.GetString(devConfig.SrcKey)

	depManager, err := dep_manager.New(srcPath, binPath)
	if err != nil {
		return nil, fmt.Errorf("dep_manager.new: %w", err)
	}

	if err := ctx.SetDepManager(depManager); err != nil {
		return nil, fmt.Errorf("ctx.SetDepManager: %w", err)
	}

	return ctx, nil
}

// SetConfig sets the config engine of the given type.
// For the development context, it could be config-lib that reads the local file system.
//
// Setting up the configuration prepares the context by creating directories.
func (ctx *Context) SetConfig(engine config.Interface) {
	ctx.engine = engine
}

// Config returns the config engine in the context.
func (ctx *Context) Config() config.Interface {
	return ctx.engine
}

// SetDepManager sets the dependency manager in the context.
func (ctx *Context) SetDepManager(depManager dep_manager2.Interface) error {
	if ctx.engine == nil {
		return fmt.Errorf("no configuration")
	}

	ctx.depManager = depManager

	return nil
}

// DepManager returns the dependency manager
func (ctx *Context) DepManager() dep_manager2.Interface {
	return ctx.depManager
}

// Type returns the context type. Useful to identify contexts in the generic functions.
func (ctx *Context) Type() config2.ContextType {
	return config2.DevContext
}
