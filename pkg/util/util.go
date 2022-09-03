package util

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/pflag"
)

// ValidatableConfig defines the validatable Config
type ValidatableConfig interface {
	// Validate is used to validate config and returns error on failure
	Validate() error
}

// ValidateConfigs is used to validate validatable configs
func ValidateConfigs(configs ...ValidatableConfig) error {
	for _, config := range configs {
		if config == nil {
			return fmt.Errorf("config(%T) is nil", config)
		}
		if err := config.Validate(); err != nil {
			return fmt.Errorf("%T: %s", config, err)
		}
	}
	return nil
}

// RegistrableConfig defines the registrable config
type RegistrableConfig interface {
	// RegisterFlagsWithPrefix is used to registerer flag with prefix
	RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet)
}

func ToHttpHeader(headers map[string]interface{}) (httpHeader http.Header) {
	httpHeader = make(http.Header)
	for k, v := range headers {
		switch vv := v.(type) {
		case string:
			httpHeader.Add(k, vv)
		case []string:
			for _, vvv := range vv {
				httpHeader.Add(k, vvv)
			}
		}
	}
	return
}

func InterfaceMapToStringMap(m map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			out[k] = vv
		case []string:
			var b strings.Builder
			for _, vvv := range vv {
				b.WriteString(vvv + ";")
			}
			out[k] = b.String()
		}
	}
	return out
}

func SliceMapToStringMap(m map[string][]string) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		var b strings.Builder
		for _, vv := range v {
			b.WriteString(vv + ";")
		}
		out[k] = b.String()
	}
	return out
}
