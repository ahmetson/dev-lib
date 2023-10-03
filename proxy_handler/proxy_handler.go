// Package proxy_handler is a thread that manages the service proxies
package proxy_handler

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/datatype-lib/message"
	"github.com/ahmetson/dev-lib/dep_client"
	"github.com/ahmetson/dev-lib/source"
	"github.com/ahmetson/handler-lib/base"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"slices"
)

const (
	Category = "proxy_handler" // handler category

	//
	// Commands of the proxy handler
	//

	// SetProxyChain command sets a new proxy chain
	SetProxyChain        = "set-proxy-chain"
	ProxyChainsByRuleUrl = "proxy-chains-by-rule-url" // route command that returns list of proxy chains by url in the rule
	SetUnits             = "set-units"                // route command that sets the proxy units
	ProxyChainsByLastId  = "proxy-chains-by-last-id"  // route command that returns list of proxy chains by the id of the last proxy
	Units                = "units"                    // route command that returns a list of destination units for a rule
	LastProxies          = "last-proxies"             // route command that returns a list of last proxies in the proxy chains.
)

type ProxyHandler struct {
	*base.Handler
	proxyChains []*service.ProxyChain
	proxyUnits  map[*service.Rule][]*service.Unit
	depClient   dep_client.Interface
	engine      configClient.Interface
	serviceId   string
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
func New(engine configClient.Interface, depClient dep_client.Interface) *ProxyHandler {
	newHandler := base.New()
	return &ProxyHandler{
		Handler:     newHandler,
		proxyChains: make([]*service.ProxyChain, 0),
		proxyUnits:  make(map[*service.Rule][]*service.Unit, 0),
		engine:      engine,
		depClient:   depClient,
	}
}

// SetServiceId notifies the proxy handler with the service where it's belonged too.
func (proxyHandler *ProxyHandler) SetServiceId(id string) {
	proxyHandler.serviceId = id
}

func (proxyHandler *ProxyHandler) setUnits(rule *service.Rule, units []*service.Unit) {
	for firstRule := range proxyHandler.proxyUnits {
		if service.IsEqualRule(firstRule, rule) {
			proxyHandler.proxyUnits[firstRule] = units
			return
		}
	}
	proxyHandler.proxyUnits[rule] = units
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

// onSetUnits is a handler function that sets the list of all proxy units for each rule.
func (proxyHandler *ProxyHandler) onSetUnits(req message.RequestInterface) message.ReplyInterface {
	raw, err := req.RouteParameters().NestedValue("rule")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.RouteParameters().NestedValue('proxy_chain'): %v", err))
	}

	var rule service.Rule
	err = raw.Interface(&rule)
	if err != nil {
		return req.Fail(fmt.Sprintf("key_value.KeyValue('proxy_chain').Interface(): %v", err))
	}

	if !rule.IsValid() {
		return req.Fail("the 'rule' parameter is not valid")
	}

	rawUnits, err := req.RouteParameters().NestedListValue("units")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.RouteParameters().NestedListValue('units'): %v", err))
	}

	units := make([]*service.Unit, len(rawUnits))
	for i, rawUnit := range rawUnits {
		var unit service.Unit
		err = rawUnit.Interface(&unit)
		if err != nil {
			return req.Fail(fmt.Sprintf("rawUnits[%d].Interface: %v", i, err))
		}

		units[i] = &unit
	}

	proxyHandler.setUnits(&rule, units)

	return req.Ok(key_value.New())
}

// onProxyChainsByLastId returns list of proxy chains where the proxy id is in the proxies list
func (proxyHandler *ProxyHandler) onProxyChainsByLastId(req message.RequestInterface) message.ReplyInterface {
	id, err := req.RouteParameters().StringValue("id")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.RouteParameters().StringValue('id'): %v", err))
	}

	proxyChains := make([]*service.ProxyChain, 0, len(proxyHandler.proxyChains))
	for _, proxyChain := range proxyHandler.proxyChains {
		lastProxy := len(proxyChain.Proxies) - 1
		if lastProxy == -1 {
			continue
		}

		if proxyChain.Proxies[lastProxy].Id == id {
			proxyChains = append(proxyChains, proxyChain)
		}
	}

	params := key_value.New().Set("proxy_chains", proxyChains)

	return req.Ok(params)
}

// onUnits returns the list of units by a rule
func (proxyHandler *ProxyHandler) onUnits(req message.RequestInterface) message.ReplyInterface {
	raw, err := req.RouteParameters().NestedValue("rule")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.RouteParameters().NestedValue('proxy_chain'): %v", err))
	}

	var rule service.Rule
	err = raw.Interface(&rule)
	if err != nil {
		return req.Fail(fmt.Sprintf("key_value.KeyValue('proxy_chain').Interface(): %v", err))
	}

	if !rule.IsValid() {
		return req.Fail("the 'rule' parameter is not valid")
	}

	units := make([]*service.Unit, 0)

	for firstRule := range proxyHandler.proxyUnits {
		if service.IsEqualRule(firstRule, &rule) {
			units = proxyHandler.proxyUnits[firstRule]
			break
		}
	}

	params := key_value.New().Set("units", units)

	return req.Ok(params)
}

// onLastProxies returns the list of the proxies.
// The duplicate proxies are compacted
func (proxyHandler *ProxyHandler) onLastProxies(req message.RequestInterface) message.ReplyInterface {
	proxies := service.LastProxies(proxyHandler.proxyChains)

	params := key_value.New().Set("proxies", proxies)

	return req.Ok(params)
}

