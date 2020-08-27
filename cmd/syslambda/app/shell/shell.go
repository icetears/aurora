package shell

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

func Shell(url *string, certFile *string, keyFile *string, rootcaFile *string) {
	config := tls.Config{InsecureSkipVerify: false}
	if *certFile != "" && *keyFile != "" {
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}
	if *rootcaFile != "" {
		serverCert, _ := ioutil.ReadFile(*rootcaFile)
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(serverCert)
		config.RootCAs = caCertPool
	}

	var dia = websocket.DefaultDialer
	dia.TLSClientConfig = &config
	conn, _, err := dia.Dial(*url, nil)
	if err != nil {
		log.Println(err)
	}

	c := exec.Command("bash", "-i")
	f, err := pty.Start(c)

	go pumpStdout(conn, f)
	pumpStdin(conn, f)
	log.Printf("end\n")
	return

}

func pumpStdin(tlsConn *websocket.Conn, w io.Writer) {
	var (
		message = make([]byte, 400)
		err     error
		n       int
	)
	for {
		if err := tlsConn.SetReadDeadline(time.Now().Add(time.Minute * 15)); err != nil {
			return
		}
		n, message, err = tlsConn.ReadMessage()
		if err != nil {
			log.Println("msg read: ", err)
			return
		}
		if err := tlsConn.SetReadDeadline(time.Time{}); err != nil {
			return
		}
		if _, err := w.Write([]byte(message[:n])); err != nil {
			log.Println("bash write: ", err)
			return
		}
	}

}

func pumpStdout(tlsConn *websocket.Conn, r io.Reader) {
	//io.WriteString(tlsConn, fmt.Sprintf("%s\r\n", "start shell"))
	var (
		message = make([]byte, 1)
		err     error
		n       int
	)
	for {
		if err = tlsConn.SetReadDeadline(time.Time{}); err != nil {
			os.Exit(1)
		}
		if n, err = r.Read(message); err != nil {
			log.Println("bash read: ", err)
			os.Exit(1)
		}
		if err = tlsConn.SetReadDeadline(time.Now().Add(time.Minute * 15)); err != nil {
			os.Exit(1)
		}
		err = tlsConn.WriteMessage(websocket.TextMessage, message[:n])
		if err != nil {
			log.Println("msg write: ", err)
			os.Exit(1)
		}
	}

}
