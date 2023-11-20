package chat

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const GRANT = "grant"

func init() {
	Index.MergeCommands(ice.Commands{
		GRANT: {Name: "grant space auto", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			web.SPACE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					link := tcp.PublishLocalhost(m, m.MergePodCmd("", "", web.SPACE, m.Option(mdb.NAME)))
					m.Sleep30ms(web.SPACE, m.Option(mdb.NAME), cli.PWD, m.Option(mdb.NAME), link, m.Cmdx(cli.QRCODE, link))
				})
			}},
			"home": {Help: "首页", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(web.MergeLink(m, "/chat/portal/"))
			}},
			aaa.CONFIRM: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.R.Method == http.MethodGet, ice.ErrNotAllow) {
					return
				} else if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) || m.Warn(m.Option(web.SPACE) == "", ice.ErrNotValid, web.SPACE) {
					return
				} else if msg := m.Cmd(web.SPACE, m.Option(web.SPACE)); m.Warn(msg.Append(mdb.TYPE) != aaa.LOGIN, ice.ErrNotFound, m.Option(web.SPACE)) {
					return
				} else {
					if m.IsWeixinUA() {
						m.Option(ice.MSG_USERUA, msg.Append(ice.MSG_USERUA))
						m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
						m.Echo(ice.SUCCESS)
					} else {
						kit.If(m.Option(ice.MSG_SESSID) == "", func() { web.RenderCookie(m, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME))) })
						m.Option(ice.MSG_USERUA, msg.Append(ice.MSG_USERUA))
						m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
						m.ProcessLocation(web.MergeURL2(m, msg.Append(mdb.TEXT)))
					}
				}
			}},
		}, aaa.RoleAction(aaa.CONFIRM), gdb.EventsAction(web.SPACE_LOGIN)), Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(web.SPACE, m.Option(web.SPACE, arg[0]))
			m.Option(tcp.HOSTNAME, ice.Info.Hostname)
			m.Option(nfs.PATH, msg.Append(mdb.TEXT))
			if !m.Warn(m.Option(nfs.PATH) == "", ice.ErrNotFound, arg[0]) {
				if m.IsWeixinUA() {
					m.Push(aaa.IP, msg.Append(aaa.IP))
					m.Push(web.SPACE, arg[0])
					m.PushAction(aaa.CONFIRM)
				} else {
					m.Option(aaa.UA, msg.Append(aaa.UA))
					m.Option(aaa.IP, msg.Append(aaa.IP))
					m.Echo(nfs.Template(m, "auth.html"))
				}
			}
		}},
	})
}
