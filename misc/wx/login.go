package wx

import (
	"crypto/sha1"
	"encoding/xml"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _wx_check(m *ice.Message) bool {
	check := kit.Sort([]string{m.Conf(ACCESS, "meta.tokens"), m.Option("timestamp"), m.Option("nonce")})
	if sig := kit.Format(sha1.Sum([]byte(strings.Join(check, "")))); m.Warn(sig != m.Option("signature"), ice.ErrNotRight, check) {
		return false // 验证失败
	}
	if m.Option("echostr") != "" {
		m.RenderResult(m.Option("echostr"))
		return false // 绑定验证
	}
	return true
}
func _wx_parse(m *ice.Message) {
	data := struct {
		FromUserName string
		ToUserName   string
		CreateTime   int64
		MsgID        int64
		Event        string
		MsgType      string
		Content      string
	}{}
	xml.NewDecoder(m.R.Body).Decode(&data)
	m.Debug("data: %#v", data)

	m.Option("FromUserName", data.FromUserName)
	m.Option("ToUserName", data.ToUserName)
	m.Option("CreateTime", data.CreateTime)
	m.Option("MsgID", data.MsgID)

	m.Option("Event", data.Event)
	m.Option("MsgType", data.MsgType)
	m.Option("Content", data.Content)
}
func _wx_reply(m *ice.Message, tmpl string) {
	if res, err := kit.Render(m.Config(kit.MDB_TEMPLATE), m); err == nil {
		m.Set(ice.MSG_RESULT).RenderResult(string(res))
	}
}

const LOGIN = "login"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		LOGIN: {Name: LOGIN, Help: "登录", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		"/login/": {Name: "/login/", Help: "认证", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if !_wx_check(m) {
				return // 验签失败
			}

			// 解析数据
			_wx_parse(m)

			// 用户登录
			m.Option(ice.MSG_USERZONE, WX)
			aaa.UserLogin(m, m.Append("FromUserName"), "")

			switch m.Option("MsgType") {
			case EVENT: // 事件
				m.Cmdy(EVENT, m.Option("Event"))

			case TEXT: // 文本
				cmds := kit.Split(m.Option("Content"))
				if !m.Right(cmds) {
					cmds = []string{MENU, mdb.CREATE}
				}
				m.Cmdy(TEXT, cmds)
			}
		}},
	}})
}
