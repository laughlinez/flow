//Package api adds a simple interface to extending flow gadgets using a basic Dependency Injection mechanism
package api

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"reflect"
	"strings"
)

//Holds the basic Interface API's that flow can provide - (this should never grow too big in reality)
//naming convention: I<apibasename>API
//
//A Gadget can attempt to 'Provide' services to the API by using struct tag 'flowapi' such as:
// type SomeProviderGadget struct {
//     Gadget
//     In flow.Input
//     Out flow.Output
//     AThing api.IThingAPI `flowapi:"ThingAPI"`
// }
//
// Where the parameter to flowapi is the api name provided through the FlowAPI struct
//   additionally, providers can include comma seperated list of 'modifiers' such as
//   `flowapi:"SettingsAPI,new"`
//   where 'new' specifies that the Consumer should instead get a NEW instance of object implementing the SettingsAPI
//
//A Gadget can attempt to 'Consume' services provided by the API using struct tag 'gadget' such as:
// type SomeUserGadget struct {
//     Gadget
//     In flow.Input
//     Out flow.Output
//     Settings api.ISettingsAPI `gadget:"SettingsAPI"`
// }
//
// Where the parameter to gadget is the api name provided through the FlowAPI struct
// In the case above, the provider specified 'new' so the consumer will be provided its own 'instance' of
// the SettingAPI object, initialised specifically for the caller gadget
//
// Note: If 'new' is not specified by the provider, all gadgets will get the *same* instance of the provider.
//
// Initialization:
// The framework looks for the method InitAPI(...interface{}) on each of the API interfaces it provides.
// If this is found, it is called with the following parameters:
// gadget-name string, gadget-path string
// This is espacially useful for providers that require use of 'new' as it can be used to distinguish each gadget.
//

//The current API's that flow api will arbitrate
type FlowAPI struct {
	SettingsAPI ISettingsAPI
	//DBReadAPI IDBReadAPI
	//DBWriteAPI IDBWriteAPI
	DBReadWriteAPI IDBReadWriteAPI
	//FilesystemAPI IFileSystemAPI
}

//allows us to reflect over the 'current' api.
var api *FlowAPI
var dict map[string]dictEntry

type dictEntry struct {
	reflect.Value
	Props []string
}

func init() {
	api = new(FlowAPI)
	dict = make(map[string]dictEntry)
}

type FlowAPIOptions struct {
	ErrorOnProviderOffering   bool //if a provider offers a service we can't take, we produce error
	ErrorOnProviderAssignment bool //if a provider assignment fails do we produce error

	ErrorOnConsumerRequest    bool //if a provider offers a service we can't take, we produce error
	ErrorOnConsumerAssignment bool //if a provider assignment fails do we produce error
}

//utility shortcut for suggested option defaults
func NewFlowAPIOptions() FlowAPIOptions {
	opt := FlowAPIOptions{}
	opt.ErrorOnConsumerRequest = true
	opt.ErrorOnConsumerAssignment = true
	return opt
}

//inject any api services the gadget needs (services are provided by gadget 'Providers')
func InjectAPI(c interface{}, opts FlowAPIOptions) error {

	//These are of course all *Gadgets
	inst := reflect.ValueOf(c)

	name := inst.MethodByName("Name").Call(nil)[0]
	path := inst.MethodByName("Path").Call(nil)[0]

	if glog.V(2) {
		glog.Infoln("Injector called for %s %s %s\n", inst, path, name)
	}

	//see if the gadget [requests] any services from 'flow/api'
	//TODO: Implement a runtime 'cache' to speed up subsequent gadget lookups
	st := inst.Elem().Type()
	for i := 0; i < st.NumField(); i++ {

		field := st.Field(i)

		tparams := field.Tag.Get("gadget") //tag we look at for 'consumer' gadgets
		props := strings.Split(tparams, ",")
		apiname := props[0]

		if apiname != "" {
			if glog.V(2) {
				glog.Infoln("Gadget requests: %s\n", apiname)
			}

			apival := reflect.Indirect(reflect.ValueOf(api))

			f := apival.FieldByName(apiname)

			if !f.IsValid() { //we dont provide this api
				if opts.ErrorOnConsumerRequest {
					return errors.New(fmt.Sprintf("FlowAPI does not provide %s", apiname))
				}
				continue
			}

			if !f.Type().AssignableTo(field.Type) { //the client cannot accept the api it requests
				if opts.ErrorOnConsumerAssignment {
					return errors.New(fmt.Sprintf("FlowAPI cannot provide this service - Gadget API incorrect for %s", apiname))
				}
				continue
			}

			vfield := inst.Elem().Field(i)

			src, ok := dict[apiname]
			if !ok { //no provider that has been seen, can provide this api
				if opts.ErrorOnConsumerRequest {
					return errors.New(fmt.Sprintf("FlowAPI missing provider %s", apiname))
				}
				continue
			}

			trg := reflect.Value{}

			//TODO: Error checking
			if contains(src.Props, "new") {
				base := src.Elem()
				trg = reflect.New(base.Type())
			} else {
				trg = f
			}

			vfield.Set(trg)

			//Attempt to invoke the InitAPI(..interface{}) if its present
			//we pass name, path (of the gadget within the circuit)
			//this will help provide some 'scope' of the gadget using the API to the provider
			m := vfield.MethodByName("InitAPI")
			if m.IsValid() {
				in := []reflect.Value{name, path}
				m.Call(in)
			}

		}

	}

	return nil
}

//determine if the Gadget is a 'Provider' and wants to offer services to the flow API
func IsAPIProvider(c interface{}, opts FlowAPIOptions) error {

	inst := reflect.ValueOf(c)

	name := inst.MethodByName("Name").Call(nil)[0]
	path := inst.MethodByName("Path").Call(nil)[0]

	if glog.V(2) {
		glog.Infoln("Provider called for %s %s %s\n", inst, path, name)
	}

	//TODO: Implement a runtime 'cache' to speed up subsequent gadget lookups
	//see if the gadget [provides] any services to 'flow/api'
	st := inst.Elem().Type()
	for i := 0; i < st.NumField(); i++ {

		field := st.Field(i)

		tparams := field.Tag.Get("flowapi") //tag we look at for 'consumer' gadgets
		props := strings.Split(tparams, ",")
		apiname := props[0]

		if apiname != "" {
			if glog.V(2) {
				glog.Infoln("Gadget provides: %s\n", apiname)
			}
			apival := reflect.Indirect(reflect.ValueOf(api))

			f := apival.FieldByName(apiname)

			if !f.IsValid() {
				if opts.ErrorOnProviderOffering {
					return errors.New(fmt.Sprintf("FlowAPI does not accept %s", apiname))
				}
				continue
			}

			vfield := inst.Elem().Field(i)

			if !vfield.Type().AssignableTo(f.Type()) {
				if opts.ErrorOnProviderAssignment {
					return errors.New(fmt.Sprintf("FlowAPI cannot accept this service - Gadget API incorrect for %s", apiname))
				}
				continue
			}

			//Important to infer real 'type' AND track modifiers
			de := dictEntry{vfield, props[1:]}
			dict[apiname] = de

			f.Set(vfield) //store this to the API (infers its been provided and matches API)

			if glog.V(0) {
				glog.Infoln("Provider installed for: %s via %s%s \n", apiname, path, name)
			}

		}

	}

	return nil
}

func contains(list []string, s string) bool {
	for _, a := range list {
		if a == s {
			return true
		}
	}
	return false
}
