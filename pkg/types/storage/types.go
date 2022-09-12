package storage

type Mysql struct {
	Enabled  bool   `yaml:"enabled"`
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Local    string `yaml:"local"`
	Charset  string `yaml:"charset"`
}

type Redis struct {
	Enabled bool   `yaml:"enabled"`
	Port    string `yaml:"port"`
	Host    string `yaml:"host"`
	Auth    string `yaml:"auth"`
	DB      int    `yaml:"db"`
}
