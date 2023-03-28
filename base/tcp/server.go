package tcp

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

type Stat struct{ nc, nr, nw int }

type Listener struct {
	net.Listener
	m *ice.Message
	h string
	s *Stat
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
	l, e := net.Listen(TCP, m.Option(HOST)+ice.DF+m.Option(PORT))
	l = &Listener{Listener: l, m: m, h: mdb.HashCreate(m, arg, kit.Dict(mdb.TARGET, l), STATUS, kit.Select(ERROR, OPEN, e == nil), ERROR, kit.Format(e)), s: &Stat{}}
	defer kit.If(e == nil, func() { l.Close() })
	switch cb := m.OptionCB("").(type) {
	case func(net.Listener):
		m.Assert(e)
		cb(l)
	case func(net.Conn):
		m.Assert(e)
		for {
			if c, e := l.Accept(); !m.Warn(e) {
				cb(c)
			} else {
				break
			}
		}
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	PROTOCOL = "protocol"
	HOSTPORT = "hostport"
	HOSTNAME = "hostname"
	NODENAME = "nodename"
)
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
	LISTEN = "listen"
)
const SERVER = "server"

func init() {
	Index.MergeCommands(ice.Commands{
		SERVER: {Name: "server hash auto prunes", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			LISTEN: {Name: "listen type name port=9030 host=", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case UDP4:
					_server_udp(m, arg...)
				default:
					_server_listen(m, arg...)
				}
			}},
		}, mdb.StatusHashAction(mdb.FIELD, "time,hash,status,type,name,host,port,error"), mdb.ClearOnExitHashAction())},
	})
}
