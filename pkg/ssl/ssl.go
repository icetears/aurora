package ssl

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

type PKey struct {
	PrivateKey interface{}
}

//GenerateKey  EC: P256
//RSA: NULL
func GenerateKey(t string) (*PKey, error) {
	var (
		priv PKey
		err  error
	)
	switch t {
	case "P256":
		priv.PrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
	}
	return &priv, nil
}

func (p *PKey) EncodeToPem() []byte {
	bytes, _ := x509.MarshalECPrivateKey(p.PrivateKey.(*ecdsa.PrivateKey))
	block := pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: bytes,
	}
	return pem.EncodeToMemory(&block)
}
