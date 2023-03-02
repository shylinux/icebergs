package web

import (
	"net"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
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
func _broad_send(m *ice.Message, host, port string, remote_host, remote_port string, arg ...string) {
	if s, e := net.DialUDP("udp4", nil, _broad_addr(m, remote_host, remote_port)); !m.Warn(e, ice.ErrNotValid) {
		defer s.Close()
		msg := m.Spawn(kit.Dict(tcp.HOST, host, tcp.PORT, port), kit.Dict(arg))
		m.Logs(tcp.SEND, BROAD, msg.FormatMeta(), nfs.TO, remote_host+ice.DF+remote_port)
		s.Write([]byte(msg.FormatMeta()))
	}
}
func _broad_serve(m *ice.Message, port string) {
	if s, e := net.ListenUDP("udp4", _broad_addr(m, "0.0.0.0", port)); m.Assert(e) {
		defer s.Close()
		m.Go(func() {
			m.Sleep("10ms").Cmd(tcp.HOST, func(values ice.Maps) {
				_broad_send(m, values[aaa.IP], port, "255.255.255.255", "9020", mdb.TYPE, ice.Info.NodeType, mdb.NAME, ice.Info.NodeName)
			})
		})
		buf := make([]byte, ice.MOD_BUFS)
		for {
			n, from, e := s.ReadFromUDP(buf[:])
			if e != nil {
				break
			}
			m.Logs(tcp.RECV, BROAD, string(buf[:n]), nfs.FROM, from)
			msg := m.Spawn(buf[:n])
			if msg.Option(mdb.ZONE) == "echo" {
				_broad_save(m, msg)
				continue
			}
			if remote := _broad_addr(m, msg.Option(tcp.HOST), msg.Option(tcp.PORT)); remote != nil {
				m.Cmd(BROAD, func(value ice.Maps) {
					m.Logs(tcp.SEND, BROAD, kit.Format(value), nfs.TO, kit.Format(remote))
					s.WriteToUDP([]byte(m.Spawn(value, kit.Dict(mdb.ZONE, "echo")).FormatMeta()), remote)
				})
				_broad_save(m, msg)
			}
		}
	}
}
func _broad_save(m, msg *ice.Message) {
	save := false
	m.Cmd(tcp.HOST, func(values ice.Maps) {
		if strings.Split(msg.Option(tcp.HOST), ice.PT)[0] == strings.Split(values[aaa.IP], ice.PT)[0] {
			save = true
		}
	})
	if save {
		mdb.HashCreate(m, msg.OptionSimple(kit.Simple(msg.Optionv(ice.MSG_OPTION))...))
	}
}

const BROAD = "broad"

func init() {
	Index.MergeCommands(ice.Commands{
		BROAD: {Name: "broad hash auto", Help: "广播", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == m.CommandKey() || arg[0] == mdb.FOREACH && arg[1] == "" {
					host := m.Cmd(tcp.HOST).Append(aaa.IP)
					domain := OptionUserWeb(m).Hostname()
					m.Cmd("", ice.Maps{ice.MSG_FIELDS: ""}, func(values ice.Maps) {
						if values[tcp.HOST] == host {
							values[tcp.HOST] = domain
						}
						switch values[mdb.TYPE] {
						case "sshd":
							m.PushSearch(mdb.NAME, ice.Render(m, ice.RENDER_SCRIPT, kit.Format("ssh -p %s %s@%s", values[tcp.PORT], m.Option(ice.MSG_USERNAME), values[tcp.HOST])),
								mdb.TEXT, kit.Format("http://%s:%s", values[tcp.HOST], values[tcp.PORT]), values)
						default:
							m.PushSearch(mdb.TEXT, kit.Format("http://%s:%s", values[tcp.HOST], values[tcp.PORT]), values)
						}
					})
				}
			}},
			SERVE: {Name: "serve port=9020", Hand: func(m *ice.Message, arg ...string) {
				_broad_serve(m, m.Option(tcp.PORT))
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessOpen(m, kit.Format("http://%s:%s", m.Option(tcp.HOST), m.Option(tcp.PORT)))
			}},
			tcp.SEND: {Hand: func(m *ice.Message, arg ...string) {
				_broad_send(m, "", "", "255.255.255.255", "9020", arg...)
			}},
		}, mdb.HashAction(mdb.SHORT, "host,port", mdb.FIELD, "time,hash,type,name,host,port", mdb.ACTION, OPEN), mdb.ClearHashOnExitAction())},
	})
}
