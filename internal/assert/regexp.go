package assert

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const regExpPrefix = "@regExp:"

func IsRegExpPattern(pattern interface{}) bool {
	if pattern == nil {
		return false
	}
	if reflect.TypeOf(pattern).Kind() != reflect.String {
		return false
	}
	return strings.HasPrefix(pattern.(string), regExpPrefix)
}

func removeRegExpPrefix(str string) string {
	return str[8:]
}

func Match(pattern string, matchValue interface{}) error {
	pattern = removeRegExpPrefix(pattern)
	s, err := toString(matchValue)
	if err != nil {
		return err
	}
	b, err := regexp.MatchString(pattern, s)
	if err != nil {
		return err
	}
	if !b {
		return fmt.Errorf("the regular expression does not match, regular: %s, matchValue :%v", pattern, s)
	}
	return nil
}

func toString(value interface{}) (string, error) {
	valueRv := reflect.ValueOf(value)
	switch valueRv.Kind() {
	case reflect.String:
		return valueRv.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return formatBaseInt(valueRv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return formatBaseUint(valueRv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return formatFloat(valueRv.Float()), nil
	case reflect.Bool:
		return strconv.FormatBool(valueRv.Bool()), nil
	default:
		return "", fmt.Errorf("this type %v conversion is not currently supported", valueRv.Type().String())
	}
}

func formatBaseInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func formatBaseUint(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func formatFloat(v float64) string {
	if v == float64(int64(v)) {
		return formatBaseInt(int64(v))
	} else if v == float64(uint64(v)) {
		return formatBaseUint(uint64(v))
	}
	// Float To String.
	// reference: https://www.includehelp.com/golang/strconv-formatfloat-function-with-examples.aspx
	return strconv.FormatFloat(v, 'f', 2, 64)
}
