package tcp

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
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
func (c *Conn) Close() error {
	return c.Conn.Close()
}

func _client_dial(m *ice.Message, arg ...string) {
	c, e := net.Dial(TCP, m.Option(HOST)+ice.DF+m.Option(PORT))

	c = &Conn{m: m, s: &Stat{}, Conn: c}
	if e == nil {
		defer c.Close()
	}

	switch cb := m.OptionCB(CLIENT).(type) {
	case func(net.Conn, error):
		cb(c, e)
	case func(net.Conn):
		m.Assert(e)
		cb(c)
	case func(net.Conn, []byte, error):
		b := make([]byte, ice.MOD_BUFS)
		for {
			n, e := c.Read(b)
			if cb(c, b[:n], e); e != nil {
				break
			}
		}
	default:
		c.Write([]byte("hello world\n"))
	}
}

const (
	PROTO  = "proto"
	STATUS = "status"
	ERROR  = "error"
	OPEN   = "open"
	CLOSE  = "close"
	START  = "start"
	STOP   = "stop"
)
const (
	DIAL = "dial"
)
const CLIENT = "client"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CLIENT: {Name: CLIENT, Help: "客户端", Value: kit.Data(
			mdb.FIELD, "time,hash,status,type,name,host,port,error,nread,nwrite",
		)},
	}, Commands: map[string]*ice.Command{
		CLIENT: {Name: "client hash auto prunes", Help: "客户端", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Richs(CLIENT, "", mdb.FOREACH, func(key string, value map[string]interface{}) {
					kit.Value(value, kit.Keym(STATUS), CLOSE)
				})
				m.Cmdy(SERVER, mdb.PRUNES)
			}},
			DIAL: {Name: "dial type name port=9010 host=", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				_client_dial(m, arg...)
			}},
		}, mdb.HashActionStatus()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(index int, value map[string]string, head []string) {
				m.PushButton(kit.Select("", mdb.REMOVE, value[STATUS] == OPEN))
			})
		}},
	}})
}
