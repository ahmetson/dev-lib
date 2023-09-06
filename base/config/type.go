// Package config defines the specific parameters of the Contexts and Dev Context
package config

type ContextType = string

const (
	// DevContext indicates that all dependency proxies are in the local machine
	DevContext ContextType = "development"
	// DefaultContext indicates that the context is unspecified.
	DefaultContext ContextType = "default"

	ContextFlag = "context"
)
