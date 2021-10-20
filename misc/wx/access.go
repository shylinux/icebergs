package wx

import (
	"crypto/sha1"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _wx_sign(m *ice.Message, nonce, stamp string) string {
	return kit.Format(sha1.Sum([]byte(kit.Join(kit.Sort([]string{
		kit.Format("jsapi_ticket=%s", m.Cmdx(ACCESS, TICKET)),
		kit.Format("url=%s", m.Option(ice.MSG_USERWEB)),
		kit.Format("timestamp=%s", stamp),
		kit.Format("noncestr=%s", nonce),
	}), "&"))))
}
func _wx_config(m *ice.Message, nonce string) {
	m.Option(APPID, m.Config(APPID))
	m.Option(ssh.SCRIPT, m.Config(ssh.SCRIPT))
	m.Option("signature", _wx_sign(m, m.Option("noncestr", nonce), m.Option("timestamp", kit.Format(time.Now().Unix()))))
}

const (
	APPID   = "appid"
	APPMM   = "appmm"
	TOKEN   = "token"
	EXPIRE  = "expire"
	TICKET  = "ticket"
	EXPIRES = "expires"
)
const (
	WEIXIN  = "weixin"
	ERRCODE = "errcode"
	ERRMSG  = "errmsg"
)
const ACCESS = "access"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ACCESS: {Name: ACCESS, Help: "认证", Value: kit.Data(
			ssh.SCRIPT, "https://res.wx.qq.com/open/js/jweixin-1.6.0.js",
			tcp.SERVER, "https://api.weixin.qq.com",
			APPID, "", APPMM, "", "tokens", "",
		)},
	}, Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(web.SPIDE, mdb.CREATE, WEIXIN, m.Config(tcp.SERVER))
		}},
		ACCESS: {Name: "access appid auto ticket token login", Help: "认证", Action: map[string]*ice.Action{
			LOGIN: {Name: "login appid appmm token", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Config(APPID, m.Option(APPID))
				m.Config(APPMM, m.Option(APPMM))
				m.Config("tokens", m.Option(TOKEN))
			}},
			TOKEN: {Name: "token", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Config(TOKEN) == "" || now > kit.Int64(m.Config(EXPIRE)) {
					msg := m.Cmd(web.SPIDE, WEIXIN, web.SPIDE_GET, "/cgi-bin/token?grant_type=client_credential",
						APPID, m.Config(APPID), "secret", m.Config(APPMM))
					if m.Warn(msg.Append(ERRCODE) != "", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
						return
					}

					m.Config(EXPIRE, now+kit.Int64(msg.Append("expires_in")))
					m.Config(TOKEN, msg.Append("access_token"))
				}
				m.Echo(m.Config(TOKEN))
			}},
			TICKET: {Name: "ticket", Help: "票据", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Conf(TICKET) == "" || now > kit.Int64(m.Config(EXPIRES)) {
					msg := m.Cmd(web.SPIDE, WEIXIN, web.SPIDE_GET, "/cgi-bin/ticket/getticket?type=jsapi", "access_token", m.Cmdx(ACCESS, TOKEN))
					if m.Warn(msg.Append(ERRCODE) != "0", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
						return
					}

					m.Config(EXPIRES, now+kit.Int64(msg.Append("expires_in")))
					m.Config(TICKET, msg.Append(TICKET))
				}
				m.Echo(m.Config(TICKET))
			}},
			ctx.CONFIG: {Name: "config", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				_wx_config(m, "some")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(kit.Formats(m.Confv(ACCESS)))
		}},
	}})
}
