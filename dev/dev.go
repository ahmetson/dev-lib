// Package dev sets up the developer context.
package dev

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	configHandler "github.com/ahmetson/config-lib/handler"
	config2 "github.com/ahmetson/dev-lib/base/config"
	devConfig "github.com/ahmetson/dev-lib/config"
	"github.com/ahmetson/dev-lib/dep_handler"
	"github.com/ahmetson/dev-lib/dep_manager"
	"github.com/ahmetson/handler-lib/base"
)

// A Context handles the config of the contexts
type Context struct {
	configClient configClient.Interface
	depClient    configClient.Interface
	controller   base.Interface
	serviceReady bool
}

// New creates Developer context.
// Loads it with the Dev Configuration and Dev DepManager Manager.
func New() (*Context, error) {
	ctx := &Context{}

	// todo make it call after config handler initiated.
	// config handler broadcasts the status.
	// once its ready, the client is created.
	// then dep manager is created.
	socket, err := configClient.New()
	if err != nil {
		return nil, fmt.Errorf("configClient.New: %w", err)
	}
	ctx.SetConfig(socket)

	// todo once we have depClient, use it here
	//if err := ctx.SetDepManager(depManager); err != nil {
	//	return nil, fmt.Errorf("ctx.SetDepManager: %w", err)
	//}

	return ctx, nil
}

// SetConfig sets the config engine of the given type.
// For the development context, it could be config-lib that reads the local file system.
//
// Setting up the configuration prepares the context by creating directories.
func (ctx *Context) SetConfig(socket configClient.Interface) {
	ctx.configClient = socket
}

// Config returns the config engine in the context.
func (ctx *Context) Config() configClient.Interface {
	return ctx.configClient
}

// SetDepManager sets the dependency manager in the context.
func (ctx *Context) SetDepManager(depClient configClient.Interface) error {
	if ctx.configClient == nil {
		return fmt.Errorf("no configuration")
	}

	ctx.depClient = depClient

	return nil
}

// DepManager returns the dependency manager
func (ctx *Context) DepManager() configClient.Interface {
	return ctx.depClient
}

// Type returns the context type. Useful to identify contexts in the generic functions.
func (ctx *Context) Type() config2.ContextType {
	return config2.DevContext
}

// Start the context
func (ctx *Context) Start() error {
	engine, err := configHandler.New()
	if err != nil {
		return fmt.Errorf("configHandler.New: %w", err)
	}

	err = engine.Start()
	if err != nil {
		return fmt.Errorf("configHandler.Start: %w", err)
	}

	if err := devConfig.SetDevDefaults(ctx.configClient); err != nil {
		return fmt.Errorf("config.SetDevDefaults: %w", err)
	}

	binPath, err := ctx.configClient.String(devConfig.BinKey)
	if err != nil {
		return fmt.Errorf("configClient.String(%s): %w", devConfig.BinKey, err)
	}
	srcPath, err := ctx.configClient.String(devConfig.SrcKey)
	if err != nil {
		return fmt.Errorf("configClient.String(%s): %w", devConfig.SrcKey, err)
	}

	depManager := dep_manager.New()
	if err := depManager.SetPaths(binPath, srcPath); err != nil {
		return fmt.Errorf("depManager.SetPaths('%s', '%s'): %w", binPath, srcPath, err)
	}
	depHandler, err := dep_handler.New(depManager)
	if err != nil {
		return fmt.Errorf("dep_handler.New: %w", err)
	}

	err = depHandler.Start()
	if err != nil {
		return fmt.Errorf("depHandler: %w", err)
	}

	return nil
}
