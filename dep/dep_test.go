package dep

import (
	"github.com/ahmetson/log-lib"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestDepSuite struct {
	suite.Suite

	logger     *log.Logger
	src        *Src
	currentDir string
	url        string
}

// Make sure that Account is set to five
// before each test
func (test *TestDepSuite) SetupTest() {
	test.url = "github.com/ahmetson/test-manager"

	// Make sure that the folders don't exist. They will be added later
	test.src = &Src{
		Url: test.url,
	}
}

// TestNew tests the creation of the DepManager managers
func (test *TestDepSuite) Test_0_New() {
	s := &test.Suite

	// If we create the DepManager manager with 'NewDev,' it will create the folders.
	expected := "https://github.com/ahmetson/test-manager.git"
	src, err := New(test.src.Url)
	s.NoError(err)

	s.Equal(expected, src.GitUrl)

	test.src = src
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

func (test *TestDepSuite) Test_1_SetBranch() {
	s := &test.Suite
	branch := "main"

	// It should be empty since we didn't set any branch yet
	s.Empty(test.src.Branch)

	// Let's update the branch
	test.src.SetBranch(branch)

	// Should be updated in the struct
	s.Equal(branch, test.src.Branch)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDep(t *testing.T) {
	suite.Run(t, new(TestDepSuite))
}
