package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/icetears/aurora/pkg/requests"
	"github.com/icetears/aurora/pkg/ssl"
	"github.com/icetears/aurora/pkg/sys"
	"io/ioutil"
	"net"
	"os"
	"time"
)

type Node struct {
	ID       string
	HostName string
	SN       string
	IPAddr   []net.IP
}

var node Node

func GetDeviceCertificate() (tls.Certificate, error) {
	var (
		certPath = BaseDir + "/conf/certs/client.crt"
		keyPath  = BaseDir + "/conf/private_keys/client.key"
		cert     tls.Certificate
	)
	certPemBlock, err := ioutil.ReadFile(certPath)
	if err != nil {
		cert, err = enroll(certPath, keyPath)
		if err != nil {
			return tls.Certificate{}, err
		}
		return cert, err
	}
	keyPemBlock, err := ioutil.ReadFile(keyPath)
	if err != nil {
		cert, err = enroll(certPath, keyPath)
		if err != nil {
			return tls.Certificate{}, err
		}
		return cert, err
	}
	b, _ := pem.Decode(certPemBlock)
	c, err := x509.ParseCertificate(b.Bytes)
	timeNow := time.Now()
	if expires := int64(c.NotAfter.Sub(timeNow).Hours()) / 24; expires < 1 {
		cert, err = enroll(certPath, keyPath)
		if err != nil {
			return tls.Certificate{}, err
		}
	}
	node.ID = c.Subject.CommonName
	cert, err = tls.X509KeyPair(certPemBlock, keyPemBlock)
	return cert, err
}

func enroll(c string, k string) (tls.Certificate, error) {
	fmt.Println("Enroll")
	NodeInit()

	p, err := ssl.GenerateKey("P256")
	if err != nil {
		os.Exit(0)
	}
	csr := ssl.GenerateCsr(p.PrivateKey, node.SN, "aurora edge certificate")
	csr.Template.DNSNames = append(csr.Template.DNSNames, node.HostName)
	csr.Template.IPAddresses = node.IPAddr
	fmt.Println(csr.Template.IPAddresses)
	d, err := csr.GenCsrPem()
	if err != nil {
		return tls.Certificate{}, err
	}

	endpoint := "https://www.icetears.com/v1/certificates/enroll"
	r := requests.New()
	r.Post(endpoint, d)

	body := r.Response.Body
	err = ioutil.WriteFile(c, body, 0600)
	if err != nil {
		return tls.Certificate{}, err
	}

	pemCert, _ := pem.Decode(body)
	x509Cert, err := x509.ParseCertificate(pemCert.Bytes)
	node.ID = x509Cert.Subject.CommonName

	err = ioutil.WriteFile(k, p.EncodeToPem(), 0600)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := tls.X509KeyPair(body, p.EncodeToPem())
	if err != nil {
		return tls.Certificate{}, err
	}
	return cert, nil
}

func NodeInit() error {
	eth, err := sys.ProbeNetworkInterface()
	if err != nil {
		return err
	}
	node.HostName, err = os.Hostname()
	if err != nil {
		return err
	}
	node.SN = eth.HWAddr
	node.IPAddr = eth.IPAddr
	return nil
}
