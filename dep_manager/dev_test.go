package dep_manager

import (
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/dev-lib/dep"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestDepSuite struct {
	suite.Suite

	logger     *log.Logger
	dep        *DepManager
	currentDir string
	url        string
}

// Make sure that Account is set to five
// before each test
func (test *TestDepSuite) SetupTest() {
	logger, _ := log.New("test dep manager", false)
	test.logger = logger

	currentDir, err := path.CurrentDir()
	test.Suite.NoError(err)
	test.currentDir = currentDir

	srcPath := path.AbsDir(currentDir, "_sds/src")
	binPath := path.AbsDir(currentDir, "_sds/bin")

	// Make sure that the folders don't exist. They will be added later
	test.dep = &DepManager{
		Src: srcPath,
		Bin: binPath,
	}

	// A valid source code that we want to download
	test.url = "github.com/ahmetson/test-manager"
}

// Test_0_New tests the creation of the DepManager managers
func (test *TestDepSuite) Test_0_New() {
	s := &test.Suite

	// Before testing, we make sure that the files don't exist
	exist, err := path.DirExist(test.dep.Bin)
	s.NoError(err)
	s.False(exist)

	exist, err = path.DirExist(test.dep.Src)
	s.NoError(err)
	s.False(exist)

	// If we create the DepManager manager with 'NewDev,' it will create the folders.
	depManager, err := NewDev(test.dep.Src, test.dep.Bin)
	s.NoError(err)

	// Now we can check for the directories
	exist, _ = path.DirExist(depManager.Src)
	s.True(exist)

	exist, _ = path.DirExist(depManager.Bin)
	s.True(exist)

	test.dep = depManager
}

// Test_1_UrlToFileName tests the utility function that converts the URL into the file name.
func (test *TestDepSuite) Test_1_UrlToFileName() {
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

// Test_2_SourcePath tests the utility functions related to the paths
func (test *TestDepSuite) Test_12_SourcePath() {
	url := "github.com/ahmetson/test-manager"
	expected := filepath.Join(test.dep.Src, "github.com.ahmetson.test-manager")
	test.Suite.Equal(expected, test.dep.srcPath(url))
}

// Test_3_Download makes sure to downloadSrc the remote repository into the context.
// This is the first part of Install.
// The second part of Install is building.
//
// Tests DepManager.downloadSrc and srcExist.
func (test *TestDepSuite) Test_13_Download() {
	s := &test.Suite

	// There should not be any source code before downloading
	exist, err := test.dep.srcExist(test.url)
	s.NoError(err)
	s.False(exist)

	src, err := dep.New(test.url)
	s.NoError(err)

	// download the source code
	err = test.dep.downloadSrc(src, test.logger)
	s.NoError(err)

	// There should be a source code
	exist, _ = test.dep.srcExist(test.url)
	s.True(exist)

	//
	// Testing the failures
	//
	url := "github.com/ahmetson/no-repo" // this repo doesn't exist
	src, err = dep.New(url)
	s.NoError(err)
	err = test.dep.downloadSrc(src, test.logger)
	s.Error(err)
}

// Test_4_Build will compile the source code downloaded in Test_3_Download
func (test *TestDepSuite) Test_14_Build() {
	s := &test.Suite

	// There should not be any binary before building
	exist := test.dep.Installed(test.url)
	s.False(exist)

	// build the binaries
	err := test.dep.build(test.url, test.logger)
	s.NoError(err)

	// There should be a binary after testing
	exist = test.dep.Installed(test.url)
	s.True(exist)
}

// Test_5_DeleteSrc deletes the dependency's source code.
// The dependency was downloaded in Test_3_Download
func (test *TestDepSuite) Test_15_DeleteSrc() {
	s := &test.Suite

	// There should be a source code
	exist, _ := test.dep.srcExist(test.url)
	s.True(exist)

	// Delete the source code
	err := test.dep.deleteSrc(test.url)
	s.NoError(err)

	// There should not be a source code
	exist, err = test.dep.srcExist(test.url)
	s.NoError(err)
	s.False(exist)
}

// Test_6_DeleteBin deletes the dependency's binary.
// The binary was created by Test_4_Build
func (test *TestDepSuite) Test_16_DeleteBin() {
	s := &test.Suite

	// The binary should be presented
	// There should not be any binary before building
	exist := test.dep.Installed(test.url)
	s.True(exist)

	// Delete the binary
	err := test.dep.deleteBin(test.url)
	s.NoError(err)

	// The binary should be removed from the file
	exist = test.dep.Installed(test.url)
	s.False(exist)
}

// Test_7_Install is the combination of Test_3_Download and Test_4_Build.
func (test *TestDepSuite) Test_17_Install() {
	s := &test.Suite

	src, err := dep.New(test.url)
	s.NoError(err)

	// There should not be installed binary
	// The binary should be presented
	// There should not be any binary before building
	exist := test.dep.Installed(test.url)
	s.False(exist)

	// Install the dependency
	err = test.dep.Install(src, test.logger)
	s.NoError(err)

	// The binary should exist
	exist = test.dep.Installed(test.url)
	s.True(exist)
}

// Test_8_Uninstall is the combination of Test_5_DeleteSrc and Test_6_DeleteBin.
func (test *TestDepSuite) Test_18_Uninstall() {
	s := &test.Suite

	src, err := dep.New(test.url)
	s.NoError(err)

	// Test_7_Install should install the binary.
	exist := test.dep.Installed(test.url)
	s.Require().True(exist)

	// Uninstall
	err = test.dep.Uninstall(src)
	s.NoError(err)

	// After uninstallation, we should not have the binary
	exist = test.dep.Installed(test.url)
	s.False(exist)
}

// Test_8_Uninstall is the combination of Test_5_DeleteSrc and Test_6_DeleteBin.
func (test *TestDepSuite) Test_19_InvalidCompile() {
	s := &test.Suite

	src, err := dep.New(test.url)
	s.NoError(err)
	src.SetBranch("uncompilable")

	// download the src
	err = test.dep.downloadSrc(src, test.logger)
	s.NoError(err)

	// building must fail, since "uncompilable" branch code is not buildable
	err = test.dep.build(src.Url, test.logger)
	s.Error(err)

	// delete the source code
	err = test.dep.deleteSrc(src.Url)
	s.NoError(err)
}

// Test_10_Run runs the given binary.
func (test *TestDepSuite) Test_20_Run() {
	s := &test.Suite

	id := "test-manager"
	parent := &clientConfig.Client{
		Url:  "dev-lib",
		Id:   "parent",
		Port: 120,
	}

	src, err := dep.New(test.url)
	s.Require().NoError(err)

	// First, install the manager
	err = test.dep.Install(src, test.logger)
	s.NoError(err)

	// Let's run it, it should exit immediately
	err = test.dep.Run(src.Url, id, parent, test.logger)
	s.Require().NoError(err)

	// Just to see the exit message
	time.Sleep(time.Second)
	s.Require().NoError(test.dep.exitErr)

	// Clean out the installed files
	err = test.dep.Uninstall(src)
	s.NoError(err)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDep(t *testing.T) {
	suite.Run(t, new(TestDepSuite))
}
