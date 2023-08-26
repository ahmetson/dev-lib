package dev

//
// The orchestra handler has only one command.
//
// Close
// this command has no arguments. And when it's given, it will close all the dependencies it has
//
// ServerReady
// marks the service as ready.

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/common-lib/message"
	"github.com/ahmetson/handler-lib/config"
	"github.com/ahmetson/handler-lib/sync_replier"
	"github.com/ahmetson/log-lib"
	"github.com/pebbe/zmq4"
)

const (
	CtxManager    = "context_handler"
	CtxManagerUrl = "inproc://context_handler" // only one context per
)

// onClose closing single the dependency in the orchestra.
func (ctx *Context) onClose(request message.Request) message.Reply {
	kv, err := request.Parameters.GetKeyValue("dep")
	if err != nil {
		return request.Fail(fmt.Sprintf("request.Parameters.GetKeyValue('dep'): %v", err))
	}

	var c *clientConfig.Client
	err = kv.Interface(c)
	if err != nil {
		return request.Fail(fmt.Sprintf("key_value.KeyValue('dep').Interface(): %v", err))
	}

	err = ctx.depManager.Close(c)
	if err != nil {
		return request.Fail(fmt.Sprintf("ctx.depManager.Close(): %v", err))
	}

	// since we closed the service, for the orchestra the service is not ready.
	// the service should call itself
	ctx.serviceReady = false

	return request.Ok(key_value.Empty())
}

// onServiceReady marks the main service to be ready.
func (ctx *Context) onServiceReady(request message.Request) message.Reply {
	if ctx.serviceReady {
		return request.Fail("main service was set as true in the orchestra")
	}
	ctx.serviceReady = true
	return request.Ok(key_value.Empty())
}

// Run the orchestra in the background. If it failed to run, then return an error.
// The url request is the main service to which this orchestra belongs too.
//
// The logger is the handler logger as it is. The orchestra will create its own logger from it.
func (ctx *Context) Run(logger *log.Logger) error {
	replier := sync_replier.New()

	replier.SetConfig(config.NewInternalHandler(zmq4.REP, CtxManager))

	err := replier.Route("close", ctx.onClose)
	if err != nil {
		return fmt.Errorf(`replier.AddRoute("close"): %w`, err)
	}
	err = replier.Route("service_ready", ctx.onServiceReady)
	if err != nil {
		return fmt.Errorf(`replier.AddRoute("service-ready"): %w`, err)
	}

	ctx.controller = replier
	go func() {
		if err := ctx.controller.Start(); err != nil {
			logger.Fatal("orchestra.handler.Run: %w", err)
		}
	}()

	return nil
}

// Close sends a close signal to the dependency through dependency manager
//
// todo move out from the handler
func (ctx *Context) Close(c *clientConfig.Client, logger *log.Logger) error {
	if ctx.controller == nil {
		logger.Warn("skipping, since orchestra.ControllerCategory is not initialised", "todo", "call orchestra.Run()")
		return nil
	}

	kv, err := key_value.NewFromInterface(c)
	if err != nil {
		return fmt.Errorf("key_value.NewFromInterface: %w", err)
	}

	contextClient, err := client.NewRaw(zmq4.REP, CtxManagerUrl)
	if err != nil {
		logger.Error("client.NewReq", "error", err)
		return fmt.Errorf("close the service by hand. client.NewReq: %w", err)
	}

	closeRequest := &message.Request{
		Command:    "close",
		Parameters: key_value.Empty().Set("dep", kv),
	}

	if err := contextClient.Submit(closeRequest); err != nil {
		logger.Error("contextClient.RequestRemoteService", "error", err)
		return fmt.Errorf("contextClient.Submit('close'): %w", err)
	}

	// release the orchestra parameters
	if err := contextClient.Close(); err != nil {
		logger.Error("contextClient.Close", "error", err)
		return fmt.Errorf("contextClient.Close: %w", err)
	}

	return nil
}

// ServiceReady sends a signal marking that the main service is ready.
func (ctx *Context) ServiceReady(logger *log.Logger) error {
	if ctx.controller == nil {
		logger.Warn("orchestra.ControllerCategory is not initialised", "todo", "call orchestra.Run()")
		return nil
	}
	contextClient, err := client.NewRaw(zmq4.REP, CtxManagerUrl)
	if err != nil {
		return fmt.Errorf("close the service by hand. client.NewReq: %w", err)
	}

	closeRequest := &message.Request{
		Command:    "service-ready",
		Parameters: key_value.Empty(),
	}

	if err := contextClient.Submit(closeRequest); err != nil {
		return fmt.Errorf("contextClient.Submit('service-ready'): %w", err)
	}

	// release the orchestra parameters
	if err := contextClient.Close(); err != nil {
		return fmt.Errorf("contextClient.Close: %w", err)
	}

	return nil
}
