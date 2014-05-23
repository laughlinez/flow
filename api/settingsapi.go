package api

//Provide a basic settings API for flow gadget providers
type ISettingsAPI interface {
	//Generic initialisation stub (per interface) - allows some 'context' for provider.
	InitAPI(...interface{})
	//Retrieve all setting keys for gadget instance
	Keys(prefix string) ([]string, error)
	//get a setting for 'key'
	Get(key string) (interface{}, error)
	//store a setting for 'key'
	Put(key string, value interface{}) error
}
