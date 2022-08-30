package pluginregistry

// Provide environmental support at runtime. (mysql, redis, ....)
type EnvPlugin interface {
	Plugin
	// GetTypeName the plugin type
	GetTypeName() string
	// Run setup run
	Run(commands []string) error
}
