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
	"testing"

	"github.com/alsritter/middlebaby/pkg/types/mbcase"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

/*
e.g.
[{
	"header": {
		"Date": "Sun, 11 Sep 2022 01:42:38 GMT",
		"Content-Length": "42",
		"Content-Type": "text/plain; charset=utf-8"
	},
	"data": "{\"name\":\"John\",\"color\":\"Purples\",\"age\":55}",
	"statusCode": 200
}]
*/

func Test_jsAssertPlugin_Assert(t *testing.T) {
	type fields struct {
		Logger logger.Logger
	}
	type args struct {
		resp    *mbcase.Response
		asserts []mbcase.CommonAssert
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "测试 js 断言",
			fields: fields{
				Logger: logger.NewDefault("test"),
			},
			args: args{
				resp: &mbcase.Response{
					Header: map[string]string{
						"Date":           "Sun, 11 Sep 2022 01:42:38 GMT",
						"Content-Length": "42",
						"Content-Type":   "text/plain; charset=utf-8",
					},
					Data:       "{\"name\":\"John\",\"color\":\"Purples\",\"age\":55}",
					StatusCode: 200,
				},
				asserts: []mbcase.CommonAssert{
					{
						TypeName: "js",
						Actual:   "assert.data.statusCode == 200",
						Expected: true,
					},
					{
						TypeName: "js",
						Actual:   "assert.data.data.color == 'Purples'",
						Expected: true,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsAssertPlugin{
				Logger: tt.fields.Logger,
			}
			if err := j.Assert(tt.args.resp, tt.args.asserts); (err != nil) != tt.wantErr {
				t.Errorf("jsAssertPlugin.Assert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
