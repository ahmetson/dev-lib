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
	Set(chain *service.ProxyChain) error
}

type Client struct {
	*client.Socket
	proxyChains []*service.ProxyChain // list of proxy chains to add on the start
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

	return &Client{socket, make([]*service.ProxyChain, 0)}, nil
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
