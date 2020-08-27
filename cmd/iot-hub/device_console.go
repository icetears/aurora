package main

import (
	"../../proto"
	"github.com/gin-gonic/gin"
	//"github.com/jinzhu/gorm"
	//"github.com/google/uuid"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func DeviceConsoleHandler(c *gin.Context) {
	//var dev Device
	//dev.ID = c.Param("id")
	//db.Save(&dev)
	var id string
	conn, err := websocket.Upgrade(c.Writer, c.Request, c.Writer.Header(), 1024, 1024)
	if err != nil {
		c.Status(http.StatusBadRequest)
	}

	if len(c.Request.TLS.PeerCertificates) > 0 {
		for _, v := range c.Request.TLS.PeerCertificates {
			//P_KEY, _ := x509.MarshalPKIXPublicKey(v.PublicKey)
			//log.Print(P_KEY)
			id = v.Subject.CommonName
			log.Println("sssl ", id)
			go ClientHandler(conn, id)
			return
		}

	} else {
		id = c.Param("id")
		fmt.Println("normal ", id)
	}
	fmt.Println("browser", id)

	hub.addClient <- id
	fmt.Println("browser2")

	hello := "******************************************\r\n******   Welcome to \x1B[1;33mN\x1B[1;34mX\x1B[1;36mP\x1B[1;0m Cloud Lab   ******\r\n******************************************\r\n"
	err = conn.WriteMessage(websocket.TextMessage, []byte(hello))
	if err != nil {
		log.Println(err)
	}

	go pumpShellStdout(conn, id)
	if err == nil {
		p := pb.ThingMSG{
			MsgId:   111,
			Cid:     "test",
			MsgType: pb.ThingMSG_SYSLAMBDA,
			Func:    "shell",
		}
		d, _ := proto.Marshal(&p)
		SendMSG(fmt.Sprintf("system/devices/%s", id), d, 2)
	}

	pumpShellStdin(conn, id)
}

func pumpShellStdin(conn *websocket.Conn, id string) {
	for {
		if hub.clientsIn[id] != nil {
			break
		}
	}

	for {
		select {
		case msgIn, ok := <-hub.clientsIn[id]:
			//out:=make([]byte,base64.StdEncoding.EncodedLen(len(msgIn)))
			//base64.StdEncoding.Encode(out,msgIn)
			//err := conn.WriteMessage(websocket.TextMessage, out)
			if ok {
				err := conn.WriteMessage(websocket.TextMessage, msgIn)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func pumpShellStdout(conn *websocket.Conn, id string) {
	var (
		message = make([]byte, 1)
		err     error
		n       int
	)
	for {
		n, message, err = conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading json.", err)
			hub.removeClient <- id
			return
		}
		hub.clientsOut[id] <- message[:n]
	}

}

//func tmp() {
//	go hub.Init()
//	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
//	if err != nil {
//		log.Fatalf("server: loadkeys: %s", err)
//	}
//	config := tls.Config{Certificates: []tls.Certificate{cert}}
//	config.Rand = rand.Reader
//	//config.VerifyPeerCertificate =
//
//	//const (
//	//        NoClientCert ClientAuthType = iota
//	//        RequestClientCert
//	//        RequireAnyClientCert
//	//        VerifyClientCertIfGiven
//	//        RequireAndVerifyClientCert
//	//)
//	config.ClientAuth = 5
//
//	ClientCA, _ := ioutil.ReadFile("certs/trust.pem")
//	caCertPool := x509.NewCertPool()
//	caCertPool.AppendCertsFromPEM(ClientCA)
//	config.ClientCAs = caCertPool
//
//	service := "0.0.0.0:8000"
//	listener, err := tls.Listen("tcp", service, &config)
//	if err != nil {
//		log.Fatalf("server: listen: %s", err)
//	}
//	log.Print("server: listening")
//	var id string
//	for {
//		Conn, err := listener.Accept()
//		if err != nil {
//			log.Printf("server: accept: %s", err)
//			break
//		}
//		defer Conn.Close()
//		log.Printf("server: accepted from %s", Conn.RemoteAddr())
//		tlsConn, ok := Conn.(*tls.Conn)
//		tlsConn.Handshake()
//		if ok {
//			log.Print("ok=true")
//			state := tlsConn.ConnectionState()
//			for _, v := range state.PeerCertificates {
//				//P_KEY, _ := x509.MarshalPKIXPublicKey(v.PublicKey)
//				//log.Print(P_KEY)
//				id = v.Subject.CommonName
//				log.Println(id)
//			}
//		}
//		go ClientHandler(Conn, id)
//	}
//}

func ClientHandler(Conn *websocket.Conn, id string) {
	defer Conn.Close()
	defer log.Println("client : closed")
	log.Printf("client %s handler", id)

	c := make(chan bool)

	go func() {
		var (
			message = make([]byte, 1)
			err     error
			n       int
		)

		for {
			n, message, err = Conn.ReadMessage()
			if err != nil {
				log.Printf("server: Conn: read: %s", err)
				c <- true
				return
			}
			hub.clientsIn[id] <- message[:n]
			hub.clientsIn[id] <- nil
		}
	}()

	if hub.clientsOut[id] != nil {
		for {
			select {
			case msgOut, ok := <-hub.clientsOut[id]:
				if ok {
					err := Conn.WriteMessage(websocket.TextMessage, msgOut)
					if err != nil {
						log.Println("server conn", err)
						Conn.Close()
						return
					}
				} else {
					log.Println("server conn", ok)
					Conn.Close()
					return
				}
			case <-c:
				return
			}
		}
	}
}
