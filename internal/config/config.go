package config

type Mock struct {
	Port int32
}

var GlobalConfigVar GlobalConfig

type GlobalConfig struct {
	HttpFiles []string `json:"httpFiles"` // mock 的 http 请求数据所在文件名
}
