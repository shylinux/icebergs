package wx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/icebergs/core/wiki"
	"github.com/shylinux/toolkits"

	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

func _wx_sign(m *ice.Message, nonce, stamp string) string {
	b := sha1.Sum([]byte(strings.Join(kit.Sort([]string{
		fmt.Sprintf("noncestr=%s", nonce),
		fmt.Sprintf("timestamp=%s", stamp),
		fmt.Sprintf("url=%s", m.Option(ice.MSG_USERWEB)),
		fmt.Sprintf("jsapi_ticket=%s", m.Cmdx(ACCESS, TICKET)),
	}), "&")))
	return hex.EncodeToString(b[:])
}
func _wx_config(m *ice.Message, nonce string) {
	m.Option(APPID, m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPID)))
	m.Option("signature", _wx_sign(m, m.Option("noncestr", nonce), m.Option("timestamp", kit.Format(time.Now().Unix()))))
}
func _wx_parse(m *ice.Message) {
	data := struct {
		FromUserName string
		ToUserName   string
		CreateTime   int64
		MsgId        int64
		MsgType      string
		Content      string
		Event        string
	}{}
	xml.NewDecoder(m.R.Body).Decode(&data)
	m.Debug("data: %#v", data)

	m.Option("FromUserName", data.FromUserName)
	m.Option("ToUserName", data.ToUserName)
	m.Option("CreateTime", data.CreateTime)

	m.Option("MsgId", data.MsgId)
	m.Option("MsgType", data.MsgType)
	m.Option("Content", data.Content)
	m.Option("Event", data.Event)
}
func _wx_reply(m *ice.Message, tmpl string) {
	m.Render(m.Conf(LOGIN, kit.Keys("meta.template", tmpl)))
}
func _wx_action(m *ice.Message) {
	m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)

	m.Echo(`<xml>
<FromUserName><![CDATA[%s]]></FromUserName>
<ToUserName><![CDATA[%s]]></ToUserName>
<CreateTime>%s</CreateTime>
<MsgType><![CDATA[%s]]></MsgType>
`, m.Option("ToUserName"), m.Option("FromUserName"), m.Option("CreateTime"), "news")

	count := 0
	m.Table(func(index int, value map[string]string, head []string) { count++ })
	m.Echo(`<ArticleCount>%d</ArticleCount>`, count)

	m.Echo(`<Articles>`)
	m.Table(func(index int, value map[string]string, head []string) {
		m.Echo(`<item>
	<Title><![CDATA[%s]]></Title>
	<Description><![CDATA[%s]]></Description>
	<PicUrl><![CDATA[%s]]></PicUrl>
	<Url><![CDATA[%s]]></Url>
</item>
`, value[wiki.TITLE], value[wiki.SPARK], value[wiki.IMAGE], value[wiki.REFER])
	})
	m.Echo(`</Articles>`)
	m.Echo(`</xml>`)

	m.Debug("echo: %v", m.Result())
}

const (
	LOGIN  = "login"
	APPID  = "appid"
	APPMM  = "appmm"
	TOKEN  = "token"
	TICKET = "ticket"
	ACCESS = "access"
	CONFIG = "config"
	WEIXIN = "weixin"
)
const WX = "wx"

