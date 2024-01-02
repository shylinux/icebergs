package chat

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/log"
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
				m.GoSleep30ms(func() {
					p := m.Cmdx(web.SPACE, web.DOMAIN)
					link := tcp.PublishLocalhost(m, m.Options(ice.MSG_USERWEB, p).MergePodCmd("", "", web.SPACE, kit.Keys(web.ParseLink(m, p)[ice.POD], m.Option(mdb.NAME)), log.DEBUG, m.Option(ice.MSG_DEBUG)))
					m.Cmd(web.SPACE, m.Option(mdb.NAME), cli.PWD, m.Option(mdb.NAME), link, m.Cmdx(cli.QRCODE, link))
				})
			}},
			web.HOME: {Help: "首页", Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(web.MergeLink(m, web.CHAT_PORTAL)) }},
			aaa.CONFIRM: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.R.Method == http.MethodGet, ice.ErrNotAllow) {
					return
				} else if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) || m.Warn(m.Option(web.SPACE) == "", ice.ErrNotValid, web.SPACE) {
					return
				} else if msg := m.Cmd(web.SPACE, m.Option(web.SPACE)); m.Warn(msg.Append(mdb.TYPE) != aaa.LOGIN, ice.ErrNotFound, m.Option(web.SPACE)) {
					return
				} else {
					web.RenderCookie(m, aaa.SessValid(m))
					m.Option(ice.MSG_USERUA, msg.Append(aaa.UA))
					if ls := kit.Split(m.Option(web.SPACE), nfs.PT); len(ls) > 1 {
						m.Option(ice.MSG_SESSID, m.Cmdx(web.SPACE, kit.Keys(kit.Slice(ls, 0, -1)), aaa.SESS, mdb.CREATE, m.Option(ice.MSG_USERNAME)))
					} else {
						aaa.SessCreate(m, m.Option(ice.MSG_USERNAME))
					}
					m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, m.Option(ice.MSG_SESSID))
					m.ProcessLocation(web.MergeLink(m, msg.Append(mdb.TEXT)))
					kit.If(m.IsWeixinUA(), func() { m.Echo(ice.SUCCESS) })
					gdb.Event(m, web.SPACE_GRANT, m.OptionSimple(web.SPACE))
				}
			}},
		}, aaa.RoleAction(aaa.CONFIRM), gdb.EventsAction(web.SPACE_LOGIN)), Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(web.SPACE, m.Option(web.SPACE, arg[0]))
			m.Options(tcp.HOSTNAME, ice.Info.Hostname, nfs.PATH, msg.Append(mdb.TEXT))
			if !m.Warn(m.Option(nfs.PATH) == "", ice.ErrNotFound, arg[0]) {
				if m.EchoInfoButton(nfs.Template(m, "auth.html"), aaa.CONFIRM); m.IsWeixinUA() {
					m.OptionFields(mdb.DETAIL)
					m.Push(web.SPACE, arg[0])
					m.Push(aaa.IP, msg.Append(aaa.IP))
					m.Push(aaa.UA, msg.Append(aaa.UA))
					m.PushAction(aaa.CONFIRM)
				}
			}
		}},
	})
}
