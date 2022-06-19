package tcp

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
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
	l.m.Cmd(mdb.MODIFY, SERVER, "", mdb.HASH, mdb.HASH, l.h, STATUS, CLOSE)
	return l.Listener.Close()
}

func _server_listen(m *ice.Message, arg ...string) {
	l, e := net.Listen(TCP, m.Option(HOST)+":"+m.Option(PORT))
	h := m.Cmdx(mdb.INSERT, SERVER, "", mdb.HASH, arg,
		STATUS, kit.Select(ERROR, OPEN, e == nil), ERROR, kit.Format(e))

	l = &Listener{m: m, h: h, s: &Stat{}, Listener: l}
	if e == nil {
		defer l.Close()
	}

	switch cb := m.OptionCB(SERVER).(type) {
	case func(net.Listener, error):
		cb(l, e)
	case func(net.Listener):
		m.Assert(e)
		cb(l)
	case func(net.Conn):
		for {
			if c, e := l.Accept(); e == nil {
				cb(c)
			} else {
				break
			}
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
}

const (
	PROTOCOL = "protocol"
	HOSTPORT = "hostport"
	HOSTNAME = "hostname"
)
const (
	LISTEN = "listen"
)
const SERVER = "server"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SERVER: {Name: SERVER, Help: "服务器", Value: kit.Data(
			mdb.FIELD, "time,hash,status,type,name,host,port,error,nconn",
		)},
	}, Commands: map[string]*ice.Command{
		SERVER: {Name: "server hash auto prunes", Help: "服务器", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Richs(SERVER, "", mdb.FOREACH, func(key string, value ice.Map) {
					kit.Value(value, kit.Keym(STATUS), CLOSE)
				})
				m.Cmdy(SERVER, mdb.PRUNES)
			}},
			LISTEN: {Name: "LISTEN type name port=9030 host=", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				_server_listen(m, arg...)
			}},
		}, mdb.HashActionStatus()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(index int, value map[string]string, head []string) {
				m.PushButton(kit.Select("", mdb.REMOVE, value[STATUS] == CLOSE))
			})
		}},
	}})
}