var Index = &ice.Context{Name: WX, Help: "公众号",
	Configs: map[string]*ice.Config{
		LOGIN: {Name: LOGIN, Help: "认证", Value: kit.Data(
			WEIXIN, "https://api.weixin.qq.com", APPID, "", APPMM, "", TOKEN, "",
			"template", kit.Dict("text", `<xml>
				<FromUserName><![CDATA[{{.Option "ToUserName"}}]]></FromUserName>
				<ToUserName><![CDATA[{{.Option "FromUserName"}}]]></ToUserName>
				<CreateTime>{{.Option "CreateTime"}}</CreateTime>
				<MsgType><![CDATA[text]]></MsgType>
				<Content><![CDATA[{{.Result}}]]></Content>
				</xml>`),
			"menu", []interface{}{
				kit.Dict(wiki.TITLE, "主页", wiki.SPARK, "点击进入", wiki.IMAGE, "https://shylinux.com/static/volcanos/favicon.ico", wiki.REFER, "https://shylinux.com"),
				kit.Dict(wiki.TITLE, "产品", wiki.SPARK, "工具", wiki.IMAGE, "https://shylinux.com/static/volcanos/favicon.ico", wiki.REFER, "https://shylinux.com?river=product"),
				kit.Dict(wiki.TITLE, "研发", wiki.SPARK, "工具", wiki.IMAGE, "https://shylinux.com/static/volcanos/favicon.ico", wiki.REFER, "https://shylinux.com?river=project"),
			},
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(web.SPIDE, mdb.CREATE, WEIXIN, m.Conf(LOGIN, kit.Keys(kit.MDB_META, WEIXIN)))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},
		ACCESS: {Name: "access appid auto ticket token login", Help: "认证", Action: map[string]*ice.Action{
			LOGIN: {Name: "login appid appmm token", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPID), m.Option(APPID))
				m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPMM), m.Option(APPMM))
				m.Conf(LOGIN, kit.Keys(kit.MDB_META, TOKEN), m.Option(TOKEN))
			}},
			TOKEN: {Name: "token", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Conf(LOGIN, "meta.access.token") == "" || now > kit.Int64(m.Conf(LOGIN, "meta.access.expire")) {
					msg := m.Cmd(web.SPIDE, "weixin", web.SPIDE_GET, "/cgi-bin/token?grant_type=client_credential",
						APPID, m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPID)), "secret", m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPMM)))

					m.Conf(LOGIN, "meta.access.token", msg.Append("access_token"))
					m.Conf(LOGIN, "meta.access.expire", now+kit.Int64(msg.Append("expires_in")))
				}
				m.Echo(m.Conf(LOGIN, "meta.access.token"))
			}},
			TICKET: {Name: "ticket", Help: "票据", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Conf(LOGIN, "meta.access.ticket") == "" || now > kit.Int64(m.Conf(LOGIN, "meta.access.expires")) {
					msg := m.Cmd(web.SPIDE, "weixin", web.SPIDE_GET, "/cgi-bin/ticket/getticket?type=jsapi",
						"access_token", m.Cmdx(ACCESS, TOKEN))

					m.Conf(LOGIN, "meta.access.ticket", msg.Append(TICKET))
					m.Conf(LOGIN, "meta.access.expires", now+kit.Int64(msg.Append("expires_in")))
				}
				m.Echo(m.Conf(LOGIN, "meta.access.ticket"))
			}},
			CONFIG: {Name: "config", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				_wx_config(m, "some")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Push(APPID, m.Conf(LOGIN, kit.Keys(kit.MDB_META, APPID)))
		}},

		"menu": {Name: "menu name auto", Help: "菜单", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				share := m.Cmdx(web.SHARE, mdb.CREATE, kit.MDB_TYPE, "login")
				kit.Fetch(m.Confv(LOGIN, "meta.menu"), func(index int, value map[string]interface{}) {
					m.Push("", value, kit.Split("title,spark,image"))
					m.Push(wiki.REFER, kit.MergeURL(kit.Format(value[wiki.REFER]), web.SHARE, share))
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			kit.Fetch(m.Confv(LOGIN, "meta.menu"), func(index int, value map[string]interface{}) {
				m.Push("", value, kit.Split("title,spark,image,refer"))
			})
		}},

		"/login/": {Name: "/login/", Help: "认证", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			check := kit.Sort([]string{m.Conf(LOGIN, "meta.token"), m.Option("timestamp"), m.Option("nonce")})
			if b := sha1.Sum([]byte(strings.Join(check, ""))); m.Warn(m.Option("signature") != hex.EncodeToString(b[:]), ice.ErrNotRight) {
				return // 验证失败
			}
			if m.Option("echostr") != "" {
				m.Render(ice.RENDER_RESULT, m.Option("echostr"))
				return // 绑定验证
			}

			// 解析数据
			_wx_parse(m)

			// 用户登录
			m.Option(ice.MSG_USERZONE, WX)
			aaa.UserLogin(m, m.Append("FromUserName"), "")

			switch m.Option("MsgType") {
			case "event":
				switch m.Option("Event") {
				case "subscribe":
					// 应用列表
					_wx_action(m.Cmdy("menu", mdb.CREATE))
				case "unsubscribe":
				}

			case "text":
				if cmds := kit.Split(m.Option("Content")); m.Warn(!m.Right(cmds), ice.ErrNotRight) {
					_wx_action(m.Cmdy("menu", mdb.CREATE))
					break // 没有权限
				} else {
					switch cmds[0] {
					case "menu":
						// 应用列表
						_wx_action(m.Cmdy("menu", mdb.CREATE))

					default:
						// 执行命令
						if m.Cmdy(cmds); len(m.Appendv(ice.MSG_APPEND)) == 0 && len(m.Resultv()) == 0 {
							m.Cmdy(cli.SYSTEM, cmds)
						} else if len(m.Resultv()) == 0 {
							m.Table()
						}

						// 返回结果
						_wx_reply(m, "text")
					}
				}
			}
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
