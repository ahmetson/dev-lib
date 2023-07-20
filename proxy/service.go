// Package proxy defines the script that acts as the middleware
package proxy

import (
	"fmt"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/service-lib/configuration"
	"github.com/ahmetson/service-lib/controller"
	"github.com/ahmetson/service-lib/log"
	"sync"
)

// Service of the proxy type
type Service struct {
	configuration configuration.Service
	sources       key_value.KeyValue
	controller    *Controller
}

// SourceName of this type should be listed within the controllers in the configuration
const SourceName = "source"

// DestinationName of this type should be listed within the controllers in the configuration
const DestinationName = "destination"

// extension creates the configuration of the proxy controller.
// The proxy controller itself is added as the extension to the source controllers,
// to the request handlers and to the reply handlers.
func extension() *configuration.Extension {
	return configuration.NewInternalExtension(ControllerName)
}

func validateConfiguration(service configuration.Service) error {
	if len(service.Controllers) < 2 {
		return fmt.Errorf("not enough controllers were given. atleast 'source' and 'destination' should be")
	}

	sourceFound := false
	destinationFound := false
	for _, c := range service.Controllers {
		if c.Name == SourceName {
			sourceFound = true
		} else if c.Name == DestinationName {
			destinationFound = true
		}
	}

	if !sourceFound {
		return fmt.Errorf("proxy service '%s' in seascape.yml doesn't have '%s' controller", service.Name, SourceName)
	}

	if !destinationFound {
		return fmt.Errorf("proxy service '%s' in seascape.yml doesn't have '%s' controller", service.Name, DestinationName)
	}

	return nil
}

// registerNonSources registers the controller instances as the destination.
// it skips the SourceName named controllers as the destination.
func registerNonSources(controllers []configuration.Controller, proxyController *Controller) error {
	for _, c := range controllers {
		if c.Name == SourceName {
			continue
		}

		for _, instance := range c.Instances {
			err := proxyController.RegisterDestination(&instance)
			if err != nil {
				return fmt.Errorf("proxyController.RegistartionDestination: %w", err)
			}
		}
	}

	return nil
}

// New proxy service based on the configurations
func New(serviceConf configuration.Service, logger log.Logger) (*Service, error) {
	if serviceConf.Type != configuration.ProxyType {
		return nil, fmt.Errorf("service type in the configuration is not Independent. It's '%s'", serviceConf.Type)
	}
	if err := validateConfiguration(serviceConf); err != nil {
		return nil, fmt.Errorf("validateConfiguration: %w", err)
	}

	proxyController, err := newController(logger)
	if err != nil {
		return nil, fmt.Errorf("newController: %w", err)
	}
	err = registerNonSources(serviceConf.Controllers, proxyController)
	if err != nil {
		return nil, fmt.Errorf("registerNonSources: %w", err)
	}

	service := Service{
		configuration: serviceConf,
		sources:       key_value.Empty(),
		controller:    proxyController,
	}

	return &service, nil
}

// NewSourceController creates a source controller of the given type.
//
// It loads the source name automatically.
func (service *Service) NewSourceController(controllerType configuration.Type) error {
	var source controller.Interface
	if controllerType == configuration.ReplierType {
		sourceController, err := controller.NewReplier(service.controller.logger)
		if err != nil {
			fmt.Errorf("failed to create a source as controller.NewReplier: %w", err)
		}
		source = sourceController
	} else if controllerType == configuration.PusherType {
		sourceController, err := controller.NewPull(service.controller.logger)
		if err != nil {
			fmt.Errorf("failed to create a source as controller.NewPull: %w", err)
		}
		source = sourceController
	} else {
		return fmt.Errorf("the '%s' controller type not supported", controllerType)
	}

	err := service.AddSourceController(SourceName, source)
	if err != nil {
		return fmt.Errorf("failed to add source controller: %w", err)
	}

	return nil
}

// SetRequestHandler sets the handler for all incoming requestMessages
func (service *Service) SetRequestHandler(handler RequestHandler) {
	service.controller.SetRequestHandler(handler)
}

func (service *Service) SetReplyHandler(handler ReplyHandler) {
	service.controller.SetReplyHandler(handler)
}

// AddSourceController sets the source controller, and invokes the source controller's
func (service *Service) AddSourceController(name string, source controller.Interface) error {
	controllerConf, err := service.configuration.GetController(name)
	if err != nil {
		return fmt.Errorf("the '%s' controller configuration wasn't found: %v", name, err)
	}
	source.AddConfig(controllerConf)
	service.sources.Set(name, source)

	return nil
}

// Run the independent service.
func (service *Service) Run() {
	var wg sync.WaitGroup

	proxyExtension := extension()

	// Run the sources
	for _, c := range service.configuration.Controllers {
		if err := service.sources.Exist(c.Name); err != nil {
			fmt.Println("the source is not included", c, "error", err)
			continue
		}
		controllerList := service.sources.Map()
		var c, ok = controllerList[c.Name].(controller.Interface)
		if !ok {
			fmt.Println("interface -> key-value", c)
			continue
		}

		// add the extensions required by the source controller
		requiredExtensions := c.RequiredExtensions()
		for _, name := range requiredExtensions {
			extension, err := service.configuration.GetExtension(name)
			if err != nil {
				log.Fatal("extension required by the controller doesn't exist in the configuration", "error", err)
			}

			c.AddExtensionConfig(extension)
		}

		// The proxy adds itself as the extension to the sources
		c.RequireExtension(proxyExtension.Name)
		c.AddExtensionConfig(proxyExtension)

		wg.Add(1)
		go func() {
			err := c.Run()
			wg.Done()
			if err != nil {
				log.Fatal("failed to run the controller", "error", err)
			}
		}()
	}

	// Run the proxy controller. Service controller itself on the other hand
	// will run the destination clients
	wg.Add(1)
	go func() {
		service.controller.Run()
		wg.Done()
	}()

	println("waiting for the wait group")
	wg.Wait()
}