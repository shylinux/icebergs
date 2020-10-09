package tcp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net"
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
	// c.m.Conf(CLIENT, kit.Keys(kit.MDB_HASH, c.h, kit.MDB_META, "nread"), c.s.nr)
	return n, e
}
func (c *Conn) Write(b []byte) (int, error) {
	n, e := c.Conn.Write(b)
	c.s.nw += n
	// c.m.Conf(CLIENT, kit.Keys(kit.MDB_HASH, c.h, kit.MDB_META, "nwrite"), c.s.nw)
	return n, e
}
func (c *Conn) Close() error {
	c.m.Cmd(mdb.MODIFY, CLIENT, "", mdb.HASH, kit.MDB_HASH, c.h, kit.MDB_STATUS, CLOSE, "nread", c.s.nr, "nwrite", c.s.nw)
	return c.Conn.Close()
}

const (
	OPEN  = "open"
	CLOSE = "close"
	ERROR = "error"
)
const (
	DIAL_CB = "dial.cb"
	DIAL    = "dial"
)

const CLIENT = "client"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CLIENT: {Name: CLIENT, Help: "客户端", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			CLIENT: {Name: "client hash auto 清理", Help: "客户端", Action: map[string]*ice.Action{
				DIAL: {Name: "dial type name port=9010 host=", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
					c, e := net.Dial(TCP, m.Option(HOST)+":"+m.Option(PORT))
					h := m.Cmdx(mdb.INSERT, CLIENT, "", mdb.HASH, PORT, m.Option(PORT), HOST, m.Option(HOST),
						kit.MDB_TYPE, m.Option(kit.MDB_TYPE), kit.MDB_NAME, m.Option(kit.MDB_NAME),
						kit.MDB_STATUS, kit.Select(ERROR, OPEN, e == nil), kit.MDB_ERROR, kit.Format(e))

					c = &Conn{m: m, h: h, s: &Stat{}, Conn: c}
					if e == nil {
						defer c.Close()
					}

					switch cb := m.Optionv(DIAL_CB).(type) {
					case func(net.Conn, error):
						cb(c, e)
					case func(net.Conn):
						m.Assert(e)
						cb(c)
					case func(net.Conn, []byte, error):
						b := make([]byte, 4096)
						for {
							n, e := c.Read(b)
							if cb(c, b[:n], e); e != nil {
								break
							}
						}
					default:
						c.Write([]byte("hello world\n"))
					}
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, CLIENT, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, CLIENT, "", mdb.HASH, kit.MDB_STATUS, ERROR)
					m.Cmdy(mdb.PRUNES, CLIENT, "", mdb.HASH, kit.MDB_STATUS, CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,status,type,name,host,port,error,nread,nwrite", mdb.DETAIL, len(arg) > 0))
				if m.Cmdy(mdb.SELECT, CLIENT, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) == 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushButton(kit.Select("", "删除", value[kit.MDB_STATUS] == OPEN))
					})
				}
			}},
		},
	}, nil)
}
