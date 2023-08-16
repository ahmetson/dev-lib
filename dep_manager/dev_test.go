package dep_manager

import (
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestDepSuite struct {
	suite.Suite

	logger     *log.Logger
	dep        *Dep
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
	test.dep = &Dep{
		Src: srcPath,
		Bin: binPath,
	}

	// A valid source code that we want to download
	test.url = "github.com/ahmetson/test-manager"
}

// TestNew tests the creation of the Dep managers
func (test *TestDepSuite) Test_0_New() {
	s := &test.Suite

	// Before testing, we make sure that the files don't exist
	exist, err := path.DirExist(test.dep.Bin)
	s.NoError(err)
	s.False(exist)

	exist, err = path.DirExist(test.dep.Src)
	s.NoError(err)
	s.False(exist)

	// If we create the Dep manager with 'NewDev,' it will create the folders.
	dep, err := NewDev(test.dep.Src, test.dep.Bin)
	s.NoError(err)

	// Now we can check for the directories
	exist, _ = path.DirExist(dep.Src)
	s.True(exist)

	exist, _ = path.DirExist(dep.Bin)
	s.True(exist)

	test.dep = dep
}

// TestConvertToGitUrl tests converting url to git url.
// Since dev dep manager uses git for loading the files.
func (test *TestDepSuite) Test_1_ConvertToGitUrl() {
	s := &test.Suite

	// valid
	url := "github.com/ahmetson/test"
	expected := "https://github.com/ahmetson/test.git"
	gitUrl, err := convertToGitUrl(url)
	s.NoError(err)
	s.Equal(expected, gitUrl)

	// invalid url
	url = "../local_dir"
	_, err = convertToGitUrl(url)
	s.Error(err)

	// having a schema prefix will fail
	url = "file://file"
	_, err = convertToGitUrl(url)
	s.Error(err)

}

// TestUrlToFileName tests the utility function that converts the URL into the file name.
func (test *TestDepSuite) Test_2_UrlToFileName() {
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

// TestSourcePath tests the utility functions related to the paths
func (test *TestDepSuite) Test_3_SourcePath() {
	url := "github.com/ahmetson/test-manager"
	expected := filepath.Join(test.dep.Src, "github.com.ahmetson.test-manager")
	test.Suite.Equal(expected, test.dep.srcPath(url))
}

// TestDownload makes sure to downloadSrc the remote repository into the context.
// This is the first part of Install.
// The second part of Install is building.
//
// Tests Dep.downloadSrc and srcExist.
func (test *TestDepSuite) Test_4_Download() {
	s := &test.Suite

	// There should not be any source code before downloading
	exist, err := test.dep.srcExist(test.url)
	s.NoError(err)
	s.False(exist)

	// download the source code
	err = test.dep.downloadSrc(test.url, test.logger)
	s.NoError(err)

	// There should be a source code
	exist, _ = test.dep.srcExist(test.url)
	s.True(exist)

	//
	// Testing the failures
	//
	url := "github.com/ahmetson/no-repo" // this repo doesn't exist
	err = test.dep.downloadSrc(url, test.logger)
	s.Error(err)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDep(t *testing.T) {
	suite.Run(t, new(TestDepSuite))
}
