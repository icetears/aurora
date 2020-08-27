package main

import (
	"flag"
	"github.com/icetears/aurora/cmd/syslambda/app/mqtt"
	"github.com/icetears/aurora/cmd/syslambda/app/shell"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type config struct {
	Level  *int
	App    *string
	URL    *string
	Cert   *string
	Key    *string
	RootCA *string
}

var cfg config

func initFlags() {
	cfg.Level = flag.Int("d", 0, "debug level")
	cfg.App = flag.String("app", "u", "syslambda")
	cfg.URL = flag.String("url", "", "syslambda")
	cfg.Cert = flag.String("cert", "", "syslambda")
	cfg.Key = flag.String("key", "", "syslambda")
	cfg.RootCA = flag.String("rootca", "", "syslambda")
	flag.Parse()
}

func main() {
	logrus.Info(filepath.Base(os.Args[0]))
	initFlags()
	switch *cfg.App {
	case "mqtt":
		logrus.Info("sdf")
		cfg := mqtt.Server{}
		go cfg.ListenAndServe("tcp://0.0.0.0:1883")
		cfg.ListenAndServe("tls://0.0.0.0:8884")
	case "shell":
		shell.Shell(cfg.URL, cfg.Cert, cfg.Key, cfg.RootCA)
		logrus.Info(cfg.Key)
		logrus.Info("shell")
	}
}
