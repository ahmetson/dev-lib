package dep_client

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/common-lib/message"
	"github.com/ahmetson/dev-lib/dep_handler"
	"github.com/ahmetson/dev-lib/source"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"time"
)

type Client struct {
	socket *client.Socket
}

type Interface interface {
	Close() error
	Timeout(duration time.Duration)
	Attempt(attempt uint8)

	CloseDep(depClient *clientConfig.Client) error
	Uninstall(src *source.Src) error
	Run(url string, id string, parent *clientConfig.Client) error
	Install(src *source.Src) error
	Running(depClient *clientConfig.Client) (bool, error)
	Installed(url string) (bool, error)
}

func New() (*Client, error) {
	configHandler := dep_handler.ServiceConfig()
	socketType := handlerConfig.SocketType(configHandler.Type)
	c := clientConfig.New("", configHandler.Id, configHandler.Port, socketType).
		UrlFunc(clientConfig.Url)

	socket, err := client.New(c)
	if err != nil {
		return nil, fmt.Errorf("client.New: %w", err)
	}

	return &Client{socket: socket}, nil
}

// Timeout of the client socket
func (c *Client) Timeout(duration time.Duration) {
	c.socket.Timeout(duration)
}

// Attempt amount for requests
func (c *Client) Attempt(attempt uint8) {
	c.socket.Attempt(attempt)
}

func (c *Client) Close() error {
	return c.socket.Close()
}

// CloseDep the running dependency
func (c *Client) CloseDep(depClient *clientConfig.Client) error {
	req := message.Request{
		Command: dep_handler.CloseDep,
		Parameters: key_value.Empty().
			Set("dep", depClient),
	}

	err := c.socket.Submit(&req)
	if err != nil {
		return fmt.Errorf("socket.Submit('%s'): %w", dep_handler.CloseDep, err)
	}

	return nil
}

// Uninstall the dependency.
func (c *Client) Uninstall(src *source.Src) error {
	req := message.Request{
		Command:    dep_handler.UninstallDep,
		Parameters: key_value.Empty().Set("src", src),
	}

	err := c.socket.Submit(&req)
	if err != nil {
		return fmt.Errorf("socket.Submit('%s'): %w", dep_handler.UninstallDep, err)
	}

	return nil
}

// Run the dependency. The url of the dependency. It's id. and the parameters of the parent to connect to.
func (c *Client) Run(url string, id string, parent *clientConfig.Client) error {
	req := message.Request{
		Command: dep_handler.RunDep,
		Parameters: key_value.Empty().
			Set("parent", parent).
			Set("url", url).
			Set("id", id),
	}
	reply, err := c.socket.Request(&req)
	if err != nil {
		return fmt.Errorf("socket.Submit('%s'): %w", dep_handler.RunDep, err)
	}

	if !reply.IsOK() {
		return fmt.Errorf("reply.Message: %s", reply.Message)
	}

	return nil
}

// Install the dependency from the source code. It compiles it.
func (c *Client) Install(src *source.Src) error {
	req := message.Request{
		Command:    dep_handler.InstallDep,
		Parameters: key_value.Empty().Set("src", src),
	}
	reply, err := c.socket.Request(&req)
	if err != nil {
		return fmt.Errorf("socket.Submit('%s'): %w", dep_handler.InstallDep, err)
	}

	if !reply.IsOK() {
		return fmt.Errorf("reply.Message: %s", reply.Message)
	}

	return nil
}

// Running checks is the service running or not
func (c *Client) Running(depClient *clientConfig.Client) (bool, error) {
	req := message.Request{
		Command: dep_handler.DepRunning,
		Parameters: key_value.Empty().
			Set("dep", depClient),
	}

	reply, err := c.socket.Request(&req)
	if err != nil {
		return false, fmt.Errorf("socket.Request('%s'): %w", dep_handler.DepRunning, err)
	}

	if !reply.IsOK() {
		return false, fmt.Errorf("reply.Message: %s", reply.Message)
	}

	res, err := reply.Parameters.GetBoolean("running")
	if err != nil {
		return false, fmt.Errorf("reply.Parameters.GetBoolean('installed'): %w", err)
	}

	return res, nil
}

// Installed checks is the service installed
func (c *Client) Installed(url string) (bool, error) {
	req := message.Request{
		Command:    dep_handler.DepInstalled,
		Parameters: key_value.Empty().Set("url", url),
	}

	reply, err := c.socket.Request(&req)
	if err != nil {
		return false, fmt.Errorf("socket.Request('%s'): %w", dep_handler.DepInstalled, err)
	}

	if !reply.IsOK() {
		return false, fmt.Errorf("reply.Message: %s", reply.Message)
	}

	res, err := reply.Parameters.GetBoolean("installed")
	if err != nil {
		return false, fmt.Errorf("reply.Parameters.GetBoolean('installed'): %w", err)
	}

	return res, nil
}
