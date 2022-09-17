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

package interact

import (
	"strings"
	"testing"
)

func TestResponse_GetByteData(t *testing.T) {
	type fields struct {
		Status  int
		Header  map[string][]string
		Body    interface{}
		Trailer map[string][]string
		Delay   *ResponseDelay
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "单个文件响应",
			fields: fields{
				Status:  0,
				Header:  map[string][]string{},
				Body:    "@file:./testdata/test01.txt",
				Trailer: map[string][]string{},
				Delay:   &ResponseDelay{},
			},
			want:    []string{"🐺 here is test 01 file~"},
			wantErr: false,
		},
		{
			name: "多个文件响应",
			fields: fields{
				Status:  0,
				Header:  map[string][]string{},
				Body:    "@multiFile:field:fieldName;file01.txt:./testdata/test01.txt;file02.txt:./testdata/test02.txt;",
				Trailer: map[string][]string{},
				Delay:   &ResponseDelay{},
			},
			want:    []string{"🐷 here is test 02", "🐺 here is test 01"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				Status:  tt.fields.Status,
				Header:  tt.fields.Header,
				Body:    tt.fields.Body,
				Trailer: tt.fields.Trailer,
				Delay:   tt.fields.Delay,
			}
			got, err := r.GetByteData()
			if (err != nil) != tt.wantErr {
				t.Errorf("Response.GetByteData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, v := range tt.want {
				if !strings.Contains(string(got), v) {
					t.Errorf("Response.GetByteData() = %v, want exists %v", got, v)
				}
			}
		})
	}
}
