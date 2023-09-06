package dep_handler

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/common-lib/message"
	"github.com/ahmetson/dev-lib/dep_manager"
	"github.com/ahmetson/dev-lib/source"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestDepHandlerSuite struct {
	suite.Suite

	logger     *log.Logger
	dep        *DepHandler          // the manager to test
	currentDir string               // executable to store the binaries and source codes
	url        string               // dependency source code
	id         string               // the id of the dependency
	parent     *clientConfig.Client // the info about the service to which dependency should connect

	client *client.Socket // imitating the service
}

// Make sure that Account is set to five
// before each test
func (test *TestDepHandlerSuite) SetupTest() {
	s := test.Suite.Require

	logger, _ := log.New("test", false)
	test.logger = logger

	currentDir, err := path.CurrentDir()
	s().NoError(err)
	test.currentDir = currentDir

	srcPath := path.AbsDir(currentDir, "_sds/src")
	binPath := path.AbsDir(currentDir, "_sds/bin")

	// Make sure that the folders don't exist. They will be added later
	manager := dep_manager.New()
	err = manager.SetPaths(srcPath, binPath)
	s().NoError(err)

	test.dep, err = New(manager)
	s().NoError(err)

	// Run the handler
	go func() {
		s().NoError(test.dep.Run())
	}()

	// wait a bit for closing
	time.Sleep(time.Millisecond * 100)

	// A valid source code that we want to download
	test.url = "github.com/ahmetson/test-manager"

	test.id = "test-manager"
	test.parent = &clientConfig.Client{
		ServiceUrl: "dev-lib",
		Id:         "parent",
		Port:       120,
		TargetType: handlerConfig.SocketType(handlerConfig.ReplierType),
	}

	config := ServiceConfig()
	socketType := handlerConfig.SocketType(config.Type)
	socket, err := client.NewRaw(socketType, fmt.Sprintf("inproc://%s", config.Id))
	s().NoError(err)

	test.client = socket
	test.client.Timeout(time.Second * 30)
	test.client.Attempt(1)
}

func (test *TestDepHandlerSuite) TearDownTest() {
	s := test.Suite.Require

	s().NoError(test.client.Close())

	s().NoError(test.dep.Close())

	// Wait a bit for the close of the handler thread.
	time.Sleep(time.Millisecond * 100)
}

// Test_10_Install checks InstallDep and DepInstalled
// This is the first part of Install.
// The second part of Install is building.
//
// Tests DepManager.downloadSrc and srcExist.
func (test *TestDepHandlerSuite) Test_10_Install() {
	s := test.Suite.Require

	src, err := source.New(test.url)
	s().NoError(err)

	// installation must fail since nothing installed
	req := message.Request{
		Command:    DepInstalled,
		Parameters: key_value.Empty().Set("url", test.url),
	}
	rep, err := test.client.Request(&req)
	s().NoError(err)
	s().True(rep.IsOK())
	res, err := rep.Parameters.GetBoolean("installed")
	s().NoError(err)
	s().False(res)

	// There should be a source code
	installReq := message.Request{
		Command:    InstallDep,
		Parameters: key_value.Empty().Set("src", src),
	}
	rep, err = test.client.Request(&installReq)
	s().NoError(err)
	s().True(rep.IsOK())

	// wait a bit until its installed
	time.Sleep(time.Millisecond * 100)

	//
	// Testing the installed after installation
	//
	rep, err = test.client.Request(&req)
	s().NoError(err)
	s().True(rep.IsOK())
	res, err = rep.Parameters.GetBoolean("installed")
	s().NoError(err)
	s().True(res)
}

// Test_11_Uninstall deletes the binary and source code installed at Test_11_Install
func (test *TestDepHandlerSuite) Test_11_Uninstall() {
	s := test.Suite.Require

	src, err := source.New(test.url)
	s().NoError(err)

	// The binary must be installed to uninstall
	req := message.Request{
		Command:    DepInstalled,
		Parameters: key_value.Empty().Set("url", test.url),
	}
	rep, err := test.client.Request(&req)
	s().NoError(err)
	s().True(rep.IsOK())
	res, err := rep.Parameters.GetBoolean("installed")
	s().NoError(err)
	s().True(res)

	// Uninstall
	uninstallReq := message.Request{
		Command:    UninstallDep,
		Parameters: key_value.Empty().Set("src", src),
	}
	rep, err = test.client.Request(&uninstallReq)
	s().NoError(err)
	s().True(rep.IsOK())

	// wait a bit for effect
	time.Sleep(time.Millisecond * 100)

	// After uninstallation, we should not have the binary
	rep, err = test.client.Request(&req)
	s().NoError(err)
	s().True(rep.IsOK())
	res, err = rep.Parameters.GetBoolean("installed")
	s().NoError(err)
	s().False(res)
}

// Test_13_Run tests DepRunning, RunDep and CloseDep commands.
func (test *TestDepHandlerSuite) Test_13_Run() {
	s := test.Suite.Require

	depClient := &clientConfig.Client{
		ServiceUrl: "test-manager",
		Id:         test.id,
		Port:       6000,
		TargetType: handlerConfig.SocketType(handlerConfig.ReplierType),
	}

	src, err := source.New(test.url)
	s().NoError(err)
	src.SetBranch("server") // the sample server is written in this branch.

	// First, install the dependency
	installReq := message.Request{
		Command:    InstallDep,
		Parameters: key_value.Empty().Set("src", src),
	}
	rep, err := test.client.Request(&installReq)
	s().NoError(err)
	s().True(rep.IsOK())

	// Let's run it
	runReq := message.Request{
		Command: RunDep,
		Parameters: key_value.Empty().
			Set("parent", test.parent).
			Set("url", src.Url).
			Set("id", test.id),
	}
	rep, err = test.client.Request(&runReq)
	s().NoError(err)
	s().True(rep.IsOK())

	// Just wait a bit for initialization of the service
	time.Sleep(time.Millisecond * 100)

	// check that service is running
	runningReq := message.Request{
		Command: DepRunning,
		Parameters: key_value.Empty().
			Set("dep", depClient),
	}
	running, err := test.client.Request(&runningReq)
	s().NoError(err)
	s().True(running.IsOK())
	result, err := running.Parameters.GetBoolean("running")
	s().NoError(err)
	s().True(result)

	// Close the service
	closeReq := message.Request{
		Command: CloseDep,
		Parameters: key_value.Empty().
			Set("dep", depClient),
	}
	running, err = test.client.Request(&closeReq)
	s().NoError(err)
	s().True(running.IsOK())

	// Wait a bit for closing the source process
	time.Sleep(time.Millisecond * 100)

	// Checking for a running source after it was closed must fail
	running, err = test.client.Request(&runningReq)
	s().NoError(err)
	s().True(running.IsOK())
	result, err = running.Parameters.GetBoolean("running")
	s().NoError(err)
	s().False(result)

	// Clean out the installed files
	installReq.Command = UninstallDep
	rep, err = test.client.Request(&installReq)
	s().NoError(err)
	s().True(rep.IsOK())
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDepManager(t *testing.T) {
	suite.Run(t, new(TestDepHandlerSuite))
}
