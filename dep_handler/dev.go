// Package dep_handler creates a thread that manages the dependencies
package dep_handler

import (
	"fmt"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/datatype-lib/message"
	"github.com/ahmetson/dev-lib/base/dep_manager"
	"github.com/ahmetson/dev-lib/source"
	"github.com/ahmetson/handler-lib/base"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"github.com/ahmetson/handler-lib/replier"
	"github.com/ahmetson/log-lib"
)

const (
	Category     = "dep_handler"   // handler category
	DepInstalled = "dep-installed" // the command to check is dependency installed
	DepRunning   = "dep-running"   // the command to check is dependency running
	InstallDep   = "install-dep"   // the command to install the dependency
	RunDep       = "run-dep"       // the command to run the dependency
	UninstallDep = "uninstall-dep" // the command to remove the dependency binary. if possible, then remove the source code as well.
	CloseDep     = "close-dep"     // the command to stop the running dependency
)

type DepHandler struct {
	handler base.Interface
	manager dep_manager.Interface
	logger  *log.Logger
}

// ServiceConfig returns the socket configuration of the handler
func ServiceConfig() *handlerConfig.Handler {
	return handlerConfig.NewInternalHandler(handlerConfig.ReplierType, Category)
}

// New dep handler returned
func New(manager dep_manager.Interface) (*DepHandler, error) {
	handler := replier.New()

	logger, err := log.New("dep_manager", true)
	if err != nil {
		return nil, fmt.Errorf("log.New('dep-handler'): %w", err)
	}

	handler.SetConfig(ServiceConfig())
	err = handler.SetLogger(logger)
	if err != nil {
		return nil, fmt.Errorf("handler.SetLogger: %w", err)
	}

	return &DepHandler{
		manager: manager,
		handler: handler,
		logger:  logger,
	}, nil
}

// onDepInstalled checks whether the dependency installed or not.
// Requires the 'url' of the dependency.
//
// Returns 'installed' boolean parameter
func (h *DepHandler) onDepInstalled(req message.RequestInterface) message.ReplyInterface {
	url, err := req.RouteParameters().StringValue("url")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('url'): %v", err))
	}

	installed := h.manager.Installed(url)
	params := key_value.New().Set("installed", installed)
	return req.Ok(params)
}

// onDepRunning checks whether the dependency is running or not.
// Requires 'dep' of the clientConfig.Client.
//
// Returns 'running' boolean result
func (h *DepHandler) onDepRunning(req message.RequestInterface) message.ReplyInterface {
	kv, err := req.RouteParameters().NestedValue("dep")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetKeyValue('client'): %v", err))
	}

	var c clientConfig.Client
	err = kv.Interface(&c)
	if err != nil {
		return req.Fail(fmt.Sprintf("kv.Interface: %v", err))
	}

	c.UrlFunc(clientConfig.Url)

	running, err := h.manager.Running(&c)
	if err != nil {
		return req.Fail(fmt.Sprintf("h.manager.Running: %v", err))
	}

	params := key_value.New().Set("running", running)
	return req.Ok(params)
}

// onInstallDep installs the dependency. if it comes with the source code, then build that as well.
//
// Requires the 'src' of a source.Src type, returns nothing.
//
// todo create a publisher that publishes the result of the installation, so user won't wait until installation.
func (h *DepHandler) onInstallDep(req message.RequestInterface) message.ReplyInterface {
	kv, err := req.RouteParameters().NestedValue("src")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetKeyValue('client'): %v", err))
	}

	var src source.Src
	err = kv.Interface(&src)
	if err != nil {
		return req.Fail(fmt.Sprintf("kv.Interface: %v", err))
	}

	err = h.manager.Install(&src, h.logger)
	if err != nil {
		return req.Fail(fmt.Sprintf("h.manager.Install: %v", err))
	}

	return req.Ok(key_value.New())
}

