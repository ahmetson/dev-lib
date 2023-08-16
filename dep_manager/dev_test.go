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

}

// TestNew tests the creation of the Dep managers
func (test *TestDepSuite) TestNew() {
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

// TestPath tests the utility functions related to the paths
func (test *TestDepSuite) TestPath() {
	url := "github.com/ahmetson/test"
	expected := filepath.Join(test.dep.Src, "github.com.ahmetson.test")
	test.Suite.Equal(expected, test.dep.srcPath(url))

	//execPath, err := path.CurrentDir()
	//test.Suite.NoError(err)
	//test.dep.Src = execPath
	//url = "config"
	//test.logger.Info("the source path", "path", test.dep.srcPath(url))
	//exist, err := test.dep.srcExist(url)
	//test.Suite.NoError(err)
	//test.Suite.True(exist)
}

// TestUrlToFileName tests the utility function that converts the URL into the file name.
func (test *TestDepSuite) TestUrlToFileName() {
	test.logger.Info("Test utils")
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

	test.logger.Info("url to file name", "url", invalid, "filename", urlToFileName(invalid))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDep(t *testing.T) {
	suite.Run(t, new(TestDepSuite))
}
