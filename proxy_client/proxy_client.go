// Package proxy_client defines a client that works with the Proxy thread.
package proxy_client

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/datatype-lib/message"
	"github.com/ahmetson/dev-lib/proxy_handler"
	handlerConfig "github.com/ahmetson/handler-lib/config"
)

type Interface interface {
	Set(chain *service.ProxyChain) error                            // Set sets a new proxy chain in the configuration. Over-write for a duplicate rule.
	ProxyChainsByRuleUrl(url string) ([]*service.ProxyChain, error) // Returns list of proxy chains by url in the destination.
	SetUnits(*service.Rule, []*service.Unit) error                  // Sets the destination units for each rule
	ProxyChainsByLastId(id string) ([]*service.ProxyChain, error)   // Returns list of proxy chains by the last proxy id
	Units(*service.Rule) ([]*service.Unit, error)                   // Returns list of destination units by a rule
	LastProxies() ([]*service.Proxy, error)                         // Returns list of proxies
}

type Client struct {
	*client.Socket
}

// The New returns a proxy client for the serviceId.
func New(serviceId string) (*Client, error) {
	configHandler := proxy_handler.HandlerConfig(serviceId)
	socketType := handlerConfig.SocketType(configHandler.Type)
	c := clientConfig.New("", configHandler.Id, configHandler.Port, socketType).
		UrlFunc(clientConfig.Url)

	socket, err := client.New(c)
	if err != nil {
		return nil, fmt.Errorf("client.New: %w", err)
	}

	return &Client{socket}, nil
}

// Set sends the proxyChain to the proxy handler
func (c *Client) Set(proxyChain *service.ProxyChain) error {
	params := key_value.New().Set("proxy_chain", proxyChain)
	req := &message.Request{
		Command:    proxy_handler.SetProxyChain,
		Parameters: params,
	}
	reply, err := c.Request(req)
	if err != nil {
		return fmt.Errorf("c.Request: %w", err)
	}
	if !reply.IsOK() {
		return fmt.Errorf("reply error message: %s", reply.ErrorMessage())
	}

	return nil
}

// ProxyChainsByRuleUrl returns the proxy chains by the destination's url field.
func (c *Client) ProxyChainsByRuleUrl(url string) ([]*service.ProxyChain, error) {
	req := &message.Request{
		Command:    proxy_handler.ProxyChainsByRuleUrl,
		Parameters: key_value.New().Set("url", url),
	}
	reply, err := c.Request(req)
	if err != nil {
		return nil, fmt.Errorf("c.Request: %w", err)
	}
	if !reply.IsOK() {
		return nil, fmt.Errorf("reply error message: %s", reply.ErrorMessage())
	}

	kvList, err := reply.ReplyParameters().NestedListValue("proxy_chains")
	if err != nil {
		return nil, fmt.Errorf("reply.ReplyParameters().NestedKeyValueList('proxy_chains'): %w", err)
	}

	proxyChains := make([]*service.ProxyChain, len(kvList))
	for i, kv := range kvList {
		err = kv.Interface(proxyChains[i])
		if err != nil {
			return nil, fmt.Errorf("kv.Interface(proxyChains[%d]): %w", i, err)
		}
	}

	return proxyChains, nil
}

// SetUnits sends the rule and units for this rule to the proxy handler
func (c *Client) SetUnits(rule *service.Rule, units []*service.Unit) error {
	params := key_value.New().
		Set("rule", rule).
		Set("units", units)
	req := &message.Request{
		Command:    proxy_handler.SetUnits,
		Parameters: params,
	}
	reply, err := c.Request(req)
	if err != nil {
		return fmt.Errorf("c.Request: %w", err)
	}
	if !reply.IsOK() {
		return fmt.Errorf("reply error message: %s", reply.ErrorMessage())
	}

	return nil
}

// ProxyChainsByLastId returns the proxy chains by the last proxy id.
func (c *Client) ProxyChainsByLastId(id string) ([]*service.ProxyChain, error) {
	req := &message.Request{
		Command:    proxy_handler.ProxyChainsByLastId,
		Parameters: key_value.New().Set("id", id),
	}
	reply, err := c.Request(req)
	if err != nil {
		return nil, fmt.Errorf("c.Request: %w", err)
	}
	if !reply.IsOK() {
		return nil, fmt.Errorf("reply error message: %s", reply.ErrorMessage())
	}

	kvList, err := reply.ReplyParameters().NestedListValue("proxy_chains")
	if err != nil {
		return nil, fmt.Errorf("reply.ReplyParameters().NestedKeyValueList('proxy_chains'): %w", err)
	}

	proxyChains := make([]*service.ProxyChain, len(kvList))
	for i, kv := range kvList {
		err = kv.Interface(proxyChains[i])
		if err != nil {
			return nil, fmt.Errorf("kv.Interface(proxyChains[%d]): %w", i, err)
		}
	}

	return proxyChains, nil
}

// The Units method returns the destination units by a rule.
func (c *Client) Units(rule *service.Rule) ([]*service.Unit, error) {
	req := &message.Request{
		Command:    proxy_handler.Units,
		Parameters: key_value.New().Set("rule", rule),
	}
	reply, err := c.Request(req)
	if err != nil {
		return nil, fmt.Errorf("c.Request: %w", err)
	}
	if !reply.IsOK() {
		return nil, fmt.Errorf("reply error message: %s", reply.ErrorMessage())
	}

	rawUnits, err := reply.ReplyParameters().NestedListValue("units")
	if err != nil {
		return nil, fmt.Errorf("reply.ReplyParameters().NestedKeyValueList('proxy_chains'): %w", err)
	}

	units := make([]*service.Unit, len(rawUnits))
	for i, rawUnit := range rawUnits {
		var unit service.Unit
		err = rawUnit.Interface(&unit)
		if err != nil {
			return nil, fmt.Errorf("rawUnits[%d].Interface: %w", i, err)
		}

		units[i] = &unit
	}

	return units, nil
}

// The LastProxies method returns the last proxies
func (c *Client) LastProxies() ([]*service.Proxy, error) {
	req := &message.Request{
		Command:    proxy_handler.LastProxies,
		Parameters: key_value.New(),
	}
	reply, err := c.Request(req)
	if err != nil {
		return nil, fmt.Errorf("c.Request: %w", err)
	}
	if !reply.IsOK() {
		return nil, fmt.Errorf("reply error message: %s", reply.ErrorMessage())
	}

	rawProxies, err := reply.ReplyParameters().NestedListValue("proxies")
	if err != nil {
		return nil, fmt.Errorf("reply.ReplyParameters().NestedKeyValueList('proxy_chains'): %w", err)
	}

	proxies := make([]*service.Proxy, len(rawProxies))
	for i, rawProxy := range rawProxies {
		var proxy service.Proxy
		err = rawProxy.Interface(&proxy)
		if err != nil {
			return nil, fmt.Errorf("rawProxies[%d].Interface: %w", i, err)
		}

		proxies[i] = &proxy
	}

	return proxies, nil
}

// The StartLastProxies method starts all the proxies
func (c *Client) StartLastProxies() error {
	req := &message.Request{
		Command:    proxy_handler.StartLastProxies,
		Parameters: key_value.New(),
	}
	reply, err := c.Request(req)
	if err != nil {
		return fmt.Errorf("c.Request: %w", err)
	}
	if !reply.IsOK() {
		return fmt.Errorf("reply error message: %s", reply.ErrorMessage())
	}

	return nil
}
