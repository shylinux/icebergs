package tcp

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

type Stat struct {
	nc, nr, nw int
}

type Conn struct {
	m *ice.Message
	h string
	s *Stat

	net.Conn
}

func (c *Conn) Read(b []byte) (int, error) {
	n, e := c.Conn.Read(b)
	c.s.nr += n
	return n, e
}
func (c *Conn) Write(b []byte) (int, error) {
	n, e := c.Conn.Write(b)
	c.s.nw += n
	return n, e
}
func (c *Conn) Close() error { return c.Conn.Close() }

func _client_dial(m *ice.Message, arg ...string) {
	c, e := net.Dial(TCP, m.Option(HOST)+ice.DF+m.Option(PORT))
	c = &Conn{m: m, s: &Stat{}, Conn: c}
	if e == nil {
		defer c.Close()
	}
	switch cb := m.OptionCB("").(type) {
	case func(*ice.Message, net.Conn):
		if !m.Warn(e) {
			cb(m, c)
		}
	case func(net.Conn):
		if !m.Warn(e) {
			cb(c)
		}
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	PROTO  = "proto"
	STATUS = "status"
	ERROR  = "error"
	START  = "start"
	OPEN   = "open"
	CLOSE  = "close"
	STOP   = "stop"
)
const (
	DIAL = "dial"
)
const CLIENT = "client"

func init() {
	Index.MergeCommands(ice.Commands{
		CLIENT: {Name: "client hash auto prunes", Help: "客户端", Actions: ice.MergeActions(ice.Actions{
			DIAL: {Name: "dial type name port=9010 host=", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				_client_dial(m, arg...)
			}},
		}, mdb.StatusHashAction(mdb.FIELD, "time,hash,status,type,name,host,port,error,nread,nwrite"), mdb.ClearHashOnExitAction())},
	})
}
