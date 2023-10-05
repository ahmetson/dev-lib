package dep_manager

import (
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/dev-lib/source"
	"github.com/ahmetson/log-lib"
)

// The Interface of the dependency manager.
//
// It doesn't have the `Stop` command.
// Because, stopping must be done by the remote call from other services.
type Interface interface {
	// Installed checks is the service installed
	Installed(url string) bool
	// Running checks is the service running or not
	Running(*clientConfig.Client) (bool, error)
	// Install the dependency from the source code. It compiles it.
	Install(src *source.Src, logger *log.Logger) error
	// Run the dependency. The url of the dependency. It's id. and the parameters of the parent to connect to.
	Run(url string, id string, parent *clientConfig.Client) error
	// Uninstall the dependency.
	Uninstall(src *source.Src) error
	// Close the given dependency service
	Close(c *clientConfig.Client) error
}
