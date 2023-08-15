package config

type ContextType = string

const (
	// DevContext indicates that all dependency proxies are in the local machine
	DevContext ContextType = "development"
	// DefaultContext indicates that all dependencies are in any machine.
	// It's unspecified.
	DefaultContext ContextType = "default"
)
