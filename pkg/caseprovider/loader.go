package caseprovider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alsritter/middlebaby/pkg/types/interact"
	"github.com/alsritter/middlebaby/pkg/types/mbcase"
	"github.com/flynn/json5"
)

type CaseLoader interface {
	LoadGlobalMockCase(filePath string) ([]*interact.ImposterMockCase, error)
	LoadItf(filePath string) (*mbcase.ItfTask, error)
}

type BasicLoader struct{}

func (l *BasicLoader) LoadGlobalMockCase(filePath string) ([]*interact.ImposterMockCase, error) {
	if !filepath.IsAbs(filePath) {
		if fp, err := filepath.Abs(filePath); err != nil {
			return nil, fmt.Errorf("to absolute representation path err: %s", err)
		} else {
			filePath = fp
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("%v: error trying to read config file: %s", err, filePath)
	}

	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)

	var imposter []*interact.ImposterMockCase
	if err := json.Unmarshal(bytes, &imposter); err != nil {
		return nil, fmt.Errorf("%v: error while unmarshal configFile file %s", err, filePath)
	}

	return imposter, nil
}

func (l *BasicLoader) LoadItf(filePath string) (*mbcase.ItfTask, error) {
	fb, err := ioutil.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("read file: %s error: %v", filePath, err)
	}

	if err != nil {
		return nil, fmt.Errorf("gets the taskserver file %s service type error: [%v]", filePath, err)
	}

	var t mbcase.ItfTask
	if err := json5.Unmarshal(fb, &t); err != nil {
		return nil, fmt.Errorf("serialization %s file error: %v", filePath, err)
	}

	return &t, nil
}
