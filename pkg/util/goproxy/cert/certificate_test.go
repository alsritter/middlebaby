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
