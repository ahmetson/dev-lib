package dep_manager

import (
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/dev-lib/dep"
	"github.com/ahmetson/log-lib"
)

// The Interface of the dependency manager.
//
// This one is initiated by the orchestra.
type Interface interface {
	// Installed checks is the service installed
	Installed(url string) bool
	// Running checks is the service running or not
	Running(*clientConfig.Client) (bool, error)
	// Install the dependency from the source code. It compiles it.
	Install(src *dep.Src, log *log.Logger) error
	// Run the dependency. The url of the dependency. It's id. and the parameters of the parent to connect to.
	Run(url string, id string, parent *clientConfig.Client, logger *log.Logger) error
	// Uninstall the dependency.
	Uninstall(src *dep.Src) error
}
