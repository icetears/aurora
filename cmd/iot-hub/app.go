package main

import (
	"crypto/tls"
	"crypto/x509"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var db *gorm.DB
var mc MQTT.Client

func main() {
	var err error
	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("x-iot-hub")
	opts.SetCleanSession(true)
	mc = MQTT.NewClient(opts)
	if token := mc.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	r := gin.New()
	r.POST("v1/devices", DeviceCreateHandler)
	r.GET("v1/devices", DeviceListHandler)
	r.DELETE("v1/devices/:id", DeviceDeleteHandler)
	r.GET("v1/devices/:id/shadow", DeviceShadowHandler)
	r.POST("v1/devices/:id/shadow", DeviceShadowUpdateHandler)
	r.Any("v1/devices/:id/console", DeviceConsoleHandler)

	r.DELETE("v1/devices/:id/app", DeviceDeleteHandler)
	r.POST("v1/devices/:id/lambda", DeviceLambdaDeployHandler)
	r.POST("v1/app/create", AppDeployHandler)

	r.POST("v1/certificates/enroll", DeviceEnrollHandler)

	http.Handle("/", r)

	db, err = gorm.Open("postgres", "host=localhost user=postgres dbname=iot_hub sslmode=disable password=password")
	if err != nil {
		panic(err)
	}
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(512)
	db.DB().SetConnMaxLifetime(100 * time.Second)
	db.AutoMigrate(&Device{})
	//r.Run("localhost:5000")

	caCert, err := ioutil.ReadFile("certs/trust.pem")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.VerifyClientCertIfGiven,
		//GetConfigForClient: getConfigForClient(tls.ClientHelloInfo.SignatureSchemes),
		//Time:       func() time.Time { return time.Unix(int64(time.Now().Second()+5000), 0) },
	}
	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		Addr:      "0.0.0.0:5000",
		TLSConfig: tlsConfig,
	}
	go hub.Init()

	log.Fatal(server.ListenAndServeTLS("certs/server.pem", "certs/server.key"))

}
