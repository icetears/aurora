package main

import (
	"../../proto"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var BaseDir string

func lambdaExec(app string) {
	x, e := exec.Command("bash", "-c", "./build/syslambda -app mqtt 9999").Output()
	logrus.Info(string(x), e)
}

//func containerExec(app ...string) {
//	container.Create(app...)
//}

func deviceMSG(client MQTT.Client, msg MQTT.Message) {
	var p pb.ThingMSG
	err := proto.Unmarshal(msg.Payload(), &p)
	if err != nil {
		return
	}
	logrus.Info(p.MsgType, p.MsgId, p.Env, p.Args)
	switch p.MsgType {
	case pb.ThingMSG_DOCKER:
		logrus.Info(p.MsgType)
		t := &DockerInputTemplate{
			Name:  p.Cid,
			Image: p.Func,
			Env:   []string{"config=yes", "dd=cc"},
		}
		dockerExec(t)
	case pb.ThingMSG_SYSLAMBDA:
		logrus.Info(p.MsgType)
		go SysLambda(p)
	case pb.ThingMSG_LAMBDA:
		logrus.Info(p.MsgType)
		go LambdaPy(p)
	}
	//container.Create(string(msg.Payload()))
}

func main() {
	logrus.Info("aurora init")
	//container.Init()
	//go lambdaExec("syslambda")
	var err error
	BaseDir, err = filepath.Abs(filepath.Dir(os.Args[0]) + "/..")
	if err != nil {
		logrus.Fatal(err)
	}

	ConfInit()

	cert, err := GetDeviceCertificate()
	if err != nil {
		logrus.Fatal(err)
	}

	tlsconfig := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: false}
	serverCert, _ := ioutil.ReadFile(Conf.TrustCA)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(serverCert)
	tlsconfig.RootCAs = caCertPool

	opts := MQTT.NewClientOptions().AddBroker(Conf.Mqtt)
	opts.SetClientID(node.ID)
	opts.SetCleanSession(false)
	opts.SetTLSConfig(&tlsconfig)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(onConnHandler)
	opts.SetPingTimeout(time.Second * 60)
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	defer c.Disconnect(250)
	select {}
}

func onConnHandler(c MQTT.Client) {
	logrus.Info("subscribe topic: ", node.ID)
	if token := c.Subscribe(fmt.Sprintf("system/devices/%s", node.ID), 2, deviceMSG); token.Wait() && token.Error() != nil {
		logrus.Error(token.Error())
		os.Exit(1)
	}
}
