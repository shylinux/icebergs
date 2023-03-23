package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
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
			INVITE: {Name: "invite email*='shylinux@163.com' subject content", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m.Spawn(), m.OptionSimple(EMAIL, SUBJECT, CONTENT), "from", m.Option(ice.MSG_USERNAME), mdb.STATUS, INVITE)
				m.Cmd(EMAIL, SEND, m.Option(EMAIL), m.OptionDefault(SUBJECT, "welcome to contents, please continue"),
					m.OptionDefault(CONTENT, ice.Render(m, ice.RENDER_ANCHOR, m.Cmdx("host", "publish", m.MergePodCmd("", "", mdb.HASH, h, gdb.DEBUG, ice.TRUE)))),
				)
			}},
			ACCEPT: {Help: "接受", Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.Option(mdb.HASH) == "", ice.ErrNotValid, mdb.HASH) {
					return
				}
				msg := m.Cmd("", m.Option(mdb.HASH))
				if ls := kit.Split(msg.Append(EMAIL), ice.AT); !m.Warn(msg.Length() == 0 || len(ls) < 2, ice.ErrNotValid, m.Option(mdb.HASH)) {
					m.Cmd(USER, mdb.CREATE, USERNICK, ls[0], USERNAME, msg.Append(EMAIL), USERZONE, ls[1])
					m.ProcessOpen(kit.MergeURL2(m.Option(ice.MSG_USERWEB), ice.PS, ice.MSG_SESSID, SessCreate(m, msg.Append(EMAIL)), mdb.HASH, "", gdb.DEBUG, ice.TRUE))
					mdb.HashModify(m, m.OptionSimple(mdb.HASH), mdb.STATUS, ACCEPT)
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,status,from,email,title,content"), RoleAction(ACCEPT)), Hand: func(m *ice.Message, arg ...string) {
			if !m.Warn(len(arg) == 0 && m.Option(ice.MSG_USERROLE) == VOID, ice.ErrNotRight) {
				kit.If(mdb.HashSelect(m, arg...).FieldsIsDetail(), func() { m.PushAction(ACCEPT) }, func() { m.Action(INVITE) })
			}
		}},
	})
}