// onStartLastProxies starts the proxies
func (proxyHandler *ProxyHandler) onStartLastProxies(req message.RequestInterface) message.ReplyInterface {
	if len(proxyHandler.serviceId) == 0 {
		return req.Fail("serviceId not set. call ProxyHandler.SetServiceId first")
	}
	if proxyHandler.engine == nil {
		return req.Fail("config engine is not set")
	}
	if proxyHandler.depClient == nil {
		return req.Fail("dependency manager is not set")
	}
	depManager := proxyHandler.depClient

	proxies := service.LastProxies(proxyHandler.proxyChains)

	if len(proxies) == 0 {
		return req.Ok(key_value.New())
	}

	serviceConfig, err := proxyHandler.engine.Service(proxyHandler.serviceId)
	if err != nil {
		return req.Fail(fmt.Sprintf("engine.Service('%s'): %v", proxyHandler.serviceId, err))
	}

	// todo make sure to run in concurrency
	for i := range proxies {
		proxy := proxies[i]

		if serviceConfig.SourceExist(proxy.Id) {
			proxySource := serviceConfig.SourceById(proxy.Id)
			if proxySource == nil || proxySource.Manager == nil {
				continue
			}

			running, err := depManager.Running(proxySource.Manager)
			if err != nil {
				return req.Fail(fmt.Sprintf("depManager.Running('%s'): %v", proxy.Id, err))
			}

			// todo make sure to update the client parameters with the rule units
			if running {
				continue
			}

			if err := depManager.Run(proxy.Url, proxy.Id, serviceConfig.Manager); err != nil {
				return req.Fail(fmt.Sprintf("depManager.Run('%s', '%s'): %v", proxy.Url, proxy.Id, err))
			}

			continue
		}

		installed, err := depManager.Installed(proxy.Url)
		if err != nil {
			return req.Fail(fmt.Sprintf("depManager.Installed('%s'): %v", proxy.Url, err))
		}

		if !installed {
			proxySrc, err := source.New(proxy.Url)
			if err != nil {
				return req.Fail(fmt.Sprintf("source.New('%s'): %v", proxy.Url, err))
			}
			err = depManager.Install(proxySrc)
			if err != nil {
				return req.Fail(fmt.Sprintf("depManager.Install: %v", err))
			}
		}

		if err := depManager.Run(proxy.Url, proxy.Id, serviceConfig.Manager); err != nil {
			return req.Fail(fmt.Sprintf("depManager.Run('%s', '%s'): %v", proxy.Url, proxy.Id, err))
		}
	}

	return req.Ok(key_value.New())
}

func (proxyHandler *ProxyHandler) setRoutes() error {
	if err := proxyHandler.Handler.Route(SetProxyChain, proxyHandler.onSetProxyChain); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Route('%s'): %w", SetProxyChain, err)
	}
	if err := proxyHandler.Handler.Route(ProxyChainsByRuleUrl, proxyHandler.onProxyChainsByRuleUrl); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Route('%s'): %w", ProxyChainsByRuleUrl, err)
	}
	if err := proxyHandler.Handler.Route(SetUnits, proxyHandler.onSetUnits); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Route('%s'): %w", SetUnits, err)
	}
	if err := proxyHandler.Handler.Route(ProxyChainsByLastId, proxyHandler.onSetUnits); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Route('%s'): %w", ProxyChainsByLastId, err)
	}
	if err := proxyHandler.Handler.Route(Units, proxyHandler.onUnits); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Route('%s'): %w", Units, err)
	}
	if err := proxyHandler.Handler.Route(LastProxies, proxyHandler.onLastProxies); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Route('%s'): %w", LastProxies, err)
	}

	return nil
}

func (proxyHandler *ProxyHandler) onClose(request message.RequestInterface) message.ReplyInterface {
	if err := proxyHandler.closeProxies(); err != nil {
		return request.Fail(fmt.Sprintf("proxyHandler.closeProxies: %v", err))
	}

	return proxyHandler.Handler.Manager.SetClose(request)
}

// closeProxies will close all proxies that have generated configuration
func (proxyHandler *ProxyHandler) closeProxies() error {
	if len(proxyHandler.serviceId) == 0 {
		return fmt.Errorf("serviceId not set. call ProxyHandler.SetServiceId first")
	}
	if proxyHandler.engine == nil {
		return nil
	}
	if proxyHandler.depClient == nil {
		return nil
	}

	serviceConfig, err := proxyHandler.engine.Service(proxyHandler.serviceId)
	if err != nil {
		return fmt.Errorf("engine.Service('%s'): %w", proxyHandler.serviceId, err)
	}

	if len(serviceConfig.Sources) == 0 {
		return nil
	}
	depManager := proxyHandler.depClient

	for i := range serviceConfig.Sources {
		sourceRule := serviceConfig.Sources[i]

		if len(sourceRule.Proxies) == 0 {
			continue
		}
		for j := range sourceRule.Proxies {
			sourceProxy := sourceRule.Proxies[j]

			if sourceProxy.Manager == nil {
				continue
			}

			if err := depManager.CloseDep(sourceProxy.Manager); err != nil {
				return fmt.Errorf("depManager.CloseDep('%s'): %w", sourceProxy.Id, err)
			}
		}
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

	if proxyHandler.Handler.Manager != nil {
		if err := proxyHandler.Handler.Manager.Route(handlerConfig.HandlerClose, proxyHandler.onClose); err != nil {
			return fmt.Errorf("manager.Route('%s'): %w", handlerConfig.HandlerClose, err)
		}
	}

	if err := proxyHandler.Handler.Start(); err != nil {
		return fmt.Errorf("proxyHandler.Handler.Start: %w", err)
	}

	return nil
}
