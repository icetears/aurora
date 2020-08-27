package message

import "io"

type Type uint8
type QosLevel uint8

type Message interface {
	Encode(w io.Writer) error
	Decode(r io.Reader, hdr Header) error
}

const (
	CONNECT = iota + 1
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

type Header struct {
	MsgType byte
	Dup     bool
	Retain  bool
	Qos     byte
	RemLen  *int
}

type Connect struct {
	Header
	ProtocolName               string
	ProtocolLevel              uint8
	WillRetain                 bool
	WillFlag                   bool
	CleanSession               bool
	WillQos                    QosLevel
	KeepAliveTimer             uint16
	ClientID                   string
	WillTopic, WillMessage     string
	UsernameFlag, PasswordFlag bool
	Username, Password         string
}

type ConnAck struct {
	RetCode uint8
}

type Subscribe struct {
	Header
	MID   uint16
	Topic string
}

type SubAck struct {
	Header
}

type Publish struct {
	Header
	MID      uint16
	Topic    string
	Payloads []byte
}

type PubAck struct {
	Header
	MID uint16
}

type PingReq struct {
	Header
}

type PingResp struct {
	Header
}
