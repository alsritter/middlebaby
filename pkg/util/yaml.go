package util

import (
	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// LoadConfig read YAML-formatted config from filename into cfg.
func LoadConfig(filename string, pointer interface{}) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return multierror.Prefix(err, "Error reading config file")
	}

	err = yaml.UnmarshalStrict(buf, pointer)
	if err != nil {
		return multierror.Prefix(err, "Error parsing config file")
	}

	return nil
}
