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

	logger *log.Logger
	dep    *Dep
}

// Make sure that Account is set to five
// before each test
func (suite *TestDepSuite) SetupTest() {
	logger, _ := log.New("test dep manager", false)
	suite.logger = logger
	suite.dep = &Dep{
		Src: "./src",
	}
}

func (suite *TestDepSuite) TestPath() {
	url := "github.com/ahmetson/test"
	expected := filepath.Join("./src/github.com.ahmetson.test")
	suite.Suite.Equal(expected, suite.dep.srcPath(url))

	execPath, err := path.GetExecPath()
	suite.Suite.NoError(err)
	suite.dep.Src = execPath
	url = "config"
	suite.logger.Info("the source path", "path", suite.dep.srcPath(url))
	exist, err := suite.dep.srcExist(url)
	suite.Suite.NoError(err)
	suite.Suite.True(exist)
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *TestDepSuite) TestUtils() {
	suite.logger.Info("Test utils")
	url := "github.com/ahmetson/test-ext"
	fileName := "github.com.ahmetson.test-ext"
	suite.Require().Equal(urlToFileName(url), fileName)

	invalid := "github.com\\ahmetson\\test-ext"
	suite.Require().Equal(urlToFileName(invalid), fileName)

	// with semicolon
	url = "::github.com/ahmetson/test-ext"
	suite.Require().Equal(urlToFileName(url), fileName)

	// with space
	url = "::github.com/ahmetson/  test-ext  "
	suite.Require().Equal(urlToFileName(url), fileName)

	suite.logger.Info("url to file name", "url", invalid, "filename", urlToFileName(invalid))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDep(t *testing.T) {
	suite.Run(t, new(TestDepSuite))
}
