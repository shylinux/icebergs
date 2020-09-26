package tcp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net"
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
			CLIENT: {Name: "client hash auto 连接 清理", Help: "客户端", Action: map[string]*ice.Action{
				DIAL: {Name: "dial host=localhost port=9010", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
					c, e := net.Dial(TCP, m.Option(HOST)+":"+m.Option(PORT))
					h := m.Cmdx(mdb.INSERT, CLIENT, "", mdb.HASH, HOST, m.Option(HOST), PORT, m.Option(PORT), kit.MDB_STATUS, kit.Select(ERROR, OPEN, e == nil), kit.MDB_ERROR, kit.Format(e))

					c = &Conn{h: h, m: m, s: &Stat{}, Conn: c}
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
				m.Option(mdb.FIELDS, kit.Select(mdb.DETAIL, "time,hash,status,host,port,error,nread,nwrite", len(arg) == 0))
				m.Cmdy(mdb.SELECT, CLIENT, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("删除")
			}},
		},
	}, nil)
}
