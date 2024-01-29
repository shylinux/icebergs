package web

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _broad_send(m *ice.Message, to_host, to_port string, host, port string, arg ...string) {
	m.Cmd(tcp.CLIENT, tcp.DIAL, mdb.TYPE, tcp.UDP4, tcp.HOST, to_host, tcp.PORT, kit.Select("9020", to_port), func(s *net.UDPConn) {
		msg := m.Spawn(kit.Dict(tcp.HOST, host, tcp.PORT, port, arg))
		msg.Logs(tcp.SEND, BROAD, msg.FormatsMeta(nil), nfs.TO, tcp.HostPort(to_host, to_port)).FormatsMeta(s)
	})
}
func _broad_serve(m *ice.Message) {
	m.GoSleep300ms(func() {
		m.Cmd(tcp.HOST, func(value ice.Maps) {
			_broad_send(m, "", "", value[aaa.IP], m.Option(tcp.PORT), kit.Simple(gdb.EVENT, tcp.LISTEN, mdb.NAME, ice.Info.Hostname, mdb.TYPE, ice.Info.NodeType, mdb.TIME, m.Time(), cli.SimpleMake())...)
		})
	})
	m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, tcp.UDP4, mdb.NAME, logs.FileLine(1), m.OptionSimple(tcp.HOST, tcp.PORT), func(from *net.UDPAddr, buf []byte) {
		msg := m.Spawn(buf).Logs(tcp.RECV, BROAD, string(buf), nfs.FROM, from)
		if m.Cmd(BROAD, mdb.CREATE, msg.OptionSimple(kit.Simple(msg.Optionv(ice.MSG_OPTION))...)); msg.Option(gdb.EVENT) == tcp.LISTEN {
			m.Cmds(BROAD, func(value ice.Maps) {
				_broad_send(m, msg.Option(tcp.HOST), msg.Option(tcp.PORT), value[tcp.HOST], value[tcp.PORT], kit.Simple(value)...)
			})
		}
	})
}

const BROAD = "broad"

func init() {
	Index.MergeCommands(ice.Commands{
		BROAD: {Help: "广播台", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
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
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) { gdb.Go(m, _broad_serve) }},
			SERVE:       {Name: "serve port=9020 host", Hand: func(m *ice.Message, arg ...string) { gdb.Go(m, _broad_serve) }},
			ADMIN: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(Domain(m.Option(mdb.NAME), m.Option(tcp.PORT)) + C(m.ActionKey()))
			}},
			DREAM: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(Domain(m.Option(mdb.NAME), m.Option(tcp.PORT)) + C(m.ActionKey()))
			}},
			VIMER: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(Domain(m.Option(mdb.NAME), m.Option(tcp.PORT)) + C(m.ActionKey()))
			}},
			OPEN:     {Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(Domain(m.Option(mdb.NAME), m.Option(tcp.PORT))) }},
			tcp.SEND: {Hand: func(m *ice.Message, arg ...string) { _broad_send(m, "", "", "", "", arg...) }},
		}, gdb.EventsAction(SERVE_START), mdb.HashAction(mdb.SHORT, "host,port",
			mdb.FIELD, "time,hash,type,name,host,port,module,version,commitTime,compileTime,bootTime,kernel,arch",
			mdb.ACTION, "admin,dream,vimer,open", mdb.SORT, "type,name,host,port"), mdb.ClearOnExitHashAction())},
	})
}
