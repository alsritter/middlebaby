package config

import "github.com/alsritter/middlebaby/pkg/storage"

// Config representation of config file yaml
type Config struct {
	HttpFiles      []string `yaml:"httpFiles"`      // http mock file.
	CaseFiles      []string `yaml:"caseFiles"`      // task paths
	Port           int      `yaml:"port"`           // proxy port
	Watcher        bool     `yaml:"watcher"`        // whether to enable file listening
	EnableDirect   bool     `yaml:"enableDirect"`   // whether the missed mock allows real requests
	TaskFileSuffix string   `yaml:"taskFileSuffix"` // the default test case suffix name. example: ".case.json"

	CORS    ConfigCORS      `yaml:"cors"`    // CORS
	Storage *storage.Config `yaml:"storage"` // mock server needs
}

// ConfigCORS representation of section CORS of the yaml
type ConfigCORS struct {
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}
