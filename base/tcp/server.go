package tcp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net"
	"strings"
)

type Stat struct {
	nc, nr, nw int
}

type Conn struct {
	h string
	m *ice.Message
	s *Stat

	net.Conn
}

func (c *Conn) Read(b []byte) (int, error) {
	n, e := c.Conn.Read(b)
	c.s.nr += n
	c.m.Conf(CLIENT, kit.Keys(kit.MDB_HASH, c.h, kit.MDB_META, "nwrite"), c.s.nw)
	return n, e
}
func (c *Conn) Write(b []byte) (int, error) {
	n, e := c.Conn.Write(b)
	c.s.nw += n
	c.m.Conf(CLIENT, kit.Keys(kit.MDB_HASH, c.h, kit.MDB_META, "nread"), c.s.nr)
	return n, e
}
func (c *Conn) Close() error {
	c.m.Cmd(mdb.MODIFY, CLIENT, "", mdb.HASH, kit.MDB_HASH, c.h, kit.MDB_STATUS, CLOSE, "nread", c.s.nr, "nwrite", c.s.nw)
	return c.Conn.Close()
}

type Listener struct {
	h string
	m *ice.Message
	s *Stat

	net.Listener
}

func (l Listener) Accept() (net.Conn, error) {
	c, e := l.Listener.Accept()
	l.s.nc += 1
	l.m.Conf(SERVER, kit.Keys(kit.MDB_HASH, l.h, kit.MDB_META, "nconn"), l.s.nc)

	ls := strings.Split(c.RemoteAddr().String(), ":")
	if strings.Contains(c.RemoteAddr().String(), "[") {
		ls = strings.Split(strings.TrimPrefix(c.RemoteAddr().String(), "["), "]:")
	}
	h := l.m.Cmdx(mdb.INSERT, CLIENT, "", mdb.HASH, HOST, ls[0], PORT, ls[1],
		kit.MDB_NAME, l.m.Option(kit.MDB_NAME), kit.MDB_STATUS, kit.Select(ERROR, OPEN, e == nil), kit.MDB_ERROR, kit.Format(e))

	c = &Conn{h: h, m: l.m, s: &Stat{}, Conn: c}
	return c, e
}
func (l Listener) Close() error {
	l.m.Cmd(mdb.MODIFY, SERVER, "", mdb.HASH, kit.MDB_HASH, l.h, kit.MDB_STATUS, CLOSE)
	return l.Listener.Close()
}

const (
	OPEN  = "open"
	CLOSE = "close"
	ERROR = "error"
)
const (
	LISTEN_CB = "listen.cb"
	LISTEN    = "listen"
)

const SERVER = "server"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVER: {Name: SERVER, Help: "服务器", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			SERVER: {Name: "server hash auto 监听 清理", Help: "服务器", Action: map[string]*ice.Action{
				LISTEN: {Name: "LISTEN host=localhost port=9010", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
					l, e := net.Listen(TCP, m.Option(HOST)+":"+m.Option(PORT))
					h := m.Option(kit.MDB_HASH)
					if h == "" {
						h = m.Cmdx(mdb.INSERT, SERVER, "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME), HOST, m.Option(HOST), PORT, m.Option(PORT), kit.MDB_STATUS, kit.Select(ERROR, OPEN, e == nil), kit.MDB_ERROR, kit.Format(e))
					}

					l = &Listener{h: h, m: m, s: &Stat{}, Listener: l}
					if e == nil {
						defer l.Close()
					}

					switch cb := m.Optionv(LISTEN_CB).(type) {
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
							if e == nil {
								b := make([]byte, 1024)
								if n, e := c.Read(b); e == nil {
									m.Info("nonce", string(b[:n]))
									c.Write(b[:n])
								}
							} else {
								break
							}
							c.Close()
						}
					}
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SERVER, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, SERVER, "", mdb.HASH, kit.MDB_STATUS, ERROR)
					m.Cmdy(mdb.PRUNES, SERVER, "", mdb.HASH, kit.MDB_STATUS, CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select(mdb.DETAIL, "time,hash,status,name,host,port,error,nconn", len(arg) == 0))
				m.Cmdy(mdb.SELECT, SERVER, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("删除")
			}},
		},
	}, nil)
}
