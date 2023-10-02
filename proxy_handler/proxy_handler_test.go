package proxy_handler

import (
	"github.com/ahmetson/client-lib"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/datatype-lib/message"
	"github.com/ahmetson/handler-lib/manager_client"
	"github.com/ahmetson/handler-lib/route"
	"github.com/ahmetson/log-lib"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestProxyHandlerSuite struct {
	suite.Suite

	logger       *log.Logger
	proxyHandler *ProxyHandler // the manager to test
	url          string        // dependency source code
	id           string        // the id of the service
	handlerId    string        // the id of the proxy handler
	handleFunc   route.HandleFunc0

	client *client.Socket // imitating the service
}

// Make sure that Account is set to five
// before each test
func (test *TestProxyHandlerSuite) SetupTest() {
	logger, _ := log.New("test", false)
	test.logger = logger

	// A valid source code that we want to download
	test.url = "github.com/ahmetson/test-service"
	test.id = "test_service"
	test.handlerId = "test_service_proxy_handler"
	test.handleFunc = func(req message.RequestInterface) message.ReplyInterface {
		return req.Ok(key_value.New())
	}
}

func (test *TestProxyHandlerSuite) TearDownTest() {}

// Test_10_Id tests generation of the handler id with Id function
func (test *TestProxyHandlerSuite) Test_10_Id() {
	s := test.Require

	actualId := Id(test.id)
	s().Equal(test.handlerId, actualId)
}

// Test_11_HandlerConfig tests HandlerConfig method
func (test *TestProxyHandlerSuite) Test_11_HandlerConfig() {
	s := test.Require

	inprocConfig := HandlerConfig(test.id)
	s().True(inprocConfig.IsInproc())
}

// Test_12_ProxyHandler_SetService tests ProxyHandler.SetService method
func (test *TestProxyHandlerSuite) Test_12_ProxyHandler_SetService() {
	s := test.Suite.Require

	handler := New()

	// By default, the service parameters are empty
	s().Empty(handler.serviceId)
	s().Empty(handler.serviceUrl)

	// After setting, the service parameters must be available
	handler.SetService(test.id, test.url)
	s().Equal(test.id, handler.serviceId)
	s().Equal(test.url, handler.serviceUrl)
}

// Test_13_ProxyHandler_Route tests that routing is not available from out
func (test *TestProxyHandlerSuite) Test_13_ProxyHandler_Route() {
	s := test.Require

	handler := New()

	err := handler.Route("cmd_1", test.handleFunc)
	s().Error(err)
}

// Test_11_Start tests preparation of the proxy
func (test *TestProxyHandlerSuite) Test_11_Start() {
	s := test.Suite.Require

	handler := New()

	// No service id and service url must fail
	err := handler.Start()
	s().Error(err)
	handler.SetService(test.id, test.url)

	// No configuration must fail
	err = handler.Start()
	s().Error(err)
	inprocConfig := HandlerConfig(test.id)
	handler.SetConfig(inprocConfig)

	// No logger must fail
	err = handler.Start()
	s().Error(err)
	err = handler.SetLogger(test.logger)
	s().NoError(err)

	// Routes are set by the user, so it must fail
	err = handler.Handler.Route("cmd_1", test.handleFunc)
	s().NoError(err)
	err = handler.Start()
	s().Error(err)

	// No routes and all parameters are set must start the handler
	handler.Routes = key_value.New()
	err = handler.Start()
	s().NoError(err)

	// Wait a bit for initialization
	time.Sleep(time.Millisecond * 100)

	// Close the service
	manager, err := manager_client.New(inprocConfig)
	s().NoError(err)
	err = manager.Close()
	s().NoError(err)

	// Wait a bit for closing the threads
	time.Sleep(time.Millisecond * 100)
}

func TestProxyHandler(t *testing.T) {
	suite.Run(t, new(TestProxyHandlerSuite))
}
