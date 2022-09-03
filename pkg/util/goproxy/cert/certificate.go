// Package cert 证书管理
package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"

	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

var (
	defaultRootCAPem = []byte(`-----BEGIN CERTIFICATE-----
MIICMzCCAdqgAwIBAgIITWWCIQf8/VIwCgYIKoZIzj0EAwIwWzEOMAwGA1UEBhMF
Q2hpbmExEjAQBgNVBAgTCUd1YW5nZG9uZzERMA8GA1UEBxMIU2hlbnpoZW4xEzAR
BgNVBAoTCm1pZGRsZWJhYnkxDTALBgNVBAMTBE1hcnMwHhcNMjIwODAzMTYwOTM3
WhcNNDIwOTAzMTYwOTM3WjBbMQ4wDAYDVQQGEwVDaGluYTESMBAGA1UECBMJR3Vh
bmdkb25nMREwDwYDVQQHEwhTaGVuemhlbjETMBEGA1UEChMKbWlkZGxlYmFieTEN
MAsGA1UEAxMETWFyczBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABFVoa0AEathz
me8b2JJ1/miuMjogK2yKx4nPutoSwFJgkyTHzrRASV/1w1akT7y2qq/7mIB4xV6Q
EQjVTm0YHWCjgYcwgYQwDgYDVR0PAQH/BAQDAgEGMB0GA1UdJQQWMBQGCCsGAQUF
BwMCBggrBgEFBQcDATASBgNVHRMBAf8ECDAGAQH/AgECMB0GA1UdDgQWBBR2vEX9
wf6DKsfgOlEF5NEpXOSZ4zAgBgNVHREEGTAXgRVhbHNyaXR0ZXJAb3V0bG9vay5j
b20wCgYIKoZIzj0EAwIDRwAwRAIgToo/zb4vWYW7m/9/zjqETBN0VqkaohHoFHmF
VPj8ltkCIDfD4yflWny/XoZkEtY7MzSOwRU337xeiv6yRQKnBKN7
-----END CERTIFICATE-----
`)
	defaultRootKeyPem = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIOMrMDPgWm4EdydF3RWkr70uiH5OKJ3+xsOOpw5jX4lGoAoGCCqGSM49
AwEHoUQDQgAEVWhrQARq2HOZ7xvYknX+aK4yOiArbIrHic+62hLAUmCTJMfOtEBJ
X/XDVqRPvLaqr/uYgHjFXpARCNVObRgdYA==
-----END EC PRIVATE KEY-----
`)
)

var (
	defaultRootCA  *x509.Certificate
	defaultRootKey *ecdsa.PrivateKey
)

func init() {
	var err error
	block, _ := pem.Decode(defaultRootCAPem)
	defaultRootCA, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("加载根证书失败: %s", err))
	}
	block, _ = pem.Decode(defaultRootKeyPem)
	defaultRootKey, err = x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("加载根证书私钥失败: %s", err))
	}
}

// Certificate 证书管理
type Certificate struct {
	cache             Cache
	defaultPrivateKey *ecdsa.PrivateKey
}

type Pair struct {
	Cert            *x509.Certificate
	CertBytes       []byte
	PrivateKey      *ecdsa.PrivateKey
	PrivateKeyBytes []byte
}

func NewCertificate(cache Cache, useDefaultPrivateKey ...bool) *Certificate {
	c := &Certificate{
		cache: cache,
	}
	if len(useDefaultPrivateKey) > 0 && useDefaultPrivateKey[0] {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		if err != nil {
			panic(err)
		}
		c.defaultPrivateKey = priv
	}

	return c
}

// GenerateTlsConfig 生成TLS配置
func (c *Certificate) GenerateTlsConfig(host string) (*tls.Config, error) {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if c.cache != nil {
		fields := strings.Split(host, ".")
		sudDomains := []string{host, strings.Join(fields[1:], ".")}
		for _, item := range sudDomains {
			// 先从缓存中查找证书
			if cert := c.cache.Get(item); cert != nil {
				tlsConf := &tls.Config{
					Certificates: []tls.Certificate{*cert},
				}

				return tlsConf, nil
			}
		}
	}
	pair, err := c.GeneratePem(host, 1, defaultRootCA, defaultRootKey)
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(pair.CertBytes, pair.PrivateKeyBytes)
	if err != nil {
		return nil, err
	}
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if c.cache != nil {
		// 缓存证书
		if len(pair.Cert.IPAddresses) > 0 {
			c.cache.Set(host, &cert)
		}
		for _, item := range pair.Cert.DNSNames {
			item = strings.TrimPrefix(item, "*.")
			c.cache.Set(item, &cert)
		}

	}

	return tlsConf, nil
}

// GeneratePem 生成证书
func (c *Certificate) GeneratePem(host string, expireDays int, rootCA *x509.Certificate, rootKey *ecdsa.PrivateKey) (*Pair, error) {
	var priv *ecdsa.PrivateKey
	var err error

	if c.defaultPrivateKey != nil {
		priv = c.defaultPrivateKey
	} else {
		priv, err = ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	}
	if err != nil {
		return nil, err
	}
	tmpl := c.template(host, expireDays)
	derBytes, err := x509.CreateCertificate(crand.Reader, tmpl, rootCA, &priv.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	serverCert := pem.EncodeToMemory(certBlock)
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	keyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	}
	serverKey := pem.EncodeToMemory(keyBlock)

	p := &Pair{
		Cert:            tmpl,
		CertBytes:       serverCert,
		PrivateKey:      priv,
		PrivateKeyBytes: serverKey,
	}

	return p, nil
}

// GenerateCA 生成根证书
func (c *Certificate) GenerateCA() (*Pair, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			CommonName:   "Mars",
			Country:      []string{"China"},
			Organization: []string{"middlebaby"},
			Province:     []string{"Guangdong"},
			Locality:     []string{"Shenzhen"},
		},
		NotBefore:             time.Now().AddDate(0, -1, 0),
		NotAfter:              time.Now().AddDate(20, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		EmailAddresses:        []string{"alsritter@outlook.com"},
	}

	derBytes, err := x509.CreateCertificate(crand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	serverCert := pem.EncodeToMemory(certBlock)

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	keyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	}
	serverKey := pem.EncodeToMemory(keyBlock)

	p := &Pair{
		Cert:            tmpl,
		CertBytes:       serverCert,
		PrivateKey:      priv,
		PrivateKeyBytes: serverKey,
	}

	return p, nil
}

func (c *Certificate) template(host string, expireYears int) *x509.Certificate {
	fv := fnv.New32a()
	_, _ = fv.Write([]byte(host))

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(int64(fv.Sum32())),
		Subject: pkix.Name{
			CommonName:   host,
			Country:      []string{"China"},
			Organization: []string{"middlebaby"},
			Province:     []string{"Guangdong"},
			Locality:     []string{"Shenzhen"},
		},
		NotBefore:             time.Now().AddDate(-1, 0, 0),
		NotAfter:              time.Now().AddDate(expireYears, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment,
		EmailAddresses:        []string{"alsritter@outlook.com"},
	}
	hosts := strings.Split(host, ",")
	for _, item := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
			continue
		}

		fields := strings.Split(item, ".")
		fieldNum := len(fields)
		for i := 0; i <= (fieldNum - 2); i++ {
			cert.DNSNames = append(cert.DNSNames, "*."+strings.Join(fields[i:], "."))
		}
		if fieldNum == 2 {
			cert.DNSNames = append(cert.DNSNames, item)
		}
	}

	return cert
}

// DefaultRootCAPem 根证书
func DefaultRootCAPem() []byte {
	return defaultRootCAPem
}
