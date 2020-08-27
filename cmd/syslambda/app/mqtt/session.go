package mqtt

import (
	"github.com/google/uuid"
	"net"
)

type Client struct {
	C  net.Conn
	ID string
}

type Session struct {
	Clients map[string]Client
}

var ses Session

func (s *Session) Add(id string, c net.Conn) {
	hid := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(c.RemoteAddr().String())).String()
	s.Clients[hid] = Client{
		C:  c,
		ID: id,
	}
}

func (s *Session) Delete(c net.Conn) {
	hid := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(c.RemoteAddr().String())).String()
	delete(ses.Clients, hid)
}
