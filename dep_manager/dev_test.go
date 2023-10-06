package dep_manager

import (
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"github.com/stretchr/testify/suite"
	"testing"
)

// todo for public functions test with the nil values
// todo for public functions test with non linted dep

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestDepManagerSuite struct {
	suite.Suite

	logger     *log.Logger
	depManager *DepManager          // the manager to test
	currentDir string               // executable to store the binaries and source codes
	url        string               // dependency source code
	id         string               // the id of the dependency
	parent     *clientConfig.Client // the info about the service to which dependency should connect
}

// Make sure that Account is set to five
// before each test
func (test *TestDepManagerSuite) SetupTest() {
	logger, _ := log.New("TestDepManagerSuite", false)
	test.logger = logger

	currentDir, err := path.CurrentDir()
	test.Suite.NoError(err)
	test.currentDir = currentDir

	srcPath := path.AbsDir(currentDir, "_sds/src")
	binPath := path.AbsDir(currentDir, "_sds/bin")

	// Make sure that the folders don't exist. They will be added later
	test.depManager = &DepManager{
		Src:         srcPath,
		Bin:         binPath,
		runningDeps: make(map[string]*Dep, 0),
	}

	// A valid source code that we want to download
	test.url = "github.com/ahmetson/test-manager"

	test.id = "test-manager"
	test.parent = &clientConfig.Client{
		ServiceUrl: "dev-lib",
		Id:         "parent",
		Port:       120,
	}
}

// Test_0_New tests the creation of the DepManager managers
func (test *TestDepManagerSuite) Test_0_New() {
	s := test.Require

	// Before testing, we make sure that the files don't exist
	exist, err := path.DirExist(test.depManager.Bin)
	s().NoError(err)
	s().False(exist)

	exist, err = path.DirExist(test.depManager.Src)
	s().NoError(err)
	s().False(exist)

	// If we create the DepManager manager with 'New,' it will create the folders.
	depManager := New()
	err = depManager.SetPaths(test.depManager.Src, test.depManager.Bin)
	s().NoError(err)

	// Now we can check for the directories
	exist, _ = path.DirExist(depManager.Src)
	s().True(exist)

	exist, _ = path.DirExist(depManager.Bin)
	s().True(exist)

	test.depManager = depManager
}

// Test_1_UrlToFileName tests the utility function that converts the URL into the file name.
func (test *TestDepManagerSuite) Test_1_UrlToFileName() {
	url := "github.com/ahmetson/test-ext"
	fileName := "github.com.ahmetson.test-ext"
	test.Require().Equal(urlToFileName(url), fileName)

	invalid := "github.com\\ahmetson\\test-ext"
	test.Require().Equal(urlToFileName(invalid), fileName)

	// with semicolon
	url = "::github.com/ahmetson/test-ext"
	test.Require().Equal(urlToFileName(url), fileName)

	// with space
	url = "::github.com/ahmetson/  test-ext  "
	test.Require().Equal(urlToFileName(url), fileName)
}

//// todo change with testing Lint.
// Test_12_SourcePath tests the utility functions related to the paths
//func (test *TestDepManagerSuite) Test_12_SourcePath() {
//	url := "github.com/ahmetson/test-manager"
//	expected := filepath.Join(test.depManager.Src, "github.com.ahmetson.test-manager")
//	test.Suite.Equal(expected, test.depManager.srcPath(url))
//}

// Test_13_downloadSrc makes sure to downloadSrc the remote repository into the context.
// This is the first part of Install.
// The second part of Install is building.
//
// Tests DepManager.downloadSrc and srcExist.
func (test *TestDepManagerSuite) Test_13_downloadSrc() {
	s := test.Require

	dep, err := NewDep(test.url, "", "'")
	s().NoError(err)

	s().False(dep.IsLinted())
	test.depManager.Lint(dep)
	s().True(dep.IsLinted())

	// There should not be any source code before downloading
	exist, err := test.depManager.srcExist(dep)
	s().NoError(err)
	s().False(exist)

	// download the source code
	err = test.depManager.downloadSrc(dep, test.logger)
	s().NoError(err)

	// There should be a source code
	exist, _ = test.depManager.srcExist(dep)
	s().True(exist)

	//
	// Testing the failures
	//
	url := "github.com/ahmetson/no-repo" // this repo doesn't exist
	dep, err = NewDep(url, "", "")
	s().NoError(err)
	err = test.depManager.downloadSrc(dep, test.logger)
	s().Error(err)
}

