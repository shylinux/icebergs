package mp

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
)

const (
	APPID  = "appid"
	SECRET = "secret"

	ACCESS  = "access"
	OPENID  = "openid"
	TOKENS  = "tokens"
	EXPIRES = "expires"
	QRCODE  = "qrcode"
)
const (
	ERRCODE = "errcode"
	ERRMSG  = "errmsg"
)
const LOGIN = "login"

func init() {
	Index.MergeCommands(ice.Commands{
		web.PP(LOGIN): {Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, MP, mdb.Config(m, tcp.SERVER))
			}},
		}, aaa.WhiteAction(aaa.SESS, aaa.USER), ctx.ConfAction(tcp.SERVER, "https://api.weixin.qq.com"))},

		LOGIN: {Name: "login list", Help: "登录", Actions: ice.Actions{
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) { m.Cmd("web.chat.wx.access", mdb.CREATE, arg) }},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.chat.wx.access") }},
	})
}
