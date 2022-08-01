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
	return &Conn{m: l.m, s: &Stat{}, Conn: c}, e
}
func (l Listener) Close() error {
	l.m.Cmd(mdb.MODIFY, SERVER, "", mdb.HASH, mdb.HASH, l.h, STATUS, CLOSE)
	return l.Listener.Close()
}

func _server_listen(m *ice.Message, arg ...string) {
	l, e := net.Listen(TCP, m.Option(HOST)+":"+m.Option(PORT))
	l = &Listener{m: m, h: m.Cmdx(mdb.INSERT, SERVER, "", mdb.HASH,
		arg, STATUS, kit.Select(ERROR, OPEN, e == nil), ERROR, kit.Format(e), kit.Dict(mdb.TARGET, l)), s: &Stat{}, Listener: l}
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
	Index.MergeCommands(ice.Commands{
		SERVER: {Name: "server hash auto prunes", Help: "服务器", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Conf("", mdb.HASH, "")
			}},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectValue(m, func(target ice.Any) {
					if l, ok := target.(net.Listener); ok {
						l.Close()
					}
					if l, ok := target.(*Listener); ok {
						l.Close()
					}
				})
			}},
			LISTEN: {Name: "LISTEN type name port=9030 host=", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				_server_listen(m, arg...)
			}},
		}, mdb.HashActionStatus(mdb.FIELD, "time,hash,status,type,name,host,port,error,nconn")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Tables(func(value ice.Maps) {
				m.PushButton(kit.Select("", mdb.REMOVE, value[STATUS] == CLOSE))
			})
		}},
	})
}