// todo change code to test Installed separately, and installed with manageable and non manageable code.
// Test_14_Build will compile the source code downloaded in Test_3_Download
func (test *TestDepManagerSuite) Test_14_Build() {
	s := test.Require

	dep, err := NewDep(test.url, "", "")
	s().NoError(err)

	test.depManager.Lint(dep)
	s().True(dep.manageableBin)

	// There should not be any binary before building
	exist := test.depManager.Installed(dep)
	s().False(exist)

	// build the binaries
	err = test.depManager.build(dep, test.logger)
	s().NoError(err)

	// There should be a binary after testing
	exist = test.depManager.Installed(dep)
	s().True(exist)
}

// Test_15_DeleteSrc deletes the dependency's source code.
// The dependency was downloaded in Test_3_Download
func (test *TestDepManagerSuite) Test_15_DeleteSrc() {
	s := test.Require

	dep, err := NewDep(test.url, "", "")
	s().NoError(err)
	test.depManager.Lint(dep)

	// There should be a source code
	exist, _ := test.depManager.srcExist(dep)
	s().True(exist)

	// Delete the source code
	err = test.depManager.deleteSrc(dep)
	s().NoError(err)

	// There should not be a source code
	exist, err = test.depManager.srcExist(dep)
	s().NoError(err)
	s().False(exist)
}

// Test_16_DeleteBin deletes the dependency's binary.
// The binary was created by Test_4_Build
func (test *TestDepManagerSuite) Test_16_DeleteBin() {
	s := test.Require

	dep, err := NewDep(test.url, "", "")
	s().NoError(err)
	test.depManager.Lint(dep)

	// The binary should be presented
	// There should not be any binary before building
	exist := test.depManager.Installed(dep)
	s().True(exist)

	// Delete the binary
	err = test.depManager.deleteBin(dep)
	s().NoError(err)

	// The binary should be removed from the file
	exist = test.depManager.Installed(dep)
	s().False(exist)
}

// Test_17_Install is the combination of Test_3_Download and Test_4_Build.
// It's a functional test.
func (test *TestDepManagerSuite) Test_17_Install() {
	s := test.Require

	dep, err := NewDep(test.url, "", "")
	s().NoError(err)
	test.depManager.Lint(dep)

	// There should not be installed binary
	// The binary should be presented
	// There should not be any binary before building
	exist := test.depManager.Installed(dep)
	s().False(exist)

	// Install the dependency
	err = test.depManager.Install(dep, test.logger)
	s().NoError(err)

	// The binary should exist
	exist = test.depManager.Installed(dep)
	s().True(exist)
}

// Test_18_Uninstall is the combination of Test_5_DeleteSrc and Test_6_DeleteBin.
func (test *TestDepManagerSuite) Test_18_Uninstall() {
	s := test.Require

	dep, err := NewDep(test.url, "", "")
	s().NoError(err)
	test.depManager.Lint(dep)

	// Test_7_Install should install the binary.
	exist := test.depManager.Installed(dep)
	s().True(exist)

	// Uninstall
	err = test.depManager.Uninstall(dep)
	s().NoError(err)

	// After uninstallation, we should not have the binary
	exist = test.depManager.Installed(dep)
	s().False(exist)
}

// Test_19_Uninstall is the combination of Test_5_DeleteSrc and Test_6_DeleteBin.
func (test *TestDepManagerSuite) Test_19_InvalidCompile() {
	s := test.Require

	uncompilableDep, err := NewDep(test.url, "", "")
	s().NoError(err)
	uncompilableDep.SetBranch("uncompilable")
	test.depManager.Lint(uncompilableDep)

	// todo use the local source code
	// download the src
	err = test.depManager.downloadSrc(uncompilableDep, test.logger)
	s().NoError(err)

	// building must fail, since "uncompilable" branch code is not buildable
	err = test.depManager.build(uncompilableDep, test.logger)
	s().Error(err)

	// todo don't delete the local source code
	// delete the source code
	err = test.depManager.deleteSrc(uncompilableDep)
	s().NoError(err)
}

