package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/term/termios"
	"io"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"
)

func Console(id string) {
	//cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	//if err != nil {
	//	log.Fatalf("server: loadkeys: %s", err)
	//}
	fmt.Println("==============================================")
	fmt.Println("=    Web Shell,  send ctrl + \\\\\\ to exit     =")
	fmt.Println("==============================================")
	fmt.Println()
	config := tls.Config{Certificates: nil, InsecureSkipVerify: true, ServerName: "www.icetears.com"}
	serverCert, _ := ioutil.ReadFile("certs/server.pem")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(serverCert)
	config.RootCAs = caCertPool

	var dia = websocket.DefaultDialer
	dia.TLSClientConfig = &config
	conn, _, err := dia.Dial(fmt.Sprintf("wss://www.icetears.com/v1/devices/%s/console", id), nil)
	if err != nil {
		log.Println("console connect:", err)
	}

	go pumpStdout(conn, os.Stdout)
	pumpStdin(conn, os.Stdin)
	log.Printf("console terminate\n")
	return

}

func pumpStdout(conn *websocket.Conn, w io.Writer) {
	var (
		message = make([]byte, 4)
		err     error
		n       int
	)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(time.Minute * 15)); err != nil {
			return
		}
		n, message, err = conn.ReadMessage()
		if err != nil {
			log.Println("msg read: ", err)
			return
		}
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			return
		}
		if _, err := w.Write([]byte(message[:n])); err != nil {
			log.Println("bash write: ", err)
			return
		}
		time.Sleep(time.Nanosecond)
	}

}

func pumpStdin(conn *websocket.Conn, r io.Reader) {
	var (
		err error
		n   int
		buf = make([]byte, 1)
		nl  syscall.Termios
		t   = 0
		i   = 2
	)
	fd, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		log.Println(err)
		return
	}
	termios.Tcgetattr(uintptr(fd), &nl)
	if err != nil {
		log.Println(err)
		return
	}

	nl.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK |
		syscall.ISTRIP | syscall.INLCR | syscall.IGNCR |
		syscall.ICRNL | syscall.IXON
	nl.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG | syscall.ECHONL
	nl.Cflag &^= syscall.CSIZE | syscall.PARENB
	nl.Cc[syscall.VMIN] = 1
	nl.Cc[syscall.VTIME] = 0
	termios.Tcsetattr(uintptr(fd), termios.TCSANOW, (*syscall.Termios)(&nl))
	for {
		n, err = syscall.Read(fd, buf)
		if err != nil {
			log.Println("tty read:", err)
		}
		time.Sleep(time.Nanosecond)

		if err = conn.SetReadDeadline(time.Now().Add(time.Minute * 15)); err != nil {
			return
		}
		if buf[0] == 28 {
			if time.Now().Second()-t < 3 {
				t = time.Now().Second()
				i--
			} else {
				i = 2
				t = time.Now().Second()
			}
			if i == 0 {
				conn.Close()
				break
			}
		}
		err = conn.WriteMessage(websocket.TextMessage, buf[:n])
		if err != nil {
			log.Println("msg write: ", err)
			return
		}
	}

}
