package dep_client

import (
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/dev-lib/dep_handler"
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
type TestDepClientSuite struct {
	suite.Suite

	logger     *log.Logger
	dep        *dep_handler.DepHandler // the manager to test
	currentDir string                  // executable to store the binaries and source codes
	url        string                  // dependency source code
	id         string                  // the id of the dependency
	parent     *clientConfig.Client    // the info about the service to which dependency should connect

	client *Client
}

// Make sure that Account is set to five
// before each test
func (test *TestDepClientSuite) SetupTest() {
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

	test.dep, err = dep_handler.New(manager)
	s().NoError(err)

	// Start the handler
	go func() {
		s().NoError(test.dep.Start())
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

	socket, err := New()
	s().NoError(err)

	test.client = socket
	test.client.Timeout(time.Second * 30)
	test.client.Attempt(1)
}

func (test *TestDepClientSuite) TearDownTest() {
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
func (test *TestDepClientSuite) Test_10_Install() {
	s := test.Suite.Require

	src, err := source.New(test.url)
	s().NoError(err)

	// installation must fail since nothing installed
	installed, err := test.client.Installed(test.url)
	s().NoError(err)
	s().False(installed)

	// There should be a source code
	err = test.client.Install(src)
	s().NoError(err)

	// wait a bit until its installed
	time.Sleep(time.Millisecond * 100)

	//
	// Testing the installed after installation
	//
	installed, err = test.client.Installed(test.url)
	s().NoError(err)
	s().True(installed)
}

// Test_11_Uninstall deletes the binary and source code installed at Test_11_Install
func (test *TestDepClientSuite) Test_11_Uninstall() {
	s := test.Suite.Require

	src, err := source.New(test.url)
	s().NoError(err)

	// The binary must be installed to uninstall
	installed, err := test.client.Installed(test.url)
	s().NoError(err)
	s().True(installed)

	// Uninstall
	err = test.client.Uninstall(src)
	s().NoError(err)

	// wait a bit for effect
	time.Sleep(time.Millisecond * 100)

	// After uninstallation, we should not have the binary
	installed, err = test.client.Installed(test.url)
	s().NoError(err)
	s().False(installed)
}

// Test_13_Run tests DepRunning, RunDep and CloseDep commands.
func (test *TestDepClientSuite) Test_13_Run() {
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
	err = test.client.Install(src)
	s().NoError(err)

	// Let's run the dependency
	err = test.client.Run(src.Url, test.id, test.parent)
	s().NoError(err)

	// Just wait a bit for initialization of the service
	time.Sleep(time.Millisecond * 100)

	// check that service is running
	running, err := test.client.Running(depClient)
	s().NoError(err)
	s().True(running)

	// CloseDep the service
	err = test.client.CloseDep(depClient)
	s().NoError(err)

	// Wait a bit for closing the source process
	time.Sleep(time.Millisecond * 100)

	// Checking for a running source after it was closed must fail
	running, err = test.client.Running(depClient)
	s().NoError(err)
	s().False(running)

	// Clean out the installed files
	err = test.client.Uninstall(src)
	s().NoError(err)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDepClient(t *testing.T) {
	suite.Run(t, new(TestDepClientSuite))
}