//// Test_20_Run runs the given binary.
//func (test *TestDepManagerSuite) Test_20_Run() {
//	s := &test.Suite
//
//	src, err := source.New(test.url)
//	s.Require().NoError(err)
//
//	// First, install the manager
//	err = test.depManager.Install(src, test.logger)
//	s.NoError(err)
//
//	// Let's run it, it should exit immediately
//	for i := 0; i < 30; i++ {
//		err = test.depManager.Run(src.Url, test.id, test.parent)
//		if err == nil {
//			break
//		}
//
//		time.Sleep(time.Second)
//	}
//	s.Require().NoError(err)
//
//	// Just to see the exit message
//	time.Sleep(time.Millisecond * 100)
//	s.Require().NoError(test.depManager.exitErr)
//
//	// Clean out the installed files
//	err = test.depManager.Uninstall(src)
//	s.NoError(err)
//}
//
//// Test_21_RunError runs the binary that exits with error.
//// Dependency manager must show it
//func (test *TestDepManagerSuite) Test_21_RunError() {
//	s := &test.Suite
//
//	src, err := source.New(test.url)
//	s.Require().NoError(err)
//	src.SetBranch("error-exit") // this branch intentionally exits the program with an error.
//
//	// First, install the manager
//	err = test.depManager.Install(src, test.logger)
//	s.NoError(err)
//
//	// Let's run it
//	for i := 0; i < 30; i++ {
//		err = test.depManager.Run(src.Url, test.id, test.parent)
//		if err == nil {
//			break
//		}
//
//		time.Sleep(time.Second)
//	}
//	s.Require().NoError(err)
//
//	// Just to see the exit message.
//	// The 0.1 seconds.
//	// That's how long the program waits before exit.
//	// Other 0.2 seconds are for some end of the background work.
//	time.Sleep(time.Millisecond * 300)
//	test.logger.Info("exit status", "err", test.depManager.exitErr)
//	s.Require().Error(test.depManager.exitErr)
//
//	// Clean out the installed files
//	err = test.depManager.Uninstall(src)
//	s.NoError(err)
//}
//
//// Test_22_Running checks that service is running
//func (test *TestDepManagerSuite) Test_22_Running() {
//	s := &test.Suite
//
//	client := &clientConfig.Client{
//		ServiceUrl: "test-manager",
//		Id:         test.id,
//		Port:       6000,
//	}
//
//	src, err := source.New(test.url)
//	s.Require().NoError(err)
//	src.SetBranch("server") // the sample server is written in this branch.
//
//	// First, install the manager
//	err = test.depManager.Install(src, test.logger)
//	s.NoError(err)
//
//	// Let's run it
//	for i := 0; i < 30; i++ {
//		err = test.depManager.Run(src.Url, test.id, test.parent)
//		if err == nil {
//			break
//		}
//
//		time.Sleep(time.Second)
//	}
//	s.Require().NoError(err)
//
//	// waiting for initialization...
//	time.Sleep(time.Millisecond * 200)
//	s.Require().NotNil(test.depManager.cmd[test.id]) // cmd == nil indicates that the program was closed
//
//	// Check is the service running
//	, err := test.depManager.Running(client)
//	s.Require().NoError(err)
//	s.True(running)
//
//	// service is running two seconds. after that running should return false
//	time.Sleep(time.Second * 3)
//	s.Require().Nil(test.depManager.cmd[test.id]) // cmd == nil indicates that the program was closed
//	running, err = test.depManager.Running(client)
//	s.Require().NoError(err)
//	s.False(running)
//
//	// Clean out the installed files
//	err = test.depManager.Uninstall(src)
//	s.NoError(err)
//}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDepManager(t *testing.T) {
	suite.Run(t, new(TestDepManagerSuite))
}
