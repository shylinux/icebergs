package tcp

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _server_udp(m *ice.Message, arg ...string) {
	l, e := net.ListenUDP(UDP4, UDPAddr(m, "0.0.0.0", m.Option(PORT)))
	defer kit.If(e == nil, func() { l.Close() })
	mdb.HashCreate(m, arg, kit.Dict(mdb.TARGET, l), STATUS, kit.Select(ERROR, OPEN, e == nil), ERROR, kit.Format(e))
	switch cb := m.OptionCB("").(type) {
	case func(*net.UDPAddr, []byte):
		m.Assert(e)
		buf := make([]byte, 2*ice.MOD_BUFS)
		for {
			if n, from, e := l.ReadFromUDP(buf[:]); !m.WarnNotValid(e) {
				cb(from, buf[:n])
			} else {
				break
			}
		}
	default:
		m.ErrorNotImplement(cb)
	}
}
func _client_dial_udp4(m *ice.Message, arg ...string) {
	c, e := net.DialUDP(UDP4, nil, UDPAddr(m, kit.Select("255.255.255.255", m.Option(HOST)), m.Option(PORT)))
	defer kit.If(e == nil, func() { c.Close() })
	switch cb := m.OptionCB("").(type) {
	case func(*net.UDPConn):
		kit.If(!m.WarnNotValid(e), func() { cb(c) })
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	UDP4   = "udp4"
	SEND   = "send"
	RECV   = "recv"
	DIRECT = "direct"
)

func UDPAddr(m *ice.Message, host, port string) *net.UDPAddr {
	if addr, e := net.ResolveUDPAddr(UDP4, host+nfs.DF+port); !m.WarnNotValid(e, host, port, logs.FileLineMeta(2)) {
		return addr
	}
	return nil
}
func HostPort(host, port string) string { return host + nfs.DF + port }
