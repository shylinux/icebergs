package wx

import (
	"crypto/sha1"
	"encoding/xml"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

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
func _wx_action(m *ice.Message) {
	m.Set(ice.MSG_RESULT).RenderResult()

	m.Echo(`<xml>
<FromUserName><![CDATA[%s]]></FromUserName>
<ToUserName><![CDATA[%s]]></ToUserName>
<CreateTime>%s</CreateTime>
<MsgType><![CDATA[%s]]></MsgType>
`, m.Option("ToUserName"), m.Option("FromUserName"), m.Option("CreateTime"), "news")

	count := 0
	m.Table(func(index int, value map[string]string, head []string) { count++ })
	m.Echo(`<ArticleCount>%d</ArticleCount>`, count)

	share := m.Cmdx(web.SHARE, mdb.CREATE, kit.MDB_TYPE, web.LOGIN)

	m.Echo(`<Articles>`)
	m.Table(func(index int, value map[string]string, head []string) {
		m.Echo(`<item>
<Title><![CDATA[%s]]></Title>
<Description><![CDATA[%s]]></Description>
<PicUrl><![CDATA[%s]]></PicUrl>
<Url><![CDATA[%s]]></Url>
</item>
`, value[wiki.TITLE], value[wiki.SPARK], value[wiki.IMAGE],
			kit.MergeURL(kit.Format(value[wiki.REFER]), web.SHARE, share))
	})
	m.Echo(`</Articles>`)
	m.Echo(`</xml>`)

	m.Debug("echo: %v", m.Result())
}

const LOGIN = "login"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		LOGIN: {Name: LOGIN, Help: "登录", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		"/login/": {Name: "/login/", Help: "认证", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			check := kit.Sort([]string{m.Conf(ACCESS, "meta.tokens"), m.Option("timestamp"), m.Option("nonce")})
			if sig := kit.Format(sha1.Sum([]byte(strings.Join(check, "")))); m.Warn(sig != m.Option("signature"), ice.ErrNotRight) {
				return // 验证失败
			}
			if m.Option("echostr") != "" {
				m.RenderResult(m.Option("echostr"))
				return // 绑定验证
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
				if m.Warn(!m.Right(cmds), ice.ErrNotRight) {
					cmds = []string{MENU, mdb.CREATE}
				}
				m.Cmdy(TEXT, cmds)
			}
		}},
	}})
}
