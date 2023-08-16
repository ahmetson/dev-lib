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
	"github.com/ahmetson/dev-lib/dep_manager"
	"github.com/ahmetson/handler-lib"
	"github.com/ahmetson/log-lib"
)

// A Context handles the config of the contexts
type Context struct {
	engine       config.Interface
	depManager   dep_manager.Interface
	controller   *handler.Controller
	serviceReady bool
	deps         map[string]string // id => url
}

// New creates Developer context.
// Loads it with the Dev Configuration and Dev Dep Manager.
func New() (*Context, error) {
	ctx := &Context{
		deps:       make(map[string]string),
		controller: nil,
	}

	logger, err := log.New(ctxConfig.DevContext, true)
	if err != nil {
		return nil, fmt.Errorf("log.New")
	}

	engine, err := config.New(logger)
	if err != nil {
		return nil, fmt.Errorf("config.NewDev: %w", err)
	}

	ctx.SetConfig(engine)
	if err := ctxConfig.SetDevDefaults(engine); err != nil {
		return nil, fmt.Errorf("config.SetDevDefaults: %w", err)
	}

	binPath := engine.GetString(ctxConfig.BinKey)
	srcPath := engine.GetString(ctxConfig.SrcKey)

	depManager, err := dep_manager.NewDev(srcPath, binPath)
	if err != nil {
		return nil, fmt.Errorf("dep_manager.new: %w", err)
	}

	if err := ctx.SetDepManager(depManager); err != nil {
		return nil, fmt.Errorf("ctx.SetDepManager: %w", err)
	}

	return ctx, nil
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

// SetDepManager sets the dependency manager in the context.
func (ctx *Context) SetDepManager(depManager dep_manager.Interface) error {
	if ctx.engine == nil {
		return fmt.Errorf("no configuration")
	}

	ctx.depManager = depManager

	return nil
}

// DepManager returns the dependency manager
func (ctx *Context) DepManager() dep_manager.Interface {
	return ctx.depManager
}

// Type returns the context type. Useful to identify contexts in the generic functions.
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
//	yamlConfig = key_value.NewDev(kv)
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
//
//// validateServicePath returns an error if the path is not a valid .yml link
//func validateServicePath(path string) error {
//	if len(path) < 5 || len(filepath.Base(path)) < 5 {
//		return fmt.Errorf("path is too short")
//	}
//	_, found := strings.CutSuffix(path, ".yml")
//	if !found {
//		return fmt.Errorf("the path should end with '.yml'")
//	}
//
//	return nil
//}

//func createYaml(configs ...*config.Service) key_value.KeyValue {
//	var services = configs
//	kv := key_value.Empty()
//	kv.Set("Services", services)
//
//	return kv
//}