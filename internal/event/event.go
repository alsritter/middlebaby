package event

import "github.com/asaskevich/EventBus"

// event center
var Bus EventBus.Bus

const (
	CLOSE = "PROCESS_CLOSE"
)

func init() {
	Bus = EventBus.New()
}
