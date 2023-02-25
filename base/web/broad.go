package web

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _broad_addr(m *ice.Message, host, port string) *net.UDPAddr {
	if addr, e := net.ResolveUDPAddr("udp4", kit.Format("%s:%s", host, port)); !m.Warn(e, ice.ErrNotValid, host, port, logs.FileLineMeta(2)) {
		return addr
	}
	return nil
}
func _broad_send(m *ice.Message, host, port string, remote_host, remote_port string) {
	if s, e := net.DialUDP("udp4", nil, _broad_addr(m, remote_host, remote_port)); !m.Warn(e, ice.ErrNotValid) {
		defer s.Close()
		msg := m.Spawn(kit.Dict(tcp.HOST, host, tcp.PORT, port))
		m.Logs(mdb.EXPORT, BROAD, msg.FormatMeta(), "to", remote_host+ice.DF+remote_port)
		s.Write([]byte(msg.FormatMeta()))
	}
}
func _broad_serve(m *ice.Message, host, port string) {
	_broad_send(m, host, port, "255.255.255.255", "9020")
	if s, e := net.ListenUDP("udp4", _broad_addr(m, "0.0.0.0", port)); m.Assert(e) {
		defer s.Close()
		defer mdb.HashCreateDeferRemove(m, tcp.HOST, host, tcp.PORT, port, kit.Dict(mdb.TARGET, s))()
		buf := make([]byte, ice.MOD_BUFS)
		for {
			n, from, e := s.ReadFromUDP(buf[:])
			if e != nil {
				break
			}
			m.Logs(mdb.IMPORT, BROAD, string(buf[:n]), "from", from)
			msg := m.Spawn(buf[:n])
			if remote := _broad_addr(m, msg.Option(tcp.HOST), msg.Option(tcp.PORT)); remote != nil {
				m.Cmd(BROAD, func(value ice.Maps) {
					m.Logs(mdb.EXPORT, BROAD, kit.Format(value), "to", kit.Format(remote))
					s.WriteToUDP([]byte(m.Spawn(value).FormatMeta()), remote)
				})
				if m.Cmd(BROAD, kit.Format("%s,%s", msg.Option(tcp.HOST), msg.Option(tcp.PORT))).Length() > 0 {
					continue
				}
				mdb.HashCreate(m, msg.OptionSimple(tcp.HOST, tcp.PORT))
			}
		}
	}
}

const BROAD = "broad"

func init() {
	Index.MergeCommands(ice.Commands{
		BROAD: {Name: "broad hash auto", Help: "广播", Actions: ice.MergeActions(ice.Actions{
			SERVE: {Name: "serve port=9020", Hand: func(m *ice.Message, arg ...string) {
				_broad_serve(m, m.Cmd(tcp.HOST).Append(aaa.IP), m.Option(tcp.PORT))
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessOpen(m, kit.Format("http://%s:%s", m.Option(tcp.HOST), m.Option(tcp.PORT)))
			}},
		}, mdb.HashAction(mdb.SHORT, "host,port", mdb.FIELD, "time,hash,host,port", mdb.ACTION, OPEN), mdb.ClearHashOnExitAction())},
	})
}
