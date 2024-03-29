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

package assert

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

// So Verify that the input value is consistent with the expected value
func So(log logger.Logger, assertType string, actual interface{}, expected interface{}) error {
	return NewAssert(log, assertType, actual, expected).assert()
}

var (
	ErrorTypeNotEqual    = errors.New("type inconsistency")
	ErrorMapKeyInvalided = errors.New("key is invalid")
	ErrorUnKnownType     = errors.New("unknown type comparison")
	ErrorLengthNotEqual  = errors.New("data length is inconsistent")
)

// cases assert error types
type AssertError struct {
	Type      string
	Err       error
	FieldName string
	Actual    interface{}
	Expected  interface{}
}

func (e *AssertError) Error() string {
	bf := bytes.Buffer{}
	bf.WriteString("[" + e.Type + "]")
	if e.Err != nil {
		bf.WriteString(" error: " + e.Err.Error())
	}

	bf.WriteString(fmt.Sprintf(`
		expected return value: [%v] 
		actual return value: [%v]
	`, e.Expected, e.Actual))
	if e.FieldName != "" {
		bf.WriteString(fmt.Sprintf(" wrong field: [%s]", e.FieldName))
	}
	return bf.String()
}

func NewAssertError(assertType string, err error, actual interface{}, expected interface{}, fieldName string) *AssertError {
	return &AssertError{Type: assertType, Err: err, Actual: actual, Expected: expected, FieldName: fieldName}
}

type Assert struct {
	assertType string
	actual     interface{}
	expected   interface{}

	log logger.Logger
}

func NewAssert(log logger.Logger, assertType string, actual interface{}, expected interface{}) *Assert {
	return &Assert{assertType: assertType, actual: actual, expected: expected, log: log}
}

// an entry function for an assertion
func (a *Assert) assert() error {
	a.before()
	return a.so(a.assertType, a.actual, a.expected)
}

// convert the expected and actual values to JSON
func (a *Assert) before() {
	if canToJson, actualInterface := a.toJsonInterface(a.actual); canToJson {
		a.actual = actualInterface
	}
	if canToJson, expectedInterface := a.toJsonInterface(a.expected); canToJson {
		a.expected = expectedInterface
	}
}

func (*Assert) toJsonInterface(ifc interface{}) (bool, interface{}) {
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

// Link comparison field name
func (*Assert) appendFieldName(parentFieldName string, afterFieldName string) string {
	if parentFieldName == "" {
		return afterFieldName
	}
	return parentFieldName + "." + afterFieldName
}

func (a *Assert) getFloat64Value(v reflect.Value) (bool, float64) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true, float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true, float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return true, v.Float()
	}
	return false, 0
}

// split all the values of a field and compare the values set in the expected value, focusing on whether the expected value corresponds to the actual return value
func (a *Assert) so(fieldName string, actual interface{}, expected interface{}) error {
	// when the corresponding Expected is nil, it is considered that the actual data does not need to be judged, and it is directly considered to be matched
	if expected == nil {
		return nil
	}

	actualRv := reflect.ValueOf(actual)
	expectedRv := reflect.ValueOf(expected)

	retErrFun := func(err error) error {
		return NewAssertError(a.assertType, err, actual, expected, fieldName)
	}

	if IsRegExpPattern(expected) {
		a.log.Debug(nil, "starts matching the regular expression, pattern: %s, actual: %v \n", expected.(string), actual)
		if err := Match(expected.(string), actual); err != nil {
			a.log.Debug(nil, "error matching regular expression, pattern: %s, actual: %v, err: %v \n", expected.(string), actual, err)
			return retErrFun(err)
		}
		return nil
	}

	ok1, actualFloat := a.getFloat64Value(actualRv)
	ok2, expectedFloat := a.getFloat64Value(expectedRv)
	if ok1 && ok2 {
		if actualFloat != expectedFloat {
			return retErrFun(nil)
		}
		return nil
	}

	if actualRv.Kind() != expectedRv.Kind() {
		return retErrFun(fmt.Errorf("%v %s != %s", ErrorTypeNotEqual, actualRv.Kind().String(), expectedRv.Kind().String()))
	}

	switch actualRv.Kind() {
	case reflect.String:
		if expectedRv.String() != actualRv.String() {
			return retErrFun(nil)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if expectedRv.Int() != actualRv.Int() {
			return retErrFun(nil)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if expectedRv.Uint() != actualRv.Uint() {
			return retErrFun(nil)
		}
	case reflect.Float32, reflect.Float64:
		if expectedRv.Float() != actualRv.Float() {
			return retErrFun(nil)
		}
	case reflect.Bool:
		if expectedRv.Bool() != actualRv.Bool() {
			return retErrFun(nil)
		}
	// In the case of structures, you need to recurse to the last level of base elements and compare them.
	// The following arrays and maps are also recursive.
	case reflect.Struct:
		numField := expectedRv.Type().NumField()
		for i := 0; i < numField; i++ {
			// get all field.
			name := expectedRv.Type().Field(i).Name
			// recursion
			if err := a.so(a.appendFieldName(fieldName, name), actualRv.FieldByName(name).Interface(), expectedRv.FieldByName(name).Interface()); err != nil {
				return err
			}
		}
	case reflect.Map:
		expectKeys := expectedRv.MapKeys()
		for _, keyVal := range expectKeys {
			// most functions and methods never return an invalid Value.
			if !actualRv.MapIndex(keyVal).IsValid() {
				fieldName = fmt.Sprintf("%v", keyVal.Interface())
				return retErrFun(ErrorMapKeyInvalided)
			}
			if err := a.so(a.appendFieldName(fieldName, keyVal.String()), actualRv.MapIndex(keyVal).Interface(), expectedRv.MapIndex(keyVal).Interface()); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		actualLen := actualRv.Len()
		expectedLen := expectedRv.Len()
		if actualLen != expectedLen {
			return retErrFun(fmt.Errorf("%v %d != %d", ErrorLengthNotEqual, actualLen, expectedLen))
		}
		for i := 0; i < actualLen; i++ {
			if err := a.so(a.appendFieldName(fieldName, fmt.Sprintf("[%d]", i)), actualRv.Index(i).Interface(), expectedRv.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return retErrFun(fmt.Errorf("%v %s != %s", ErrorUnKnownType, actualRv.Type().String(), expectedRv.Type().String()))
	}
	return nil
}
