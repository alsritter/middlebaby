package pluginregistry

import "github.com/alsritter/middlebaby/pkg/caseprovider"

type AssertPlugin interface {
	Plugin
	// GetTypeName the plugin type
	GetTypeName() string
	Assert([]caseprovider.CommonAssert) error
}
