package lark

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	LOGIN  = "login"
	APPID  = "appid"
	APPMM  = "appmm"
	TOKEN  = "token"
	EXPIRE = "expire"
)
const APP = "app"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			APP: {Name: APP, Help: "服务配置", Value: kit.Data(
				kit.MDB_SHORT, APPID, kit.MDB_FIELD, "time,appid,appmm,duty,token,expire",
				LARK, "https://open.feishu.cn/",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, LARK, m.Conf(APP, kit.Keym(LARK)))
			}},
			APP: {Name: "app appid auto token login", Help: "应用", Action: map[string]*ice.Action{
				LOGIN: {Name: "login appid appmm duty", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(APP), "", mdb.HASH, arg)
				}},
				TOKEN: {Name: "token appid", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(APP, m.Option(APPID), ice.OptionFields(m.Conf(APP, kit.META_FIELD)))
					if now := time.Now().Unix(); msg.Append(TOKEN) == "" || now > kit.Int64(msg.Append(EXPIRE)) {
						sub := m.Cmd(web.SPIDE, LARK, web.SPIDE_POST, "/open-apis/auth/v3/tenant_access_token/internal/",
							APP_ID, msg.Append(APPID), "app_secret", msg.Append(APPMM))

						m.Cmd(mdb.MODIFY, m.Prefix(APP), "", mdb.HASH, m.OptionSimple(APPID),
							TOKEN, sub.Append("tenant_access_token"), EXPIRE, now+kit.Int64(sub.Append(EXPIRE)))
						m.Echo(sub.Append("tenant_access_token"))
						return
					}
					m.Echo(msg.Append(TOKEN))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				m.Fields(len(arg), m.Conf(APP, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(APP), "", mdb.HASH, APPID, arg)
			}},
		},
	})
}