// onRunDep runs the dependency.
// Requires 'url' string parameter, 'id' string parameter, and the 'parent' of the clientConfig.Client type.
// Returns nothing.
//
// todo make it publish the result through publisher, so user won't wait for the result.
func (h *DepHandler) onRunDep(req message.RequestInterface) message.ReplyInterface {
	kv, err := req.RouteParameters().NestedValue("parent")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetKeyValue('parent'): %v", err))
	}

	var parent clientConfig.Client
	err = kv.Interface(&parent)
	if err != nil {
		return req.Fail(fmt.Sprintf("kv.Interface: %v", err))
	}

	parent.UrlFunc(clientConfig.Url)

	url, err := req.RouteParameters().StringValue("url")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('url'): %v", err))
	}

	id, err := req.RouteParameters().StringValue("id")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('id'): %v", err))
	}

	err = h.manager.Run(url, id, &parent)
	if err != nil {
		return req.Fail(fmt.Sprintf("h.manager.Start(url: '%s', id: '%s'): %v", url, id, err))
	}

	return req.Ok(key_value.New())
}

// onUninstallDep uninstalls the dependency binary. if it comes with the source code, then deletes source code as well.
//
// Requires the 'src' of a source.Src type, returns nothing.
//
// todo creates a publisher that publishes the result of the installation, so user won't wait until installation.
func (h *DepHandler) onUninstallDep(req message.RequestInterface) message.ReplyInterface {
	kv, err := req.RouteParameters().NestedValue("src")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetKeyValue('client'): %v", err))
	}

	var src source.Src
	err = kv.Interface(&src)
	if err != nil {
		return req.Fail(fmt.Sprintf("kv.Interface: %v", err))
	}

	err = h.manager.Uninstall(&src)
	if err != nil {
		return req.Fail(fmt.Sprintf("h.manager.Uninstall: %v", err))
	}

	return req.Ok(key_value.New())
}

// onCloseDep stops the dependency.
// Requires 'source' of the clientConfig.Client type.
// Returns nothing.
//
// Todo make it publish the result through publisher, so user won't wait for the result.
func (h *DepHandler) onCloseDep(req message.RequestInterface) message.ReplyInterface {
	kv, err := req.RouteParameters().NestedValue("dep")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetKeyValue('dep'): %v", err))
	}

	var c clientConfig.Client
	err = kv.Interface(&c)
	if err != nil {
		return req.Fail(fmt.Sprintf("kv.Interface: %v", err))
	}

	c.UrlFunc(clientConfig.Url)

	err = h.manager.Close(&c)
	if err != nil {
		return req.Fail(fmt.Sprintf("h.manager.Close: %v", err))
	}

	return req.Ok(key_value.New())
}

// Start the dependency handler with the available operations.
func (h *DepHandler) Start() error {
	if err := h.handler.Route(DepInstalled, h.onDepInstalled); err != nil {
		return fmt.Errorf("h.handler.Route('%s'): %v", DepInstalled, err)
	}
	if err := h.handler.Route(DepRunning, h.onDepRunning); err != nil {
		return fmt.Errorf("h.handler.Route('%s'): %v", DepRunning, err)
	}
	if err := h.handler.Route(InstallDep, h.onInstallDep); err != nil {
		return fmt.Errorf("h.handler.Route('%s'): %v", InstallDep, err)
	}
	if err := h.handler.Route(RunDep, h.onRunDep); err != nil {
		return fmt.Errorf("h.handler.Route('%s'): %v", RunDep, err)
	}
	if err := h.handler.Route(UninstallDep, h.onUninstallDep); err != nil {
		return fmt.Errorf("h.handler.Route('%s'): %v", UninstallDep, err)
	}
	if err := h.handler.Route(CloseDep, h.onCloseDep); err != nil {
		return fmt.Errorf("h.handler.Route('%s'): %v", CloseDep, err)
	}

	return h.handler.Start()
}
