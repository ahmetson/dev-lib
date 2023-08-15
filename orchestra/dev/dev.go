// Package dev handles the dependencies in the development environment,
// which means it's in the current machine.
//
// The dependencies are including the extensions and proxies.
//
// How it works?
//
// The orchestra is set up. It checks the folder. And if they are not existing, it will create them.
// >> dev.Run(orchestra)
//
// then lets work on the extension.
// User is passing an extension url.
// The service is checking whether it exists in the data or not.
// If the service exists, it gets the yaml. And return the config.
//
// If the service doesn't exist, it checks whether the service exists in the bin.
// If it exists, then it runs it with --build-config.
//
// Then, if the service doesn't exist in the bin, it checks the source.
// If the source exists, then it will call `go build`.
// Then call bin file with the generated files.
//
// Lastly, if a source doesn't exist, it will download the files from the repository using go-git.
// Then we build the binary.
// We generate config.
//
// Lastly, the service.Run() will make sure that all binaries exist.
// If not, then it will create them.
//
// -----------------------------------------------
// running the application will do the following.
// It checks the port of proxies is in use. If it's not, then it will call a run.
//
// Then it will call itself.
//
// The service will have a command to "shutdown" contexts. As well as "rebuild"
package dev

import (
	"fmt"
	"github.com/ahmetson/config-lib"
	ctxConfig "github.com/ahmetson/dev-lib/config"
	"github.com/ahmetson/dev-lib/dep"
	"github.com/ahmetson/handler-lib"
	"path/filepath"
	"strings"
)

// A Context handles the config of the contexts
type Context struct {
	engine       config.Interface
	depManager   dep.Interface
	controller   *handler.Controller
	serviceReady bool
	deps         map[string]string // id => url
}

// New creates Developer context.
// Call SetConfig() to prepare it.
func New() *Context {
	return &Context{
		deps:       make(map[string]string),
		controller: nil,
	}
}

// SetConfig sets the config engine of the given type.
// For the development context, it could be config-lib that reads the local file system.
//
// Setting up the configuration prepares the context by creating directories.
func (ctx *Context) SetConfig(engine config.Interface) {
	ctx.engine = engine
}

// Config returns the config engine in the context.
func (ctx *Context) Config() config.Interface {
	return ctx.engine
}

func (ctx *Context) SetDepManager(depManager dep.Interface) error {
	if ctx.engine == nil {
		return fmt.Errorf("no configuration")
	}

	ctx.depManager = depManager

	return nil
}

func (ctx *Context) DepManager() dep.Interface {
	return ctx.depManager
}

func (ctx *Context) Type() ctxConfig.ContextType {
	return ctxConfig.DevContext
}

//
//// GetConfig on the given path.
//// If a path is not obsolete, then it should be relative to the executable.
//// The path should have the .yml extension
//func (ctx *Context) GetConfig(url string) (*config.Service, error) {
//	path := ctx.ConfigurationPath(url)
//
//	if err := validateServicePath(path); err != nil {
//		return nil, fmt.Errorf("validateServicePath: %w", err)
//	}
//
//	bytes, err := os.ReadFile(path)
//	if err != nil {
//		return nil, fmt.Errorf("os.ReadFile of %s: %w", path, err)
//	}
//
//	yamlConfig := createYaml()
//	kv := yamlConfig.Map()
//	err = yaml.Unmarshal(bytes, &kv)
//
//	if err != nil {
//		return nil, fmt.Errorf("yaml.Unmarshal of %s: %w", path, err)
//	}
//
//	fmt.Println("service", kv)
//
//	yamlConfig = key_value.New(kv)
//	if err := yamlConfig.Exist("Services"); err != nil {
//		return nil, fmt.Errorf("no services in yaml: %w", err)
//	}
//
//	services, err := yamlConfig.GetKeyValueList("Services")
//	if err != nil {
//		return nil, fmt.Errorf("failed to get services as key value list: %w", err)
//	}
//
//	if len(services) == 0 {
//		return nil, fmt.Errorf("no services in the config")
//	}
//
//	var serviceConfig config.Service
//	err = services[0].Interface(&serviceConfig)
//	if err != nil {
//		return nil, fmt.Errorf("convert key value to Service: %w", err)
//	}
//
//	err = serviceConfig.PrepareService()
//	if err != nil {
//		return nil, fmt.Errorf("prepareService: %w", err)
//	}
//
//	return &serviceConfig, nil
//}
//
//// WriteService writes the service as the yaml on the given path.
//// If the path doesn't contain the file extension, it will through an error
//func (ctx *Context) SetConfig(url string, service *config.Service) error {
//	path := ctx.ConfigurationPath(url)
//
//	if err := validateServicePath(path); err != nil {
//		return fmt.Errorf("validateServicePath: %w", err)
//	}
//
//	kv := createYaml(service)
//
//	serviceConfig, err := yaml.Marshal(kv.Map())
//	if err != nil {
//		return fmt.Errorf("failed to marshall config.Service: %w", err)
//	}
//
//	f, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
//	_, err = f.Write(serviceConfig)
//	closeErr := f.Close()
//	if err != nil {
//		return fmt.Errorf("failed to write service into the given path: %w", err)
//	} else if closeErr != nil {
//		return fmt.Errorf("failed to close the file descriptor: %w", closeErr)
//	} else {
//		return nil
//	}
//}

// validateServicePath returns an error if the path is not a valid .yml link
func validateServicePath(path string) error {
	if len(path) < 5 || len(filepath.Base(path)) < 5 {
		return fmt.Errorf("path is too short")
	}
	_, found := strings.CutSuffix(path, ".yml")
	if !found {
		return fmt.Errorf("the path should end with '.yml'")
	}

	return nil
}

//func createYaml(configs ...*config.Service) key_value.KeyValue {
//	var services = configs
//	kv := key_value.Empty()
//	kv.Set("Services", services)
//
//	return kv
//}
