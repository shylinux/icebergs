package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	INVITE = "invite"
	ACCEPT = "accept"
)
const OFFER = "offer"

func init() {
	Index.MergeCommands(ice.Commands{
		OFFER: {Name: "offer hash auto", Help: "邀请", Actions: ice.MergeActions(ice.Actions{
			INVITE: {Name: "invite email*='shylinux@163.com' content", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m, m.OptionSimple(EMAIL, "content"), "from", m.Option(ice.MSG_USERNAME), mdb.STATUS, INVITE)
				msg := m.Cmd("web.share", mdb.CREATE, mdb.TYPE, "field", mdb.NAME, m.PrefixKey(), mdb.TEXT, kit.Format(kit.List(h)),
					kit.Dict(ice.MSG_USERNAME, m.Option(EMAIL), ice.MSG_USERNICK, VOID, ice.MSG_USERROLE, VOID))
				m.Cmd(EMAIL, SEND, m.Option(EMAIL), "welcome to contents, please continue", ice.Render(m, ice.RENDER_ANCHOR, msg.Option(mdb.LINK)))
			}},
			ACCEPT: {Help: "接受", Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.Option(mdb.HASH) == "", ice.ErrNotValid, mdb.HASH) {
					return
				}
				msg := m.Cmd("", m.Option(mdb.HASH))
				m.Cmd(USER, mdb.CREATE, USERNAME, msg.Option(EMAIL))
				m.ProcessOpen(kit.MergeURL2(m.Option(ice.MSG_USERWEB), ice.PS, ice.MSG_SESSID, SessCreate(m, msg.Option(EMAIL))))
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), mdb.STATUS, ACCEPT)
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,status,from,email,content"), RoleAction(ACCEPT)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 && m.Option(ice.MSG_USERROLE) == VOID {
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.Action(INVITE)
			} else {
				m.PushAction(ACCEPT)
			}
		}},
	})
}
