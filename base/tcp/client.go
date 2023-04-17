package tcp

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type Conn struct {
	net.Conn
	m *ice.Message
	h string
	s *Stat
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
	c, e := net.Dial(TCP, m.Option(HOST)+nfs.DF+m.Option(PORT))
	c = &Conn{Conn: c, m: m, s: &Stat{}}
	defer kit.If(e == nil, func() { c.Close() })
	switch cb := m.OptionCB("").(type) {
	case func(net.Conn):
		kit.If(!m.Warn(e), func() { cb(c) })
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	DIAL = "dial"
)
const CLIENT = "client"

func init() {
	Index.MergeCommands(ice.Commands{
		CLIENT: {Name: "client hash auto prunes", Help: "客户端", Actions: ice.MergeActions(ice.Actions{
			DIAL: {Name: "dial type name port=9010 host=", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case UDP4:
					_client_dial_udp4(m, arg...)
				default:
					_client_dial(m, arg...)
				}
			}},
		}, mdb.StatusHashAction(mdb.FIELD, "time,hash,status,type,name,host,port,error"), mdb.ClearOnExitHashAction())},
	})
}
