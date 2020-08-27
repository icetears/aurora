package mqtt

import (
	"crypto/tls"
	"fmt"
	"github.com/icetears/aurora/cmd/syslambda/app/mqtt/pkg/message"
	"github.com/sirupsen/logrus"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type Server struct {
	listener    net.Listener
	tlsListener net.Listener
}

var log = logrus.New()

func (this *Server) ListenAndServe(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		log.Info(err)
		return err
	}
	log.Info("MQTT v3.1 message broker: ", u.Scheme, "://", u.Host)
	switch {
	case u.Scheme == "tcp":
		this.listener, err = net.Listen(u.Scheme, u.Host)
		if err != nil {
			log.Error(err)
			return err
		}
		defer this.listener.Close()
		for {
			conn, err := this.listener.Accept()
			if err != nil {
				log.Info(err)
				//if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(5 * time.Second)
				conn.Close()
				continue
				//}
				return err
			}
			go this.NewConnection(conn)
			fmt.Println("end")
		}
	case u.Scheme == "tls":
		xRoot, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(xRoot)
		cer, err := tls.LoadX509KeyPair("/home/xiao/backup/edge.crt", "/home/xiao/backup/edge.key")
		if err != nil {
			fmt.Println("load certificate error")
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
		this.tlsListener, err = tls.Listen("tcp", u.Host, tlsConfig)
		defer this.tlsListener.Close()
		for {
			conn, err := this.tlsListener.Accept()
			if err != nil {
				conn.Close()
				//if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(5 * time.Second)
				continue
				//}
				return err
			}
			go this.NewConnection(conn)
		}
	}
	return nil
}

func StartIOLoop(c net.Conn) {
	for {
		//msgType, err := message.GetType(c)
		//if err != nil {
		//	log.Info(err)
		//	break
		//}
		var hdr message.Header
		err := hdr.Decode(c)
		if err != nil {
			ses.Delete(c)
			c.Close()
			log.Debug(c.RemoteAddr(), "lost")
			break
		}
		newMessage(hdr, c)
	}
}

func newMessage(hdr message.Header, c net.Conn) {
	switch hdr.MsgType {
	case message.PUBLISH:
		msg := new(message.Publish)
		resp := new(message.PubAck)
		msg.Decode(c, hdr)
		resp.MID = msg.MID
		log.Info(msg.Topic, msg.MID, string(msg.Payloads))
		router.Publish(msg.Payloads, msg.Topic)
		if hdr.Qos > 0 {
			resp.Header.Qos = hdr.Qos
			log.Info(resp.Qos)
			resp.Encode(c)
		}
		//msg = new(message.Publish)
		//msg.Payloads = "mymessage"
		//msg.Topic = "mytopic/test"
		//msg.Qos = 2
		//msg.Encode(c)
		return
	case message.PUBACK:
		log.Info("publish ack packet")
		b := make([]byte, 1024)
		c.Read(b)
	//	msg = new(message.Publish)
	case message.PUBREC:
		log.Info("publish rec packet")
		b := make([]byte, 1024)
		c.Read(b)
	//	msg = new(message.Publish)
	case message.PUBCOMP:
		log.Info("publish comp packet")
	//	msg = new(message.Publish)
	case message.SUBSCRIBE:
		msg := new(message.Subscribe)
		resp := new(message.SubAck)
		msg.Decode(c, hdr)
		router.Add(msg.Topic, msg.Qos, c)
		log.Info(msg.Topic, msg.Qos)
		resp.Encode(c)
	//case message.SUBACK:
	//	log.Info("subscribe ack packet")
	//	msg = new(message.Publish)
	//case message.UNSUBSCRIBE:
	//	log.Info("unsubscribe packet")
	//	msg = new(message.Publish)
	//case message.UNSUBACK:
	//	log.Info("unsubscribe ack packet")
	//	msg = new(message.Publish)
	case message.PINGREQ:
		//msg := new(message.PingReq)
		resp := new(message.PingResp)
		//msg.Decode(c, hdr)
		resp.Encode(c)
	case message.DISCONNECT:
		log.Info("disconnect packet")
		ses.Delete(c)
	//	msg = new(message.Publish)
	default:
		log.Info("invalid packet", hdr.MsgType)
		return
	}
}

func (this *Server) NewConnection(c net.Conn) error {
	log.Info("New client From: ", c.RemoteAddr().String(), message.CONNECT)
	if ses.Clients == nil {
		ses.Clients = make(map[string]Client)
	}
	if router.Topic == nil {
		router.Topic = make(map[string]map[string]Subscribe)
	}

	msg := new(message.Connect)
	msg.Header.Decode(c)
	msg.Decode(c)

	log.Info("Protocol: ", msg.ProtocolName, "  Level: ", msg.ProtocolLevel)
	log.Info("name flag: ", msg.UsernameFlag)
	log.Info("pass flag: ", msg.PasswordFlag)
	log.Info("will retain flag: ", msg.WillRetain)
	log.Info("will qos level: ", msg.WillQos)
	log.Info("will flag: ", msg.WillFlag)
	log.Info("cleansess flag: ", msg.CleanSession)
	log.Info("keepalive time: ", msg.KeepAliveTimer)
	log.Info("client ID: ", msg.ClientID)
	log.Info("username: ", msg.Username)
	log.Info("password: ", msg.Password)
	if true {
		ses.Add(msg.ClientID, c)
	}
	fmt.Printf("ss %v \n", ses.Clients)
	resp := new(message.ConnAck)
	resp.RetCode = 0x00
	resp.Encode(c)
	go StartIOLoop(c)

	return nil
}

func main() {
	l := os.Getenv("loglevel")
	if l == "" {
		log.SetLevel(logrus.Level(4))
	} else {
		log.SetLevel(logrus.Level(2))
	}
	cfg := Server{}
	cfg.ListenAndServe("tcp://0.0.0.0:1883")
}
