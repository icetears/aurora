package ssl

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
)

type Csr struct {
	Template *x509.CertificateRequest
	PKey     crypto.PrivateKey
}

func GenerateCsr(priv crypto.PrivateKey, sn string, commonName string) Csr {
	var c Csr
	c.PKey = priv
	c.Template = &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   commonName,
			SerialNumber: sn,
		},
		//SignatureAlgorithm: x509.SHA256WithRSA,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
	}
	return c
}

func (c Csr) GenCsrPem() ([]byte, error) {
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, c.Template, c.PKey)
	if err != nil {
		return nil, err
	}
	block := pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	}
	csrPem := pem.EncodeToMemory(&block)
	return csrPem, nil
}
