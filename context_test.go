package context

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestCtxSuite struct {
	suite.Suite
}

// Make sure that Account is set to five
// before each test
func (test *TestCtxSuite) SetupTest() {}

// Test_0_New tests the creation of the DepManager managers
func (test *TestCtxSuite) Test_0_New() {
	s := &test.Suite

	// Before testing, we make sure that the files don't exist
	_, err := New(UnknownContext)
	s.Require().Error(err, "only dev context supported")

	ctx, err := New(DevContext)
	s.Require().NoError(err)

	s.Require().Equal(DevContext, ctx.Type())
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCtx(t *testing.T) {
	suite.Run(t, new(TestCtxSuite))
}
