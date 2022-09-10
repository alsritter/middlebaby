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
	"io/ioutil"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v2"
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
