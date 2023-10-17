// Package context sets up the developer context.
package context

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	configHandler "github.com/ahmetson/config-lib/handler"
	"github.com/ahmetson/dev-lib/dep_client"
	"github.com/ahmetson/dev-lib/dep_handler"
	"github.com/ahmetson/dev-lib/dep_manager"
	"github.com/ahmetson/dev-lib/proxy_client"
	"github.com/ahmetson/dev-lib/proxy_handler"
	"github.com/ahmetson/handler-lib/manager_client"
	"github.com/ahmetson/log-lib"
)

// A Context handles the config of the contexts
type Context struct {
	configClient        configClient.Interface
	depHandler          *dep_handler.DepHandler
	depHandlerManager   manager_client.Interface
	proxyClient         proxy_client.Interface
	proxyHandler        *proxy_handler.ProxyHandler
	proxyHandlerManager manager_client.Interface
	engineStarted       bool // is the config engine started or not?
	serviceId           string
	serviceUrl          string
	depClient           *dep_client.Client
}

// NewDev creates Developer context.
// Loads it with the Dev Configuration and Dev DepManager Manager.
func NewDev() (*Context, error) {
	ctx := &Context{}

	socket, err := configClient.New()
	if err != nil {
		return nil, fmt.Errorf("configClient.New: %w", err)
	}
	ctx.SetConfig(socket)

	return ctx, nil
}

func (ctx *Context) IsRunning() bool {
	return ctx.engineStarted && ctx.depHandlerManager != nil && ctx.proxyHandlerManager != nil
}

func (ctx *Context) IsConfigRunning() bool {
	return ctx.engineStarted
}

func (ctx *Context) IsDepManagerRunning() bool {
	return ctx.depHandlerManager != nil
}

func (ctx *Context) IsProxyHandlerRunning() bool {
	return ctx.proxyHandlerManager != nil
}

func (ctx *Context) SetDepClient(dc dep_client.Interface) error {
	devDc, ok := dc.(*dep_client.Client)
	if !ok {
		return fmt.Errorf("only dev context dep client supported")
	}
	ctx.depClient = devDc

	return nil
}

