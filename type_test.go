package context

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestTypeSuite struct {
	suite.Suite
}

// Make sure that Account is set to five
// before each test
func (suite *TestTypeSuite) SetupTest() {}

func (suite *TestTypeSuite) TestConstants() {
	fmt.Printf("Context Types: %s, %s\n", DevContext, UnknownContext)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestType(t *testing.T) {
	suite.Run(t, new(TestTypeSuite))
}
