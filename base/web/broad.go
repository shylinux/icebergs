package web

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _broad_send(m *ice.Message, to_host, to_port string, host, port string, arg ...string) {
	m.Cmd(tcp.CLIENT, tcp.DIAL, mdb.TYPE, tcp.UDP4, tcp.HOST, to_host, tcp.PORT, kit.Select("9020", to_port), func(s *net.UDPConn) {
		msg := m.Spawn(kit.Dict(tcp.HOST, host, tcp.PORT, port, arg))
		msg.Logs(tcp.SEND, BROAD, msg.FormatsMeta(nil), nfs.TO, tcp.HostPort(to_host, to_port)).FormatsMeta(s)
	})
}
func _broad_serve(m *ice.Message) {
	m.GoSleep("10ms", tcp.HOST, func(value ice.Maps) {
		_broad_send(m, "", "", value[aaa.IP], m.Option(tcp.PORT), gdb.EVENT, tcp.LISTEN, mdb.NAME, ice.Info.NodeName, mdb.TYPE, ice.Info.NodeType)
	})
	m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, tcp.UDP4, m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT), func(from *net.UDPAddr, buf []byte) {
		msg := m.Spawn(buf).Logs(tcp.RECV, BROAD, string(buf), nfs.FROM, from)
		if mdb.HashCreate(m, msg.OptionSimple(kit.Simple(msg.Optionv(ice.MSG_OPTION))...)); msg.Option(gdb.EVENT) == tcp.LISTEN {
			m.Cmds("", func(value ice.Maps) {
				_broad_send(m, msg.Option(tcp.HOST), msg.Option(tcp.PORT), value[tcp.HOST], value[tcp.PORT], mdb.TYPE, value[mdb.TYPE], mdb.NAME, value[mdb.NAME])
			})
		}
	})
}

const BROAD = "broad"

func init() {
	Index.MergeCommands(ice.Commands{
		BROAD: {Name: "broad hash auto", Help: "广播", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					host, domain := m.Cmdv(tcp.HOST, aaa.IP), UserWeb(m).Hostname()
					m.Cmds("", func(value ice.Maps) {
						switch kit.If(value[tcp.HOST] == host, func() { value[tcp.HOST] = domain }); value[mdb.TYPE] {
						case "sshd":
							m.PushSearch(mdb.NAME, Script(m, "ssh -p %s %s@%s", value[tcp.PORT], m.Option(ice.MSG_USERNAME), value[tcp.HOST]), mdb.TEXT, Domain(value[tcp.HOST], value[tcp.PORT]), value)
						default:
							m.PushSearch(mdb.TEXT, Domain(value[tcp.HOST], value[tcp.PORT]), value)
						}
					})
				}
			}},
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) { m.Go(func() { m.Cmd("", SERVE, m.OptionSimple(tcp.PORT)) }) }},
			SERVE:       {Name: "serve port=9020", Hand: func(m *ice.Message, arg ...string) { _broad_serve(m) }},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessOpen(m, Domain(m.Option(tcp.HOST), m.Option(tcp.PORT)))
			}},
			tcp.SEND: {Hand: func(m *ice.Message, arg ...string) { _broad_send(m, "", "", "", "", arg...) }},
		}, mdb.HashAction(mdb.SHORT, "host,port", mdb.FIELD, "time,hash,type,name,host,port", mdb.ACTION, OPEN), mdb.ClearOnExitHashAction())},
	})
}
