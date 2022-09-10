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
