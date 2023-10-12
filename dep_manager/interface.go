package dep_manager

import (
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/log-lib"
)

// The Interface of the dependency manager.
//
// It doesn't have the `Stop` command.
// Because, stopping must be done by the remote call from other services.
type Interface interface {
	// Installed checks is the service binary exists
	Installed(dep *Dep) bool

	// Install the dependency from the source code. It compiles it.
	Install(dep *Dep, logger *log.Logger) error

	// Run the dependency with the given id and parent.
	Run(dep *Dep, id string, optionalParent ...*clientConfig.Client) error
	// Uninstall the dependency.
	Uninstall(dep *Dep) error

	// Lint sets the flags in the Dep if this depManager is managed by the DepManager
	Lint(*Dep)

	// Running checks is the service running or not
	Running(*clientConfig.Client) (bool, error)

	// Close the given dependency service
	Close(c *clientConfig.Client) error
}
