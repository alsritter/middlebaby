package pluginregistry

type Registry interface {
	// MockPlugins is used to return all registered mock plugins
	MockPlugins() []MockPlugin
	// RegisterMockPlugins is used to register mock plugins
	RegisterMockPlugins(...MockPlugin) error

	// MockPlugins is used to return all registered match plugins
	MatchPlugins() []MatchPlugin
	// RegisterMatchPlugins is used to register match plugins
	RegisterMatchPlugins(...MatchPlugin) error
}
