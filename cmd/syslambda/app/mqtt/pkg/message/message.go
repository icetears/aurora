package message

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
)

func GetString(r io.Reader, remlen *int) string {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		log.Error(err)
	}
	l := uint16(buf[0])<<8 | uint16(buf[1])
	buf = make([]byte, l)
	if _, err := r.Read(buf); err != nil {
		log.Error(err)
	}
	*remlen -= int(l) + 2
	return string(buf)
}

func GetU8(r io.Reader, remlen *int) uint8 {
	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		log.Error(err)
	}
	*remlen--
	return uint8(buf[0])
}

func GetU16(r io.Reader, remlen *int) uint16 {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		log.Error(err)
	}
	*remlen -= 2
	return uint16(buf[0])<<8 | uint16(buf[1])
}

func SetU8(buf *bytes.Buffer, v uint8) {
	buf.Write([]byte{v})
}

func SetU16(buf *bytes.Buffer, v uint16) {
	buf.Write([]byte{byte(v & 0xff00 >> 8), byte(v & 0x00ff)})
}

func SetString(buf *bytes.Buffer, v string) {
	len := uint16(len(v))
	SetU16(buf, len)
	log.Info("len: ", len)
	buf.WriteString(v)
}

func SetPayload(buf *bytes.Buffer, v string) {
	buf.WriteString(v)
}

func GetType(r io.Reader) (Type, error) {
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		return 0, errors.New("invalid type")
	}
	return Type(buf[0] & 0xF0 >> 4), nil
}

func (h *Header) Decode(r io.Reader) error {
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		return err
	}

	remlen := GetRemLength(r)

	*h = Header{
		Dup:     buf[0]&0x08 > 0,
		Qos:     buf[0] & 0x06 >> 1,
		Retain:  buf[0]&0x01 > 0,
		MsgType: buf[0] & 0xF0 >> 4,
		RemLen:  &remlen,
	}

	return nil
}

func (h *Header) Encode(w io.Writer) error {
	return nil
}

func GetRemLength(r io.Reader) int {
	var v int32
	buf := make([]byte, 1)
	var shift uint
	for shift < 21 {
		if _, err := io.ReadFull(r, buf); err != nil {
			log.Info(err)
		}

		v |= int32(buf[0]&0x7F) << shift
		if buf[0] < 0x80 {
			return int(v)
		}
		shift += 7
	}
	panic("malformed remaining length")
}

func (c *Connect) Decode(r net.Conn) error {
	//if t, _ := GetType(r); t != CONNECT {
	//	return nil
	//}

	//remlen := GetRemLength(r)
	c.ProtocolName = GetString(r, c.RemLen)
	c.ProtocolLevel = GetU8(r, c.RemLen)
	flags := GetU8(r, c.RemLen)
	c.UsernameFlag = flags&0x80 > 0
	c.PasswordFlag = flags&0x40 > 0
	c.WillRetain = flags&0x20 > 0
	c.WillQos = QosLevel(flags & 0x18 >> 3)
	c.WillFlag = flags&0x04 > 0
	c.CleanSession = flags&0x02 > 0

	c.KeepAliveTimer = GetU16(r, c.RemLen)
	c.ClientID = GetString(r, c.RemLen)

	if c.WillFlag {
		c.WillTopic = GetString(r, c.RemLen)
		c.WillMessage = GetString(r, c.RemLen)
	}
	if c.UsernameFlag {
		c.Username = GetString(r, c.RemLen)
	}
	if c.PasswordFlag {
		c.Password = GetString(r, c.RemLen)
	}

	return nil
}

func (c *ConnAck) Encode(w io.Writer) {
	var b bytes.Buffer
	SetU8(&b, uint8(CONNACK<<4))
	SetU8(&b, 0x02)
	SetU8(&b, 0x01)
	SetU8(&b, c.RetCode)
	w.Write(b.Bytes())
}

func (m *Subscribe) Decode(r io.Reader, h Header) error {
	//buf := make([]byte, 1)
	//_, err := r.Read(buf)
	//if err != nil {
	//	return errors.New("bad pecket")
	//}
	//remlen := int(buf[0])
	m.MID = GetU16(r, h.RemLen)
	for *h.RemLen > 0 {
		m.Topic = GetString(r, h.RemLen)
		m.Qos = GetU8(r, h.RemLen)
	}
	return nil
}

func (m *Subscribe) Encode(w io.Writer) error {
	return nil
}

func (m *SubAck) Encode(w io.Writer) error {
	var b bytes.Buffer
	SetU8(&b, uint8(SUBACK)<<4)
	SetU8(&b, 0x2)
	SetU8(&b, 0x0)
	SetU8(&b, 0x1)
	w.Write(b.Bytes())
	return nil
}

func (m *SubAck) Decode(r io.Reader, h Header) error {
	return nil
}

func (m *Publish) Encode(w io.Writer) error {
	var (
		b      bytes.Buffer
		remlen uint8
	)

	SetU8(&b, uint8(PUBLISH)<<4|m.Qos<<1)

	if m.Qos > 0 {
		remlen = uint8(len(m.Payloads) + 4 + len(m.Topic))
		SetU8(&b, remlen)
		SetString(&b, m.Topic)
		SetU16(&b, 1)
	} else {
		remlen = uint8(len(m.Payloads) + 2 + len(m.Topic))
		SetU8(&b, remlen)
		SetString(&b, m.Topic)
	}

	SetPayload(&b, string(m.Payloads))
	log.Info("write len", remlen)
	log.Info(b.Bytes())
	w.Write(b.Bytes())
	return nil
}

func (m *Publish) Decode(r io.Reader, h Header) error {
	//remlen := int(GetRemLength(r))
	m.Topic = GetString(r, h.RemLen)
	m.MID = GetU16(r, h.RemLen)
	if *h.RemLen > 0 {
		m.Payloads = make([]byte, *h.RemLen)
		if n, err := r.Read(m.Payloads); err != nil {
			log.Info(n, err)
		}
	}
	return nil
}

func (m *PubAck) Encode(w io.Writer) error {
	var b bytes.Buffer
	SetU8(&b, uint8(PUBACK)<<4)
	SetU8(&b, 0x2)
	SetU16(&b, m.MID)
	w.Write(b.Bytes())
	log.Info("pub ack")
	return nil
}

func (m *PubAck) Decode(r io.Reader, h Header) error {
	return nil
}

func (m *PingReq) Encode(w io.Writer) error {
	return nil
}

func (m *PingReq) Decode(r io.Reader, h Header) error {
	//remlen := GetRemLength(r)
	return nil
}

func (m *PingResp) Encode(w io.Writer) error {
	var b bytes.Buffer
	SetU8(&b, uint8(PINGRESP)<<4)
	SetU8(&b, 0x0)
	_, err := w.Write(b.Bytes())
	return err

}
func (m *PingResp) Decode(r io.Reader, h Header) error {
	return nil
}
