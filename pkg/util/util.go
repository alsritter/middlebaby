package util

import (
	"fmt"
	"net/http"

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