func (ctx *Context) DepClient() dep_client.Interface {
	return ctx.depClient
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

// SetProxyClient sets the client that works with proxies
func (ctx *Context) SetProxyClient(proxyClient proxy_client.Interface) error {
	if ctx.configClient == nil {
		return fmt.Errorf("no configuration")
	}

	ctx.proxyClient = proxyClient

	return nil
}

// ProxyClient returns the client that works with proxies
func (ctx *Context) ProxyClient() proxy_client.Interface {
	return ctx.proxyClient
}

// Type returns the context type. Useful to identify contexts in the generic functions.
func (ctx *Context) Type() ContextType {
	return DevContext
}

// Close the dep handler and config handler. The dep manager client is not closed
func (ctx *Context) Close() error {
	if ctx.proxyHandlerManager != nil {
		if ctx.proxyHandlerManager != nil {
			if err := ctx.proxyHandlerManager.Close(); err != nil {
				return fmt.Errorf("ctx.proxyHandlerManager.Close: %w", err)
			}
		}
		ctx.proxyHandlerManager = nil
	}

	if ctx.depHandlerManager != nil {
		if err := ctx.depHandlerManager.Close(); err != nil {
			return fmt.Errorf("ctx.depHandlerManager.Close: %w", err)
		}
		ctx.depHandlerManager = nil
	}

	if ctx.engineStarted {
		if err := ctx.Config().Close(); err != nil {
			return fmt.Errorf("ctx.Config.Close: %w", err)
		}
		ctx.engineStarted = false
	}

	return nil
}

// SetService sets the service id and url for which this context belongs too.
func (ctx *Context) SetService(id string, url string) {
	ctx.serviceId = id
	ctx.serviceUrl = url
}

// StartConfig starts the config engine.
func (ctx *Context) StartConfig() error {
	if ctx.engineStarted {
		return fmt.Errorf("config engine already started")
	}

	engine, err := configHandler.New()
	if err != nil {
		return fmt.Errorf("configHandler.New: %w", err)
	}

	if err := engine.Start(); err != nil {
		return fmt.Errorf("configHandler.Start: %w", err)
	}

	ctx.engineStarted = true

	return nil
}

// StartDepManager starts the dependency manager
func (ctx *Context) StartDepManager() error {
	if !ctx.engineStarted {
		return fmt.Errorf("config engine not started. call StartConfig first")
	}
	if ctx.depHandlerManager != nil {
		return fmt.Errorf("dep manager already started")
	}
	// Set the development context parameters
	if err := SetDevDefaults(ctx.configClient); err != nil {
		return fmt.Errorf("config.SetDevDefaults: %w", err)
	}
	binPath, err := ctx.configClient.String(BinKey)
	if err != nil {
		return fmt.Errorf("configClient.String(%s): %w", BinKey, err)
	}
	srcPath, err := ctx.configClient.String(SrcKey)
	if err != nil {
		return fmt.Errorf("configClient.String(%s): %w", SrcKey, err)
	}

	//
	// Start the dependency manager
	//
	depManager := dep_manager.New()
	if err := depManager.SetPaths(binPath, srcPath); err != nil {
		return fmt.Errorf("depManager.SetPaths('%s', '%s'): %w", binPath, srcPath, err)
	}
	ctx.depHandler, err = dep_handler.New(depManager)
	if err != nil {
		return fmt.Errorf("dep_handler.New: %w", err)
	}

	err = ctx.depHandler.Start()
	if err != nil {
		return fmt.Errorf("depHandler: %w", err)
	}

	ctx.depHandlerManager, err = manager_client.New(dep_handler.ServiceConfig())
	if err != nil {
		return fmt.Errorf("manager_client.New('dep_handler'): %w", err)
	}

	depClient, err := dep_client.New()
	if err != nil {
		return fmt.Errorf("dep_client.New: %w", err)
	}

	err = ctx.SetDepClient(depClient)
	if err != nil {
		return fmt.Errorf("ctx.SetDepClient: %w", err)
	}

	return nil
}

// StartProxyHandler starts the proxy handler
func (ctx *Context) StartProxyHandler() error {
	if len(ctx.serviceId) == 0 || len(ctx.serviceUrl) == 0 {
		return fmt.Errorf("service parameters are not set. call Context.SetService first")
	}
	if !ctx.engineStarted {
		return fmt.Errorf("config engine not started. call StartConfig first")
	}
	if ctx.proxyHandlerManager != nil {
		return fmt.Errorf("proxy handler already started")
	}
	proxyLogger, err := log.New("proxy-handler", true)
	if err != nil {
		return fmt.Errorf("log.New('proxy-handler'): %w", err)
	}

	depClient, err := dep_client.New()
	if err != nil {
		return fmt.Errorf("dep_client.New: %w", err)
	}

	proxyHandler := proxy_handler.New(ctx.configClient, depClient)
	proxyHandlerConfig := proxy_handler.HandlerConfig(ctx.serviceId)
	proxyHandler.SetConfig(proxyHandlerConfig)
	err = proxyHandler.SetLogger(proxyLogger)
	if err != nil {
		return fmt.Errorf("proxyHandler.SetLogger: %w", err)
	}
	proxyHandler.SetServiceId(ctx.serviceId)
	err = proxyHandler.Start()
	if err != nil {
		return fmt.Errorf("proxyHandler.Start: %w", err)
	}
	ctx.proxyHandler = proxyHandler

	ctx.proxyHandlerManager, err = manager_client.New(proxyHandlerConfig)
	if err != nil {
		return fmt.Errorf("manager_client.New('proxyHandlerConfig'): %w", err)
	}
	proxyClient, err := proxy_client.New(ctx.serviceId)
	if err != nil {
		return fmt.Errorf("proxy_client.New('%s'): %w", ctx.serviceId, err)
	}
	err = ctx.SetProxyClient(proxyClient)
	if err != nil {
		return fmt.Errorf("ctx.SetProxyClient: %w", err)
	}

	return nil
}
