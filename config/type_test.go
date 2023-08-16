package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestConfigSuite struct {
	suite.Suite
}

// Make sure that Account is set to five
// before each test
func (suite *TestConfigSuite) SetupTest() {}

func (suite *TestConfigSuite) TestConstants() {
	fmt.Printf("context types: %s, %s\n", DevContext, DefaultContext)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestConfig(t *testing.T) {
	suite.Run(t, new(TestConfigSuite))
}
