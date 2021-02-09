package wx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"

	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
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
	m.Option("signature", _wx_sign(m, m.Option("noncestr", nonce), m.Option("timestamp", kit.Format(time.Now().Unix()))))
}

func _wx_parse(m *ice.Message) {
	data := struct {
		FromUserName string
		ToUserName   string
		CreateTime   int64
		MsgId        int64
		Event        string
		MsgType      string
		Content      string
	}{}
	xml.NewDecoder(m.R.Body).Decode(&data)
	m.Debug("data: %#v", data)

	m.Option("FromUserName", data.FromUserName)
	m.Option("ToUserName", data.ToUserName)
	m.Option("CreateTime", data.CreateTime)
	m.Option("MsgId", data.MsgId)

	m.Option("Event", data.Event)
	m.Option("MsgType", data.MsgType)
	m.Option("Content", data.Content)
}
func _wx_reply(m *ice.Message, tmpl string) {
	m.Render(m.Conf(LOGIN, kit.Keym(kit.MDB_TEMPLATE, tmpl)))
}
func _wx_action(m *ice.Message) {
	m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
	m.Set(ice.MSG_RESULT)

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
	LOGIN   = "login"
	APPID   = "appid"
	APPMM   = "appmm"
	TOKEN   = "token"
	EXPIRE  = "expire"
	TICKET  = "ticket"
	EXPIRES = "expires"
	CONFIG  = "config"
	WEIXIN  = "weixin"
)
const (
	ACCESS = "access"
	MENU   = "menu"
)
const WX = "wx"

