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
		GRANT: {Name: "grant space auto", Help: "授权", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			web.SPACE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.GoSleep30ms(func() {
					// 	p := m.Cmdx(web.SPACE, web.DOMAIN)
					p := m.Option(ice.MSG_USERWEB)
					link := tcp.PublishLocalhost(m, m.Options(ice.MSG_USERWEB, p).MergePodCmd("", "", web.SPACE, kit.Keys(web.ParseLink(m, p)[ice.POD], m.Option(mdb.NAME))))
					m.Cmd(web.SPACE, m.Option(mdb.NAME), cli.PWD, m.Option(mdb.NAME), link, m.Cmdx(cli.QRCODE, link))
				})
			}},
			web.HOME: {Help: "首页", Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(web.C(web.ADMIN)) }},
			aaa.CONFIRM: {Help: "授权", Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				if m.WarnNotAllow(m.R.Method == http.MethodGet) {
					return
				} else if m.WarnNotLogin(m.Option(ice.MSG_USERNAME) == "") || m.WarnNotValid(m.Option(web.SPACE) == "", web.SPACE) {
					return
				} else if msg := m.Cmd(web.SPACE, m.Option(web.SPACE)); m.WarnNotFound(msg.Append(mdb.TYPE) == "", m.Option(web.SPACE)) {
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
					if m.IsWeixinUA() {
						m.Echo(ice.SUCCESS)
					} else if web.UserWeb(m).Path == "/c/web.chat.grant" {
						m.ProcessLocation(m.MergeLink(msg.Append(mdb.TEXT)))
					} else {
						m.Echo(ice.SUCCESS)
						m.ProcessInner()
					}
					gdb.Event(m, web.SPACE_GRANT, m.OptionSimple(web.SPACE))
				}
			}},
		}, gdb.EventsAction(web.SPACE_LOGIN)), Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(web.SPACE, m.Option(web.SPACE, arg[0]))
			if msg.Length() == 0 {
				return
			}
			m.Options(tcp.HOSTNAME, ice.Info.Hostname, nfs.PATH, msg.Append(mdb.TEXT))
			if !m.WarnNotValid(m.Option(nfs.PATH) == "", arg[0]) {
				m.Option(aaa.IP, msg.Append(aaa.IP))
				m.Option(ice.MSG_USERUA, msg.Append(aaa.UA))
				m.Options(web.ParseUA(m))
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
