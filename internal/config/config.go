package config

type Mock struct {
	Port int32
}

var GlobalConfigVar Config

// Config representation of config file yaml
type Config struct {
	HttpFiles []string   `yaml:"httpFiles"` // http mock file.
	Name      string     `yaml:"name"`      // Name of the HTTP request data for the mock
	Port      int        `yaml:"port"`      // proxy port
	CORS      ConfigCORS `yaml:"cors"`
	Watcher   bool       `yaml:"watcher"`
}

// ConfigCORS representation of section CORS of the yaml
type ConfigCORS struct {
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

//TODO: support auto test case
type TestCase struct{}
