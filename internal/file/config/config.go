package config

// Config representation of config file yaml
type Config struct {
	HttpFiles      []string   `yaml:"httpFiles"`      // http mock file.
	CaseFiles      []string   `yaml:"caseFiles"`      // task paths
	Port           int        `yaml:"port"`           // proxy port
	CORS           ConfigCORS `yaml:"cors"`           // CORS
	Storage        Storage    `yaml:"storage"`        // mock server needs
	Watcher        bool       `yaml:"watcher"`        // whether to enable file listening
	EnableDirect   bool       `yaml:"enableDirect"`   // whether the missed mock allows real requests
	TaskFileSuffix string     `yaml:"taskFileSuffix"` // the default test case suffix name. example: ".case.json"
}

// ConfigCORS representation of section CORS of the yaml
type ConfigCORS struct {
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

type Storage struct {
	Mysql Mysql `yaml:"mysql"`
	Redis Redis `yaml:"redis"`
}

type Mysql struct {
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Local    string `yaml:"local"`
	Charset  string `yaml:"charset"`
}

type Redis struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
	Auth string `yaml:"auth"`
	DB   int    `yaml:"db"`
}
