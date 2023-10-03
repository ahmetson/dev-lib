package proxy_handler

import (
	"github.com/ahmetson/client-lib"
	"github.com/ahmetson/config-lib/service"
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
	proxyChain   *service.ProxyChain
	proxy1       *service.Proxy
	proxy2       *service.Proxy

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

	test.proxyChain = &service.ProxyChain{Sources: []string{}, Proxies: []*service.Proxy{},
		Destination: &service.Rule{Urls: []string{}, Categories: []string{}, Commands: []string{}, ExcludedCommands: []string{}}}
	test.proxy1 = &service.Proxy{Id: "id_1", Url: "url_1", Category: "category_1"}
	test.proxy2 = &service.Proxy{Id: "id_2", Url: "url_2", Category: "category_2"}

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

// Test_12_ProxyHandler_Route tests that routing is not available from out
func (test *TestProxyHandlerSuite) Test_12_ProxyHandler_Route() {
	s := test.Require

	handler := New(nil, nil)

	err := handler.Route("cmd_1", test.handleFunc)
	s().Error(err)
}

// Test_13_ProxyHandler_Start tests the ProxyHandler.Start method.
func (test *TestProxyHandlerSuite) Test_13_ProxyHandler_Start() {
	s := test.Suite.Require

	handler := New(nil, nil)

	// No configuration must fail
	err := handler.Start()
	s().Error(err)
	inprocConfig := HandlerConfig(test.id)
	handler.SetConfig(inprocConfig)
	handler.SetServiceId(test.id)

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

// Test_14_onSetProxyChain tests ProxyHandler receiving a SetProxyChain command.
func (test *TestProxyHandlerSuite) Test_14_ProxyHandler_onSetProxyChain() {
	s := test.Require

	req := &message.Request{
		Command:    SetProxyChain,
		Parameters: key_value.New(),
	}

	handler := New(nil, nil)

	// the proxy chain doesn't exist in the request parameters
	reply := handler.onSetProxyChain(req)
	s().False(reply.IsOK())

	// empty proxy chain is invalid, so it must be not set
	proxyKv, err := key_value.NewFromInterface(test.proxyChain)
	s().NoError(err)
	req.Parameters.Set("proxy_chain", proxyKv)
	reply = handler.onSetProxyChain(req)
	s().False(reply.IsOK())

	// set a one rule and one proxy
	test.proxyChain.Destination = service.NewServiceDestination(test.url)
	test.proxyChain.Proxies = []*service.Proxy{test.proxy1}
	proxyKv, err = key_value.NewFromInterface(test.proxyChain)
	s().NoError(err)
	req.Parameters.Set("proxy_chain", proxyKv)

	s().Len(handler.proxyChains, 0)
	reply = handler.onSetProxyChain(req)
	s().True(reply.IsOK())
	s().Len(handler.proxyChains, 1)
	s().Len(handler.proxyChains[0].Proxies, 1)

	// set of the same proxy rule must over-write the previous proxy chains
	test.proxyChain.Proxies = []*service.Proxy{test.proxy2, test.proxy1}
	proxyKv, err = key_value.NewFromInterface(test.proxyChain)
	s().NoError(err)
	req.Parameters.Set("proxy_chain", proxyKv)

	reply = handler.onSetProxyChain(req)
	s().True(reply.IsOK())
	s().Len(handler.proxyChains, 1)
	s().Len(handler.proxyChains[0].Proxies, 2)
}

// Test_15_ProxyHandler_onProxyChainsByRuleUrl tests ProxyHandler receiving a ProxyChainsByRuleUrl command.
func (test *TestProxyHandlerSuite) Test_15_ProxyHandler_onProxyChainsByRuleUrl() {
	s := test.Require

	req := &message.Request{
		Command:    SetProxyChain,
		Parameters: key_value.New(),
	}

	handler := New(nil, nil)
	test.proxyChain.Destination = service.NewServiceDestination(test.url)
	test.proxyChain.Proxies = []*service.Proxy{test.proxy1, test.proxy2}
	handler.proxyChains = append(handler.proxyChains, test.proxyChain)

	// the handler has one proxy, return it
	req.Parameters.Set("url", test.url)
	reply := handler.onProxyChainsByRuleUrl(req)
	s().True(reply.IsOK())
	proxyChainKvs, ok := reply.ReplyParameters()["proxy_chains"].([]*service.ProxyChain)
	s().True(ok)
	s().Len(proxyChainKvs, 1)

	// try to get non-existing url must return an empty value
	req.Parameters.Set("url", "non_existing_service")
	reply = handler.onProxyChainsByRuleUrl(req)
	s().True(reply.IsOK())
	proxyChainKvs, ok = reply.ReplyParameters()["proxy_chains"].([]*service.ProxyChain)
	s().True(ok)
	s().Len(proxyChainKvs, 0)

	// add another proxy chain for another url
	// must not interfere to counting other services
	rule := service.NewServiceDestination("url_2")
	proxyChain2 := &service.ProxyChain{
		Sources:     []string{},
		Proxies:     []*service.Proxy{test.proxy1},
		Destination: rule,
	}
	handler.proxyChains = append(handler.proxyChains, proxyChain2)

	req.Parameters.Set("url", test.url)
	reply = handler.onProxyChainsByRuleUrl(req)
	s().True(reply.IsOK())
	proxyChainKvs, ok = reply.ReplyParameters()["proxy_chains"].([]*service.ProxyChain)
	s().True(ok)
	s().Len(proxyChainKvs, 1)

	// adding a proxy chain of the same type must return both new added proxy chain and old one
	rule = service.NewHandlerDestination(test.url, "handler_category")
	proxyChain3 := &service.ProxyChain{
		Sources:     []string{},
		Proxies:     []*service.Proxy{test.proxy1},
		Destination: rule,
	}
	handler.proxyChains = append(handler.proxyChains, proxyChain3)

	req.Parameters.Set("url", test.url)
	reply = handler.onProxyChainsByRuleUrl(req)
	s().True(reply.IsOK())
	proxyChainKvs, ok = reply.ReplyParameters()["proxy_chains"].([]*service.ProxyChain)
	s().True(ok)
	s().Len(proxyChainKvs, 2)
}

// Test_16_ProxyHandler_units tests ProxyHandler receiving a Units and SetUnits commands.
func (test *TestProxyHandlerSuite) Test_16_ProxyHandler_units() {
	s := test.Require

	handler := New(nil, nil)
	rule1 := service.NewServiceDestination(test.url)
	rule1Kv, err := key_value.NewFromInterface(rule1)
	s().NoError(err)
	unit1 := &service.Unit{
		ServiceId: "service",
		HandlerId: "handler",
		Command:   "command",
	}
	unit1Kv, err := key_value.NewFromInterface(unit1)
	s().NoError(err)

	// the units are empty
	s().Len(handler.proxyUnits, 0)

	// requesting a unit must return an empty result
	req := &message.Request{
		Command:    Units,
		Parameters: key_value.New().Set("rule", rule1Kv),
	}
	reply := handler.onUnits(req)
	s().True(reply.IsOK())
	unitRaws, ok := reply.ReplyParameters()["units"].([]*service.Unit)
	s().True(ok)
	s().Len(unitRaws, 0)

	// set the units
	units := []key_value.KeyValue{unit1Kv}
	req.Command = SetUnits
	req.Parameters.Set("rule", rule1Kv).Set("units", units)
	reply = handler.onSetUnits(req)
	s().True(reply.IsOK())

	// get the units
	req.Command = Units
	reply = handler.onUnits(req)
	s().True(reply.IsOK())
	unitRaws, ok = reply.ReplyParameters()["units"].([]*service.Unit)
	s().True(ok)
	s().Len(unitRaws, 1)
}

// Test_17_ProxyHandler_onProxyChainsByLastId tests ProxyHandler receiving a ProxyChainsByLastId command.
func (test *TestProxyHandlerSuite) Test_17_ProxyHandler_onProxyChainsByLastId() {
	s := test.Require

	req := &message.Request{
		Command:    ProxyChainsByLastId,
		Parameters: key_value.New(),
	}

	handler := New(nil, nil)
	test.proxyChain.Destination = service.NewServiceDestination(test.url)
	test.proxyChain.Proxies = []*service.Proxy{test.proxy1, test.proxy2}
	handler.proxyChains = append(handler.proxyChains, test.proxyChain)

	// the proxy1 is the first, not the last. so it must return an empty result
	req.Parameters.Set("id", test.proxy1.Id)
	reply := handler.onProxyChainsByLastId(req)
	s().True(reply.IsOK())
	proxyChainKvs, ok := reply.ReplyParameters()["proxy_chains"].([]*service.ProxyChain)
	s().True(ok)
	s().Len(proxyChainKvs, 0)

	// the proxy2 is the last, so it must return a one proxy chain
	req.Parameters.Set("id", test.proxy2.Id)
	reply = handler.onProxyChainsByLastId(req)
	s().True(reply.IsOK())
	proxyChainKvs, ok = reply.ReplyParameters()["proxy_chains"].([]*service.ProxyChain)
	s().True(ok)
	s().Len(proxyChainKvs, 1)

}

// Test_18_ProxyHandler_onLastProxies tests ProxyHandler receiving a LastProxies command.
func (test *TestProxyHandlerSuite) Test_18_ProxyHandler_onLastProxies() {
	s := test.Require

	req := &message.Request{
		Command:    ProxyChainsByLastId,
		Parameters: key_value.New(),
	}

	handler := New(nil, nil)
	test.proxyChain.Destination = service.NewServiceDestination(test.url)
	test.proxyChain.Proxies = []*service.Proxy{test.proxy1, test.proxy2}
	handler.proxyChains = append(handler.proxyChains, test.proxyChain)

	// the proxy1 is the first, not the last. so it must return an empty result
	reply := handler.onLastProxies(req)
	s().True(reply.IsOK())
	proxyChainKvs, ok := reply.ReplyParameters()["proxies"].([]*service.Proxy)
	s().True(ok)
	s().Len(proxyChainKvs, 1)

}

func TestProxyHandler(t *testing.T) {
	suite.Run(t, new(TestProxyHandlerSuite))
}
