package aaa

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _offer_create(m *ice.Message, arg ...string) {
	h := mdb.HashCreate(m.Spawn(), FROM, m.Option(ice.MSG_USERNAME), mdb.STATUS, INVITE, m.OptionSimple(EMAIL, SUBJECT, CONTENT))
	SendEmail(m.Options("link", m.Cmdx("host", "publish", m.MergePodCmd("", "", mdb.HASH, h))), m.Option(FROM), "", "")
	gdb.Event(m, OFFER_CREATE, mdb.HASH, h, EMAIL, m.Option(EMAIL))
}
func _offer_accept(m *ice.Message, arg ...string) {
	msg := mdb.HashSelect(m.Spawn(), m.Option(mdb.HASH))
	if ls := kit.Split(msg.Append(EMAIL), mdb.AT); !m.WarnNotFound(msg.Length() == 0 || len(ls) < 2, m.Option(mdb.HASH)) {
		m.Spawn().AdminCmd(USER, mdb.CREATE, USERROLE, VOID, USERNAME, msg.Append(EMAIL), USERNICK, ls[0], USERZONE, ls[1])
		mdb.HashModify(m, m.OptionSimple(mdb.HASH), mdb.STATUS, ACCEPT)
		gdb.Event(m, OFFER_ACCEPT, mdb.HASH, m.Option(mdb.HASH), EMAIL, msg.Append(EMAIL))
		m.ProcessLocation(m.MergePod("", ice.MSG_SESSID, SessValid(m.Options(ice.MSG_USERNAME, msg.Append(EMAIL)))))
	}
}

const (
	INVITE       = "invite"
	ACCEPT       = "accept"
	SUBJECT_HTML = "subject.html"
	CONTENT_HTML = "content.html"

	OFFER_CREATE = "offer.create"
	OFFER_ACCEPT = "offer.accept"
)
const APPLY = "apply"
const OFFER = "offer"

func init() {
	Index.MergeCommands(ice.Commands{
		OFFER: {Help: "邀请", Role: VOID, Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict("from", "发自")),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create from*=admin email*='shy@shylinux.com' subject content", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				_offer_create(m, arg...)
			}},
			ACCEPT: {Help: "接受", Role: VOID, Hand: func(m *ice.Message, arg ...string) {
				if !m.WarnNotValid(m.Option(mdb.HASH), mdb.HASH) {
					_offer_accept(m, arg...)
				}
			}},
		}, mdb.ImportantHashAction(
			mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,from,status,email,subject,content"), EMAIL, ADMIN,
		), Hand: func(m *ice.Message, arg ...string) {
			if m.WarnNotRight(len(arg) == 0 && !IsTechOrRoot(m)) {
				return
			} else if mdb.HashSelect(m, arg...).FieldsIsDetail() {
				if m.Option(ice.MSG_USERNAME) == "" {
					m.Option(ice.MSG_USERHOST, strings.Split(m.Option(ice.MSG_USERHOST), "://")[1])
					m.SetAppend().EchoInfoButton(m.Template(SUBJECT_HTML), ACCEPT)
				} else if strings.Contains(m.Option(ice.MSG_USERWEB), "/c/offer") {
					m.ProcessLocation(m.MergePod(""))
				}
			}
		}},
	})
}
func OfferAction() ice.Actions {
	return gdb.EventsAction(OFFER_CREATE, OFFER_ACCEPT, USER_CREATE, USER_REMOVE)
}
