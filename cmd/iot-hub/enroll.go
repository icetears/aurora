package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math/big"
	"time"
)

//type RootCA struct {
//	Certificate tls.Certificate
//	PrivateKey  crypto.PrivateKey
//	NextProtoID string
//}

func getRootCA(s string) (tls.Certificate, error) {

	cd := []byte(`-----BEGIN CERTIFICATE-----
MIICHzCCAcWgAwIBAgIUSmEbP0cURQQrbr2Xksfzr7kEiacwCgYIKoZIzj0EAwIw
ZTELMAkGA1UEBhMCQ04xCzAJBgNVBAgMAkJKMQswCQYDVQQHDAJCSjEPMA0GA1UE
CgwGQXVyb3JhMREwDwYDVQQLDAhBdXJvcmFDQTEYMBYGA1UEAwwPY2EuaWNldGVh
cnMuY29tMB4XDTIwMDgxNzE4MDY0M1oXDTQ1MTEwNjE4MDY0M1owZTELMAkGA1UE
BhMCQ04xCzAJBgNVBAgMAkJKMQswCQYDVQQHDAJCSjEPMA0GA1UECgwGQXVyb3Jh
MREwDwYDVQQLDAhBdXJvcmFDQTEYMBYGA1UEAwwPY2EuaWNldGVhcnMuY29tMFkw
EwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEP6VfVtDi74CH/LJQzdPcv8kHKFYH8TVW
EZB7FpCjdQt1S5zEOIzASffTZZqkt9uYe9H2ye4LGaWlGo1vQGlwF6NTMFEwHQYD
VR0OBBYEFI7F6lzHd6W9pGnhB8Jx9ZB/7cH1MB8GA1UdIwQYMBaAFI7F6lzHd6W9
pGnhB8Jx9ZB/7cH1MA8GA1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0EAwIDSAAwRQIh
AJyrtq9YyKkYz3xC51YMEnUbpSLQpDtIqSQ719UAnEwJAiBLbd6y20MRiVjx+vmp
rFX/tGNqXMoEAEfu1VTd8TO3PQ==
-----END CERTIFICATE-----`)
	kd := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIOI1UAPx98BG/hlcjh81TYmcUUK2NgamjcMn8k4p3jOfoAoGCCqGSM49
AwEHoUQDQgAEP6VfVtDi74CH/LJQzdPcv8kHKFYH8TVWEZB7FpCjdQt1S5zEOIzA
SffTZZqkt9uYe9H2ye4LGaWlGo1vQGlwFw==
-----END EC PRIVATE KEY-----`)

	ca, err := tls.X509KeyPair(cd, kd)
	log.Println(err)
	//var rca RootCA

	return ca, err
}

func DeviceEnrollHandler(c *gin.Context) {
	csrBytes, err := ioutil.ReadAll(c.Request.Body)
	csrPEM, _ := pem.Decode(csrBytes)
	if csrPEM == nil {
		log.Fatal("pemcheck", err)
	}
	csr, err := x509.ParseCertificateRequest(csrPEM.Bytes)
	if err != nil {
		log.Fatal("pemcheck", err)
	}

	csr.CheckSignature()
	if err != nil {
		log.Fatal("signature: ", err)
	}

	RootCA, _ := getRootCA(csr.Subject.SerialNumber)
	log.Println(RootCA.Certificate)
	certificate, err := x509.ParseCertificate(RootCA.Certificate[0])

	template := &x509.Certificate{
		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,
		SignatureAlgorithm: certificate.SignatureAlgorithm,
		//OCSPServer:            cfg.OCSPServer,
		//CRLDistributionPoints: cfg.OCSPServer,
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              csr.DNSNames,
		EmailAddresses:        csr.EmailAddresses,
		IPAddresses:           csr.IPAddresses,
	}
	template.Subject.CommonName, _ = DeviceMFT(csr.Subject.SerialNumber, csr.DNSNames[0])

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}
	log.Println(serialNumber)

	now := time.Now()
	template.SerialNumber = serialNumber
	template.NotBefore = now.UTC()
	template.NotAfter = now.Add(time.Hour * 24 * 30).UTC()
	//template.Subject.CommonName = ""

	cert, err := x509.CreateCertificate(rand.Reader, template, certificate, csr.PublicKey, RootCA.PrivateKey)
	block := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}
	certPem := pem.EncodeToMemory(&block)
	c.Writer.Header().Set("Content-Type", "application/x-x509-ca-ra-cert")
	c.Writer.Write(certPem)

}
