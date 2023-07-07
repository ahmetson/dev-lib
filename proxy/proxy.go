// Package proxy defines the script that acts as the middleware
package proxy

// For proxy, there is no controllers.
// But only two kind of functions.
// The proxy enables the request and reply handlers.
//
//
/*Package independent is used to scaffold the independent service
 */

import (
	"fmt"
	"github.com/Seascape-Foundation/sds-common-lib/data_type/key_value"
	"github.com/Seascape-Foundation/sds-service-lib/configuration"
	"github.com/Seascape-Foundation/sds-service-lib/controller"
	"github.com/Seascape-Foundation/sds-service-lib/log"
	"sync"
)

type Proxy struct {
	configuration configuration.Service
	controllers   key_value.KeyValue
	controller    *Controller
}

const sourceName = "source"
const destinationName = "destination"

func validateConfiguration(service configuration.Service) error {
	if len(service.Controllers) < 2 {
		return fmt.Errorf("not enough controllers were given. atleast 'source' and 'destination' should be")
	}

	sourceFound := false
	destinationFound := false
	for _, c := range service.Controllers {
		if c.Name == sourceName {
			sourceFound = true
		} else if c.Name == destinationName {
			destinationFound = true
		}
	}

	if !sourceFound {
		return fmt.Errorf("proxy service '%s' in seascape.yml doesn't have '%s' controller", service.Name, sourceName)
	}

	if !destinationFound {
		return fmt.Errorf("proxy service '%s' in seascape.yml doesn't have '%s' controller", service.Name, destinationName)
	}

	return nil
}

// registerNonSources registers the controller instances as the destination.
// it skips the sourceName named controllers as the destination.
func registerNonSources(controllers []configuration.Controller, proxyController *Controller) error {
	for _, c := range controllers {
		if c.Name == sourceName {
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

// New Proxy service based on the configurations
func New(serviceConf configuration.Service, logger log.Logger) (*Proxy, error) {
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

	service := Proxy{
		configuration: serviceConf,
		controllers:   key_value.Empty(),
		controller:    proxyController,
	}

	return &service, nil
}

// SetRequestHandler sets the handler for all incoming requests
func (service *Proxy) SetRequestHandler(handler HandleFunc) {
	service.controller.SetRequestHandler(handler)
}

func (service *Proxy) AddController(name string, controller *controller.Controller) error {
	controllerConf, err := service.configuration.GetController(name)
	if err != nil {
		return fmt.Errorf("the '%s' controller configuration wasn't found: %v", name, err)
	}
	controller.AddConfig(controllerConf)
	service.controllers.Set(name, controller)

	return nil
}

// Run the independent service.
func (service *Proxy) Run() {
	var wg sync.WaitGroup

	for _, c := range service.configuration.Controllers {
		if err := service.controllers.Exist(c.Name); err != nil {
			fmt.Println("the config doesn't exist", c, "error", err)
			continue
		}
		controllerList := service.controllers.Map()
		var c, ok = controllerList[c.Name].(*controller.Controller)
		if !ok {
			fmt.Println("interface -> key-value", c)
			continue
		}

		// add the extensions required by the controller
		requiredExtensions := c.RequiredExtensions()
		for _, name := range requiredExtensions {
			extension, err := service.configuration.GetExtension(name)
			if err != nil {
				log.Fatal("extension required by the controller doesn't exist in the configuration", "error", err)
			}

			c.AddExtensionConfig(extension)
		}

		wg.Add(1)
		go func() {
			err := c.Run()
			wg.Done()
			if err != nil {
				log.Fatal("failed to run the controller", "error", err)
			}
		}()
	}
	println("waiting for the wait group")
	wg.Wait()
}
