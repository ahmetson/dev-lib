// Package proxy_handler is a thread that manages the service proxies
package proxy_handler

import (
	"fmt"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/datatype-lib/message"
	"github.com/ahmetson/handler-lib/base"
	handlerConfig "github.com/ahmetson/handler-lib/config"
)

const (
	Category = "proxy_handler" // handler category
)

type ProxyHandler struct {
	*base.Handler
	serviceId   string // the parent id
	serviceUrl  string // the parent url for classification
	proxyChains []*service.ProxyChain
}

// Id of the proxy handler based on the service id
func Id(id string) string {
	return fmt.Sprintf("%s_proxy_handler", id)
}

// HandlerConfig creates a configuration of the thread
func HandlerConfig(serviceId string) *handlerConfig.Handler {
	inprocConfig := handlerConfig.NewInternalHandler(handlerConfig.SyncReplierType, Category)
	inprocConfig.Id = Id(serviceId)

	return inprocConfig
}

// New returns a proxy handler
func New() *ProxyHandler {
	newHandler := base.New()
	return &ProxyHandler{
		Handler:     newHandler,
		proxyChains: make([]*service.ProxyChain, 0),
	}
}

// SetService sets the service parameters that this context belongs too
func (proxyHandler *ProxyHandler) SetService(id string, url string) {
	proxyHandler.serviceId = id
	proxyHandler.serviceUrl = url
}

// The Route is over-written to be disabled.
func (proxyHandler *ProxyHandler) Route(string, any, ...string) error {
	return fmt.Errorf("not implemented")
}

// onSetProxyChain registers the new proxy chain
func (proxyHandler *ProxyHandler) onSetProxyChain(req message.RequestInterface) message.ReplyInterface {
	raw, err := req.RouteParameters().NestedValue("proxy_chain")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.RouteParameters().NestedValue('proxy_chain'): %v", err))
	}

	var proxyChain service.ProxyChain
	err = raw.Interface(&proxyChain)
	if err != nil {
		return req.Fail(fmt.Sprintf("key_value.KeyValue('proxy_chain').Interface(): %v", err))
	}

	// Feel the missing urls
	if proxyChain.Destination == nil {
		return req.Fail("the 'Destination' and 'Proxies' fields of 'proxy_chain' are required")
	}

	if len(proxyChain.Destination.Urls) == 0 {
		proxyChain.Destination.Urls = []string{proxyHandler.serviceUrl}
	}

	if proxyChain.Sources == nil {
		proxyChain.Sources = []string{}
	}

	if !proxyChain.IsValid() {
		return req.Fail("proxy chain is not valid")
	}

	// todo the duplicate proxy chains for a rule is not defined yet
	proxyHandler.proxyChains = append(proxyHandler.proxyChains, &proxyChain)

	return req.Ok(key_value.New())
}

// Start starts the proxy handler as a new thread
func (proxyHandler *ProxyHandler) Start() error {
	if len(proxyHandler.serviceId) == 0 || len(proxyHandler.serviceUrl) == 0 {
		return fmt.Errorf("no service id or service url. call ProxyHandler.SetService first")
	}

	if proxyHandler.Handler.Config() == nil {
		return fmt.Errorf("handler configuration wasn't set. Call ProxyHandler.SetConfig first")
	}

	if len(proxyHandler.Handler.Routes) > 0 {
		return fmt.Errorf("writing routes is not allowed")
	}

	// todo
	// add the routes to do operation on the proxy handler

	if err := proxyHandler.Handler.Start(); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Start: %w", err)
	}

	return nil
}
