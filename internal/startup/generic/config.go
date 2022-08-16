package generic

import (
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/mockserver"
	"github.com/alsritter/middlebaby/pkg/storage"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/taskserver"
)

type Config struct {
	ApiManager    *apimanager.Config
	TargetProcess *targetprocess.Config
	MockServer    *mockserver.Config
	Storage       *storage.Config `yaml:"storage"` // mock server needs
	TaskService   *taskserver.Config
}
