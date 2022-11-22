package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const OFFER = "offer"

func init() {
	const (
		ACCEPT = "accept"
	)
	Index.MergeCommands(ice.Commands{
		OFFER: {Name: "offer hash auto", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create email", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m, m.OptionSimple(EMAIL), "from", m.Option(ice.MSG_USERNAME))
				msg := m.Cmd("web.share", mdb.CREATE, mdb.TYPE, "field", mdb.NAME, m.PrefixKey(), mdb.TEXT, kit.Format(kit.List(h)),
					kit.Dict(ice.MSG_USERNAME, m.Option(EMAIL), ice.MSG_USERNICK, VOID, ice.MSG_USERROLE, VOID))
				m.Cmd(EMAIL, SEND, m.Option(EMAIL), "welcome to contents, please continue", ice.Render(m, ice.RENDER_ANCHOR, msg.Option(mdb.LINK)))
			}},
			ACCEPT: {Help: "接受", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(USER, mdb.CREATE, USERNAME, m.Option(EMAIL))
				m.ProcessOpen(kit.MergeURL2(m.Option(ice.MSG_USERWEB), ice.PS, ice.MSG_SESSID, m.Cmdx(SESS, mdb.CREATE, USERNAME, m.Option(EMAIL))))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,email,from"), RoleAction(ACCEPT)), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_USERROLE) == VOID && len(arg) == 0 {
				return
			}
			if mdb.HashSelect(m, arg...).PushAction(ACCEPT); len(arg) == 0 {
				m.Action(mdb.CREATE)
			}
		}},
	})
}