var Index = &ice.Context{Name: WX, Help: "公众号",
	Configs: map[string]*ice.Config{
		LOGIN: {Name: LOGIN, Help: "认证", Value: kit.Data(
			WEIXIN, "https://api.weixin.qq.com", APPID, "", APPMM, "", TOKEN, "",
			kit.MDB_TEMPLATE, kit.Dict(kit.MDB_TEXT, text), MENU, []interface{}{
				kit.Dict(wiki.TITLE, "网站主页", wiki.SPARK, "点击进入", wiki.REFER, "https://shylinux.com",
					wiki.IMAGE, "https://shylinux.com/share/local/usr/publish/3f265cd2455053b68976bc63bdd432d4.jpeg",
				),
				kit.Dict(wiki.TITLE, "产品工具", wiki.SPARK, "点击进入", wiki.REFER, "https://shylinux.com?river=product",
					wiki.IMAGE, "https://shylinux.com/share/local/usr/publish/3f265cd2455053b68976bc63bdd432d4.jpeg",
				),
				kit.Dict(wiki.TITLE, "研发工具", wiki.SPARK, "点击进入", wiki.REFER, "https://shylinux.com?river=project",
					wiki.IMAGE, "https://shylinux.com/share/local/usr/publish/3f265cd2455053b68976bc63bdd432d4.jpeg",
				),
				kit.Dict(wiki.TITLE, "测试工具", wiki.SPARK, "点击进入", wiki.REFER, "https://shylinux.com?river=profile",
					wiki.IMAGE, "https://shylinux.com/share/local/usr/publish/3f265cd2455053b68976bc63bdd432d4.jpeg",
				),
			},
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(web.SPIDE, mdb.CREATE, WEIXIN, m.Conf(LOGIN, kit.Keym(WEIXIN)))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},

		ACCESS: {Name: "access appid auto ticket token login", Help: "认证", Action: map[string]*ice.Action{
			LOGIN: {Name: "login appid appmm token", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(LOGIN, kit.Keym(APPID), m.Option(APPID))
				m.Conf(LOGIN, kit.Keym(APPMM), m.Option(APPMM))
				m.Conf(LOGIN, kit.Keym(TOKEN), m.Option(TOKEN))
			}},
			TOKEN: {Name: "token", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Conf(LOGIN, kit.Keym(ACCESS, TOKEN)) == "" || now > kit.Int64(m.Conf(LOGIN, kit.Keym(ACCESS, EXPIRE))) {
					msg := m.Cmd(web.SPIDE, WEIXIN, web.SPIDE_GET, "/cgi-bin/token?grant_type=client_credential",
						APPID, m.Conf(LOGIN, kit.Keym(APPID)), "secret", m.Conf(LOGIN, kit.Keym(APPMM)))
					if m.Warn(msg.Append("errcode") != "", "%v: %v", msg.Append("errcode"), msg.Append("errmsg")) {
						return
					}

					m.Conf(LOGIN, kit.Keym(ACCESS, EXPIRE), now+kit.Int64(msg.Append("expires_in")))
					m.Conf(LOGIN, kit.Keym(ACCESS, TOKEN), msg.Append("access_token"))
				}
				m.Echo(m.Conf(LOGIN, kit.Keym(ACCESS, TOKEN)))
			}},
			TICKET: {Name: "ticket", Help: "票据", Hand: func(m *ice.Message, arg ...string) {
				if now := time.Now().Unix(); m.Conf(LOGIN, kit.Keym(ACCESS, TICKET)) == "" || now > kit.Int64(m.Conf(LOGIN, kit.Keym(ACCESS, EXPIRES))) {
					msg := m.Cmd(web.SPIDE, WEIXIN, web.SPIDE_GET, "/cgi-bin/ticket/getticket?type=jsapi",
						"access_token", m.Cmdx(ACCESS, TOKEN))
					if m.Warn(msg.Append("errcode") != "0", msg.Append("errcode"), msg.Append("errmsg")) {
						return
					}

					m.Conf(LOGIN, kit.Keym(ACCESS, EXPIRES), now+kit.Int64(msg.Append("expires_in")))
					m.Conf(LOGIN, kit.Keym(ACCESS, TICKET), msg.Append(TICKET))
				}
				m.Echo(m.Conf(LOGIN, kit.Keym(ACCESS, TICKET)))
			}},
			CONFIG: {Name: "config", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				_wx_config(m, "some")
			}},

			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(m.Conf(LOGIN, kit.Keym(APPID)))
		}},

		MENU: {Name: "menu name auto", Help: "菜单", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				share := m.Cmdx(web.SHARE, mdb.CREATE, kit.MDB_TYPE, web.LOGIN)
				kit.Fetch(m.Confv(LOGIN, kit.Keym(MENU)), func(index int, value map[string]interface{}) {
					m.Push("", value, kit.Split("title,spark,image"))
					m.Push(wiki.REFER, kit.MergeURL(kit.Format(value[wiki.REFER]), web.SHARE, share))
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			kit.Fetch(m.Confv(LOGIN, kit.Keym(MENU)), func(index int, value map[string]interface{}) {
				m.Push("", value, kit.Split("title,spark,refer,image"))
			})
		}},

		"/login/": {Name: "/login/", Help: "认证", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			check := kit.Sort([]string{m.Conf(LOGIN, kit.Keym(TOKEN)), m.Option("timestamp"), m.Option("nonce")})
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
			case kit.MDB_EVENT:
				switch m.Option("Event") {
				case "subscribe": // 关注事件
					_wx_action(m.Cmdy(MENU, mdb.CREATE))
				case "unsubscribe": // 取关事件
				}

			case kit.MDB_TEXT:
				if cmds := kit.Split(m.Option("Content")); m.Warn(!m.Right(cmds), ice.ErrNotRight) {
					_wx_action(m.Cmdy(MENU, mdb.CREATE))
					break // 没有权限
				} else {
					switch cmds[0] {
					case MENU:
						// 应用列表
						_wx_action(m.Cmdy(MENU, mdb.CREATE))

					default:
						// 执行命令
						if m.Cmdy(cmds); len(m.Appendv(ice.MSG_APPEND)) == 0 && len(m.Resultv()) == 0 {
							m.Cmdy(cli.SYSTEM, cmds)
						} else if len(m.Resultv()) == 0 {
							m.Table()
						}

						// 返回结果
						_wx_reply(m, kit.MDB_TEXT)
					}
				}
			}
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }

var text = `<xml>
<FromUserName><![CDATA[{{.Option "ToUserName"}}]]></FromUserName>
<ToUserName><![CDATA[{{.Option "FromUserName"}}]]></ToUserName>
<CreateTime>{{.Option "CreateTime"}}</CreateTime>
<MsgType><![CDATA[text]]></MsgType>
<Content><![CDATA[{{.Result}}]]></Content>
</xml>`
