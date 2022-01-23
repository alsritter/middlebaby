package example

import (
	"alsritter.icu/middlebaby/pkg/plugin"
)

type PluginA struct{}

func (p *PluginA) Exec(chan<- string) error {
	return nil
}

func init() {
	plugin.Registry["plugin_a"] = &PluginA{}
}
