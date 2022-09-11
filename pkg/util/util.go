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
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/alsritter/middlebaby/pkg/util/common"
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

func ReadStreamFile(fileName string) ([]byte, error) {
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error reading file, filename: [%s], error: [%v]", fileName, err)
	}
	return buf, nil
}

func ReadMultiFile(fileListStr string) ([]byte, error) {
	field, fileList, err := getFieldAndFileList(fileListStr)
	if err != nil {
		return nil, err
	}
	bodyBuf := &bytes.Buffer{}
	bodyWrite := multipart.NewWriter(bodyBuf)
	defer bodyWrite.Close()
	for _, fileInfoStr := range fileList {
		fileInfo := strings.Split(fileInfoStr, ":")
		if len(fileInfo) != 2 {
			return nil, fmt.Errorf("the file format is incorrect, want: [fileName:filePath], got:[%s]", fileInfo)
		}
		fileWrite, err := bodyWrite.CreateFormFile(field, fileInfo[0])
		if err != nil {
			return nil, fmt.Errorf("create FormFile error, field:[%s], fileName:[%s], error:[%v]", field, fileInfo[0], err)
		}
		fileData, err := ioutil.ReadFile(fileInfo[1])
		if err != nil {
			return nil, fmt.Errorf("read file error, fileName: [%s], err: [%v]", fileInfo[1], err)
		}
		_, err = fileWrite.Write(fileData)
		if err != nil {
			return nil, fmt.Errorf("write file to response error, fileName: [%s], err: [%v]", fileInfo[1], err)
		}
	}
	return bodyBuf.Bytes(), nil
}

func getFieldAndFileList(fileListStr string) (string, []string, error) {
	fileList := strings.Split(fileListStr, ";")
	if len(fileList) <= 1 || !strings.HasPrefix(fileList[0], common.FileFieldPrefix) {
		return "", nil, fmt.Errorf("the file name format is incorrect, "+
			"want: [field:fieldName;file01:file01Path;......fileN:fileNPath;], got: [%s]", fileListStr)
	}

	if fileList[len(fileList)-1] == "" {
		fileList = fileList[:len(fileList)-1]
	}

	field := strings.ReplaceAll(fileList[0], common.FileFieldPrefix, "")
	return field, fileList[1:], nil
}
