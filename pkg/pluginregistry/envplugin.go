package pluginregistry

// Provide environmental support at runtime. (mysql, redis, ....)
type EnvPlugin interface {
	Plugin
	// Get the plugin type
	GetTypeName() string
	// Run setup run
	Run(commands []string) error
}
