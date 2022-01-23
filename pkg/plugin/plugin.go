package plugin

var Registry map[string]Plugin

type Plugin interface {
	Exec(chan<- string) error
}
