package wx

import (
	"crypto/sha1"
	"net/http"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

func _wx_sign(m *ice.Message, nonce, stamp string) string {
	return kit.Format(sha1.Sum([]byte(kit.Join(kit.Sort(kit.Simple(
		kit.Format("jsapi_ticket=%s", m.Cmdx(ACCESS, TICKET)),
		kit.Format("url=%s", m.Option(ice.MSG_USERWEB)),
		kit.Format("timestamp=%s", stamp),
		kit.Format("noncestr=%s", nonce),
	)), "&"))))
}
func _wx_config(m *ice.Message, nonce string) {
	m.Option("signature", _wx_sign(m, m.Option("noncestr", nonce), m.Option("timestamp", kit.Format(time.Now().Unix()))))
	ctx.OptionFromConfig(m, APPID, nfs.SCRIPT)
}
func _wx_check(m *ice.Message) {
	check := kit.Sort([]string{mdb.Config(m, TOKEN), m.Option("timestamp"), m.Option("nonce")})
	if sig := kit.Format(sha1.Sum([]byte(strings.Join(check, "")))); !m.Warn(sig != m.Option("signature"), ice.ErrNotRight, check) {
		kit.If(m.Option("echostr") != "", func() { m.RenderResult(m.Option("echostr")) }, func() { m.Echo(ice.TRUE) })
	}
}

const (
	APPID   = "appid"
	APPMM   = "appmm"
	TOKEN   = "token"
	TOKENS  = "tokens"
	EXPIRES = "expires"
	TICKET  = "ticket"
	EXPIRE  = "expire"
	CONFIG  = "config"
	CHECK   = "check"
)
const (
	ERRCODE = "errcode"
	ERRMSG  = "errmsg"
)
const ACCESS = "access"

func init() {
	Index.MergeCommands(ice.Commands{
		ACCESS: {Name: "access appid auto ticket tokens login", Help: "认证", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, WX, mdb.Config(m, tcp.SERVER))
				gdb.Watch(m, chat.HEADER_AGENT, m.PrefixKey())
			}},
			chat.HEADER_AGENT: {Hand: func(m *ice.Message, arg ...string) {
				if strings.Index(m.Option(ice.MSG_USERUA), "MicroMessenger") > -1 {
					_wx_config(m, mdb.Config(m, APPID))
				}
			}},
			LOGIN: {Name: "login appid appmm token", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				ctx.ConfigFromOption(m, APPID, APPMM, TOKEN)
			}},
			TOKENS: {Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); mdb.Config(m, TOKENS) == "" || now > kit.Int64(mdb.Config(m, EXPIRES)) {
					msg := m.Cmd(web.SPIDE, WX, http.MethodGet, "/cgi-bin/token?grant_type=client_credential", APPID, mdb.Config(m, APPID), "secret", mdb.Config(m, APPMM))
					if m.Warn(msg.Append(ERRCODE) != "", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
						return
					}
					mdb.Config(m, EXPIRES, now+kit.Int64(msg.Append("expires_in")))
					mdb.Config(m, TOKENS, msg.Append("access_token"))
				}
				m.Echo(mdb.Config(m, TOKENS)).Status(EXPIRES, time.Unix(kit.Int64(mdb.Config(m, EXPIRES)), 0).Format(ice.MOD_TIME))
			}},
			TICKET: {Help: "票据", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); mdb.Config(m, TICKET) == "" || now > kit.Int64(mdb.Config(m, EXPIRE)) {
					msg := m.Cmd(web.SPIDE, WX, http.MethodGet, "/cgi-bin/ticket/getticket?type=jsapi", "access_token", m.Cmdx("", TOKENS))
					if m.Warn(msg.Append(ERRCODE) != "0", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
						return
					}
					mdb.Config(m, EXPIRE, now+kit.Int64(msg.Append("expires_in")))
					mdb.Config(m, TICKET, msg.Append(TICKET))
				}
				m.Echo(mdb.Config(m, TICKET)).Status(EXPIRE, time.Unix(kit.Int64(mdb.Config(m, EXPIRE)), 0).Format(ice.MOD_TIME))
			}},
			CONFIG: {Hand: func(m *ice.Message, arg ...string) { _wx_config(m, mdb.Config(m, APPID)) }},
			CHECK:  {Hand: func(m *ice.Message, arg ...string) { _wx_check(m) }},
		}, mdb.HashAction(tcp.SERVER, "https://api.weixin.qq.com", nfs.SCRIPT, "/plugin/local/chat/wx.js")), Hand: func(m *ice.Message, arg ...string) {
			m.Echo(mdb.Config(m, APPID))
		}},
	})
}
