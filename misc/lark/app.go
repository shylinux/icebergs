package lark

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		APP: {Name: APP, Help: "应用", Value: kit.Data(
			mdb.SHORT, APPID, mdb.FIELD, "time,appid,duty,token,expire",
			tcp.SERVER, "https://open.feishu.cn/",
		)},
	}, Commands: map[string]*ice.Command{
		APP: {Name: "app appid auto token login", Help: "应用", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, LARK, m.Config(tcp.SERVER))
			}},
			LOGIN: {Name: "login appid appmm duty", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, arg)
			}},
			TOKEN: {Name: "token appid", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(APP, m.Option(APPID))
				if now := time.Now().Unix(); msg.Append(TOKEN) == "" || now > kit.Int64(msg.Append(EXPIRE)) {
					res := m.Cmd(web.SPIDE, LARK, "/open-apis/auth/v3/tenant_access_token/internal/",
						APP_ID, msg.Append(APPID), "app_secret", msg.Append(APPMM))

					m.Cmd(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(APPID),
						TOKEN, res.Append("tenant_access_token"), EXPIRE, now+kit.Int64(res.Append(EXPIRE)))
					m.Echo(res.Append("tenant_access_token"))
					return
				}
				m.Echo(msg.Append(TOKEN))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
