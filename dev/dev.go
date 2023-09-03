// Package dev sets up the developer context.
package dev

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	configHandler "github.com/ahmetson/config-lib/handler"
	config2 "github.com/ahmetson/dev-lib/base/config"
	depmanager2 "github.com/ahmetson/dev-lib/base/dep_manager"
	devConfig "github.com/ahmetson/dev-lib/config"
	"github.com/ahmetson/dev-lib/dep_manager"
	"github.com/ahmetson/handler-lib/base"
)

// A Context handles the config of the contexts
type Context struct {
	configClient configClient.Interface
	depManager   depmanager2.Interface
	controller   base.Interface
	serviceReady bool
}

// New creates Developer context.
// Loads it with the Dev Configuration and Dev DepManager Manager.
func New() (*Context, error) {
	ctx := &Context{}

	engine, err := configHandler.New()
	if err != nil {
		return nil, fmt.Errorf("configHandler.New: %w", err)
	}
	err = engine.Start()
	if err != nil {
		return nil, fmt.Errorf("configHandler.Start: %w", err)
	}

	socket, err := configClient.New()
	if err != nil {
		return nil, fmt.Errorf("configClient.New: %w", err)
	}
	ctx.SetConfig(socket)
	if err := devConfig.SetDevDefaults(socket); err != nil {
		return nil, fmt.Errorf("config.SetDevDefaults: %w", err)
	}

	binPath, err := socket.String(devConfig.BinKey)
	if err != nil {
		return nil, fmt.Errorf("configClient.String(%s): %w", devConfig.BinKey, err)
	}
	srcPath, err := socket.String(devConfig.SrcKey)
	if err != nil {
		return nil, fmt.Errorf("configClient.String(%s): %w", devConfig.SrcKey, err)
	}

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
func (ctx *Context) SetConfig(socket configClient.Interface) {
	ctx.configClient = socket
}

// Config returns the config engine in the context.
func (ctx *Context) Config() configClient.Interface {
	return ctx.configClient
}

// SetDepManager sets the dependency manager in the context.
func (ctx *Context) SetDepManager(depManager depmanager2.Interface) error {
	if ctx.configClient == nil {
		return fmt.Errorf("no configuration")
	}

	ctx.depManager = depManager

	return nil
}

// DepManager returns the dependency manager
func (ctx *Context) DepManager() depmanager2.Interface {
	return ctx.depManager
}

// Type returns the context type. Useful to identify contexts in the generic functions.
func (ctx *Context) Type() config2.ContextType {
	return config2.DevContext
}
