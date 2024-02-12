package web

import (
	"net"
	"strings"

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
	m.Cmd(tcp.CLIENT, tcp.DIAL, mdb.TYPE, tcp.UDP4, tcp.HOST, to_host, tcp.PORT, kit.Select(tcp.PORT_9020, to_port), func(s *net.UDPConn) {
		msg := m.Spawn(kit.Dict(tcp.HOST, host, tcp.PORT, port, arg))
		msg.Logs(tcp.SEND, BROAD, msg.FormatsMeta(nil), nfs.TO, tcp.HostPort(to_host, to_port)).FormatsMeta(s)
	})
}
func _broad_serve(m *ice.Message) {
	if m.Cmd(tcp.HOST).Length() == 0 {
		return
	}
	m.GoSleep300ms(func() {
		m.Cmd(tcp.HOST, func(value ice.Maps) {
			_broad_send(m, "", "", value[aaa.IP], m.Option(tcp.PORT), kit.Simple(gdb.EVENT, tcp.LISTEN, mdb.NAME, ice.Info.NodeName, mdb.TYPE, ice.Info.NodeType, mdb.TIME, m.Time(), cli.SimpleMake())...)
		})
	})
	m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, tcp.UDP4, mdb.NAME, logs.FileLine(1), m.OptionSimple(tcp.HOST, tcp.PORT), func(from *net.UDPAddr, buf []byte) {
		if m.WarnNotValid(len(buf) > 1024, "broad recv buf size too large") {
			return
		}
		msg := m.Spawn(buf).Logs(tcp.RECV, BROAD, string(buf), nfs.FROM, from)
		if strings.HasPrefix(msg.Option(tcp.HOST), "169.254") {
			return
		}
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
		BROAD: {Help: "广播台", Icon: "Podcasts.png", Actions: ice.MergeActions(ice.Actions{
			SERVE_START: {Hand: func(m *ice.Message, arg ...string) { gdb.Go(m, _broad_serve) }},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					host, domain := m.Cmdv(tcp.HOST, aaa.IP), UserWeb(m).Hostname()
					m.Cmds("", func(value ice.Maps) {
						switch kit.If(value[tcp.HOST] == host, func() { value[tcp.HOST] = domain }); value[mdb.TYPE] {
						case "sshd":
							m.PushSearch(mdb.NAME, Script(m, "ssh -p %s %s@%s", value[tcp.PORT], m.Option(ice.MSG_USERNAME), value[tcp.HOST]), mdb.TEXT, HostPort(m, value[tcp.HOST], value[tcp.PORT]), value)
						default:
							m.PushSearch(mdb.TEXT, HostPort(m, value[tcp.HOST], value[tcp.PORT]), value)
						}
					})
				}
			}},
			SERVE: {Name: "serve port=9020 host", Hand: func(m *ice.Message, arg ...string) { gdb.Go(m, _broad_serve) }},
			ADMIN: {Hand: func(m *ice.Message, arg ...string) { broadOpen(m) }},
			DREAM: {Hand: func(m *ice.Message, arg ...string) { broadOpen(m) }},
			VIMER: {Hand: func(m *ice.Message, arg ...string) { broadOpen(m) }},
			SPIDE: {Name: "spide name type=repos", Icon: "bi bi-house-add", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple(mdb.NAME), ORIGIN, HostPort(m, m.Option(tcp.HOST), m.Option(tcp.PORT)))
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(HostPort(m, m.Option(mdb.NAME), m.Option(tcp.PORT)))
			}},
			tcp.SEND: {Hand: func(m *ice.Message, arg ...string) { _broad_send(m, "", "", "", "", arg...) }},
		}, gdb.EventsAction(SERVE_START), mdb.HashAction(mdb.SHORT, "host,port",
			mdb.FIELD, "time,hash,type,name,host,port,module,version,commitTime,compileTime,bootTime,kernel,arch",
			mdb.ACTION, "admin,dream,vimer,spide,open", mdb.SORT, "type,name,host,port"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.StatusTimeCount("nodename", ice.Info.NodeName)
		}},
	})
}
func broadOpen(m *ice.Message) {
	m.ProcessOpen(HostPort(m, m.Option(mdb.NAME), m.Option(tcp.PORT)) + C(m.ActionKey()))
}
