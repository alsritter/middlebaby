package proxy

import (
	"sync"

	"alsritter.icu/middlebaby/internal/file/common"
)

type MockCenter interface {
}

// mockCenter handler all mock
type mockCenter struct {
	httpMock map[string][]common.HttpImposter
	gRpcMock map[string][]common.GRpcImposter
	sync.Mutex
}
