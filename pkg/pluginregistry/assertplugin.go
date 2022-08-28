package pluginregistry

import "github.com/alsritter/middlebaby/pkg/caseprovider"

type AssertPlugin interface {
	Plugin
	Assert(caseprovider.CommonAssert) error
}
