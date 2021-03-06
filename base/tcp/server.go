package tcp

import (
	"net"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

type Listener struct {
	m *ice.Message
	h string
	s *Stat

	net.Listener
}

func (l Listener) Accept() (net.Conn, error) {
	c, e := l.Listener.Accept()
	l.s.nc += 1
	return &Conn{m: l.m, h: "", s: &Stat{}, Conn: c}, e
}
func (l Listener) Close() error {
	l.m.Cmd(mdb.MODIFY, SERVER, "", mdb.HASH, kit.MDB_HASH, l.h, kit.MDB_STATUS, CLOSE)
	return l.Listener.Close()
}

const (
	LISTEN = "listen"
)

const SERVER = "server"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVER: {Name: SERVER, Help: "服务器", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			SERVER: {Name: "server hash auto prunes", Help: "服务器", Action: map[string]*ice.Action{
				LISTEN: {Name: "LISTEN type name port=9010 host=", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
					l, e := net.Listen(TCP, m.Option(HOST)+":"+m.Option(PORT))
					h := m.Cmdx(mdb.INSERT, SERVER, "", mdb.HASH, arg,
						kit.MDB_STATUS, kit.Select(ERROR, OPEN, e == nil), kit.MDB_ERROR, kit.Format(e))

					l = &Listener{m: m, h: h, s: &Stat{}, Listener: l}
					if e == nil {
						defer l.Close()
					}

					switch cb := m.Optionv(kit.Keycb(LISTEN)).(type) {
					case func(net.Listener, error):
						cb(l, e)
					case func(net.Listener):
						m.Assert(e)
						cb(l)
					case func(net.Conn):
						for {
							c, e := l.Accept()
							if e != nil {
								break
							}
							cb(c)
						}
					case func(net.Conn, error):
						for {
							c, e := l.Accept()
							if cb(c, e); e != nil {
								break
							}
						}
					default:
						for {
							c, e := l.Accept()
							if e != nil {
								break
							}

							b := make([]byte, ice.MOD_BUFS)
							if n, e := c.Read(b); e == nil {
								m.Info("nonce", string(b[:n]))
								c.Write(b[:n])
							}
							c.Close()
						}
					}
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SERVER, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, SERVER, "", mdb.HASH, kit.MDB_STATUS, ERROR)
					m.Cmdy(mdb.PRUNES, SERVER, "", mdb.HASH, kit.MDB_STATUS, CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) == 0, "time,hash,status,type,name,host,port,error,nconn")
				if m.Cmdy(mdb.SELECT, SERVER, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) == 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushButton(kit.Select("", mdb.REMOVE, value[kit.MDB_STATUS] == CLOSE))
					})
				}
			}},
		},
	})
}
