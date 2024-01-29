package web

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const TOKEN = "token"

func init() {
	Index.MergeCommands(ice.Commands{
		TOKEN: {Help: "令牌桶", Actions: mdb.HashAction(mdb.SHORT, mdb.UNIQ, mdb.EXPIRE, mdb.MONTH)},
	})
}

const (
	DEV_REQUEST_TEXT = "devRequestText"
	DEV_CREATE_TOKEN = "devCreateToken"
)

func DevTokenAction(name, origin string) ice.Actions {
	return ice.Actions{
		DEV_REQUEST_TEXT: {Hand: func(m *ice.Message, arg ...string) { m.Echo(UserHost(m)) }},
		DEV_CREATE_TOKEN: {Hand: func(m *ice.Message, arg ...string) {}},
		mdb.DEV_REQUEST: {Name: "dev.request", Help: "连接", Icon: "bi bi-person-down", Hand: func(m *ice.Message, arg ...string) {
			back := m.Options(ice.MSG_USERWEB, m.Option(ice.MSG_USERHOST)).MergePod("")
			m.ProcessOpen(m.Options(ice.MSG_USERWEB, m.Option(origin), ice.MSG_USERPOD, "").MergePodCmd("", m.PrefixKey(),
				ctx.ACTION, mdb.DEV_CHOOSE, cli.BACK, back, cli.DAEMON, m.Option(ice.MSG_DAEMON),
				m.OptionSimple(name), mdb.TEXT, m.Cmdx("", DEV_REQUEST_TEXT),
			))
		}},
		mdb.DEV_CHOOSE: {Hand: func(m *ice.Message, arg ...string) {
			m.EchoInfoButton(kit.JoinWord(m.PrefixKey(),
				m.Cmdx(nfs.CAT, nfs.SRC_TEMPLATE+"web.token/saveto.html"), m.Option(cli.BACK)), mdb.DEV_RESPONSE)
		}},
		mdb.DEV_RESPONSE: {Help: "确认", Hand: func(m *ice.Message, arg ...string) {
			if !m.Warn(m.Option(ice.MSG_METHOD) != http.MethodPost, ice.ErrNotAllow) {
				m.ProcessReplace(m.ParseLink(m.Option(cli.BACK)).MergePodCmd("", m.PrefixKey(), ctx.ACTION, mdb.DEV_CONFIRM, m.OptionSimple(cli.DAEMON),
					m.OptionSimple(name), TOKEN, m.Cmdx(TOKEN, mdb.CREATE, mdb.TYPE, m.CommandKey(), mdb.NAME, m.Option(ice.MSG_USERNAME), m.OptionSimple(mdb.TEXT))))
			}
		}},
		mdb.DEV_CONFIRM: {Hand: func(m *ice.Message, arg ...string) {
			m.EchoInfoButton(kit.JoinWord(m.PrefixKey(), m.Cmdx(nfs.CAT, nfs.SRC_TEMPLATE+"web.token/savefrom.html"), m.Option(name)), mdb.DEV_CREATE)
		}},
		mdb.DEV_CREATE: {Help: "创建", Hand: func(m *ice.Message, arg ...string) {
			if !m.Warn(m.Option(ice.MSG_METHOD) != http.MethodPost, ice.ErrNotAllow) {
				defer kit.If(m.Option(cli.DAEMON), func(p string) { m.Cmd(SPACE, p, html.REFRESH) })
				mdb.HashModify(m, m.OptionSimple(name, TOKEN))
				m.Cmdy("", DEV_CREATE_TOKEN).ProcessClose()
			}
		}},
	}
}
