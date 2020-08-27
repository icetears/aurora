package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

var Conf struct {
	Api     string `yaml:"api"`
	Mqtt    string `yaml:"mqtt"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
	TrustCA string `yaml:"trust_ca"`
}

func ConfInit() {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s/conf/config.yml", BaseDir))
	if err != nil {
		fmt.Println("error to open conf file", err)
		os.Exit(1)
	}
	yaml.Unmarshal(buf, &Conf)
	if !filepath.IsAbs(Conf.Cert) {
		Conf.Cert = fmt.Sprintf("%s/conf/certs/%s", BaseDir, Conf.Cert)
	}
	if !filepath.IsAbs(Conf.Key) {
		Conf.Key = fmt.Sprintf("%s/conf/certs/%s", BaseDir, Conf.Key)
	}
	if !filepath.IsAbs(Conf.TrustCA) {
		Conf.TrustCA = fmt.Sprintf("%s/conf/certs/%s", BaseDir, Conf.TrustCA)
	}
}
