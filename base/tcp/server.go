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
	l = &Listener{m: m, h: mdb.HashCreate(m, arg, kit.Dict(mdb.TARGET, l), STATUS, kit.Select(ERROR, OPEN, e == nil), ERROR, kit.Format(e)), s: &Stat{}, Listener: l}
	if e == nil {
		defer l.Close()
	}
	switch cb := m.OptionCB("").(type) {
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
	default:
		m.ErrorNotImplement(cb)
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
		SERVER: {Name: "server hash auto prunes", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			LISTEN: {Name: "listen type name port=9030 host=", Hand: func(m *ice.Message, arg ...string) {
				_server_listen(m, arg...)
			}},
		}, mdb.StatusHashAction(mdb.FIELD, "time,hash,status,type,name,host,port,error,nconn"), mdb.ClearHashOnExitAction())},
	})
}
