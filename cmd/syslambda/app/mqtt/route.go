package mqtt

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/icetears/aurora/cmd/syslambda/app/mqtt/pkg/message"
	"net"
)

type Subscribe struct {
	Qos byte
}

type Route struct {
	Topic map[string]map[string]Subscribe
}

var router Route

func (r *Route) Add(t string, qos byte, c net.Conn) {
	hid := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(c.RemoteAddr().String())).String()
	if r.Topic[t] == nil {
		r.Topic[t] = make(map[string]Subscribe)
	}
	r.Topic[t][hid] = Subscribe{Qos: qos}
	log.Println(r.Topic)
}

func (r *Route) Publish(payload []byte, topic string) {
	fmt.Println("publish")
	msg := new(message.Publish)
	msg.Payloads = payload
	msg.Topic = topic
	for idx, s := range r.Topic[topic] {
		fmt.Println(idx)
		msg.Qos = s.Qos
		if ses.Clients[idx].C == nil {
			fmt.Println("connection lost")
			continue
		}
		msg.Encode(ses.Clients[idx].C)
	}
}

func (r *Route) Delete(t string, qos byte, c net.Conn) {
}
