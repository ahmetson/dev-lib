package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestDepManagerSuite struct {
	suite.Suite
}

// Make sure that Account is set to five
// before each test
func (suite *TestDepManagerSuite) SetupTest() {}

func (suite *TestDepManagerSuite) TestConstants() {
	fmt.Printf("Configuration keys: source path: %s, bin path: %s\n", SrcKey, BinKey)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDepManager(t *testing.T) {
	suite.Run(t, new(TestDepManagerSuite))
}
