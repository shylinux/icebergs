package wx

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _wx_sign(m *ice.Message, nonce, stamp string) string {
	b := sha1.Sum([]byte(strings.Join(kit.Sort([]string{
		fmt.Sprintf("jsapi_ticket=%s", m.Cmdx(ACCESS, TICKET)),
		fmt.Sprintf("url=%s", m.Option(ice.MSG_USERWEB)),
		fmt.Sprintf("timestamp=%s", stamp),
		fmt.Sprintf("noncestr=%s", nonce),
	}), "&")))
	return hex.EncodeToString(b[:])
}
func _wx_config(m *ice.Message, nonce string) {
	m.Option(APPID, m.Conf(LOGIN, kit.Keym(APPID)))
	m.Option(SCRIPT, m.Conf(ACCESS, kit.Keym(SCRIPT)))
	m.Option("signature", _wx_sign(m, m.Option("noncestr", nonce), m.Option("timestamp", kit.Format(time.Now().Unix()))))
}

const (
	WEIXIN  = "weixin"
	EXPIRE  = "expire"
	TICKET  = "ticket"
	EXPIRES = "expires"
	SCRIPT  = "script"
	CONFIG  = "config"
)
const (
	ERRCODE = "errcode"
	ERRMSG  = "errmsg"
)
const ACCESS = "access"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ACCESS: {Name: ACCESS, Help: "认证", Value: kit.Data(
				SCRIPT, "https://res.wx.qq.com/open/js/jweixin-1.6.0.js",
				WEIXIN, "https://api.weixin.qq.com",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, WEIXIN, m.Conf(ACCESS, kit.Keym(WEIXIN)))
			}},
			ACCESS: {Name: "access appid auto ticket token login", Help: "认证", Action: map[string]*ice.Action{
				LOGIN: {Name: "login appid appmm token", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
					m.Conf(LOGIN, kit.Keym(APPID), m.Option(APPID))
					m.Conf(LOGIN, kit.Keym(APPMM), m.Option(APPMM))
					m.Conf(LOGIN, kit.Keym(TOKEN), m.Option(TOKEN))
				}},
				TOKEN: {Name: "token", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
					if now := time.Now().Unix(); m.Conf(ACCESS, kit.Keym(TOKEN)) == "" || now > kit.Int64(m.Conf(ACCESS, kit.Keym(EXPIRE))) {
						msg := m.Cmd(web.SPIDE, WEIXIN, web.SPIDE_GET, "/cgi-bin/token?grant_type=client_credential",
							APPID, m.Conf(LOGIN, kit.Keym(APPID)), "secret", m.Conf(LOGIN, kit.Keym(APPMM)))
						if m.Warn(msg.Append(ERRCODE) != "", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
							return
						}

						m.Conf(ACCESS, kit.Keym(EXPIRE), now+kit.Int64(msg.Append("expires_in")))
						m.Conf(ACCESS, kit.Keym(TOKEN), msg.Append("access_token"))
					}
					m.Echo(m.Conf(ACCESS, kit.Keym(TOKEN)))
				}},
				TICKET: {Name: "ticket", Help: "票据", Hand: func(m *ice.Message, arg ...string) {
					if now := time.Now().Unix(); m.Conf(ACCESS, kit.Keym(TICKET)) == "" || now > kit.Int64(m.Conf(ACCESS, kit.Keym(EXPIRES))) {
						msg := m.Cmd(web.SPIDE, WEIXIN, web.SPIDE_GET, "/cgi-bin/ticket/getticket?type=jsapi", "access_token", m.Cmdx(ACCESS, TOKEN))
						if m.Warn(msg.Append(ERRCODE) != "0", msg.Append(ERRCODE), msg.Append(ERRMSG)) {
							return
						}

						m.Conf(ACCESS, kit.Keym(EXPIRES), now+kit.Int64(msg.Append("expires_in")))
						m.Conf(ACCESS, kit.Keym(TICKET), msg.Append(TICKET))
					}
					m.Echo(m.Conf(ACCESS, kit.Keym(TICKET)))
				}},
				CONFIG: {Name: "config", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
					_wx_config(m, "some")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(kit.Formats(m.Confv(ACCESS)))
			}},
		}})
}
