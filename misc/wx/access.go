package wx

import (
	"crypto/sha1"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
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
	m.Option(APPID, m.Config(APPID))
	m.Option(ssh.SCRIPT, m.Config(ssh.SCRIPT))
	m.Option("signature", _wx_sign(m, m.Option("noncestr", nonce), m.Option("timestamp", kit.Format(time.Now().Unix()))))
}
func _wx_check(m *ice.Message) {
	check := kit.Sort([]string{m.Config(TOKEN), m.Option("timestamp"), m.Option("nonce")})
	if sig := kit.Format(sha1.Sum([]byte(strings.Join(check, "")))); m.Warn(sig != m.Option("signature"), ice.ErrNotRight, check) {
		return // 验证失败
	}
	if m.Option("echostr") != "" {
		m.RenderResult(m.Option("echostr"))
		return // 绑定验证
	}
	m.Echo(ice.TRUE)
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
	Index.Merge(&ice.Context{Configs: ice.Configs{
		ACCESS: {Name: ACCESS, Help: "认证", Value: kit.Data(
			tcp.SERVER, "https://api.weixin.qq.com", ssh.SCRIPT, "/plugin/local/chat/wx.js",
		)},
	}, Commands: ice.Commands{
		ACCESS: {Name: "access appid auto config ticket tokens login", Help: "认证", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, WX, m.Config(tcp.SERVER))
			}},
			LOGIN: {Name: "login appid appmm token", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Config(APPID, m.Option(APPID))
				m.Config(APPMM, m.Option(APPMM))
				m.Config(TOKEN, m.Option(TOKEN))
			}},

			TOKENS: {Name: "tokens", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Config(TOKENS) == "" || now > kit.Int64(m.Config(EXPIRES)) {
					msg := m.Cmd(web.SPIDE, WX, web.SPIDE_GET, "/cgi-bin/token?grant_type=client_credential",
						APPID, m.Config(APPID), "secret", m.Config(APPMM))
					if m.Warn(msg.Append(ERRCODE) != "", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
						return
					}

					m.Config(EXPIRES, now+kit.Int64(msg.Append("expires_in")))
					m.Config(TOKENS, msg.Append("access_token"))
				}
				m.Echo(m.Config(TOKENS))
			}},
			TICKET: {Name: "ticket", Help: "票据", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Config(TICKET) == "" || now > kit.Int64(m.Config(EXPIRE)) {
					msg := m.Cmd(web.SPIDE, WX, web.SPIDE_GET, "/cgi-bin/ticket/getticket?type=jsapi", "access_token", m.Cmdx(ACCESS, TOKENS))
					if m.Warn(msg.Append(ERRCODE) != "0", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
						return
					}

					m.Config(EXPIRE, now+kit.Int64(msg.Append("expires_in")))
					m.Config(TICKET, msg.Append(TICKET))
				}
				m.Echo(m.Config(TICKET))
			}},
			CONFIG: {Name: "config", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				_wx_config(m, m.Config(APPID))
			}},
			CHECK: {Name: "check", Help: "检验", Hand: func(m *ice.Message, arg ...string) {
				_wx_check(m)
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Echo(m.Config(APPID))
		}},
	}})
}
