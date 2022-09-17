/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package javascript

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/types/mbcase"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	jsoniter "github.com/json-iterator/go"
	"rogchap.com/v8go"
)

type jsAssertPlugin struct {
	logger.Logger
}

func New(log logger.Logger) pluginregistry.AssertPlugin {
	return &jsAssertPlugin{Logger: log.NewLogger("js-assert")}
}

// Name implements pluginregistry.AssertPlugin
func (*jsAssertPlugin) Name() string {
	return "jsAssertPlugin"
}

// Assert CommonAssert e.g.
// "assert.data.activityList.length==3",
// "assert.data.activityList[0].activityBase.activityId==1"
func (j *jsAssertPlugin) Assert(resp *mbcase.Response, asserts []mbcase.CommonAssert) error {
	ctx := context.Background()
	vm := v8go.NewContext()

	// try converting to JSON
	if canToJson, actualInterface := j.toJsonInterface(resp.Data); canToJson {
		resp.Data = actualInterface
	}

	respRaw, err := jsoniter.MarshalToString(resp)
	j.Trace(nil, "target response to json: [%s]", respRaw)

	if err != nil {
		return err
	}

	// init js params.
	_, err = j.runScript(ctx, vm, fmt.Sprintf("const assert = {data: %s}", respRaw))
	if err != nil {
		return err
	}

	for _, a := range asserts {
		value, err := j.runScript(ctx, vm, a.Actual)
		if err != nil {
			return err
		}

		str, err := value.MarshalJSON()
		if err != nil {
			return fmt.Errorf("js assert result marshal error: [%v]", err)
		}

		if err := assert.So(j, "javascript assert", str, a.Expected); err != nil {
			return err
		}
	}

	return nil
}

// GetTypeName implements pluginregistry.AssertPlugin
func (*jsAssertPlugin) GetTypeName() string {
	return "js"
}

func (*jsAssertPlugin) toJsonInterface(ifc interface{}) (bool, interface{}) {
	if sb, ok := ifc.([]byte); ok {
		var i interface{}
		if err := json.Unmarshal(sb, &i); err != nil {
			return false, nil
		}
		return true, i
	}

	if str, ok := ifc.(string); ok {
		var maybeJson interface{}
		if err := json.Unmarshal([]byte(str), &maybeJson); err != nil {
			return false, nil
		}
		return true, maybeJson
	}

	return false, nil
}

// RunScript is used to run javascript with context
func (j *jsAssertPlugin) runScript(ctx context.Context, v8Context *v8go.Context, script string) (*v8go.Value, error) {
	valCh := make(chan *v8go.Value, 1)
	errCh := make(chan error, 1)
	go func() {
		val, err := v8Context.RunScript(script, "main.js")
		if err != nil {
			errCh <- err
			return
		}
		valCh <- val
	}()
	var terminateFunc = func() error {
		vm := v8Context.Isolate()
		vm.TerminateExecution()
		return <-errCh
	}
	select {
	case val := <-valCh:
		return val, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, terminateFunc()
	}
}
