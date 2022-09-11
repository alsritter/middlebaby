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

package cert

import (
	"crypto/ecdsa"
	"testing"
)

func TestCertificate_GenerateCA(t *testing.T) {
	type fields struct {
		cache             Cache
		defaultPrivateKey *ecdsa.PrivateKey
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "生成根证书",
			fields:  fields{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Certificate{
				cache:             tt.fields.cache,
				defaultPrivateKey: tt.fields.defaultPrivateKey,
			}
			got, err := c.GenerateCA()

			t.Log(string(got.CertBytes))

			t.Log("===========================")

			t.Log(string(got.PrivateKeyBytes))

			if (err != nil) != tt.wantErr {
				t.Errorf("Certificate.GenerateCA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
