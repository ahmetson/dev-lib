// Package proxy_handler is a thread that manages the service proxies
package proxy_handler

import (
	"fmt"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/datatype-lib/message"
	"github.com/ahmetson/handler-lib/base"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"slices"
)

const (
	Category             = "proxy_handler"            // handler category
	SetProxyChain        = "set-proxy-chain"          // route command that sets a new proxy chain
	ProxyChainsByRuleUrl = "proxy-chains-by-rule-url" // route command that returns list of proxy chains by url in the rule
)

type ProxyHandler struct {
	*base.Handler
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

// The Route is over-written to be disabled.
func (proxyHandler *ProxyHandler) Route(string, any, ...string) error {
	return fmt.Errorf("not implemented")
}

// onSetProxyChain is a handler function to set the new proxy chain
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
	if proxyChain.Sources == nil {
		proxyChain.Sources = []string{}
	}

	if !proxyChain.IsValid() {
		return req.Fail("proxy chain is not valid")
	}

	params := key_value.New()

	i := slices.IndexFunc(proxyHandler.proxyChains, func(proxyChain *service.ProxyChain) bool {
		return service.IsEqualRule(proxyChain.Destination, proxyChain.Destination)
	})
	if i > -1 {
		proxyHandler.proxyChains[i] = &proxyChain
	} else {
		proxyHandler.proxyChains = append(proxyHandler.proxyChains, &proxyChain)
	}
	params.Set("overwrite", i > -1)

	return req.Ok(key_value.New())
}

// onProxyChainsByRuleUrl is a handler function to get the proxy chains by the destination.
//
// This method is intended to be called by the independent service, to return the list of proxy chains in the service.
//
// Returns empty data if no proxy chain is found.
func (proxyHandler *ProxyHandler) onProxyChainsByRuleUrl(req message.RequestInterface) message.ReplyInterface {
	url, err := req.RouteParameters().StringValue("url")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.RouteParameters().StringValue('url'): %v", err))
	}

	proxyChains := service.ProxyChainsByRuleUrl(proxyHandler.proxyChains, url)

	params := key_value.New().Set("proxy_chains", proxyChains)

	return req.Ok(params)
}

func (proxyHandler *ProxyHandler) setRoutes() error {
	if err := proxyHandler.Handler.Route(SetProxyChain, proxyHandler.onSetProxyChain); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Route('%s'): %w", SetProxyChain, err)
	}
	if err := proxyHandler.Handler.Route(ProxyChainsByRuleUrl, proxyHandler.onProxyChainsByRuleUrl); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Route('%s'): %w", ProxyChainsByRuleUrl, err)
	}

	return nil
}

// Start starts the proxy handler as a new thread
func (proxyHandler *ProxyHandler) Start() error {
	if len(proxyHandler.Handler.Routes) > 0 {
		return fmt.Errorf("writing routes is not allowed")
	}

	if err := proxyHandler.setRoutes(); err != nil {
		return fmt.Errorf("proxyHandler.setRoutes: %w", err)
	}

	if err := proxyHandler.Handler.Start(); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Start: %w", err)
	}

	return nil
}
