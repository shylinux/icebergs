package aaa

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	INVITER = "inviter"
	INVITE  = "invite"
	ACCEPT  = "accept"
)
const APPLY = "apply"
const OFFER = "offer"

func init() {
	const (
		SUBJECT_HTML = "subject.html"
		CONTENT_HTML = "content.html"
	)
	Index.MergeCommands(ice.Commands{
		OFFER: {Help: "邀请", Role: VOID, Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict("from", "发自", "inviter", "邀请人")),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create from*=admin email*='shy@shylinux.com' subject content", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m.Spawn(), m.OptionSimple(EMAIL, SUBJECT, CONTENT), INVITER, m.Option(ice.MSG_USERNAME), mdb.STATUS, INVITE)
				SendEmail(m.Options("link", m.Cmdx("host", "publish", m.MergePodCmd("", "", mdb.HASH, h))), m.Option(FROM), "", "")
				m.Cmd("count", mdb.CREATE, OFFER, m.Option(FROM), kit.Dict(ice.LOG_DISABLE, ice.TRUE))
			}},
			ACCEPT: {Help: "接受", Role: VOID, Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(m.Option(mdb.HASH) == "", ice.ErrNotValid, mdb.HASH) {
					return
				}
				msg := mdb.HashSelect(m.Spawn(), m.Option(mdb.HASH))
				if ls := kit.Split(msg.Append(EMAIL), mdb.AT); !m.Warn(msg.Length() == 0 || len(ls) < 2, ice.ErrNotValid, m.Option(mdb.HASH)) {
					ice.Info.AdminCmd(m.Spawn(), USER, mdb.CREATE, USERNICK, ls[0], USERNAME, msg.Append(EMAIL), USERZONE, ls[1])
					m.ProcessLocation(m.MergePod("", ice.MSG_SESSID, SessValid(m.Options(ice.MSG_USERNAME, msg.Append(EMAIL)))))
					mdb.HashModify(m, m.OptionSimple(mdb.HASH), mdb.STATUS, ACCEPT)
				}
			}},
		}, mdb.ImportantHashAction(EMAIL, ADMIN, mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,status,inviter,email,title,content")), Hand: func(m *ice.Message, arg ...string) {
			if !m.Warn(len(arg) == 0 && m.Option(ice.MSG_USERROLE) == VOID, ice.ErrNotRight) {
				kit.If(mdb.HashSelect(m, arg...).FieldsIsDetail(), func() {
					if m.Option(ice.MSG_USERNAME) != "" {
						m.ProcessLocation(m.MergePod(""))
						return
					}
					m.Option(ice.MSG_USERHOST, strings.Split(m.Option(ice.MSG_USERHOST), "://")[1])
					m.SetAppend().EchoInfoButton(ice.Info.Template(m, SUBJECT_HTML), ACCEPT)
				})
			}
		}},
	})
}
