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
const APPLY = "apply"
const OFFER = "offer"

func init() {
	Index.MergeCommands(ice.Commands{
		OFFER: {Help: "邀请", Role: VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create email*='shy@shylinux.com' subject content", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m.Spawn(), m.OptionSimple(EMAIL, SUBJECT, CONTENT), INVITE, m.Option(ice.MSG_USERNAME), mdb.STATUS, INVITE)
				m.Cmd(EMAIL, SEND, m.Option(EMAIL), "", m.OptionDefault(SUBJECT, "welcome to contexts, please continue"),
					m.OptionDefault(CONTENT, ice.Render(m, ice.RENDER_ANCHOR, m.Cmdx("host", "publish", m.MergePodCmd("", "", mdb.HASH, h)))),
				)
			}},
			ACCEPT: {Help: "接受", Role: VOID, Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.Option(mdb.HASH) == "", ice.ErrNotValid, mdb.HASH) {
					return
				}
				msg := mdb.HashSelect(m.Spawn(), m.Option(mdb.HASH))
				if ls := kit.Split(msg.Append(EMAIL), mdb.AT); !m.Warn(msg.Length() == 0 || len(ls) < 2, ice.ErrNotValid, m.Option(mdb.HASH)) {
					m.Cmd(USER, mdb.CREATE, USERNICK, ls[0], USERNAME, msg.Append(EMAIL), USERZONE, ls[1])
					m.ProcessLocation(kit.MergeURL2(m.Option(ice.MSG_USERWEB), ice.PS, ice.MSG_SESSID, SessValid(m.Options(ice.MSG_USERNAME, msg.Append(EMAIL))), mdb.HASH, ""))
					mdb.HashModify(m, m.OptionSimple(mdb.HASH), mdb.STATUS, ACCEPT)
				}
			}},
		}, mdb.ImportantHashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,status,invite,email,title,content")), Hand: func(m *ice.Message, arg ...string) {
			if !m.Warn(len(arg) == 0 && m.Option(ice.MSG_USERROLE) == VOID, ice.ErrNotRight) {
				kit.If(mdb.HashSelect(m, arg...).FieldsIsDetail(), func() {
					m.SetAppend()
					m.EchoInfoButton("welcome to contexts, please continue", ACCEPT)
					// m.PushAction(ACCEPT)
				})
			}
		}},
	})
}
