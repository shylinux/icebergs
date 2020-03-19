package wx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"

	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"sort"
	"strings"
)

func parse(m *ice.Message) {
	data := struct {
		ToUserName   string
		FromUserName string
		CreateTime   int
		MsgId        int64
		MsgType      string
		Content      string
	}{}
	xml.NewDecoder(m.R.Body).Decode(&data)
	m.Option("ToUserName", data.ToUserName)
	m.Option("FromUserName", data.FromUserName)
	m.Option("CreateTime", data.CreateTime)

	m.Option("MsgType", data.MsgType)
	m.Option("Content", data.Content)
}

func reply(m *ice.Message) {
	m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
	m.Render(m.Conf("login", "meta.template.text"))
}
func action(m *ice.Message) {
	m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)

	m.Echo(`<xml>
<ToUserName><![CDATA[%s]]></ToUserName>
<FromUserName><![CDATA[%s]]></FromUserName>
<CreateTime>%s</CreateTime>
<MsgType><![CDATA[news]]></MsgType>
`, m.Option("FromUserName"), m.Option("ToUserName"), m.Option("CreateTime"))

	count := 0
	m.Table(func(index int, value map[string]string, head []string) {
		count++
	})
	m.Echo(`<ArticleCount>%d</ArticleCount>
`, count)

	m.Echo(`<Articles>
`)
	m.Table(func(index int, value map[string]string, head []string) {
		m.Echo(`<item>
	<Title><![CDATA[%s]]></Title>
	<Description><![CDATA[%s]]></Description>
	<PicUrl><![CDATA[%s]]></PicUrl>
	<Url><![CDATA[%s]]></Url>
</item>
`, value["name"], value["text"], value["view"], value["link"])
	})

	m.Echo(`</Articles>
`)
	m.Echo(`</xml>
`)
	m.Info("what %v", m.Result())
}

var Index = &ice.Context{Name: "wx", Help: "公众号",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"login": {Name: "login", Help: "认证", Value: kit.Data(
			"auth", "/sns/jscode2session?grant_type=authorization_code",
			"weixin", "https://api.weixin.qq.com",
			"appid", "", "appmm", "", "token", "",
			"userrole", kit.Dict(),
			"template", kit.Dict(
				"text", `<xml>
				<ToUserName><![CDATA[{{.Option "FromUserName"}}]]></ToUserName>
				<FromUserName><![CDATA[{{.Option "ToUserName"}}]]></FromUserName>
				<CreateTime>{{.Option "CreateTime"}}</CreateTime>
				<MsgType><![CDATA[text]]></MsgType>
				<Content><![CDATA[{{.Append "reply"}}]]></Content>
				</xml>`,
			),
			"menu", []interface{}{
				kit.Dict("name", "home", "text", "主页", "view", "https://shylinux.com/static/volcanos/favicon.ico", "link", "https://shylinux.com"),
				kit.Dict("name", "sub", "text", "工具", "view", "https://shylinux.com/static/volcanos/favicon.ico", "link", "https://shylinux.com"),
			},
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Confm("login", "meta.userrole", func(key string, value string) {
				m.Cmd(ice.AAA_ROLE, value, key)
			})
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("login")
		}},

		"/login/": {Name: "/login/", Help: "认证", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			check := []string{m.Conf("login", "meta.token"), m.Option("timestamp"), m.Option("nonce")}
			sort.Strings(check)
			if b := sha1.Sum([]byte(strings.Join(check, ""))); m.Warn(m.Option("signature") != hex.EncodeToString(b[:]), "error") {
				// 验证失败
				return
			}

			if m.Option("echostr") != "" {
				// 绑定验证
				web.Render(m, m.Option("echostr"))
				return
			}

			// 解析数据
			parse(m)

			// 用户登录
			m.Option(ice.MSG_USERNAME, m.Option("FromUserName"))
			m.Option(ice.MSG_USERROLE, m.Cmdx(ice.AAA_ROLE, "check", m.Option("FromUserName")))
			m.Info("%s: %s", m.Option(ice.MSG_USERROLE), m.Option(ice.MSG_USERNAME))
			m.Option(ice.MSG_SESSID, m.Cmdx(ice.AAA_USER, "login", m.Option(ice.MSG_USERNAME)))

			switch m.Option("MsgType") {
			case "text":
				if cmds := kit.Split(m.Option("Content")); !m.Right(cmds) {
					action(m.Cmdy("menu"))
				} else {
					switch cmds[0] {
					case "menu":
						action(m.Cmdy("menu"))
					default:
						// 执行命令
						msg := m.Cmd(cmds)
						if m.Hand = false; !msg.Hand {
							msg = m.Cmd(ice.CLI_SYSTEM, cmds)
						}
						if msg.Result() == "" {
							msg.Table()
						}

						// 返回结果
						reply(m.Push("reply", msg.Result()))
					}
				}
			}

		}},

		"menu": {Name: "menu", Help: "菜单", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			kit.Fetch(m.Confv("login", "meta.menu"), func(index int, value map[string]interface{}) {
				m.Push("", value, []string{"name", "text", "view"})
				m.Push("link", kit.MergeURL(kit.Format(value["link"]), ice.MSG_SESSID, m.Option(ice.MSG_SESSID)))
			})
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
