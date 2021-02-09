package lark

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"

	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func _lark_get(m *ice.Message, bot string, arg ...interface{}) *ice.Message {
	m.Option(web.SPIDE_HEADER, "Authorization", "Bearer "+m.Cmdx(APP, TOKEN, bot), web.ContentType, web.ContentJSON)
	return m.Cmd(web.SPIDE, LARK, http.MethodGet, arg)
}
func _lark_post(m *ice.Message, bot string, arg ...interface{}) *ice.Message {
	m.Option(web.SPIDE_HEADER, "Authorization", "Bearer "+m.Cmdx(APP, TOKEN, bot), web.ContentType, web.ContentJSON)
	return m.Cmd(web.SPIDE, LARK, arg)
}
func _lark_parse(m *ice.Message) {
	data := m.Optionv(ice.MSG_USERDATA)
	if data == nil {
		json.NewDecoder(m.R.Body).Decode(&data)
		m.Optionv(ice.MSG_USERDATA, data)

		switch d := data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				switch d := v.(type) {
				case map[string]interface{}:
					for k, v := range d {
						m.Add(ice.MSG_OPTION, k, kit.Format(v))
					}
				default:
					for _, v := range kit.Simple(v) {
						m.Add(ice.MSG_OPTION, kit.Keys("msg", k), kit.Format(v))
					}
				}
			}
		}
	}
	m.Debug("msg: %v", kit.Format(data))
}

const (
	P2P_CHAT_CREATE = "p2p_chat_create"
	ADD_BOT         = "add_bot"
)
const (
	SHIP_ID      = "ship_id"
	OPEN_ID      = "open_id"
	CHAT_ID      = "chat_id"
	OPEN_CHAT_ID = "open_chat_id"
	USER_OPEN_ID = "user_open_id"
)
const (
	LOGIN  = "login"
	APPID  = "appid"
	APPMM  = "appmm"
	TOKEN  = "token"
	EXPIRE = "expire"
)
const (
	APP      = "app"
	COMPANY  = "company"
	EMPLOYEE = "employee"
	GROUP    = "group"

	SEND = "send"
	DUTY = "duty"
	HOME = "home"
	FORM = "form"
	TALK = "talk"
	RAND = "rand"
)

const LARK = "lark"

var Index = &ice.Context{Name: LARK, Help: "机器人",
	Configs: map[string]*ice.Config{
		APP: {Name: APP, Help: "服务配置", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
			LARK, "https://open.feishu.cn", DUTY, "", kit.MDB_TEMPLATE, kit.Dict(
				ADD_BOT, "我来也~", P2P_CHAT_CREATE, "让我们做好朋友吧~",
			),
		)},
		COMPANY:  {Name: COMPANY, Help: "组织配置", Value: kit.Data(kit.MDB_SHORT, SHIP_ID)},
		EMPLOYEE: {Name: EMPLOYEE, Help: "员工配置", Value: kit.Data(kit.MDB_SHORT, OPEN_ID)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(web.SPIDE, mdb.CREATE, LARK, m.Conf(APP, kit.Keym(LARK)))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},

		APP: {Name: "app name auto token login", Help: "应用", Action: map[string]*ice.Action{
			LOGIN: {Name: "login name appid appmm", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(APP), "", mdb.HASH, arg)
			}},
			TOKEN: {Name: "token name", Help: "令牌", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.FIELDS, "time,appid,appmm,token,expire")
				msg := m.Cmd(mdb.SELECT, m.Prefix(APP), "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
				if now := time.Now().Unix(); msg.Append(TOKEN) == "" || now > kit.Int64(msg.Append(EXPIRE)) {
					sub := m.Cmd(web.SPIDE, LARK, "/open-apis/auth/v3/tenant_access_token/internal/",
						"app_id", msg.Append(APPID), "app_secret", msg.Append(APPMM))

					m.Cmd(mdb.MODIFY, m.Prefix(APP), "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME),
						TOKEN, msg.Append(TOKEN, sub.Append("tenant_access_token")), EXPIRE, now+kit.Int64(sub.Append(EXPIRE)))
				}
				m.Echo(msg.Append(TOKEN))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option(mdb.FIELDS, kit.Select("time,name,appid,token,expire", mdb.DETAIL, len(arg) > 0))
			m.Cmdy(mdb.SELECT, m.Prefix(APP), "", mdb.HASH, kit.MDB_NAME, arg)
		}},
		COMPANY: {Name: "company ship_id open_id text auto", Help: "组织", Action: map[string]*ice.Action{
			"info": {Name: "info ship_id", Hand: func(m *ice.Message, arg ...string) {
				msg := _lark_get(m, "bot", "/open-apis/contact/v1/department/detail/batch_get", "department_ids", m.Option(SHIP_ID))
				kit.Fetch(kit.Value(msg.Optionv("content_data"), "data.department_infos"), func(index int, value map[string]interface{}) {
					m.Push("", value)
				})
			}},
			"list": {Name: "list ship_id", Hand: func(m *ice.Message, arg ...string) {
				msg := _lark_get(m, "bot", "/open-apis/contact/v1/department/user/list",
					"department_id", m.Option(SHIP_ID), "page_size", "100", "fetch_child", "true")

				kit.Fetch(kit.Value(msg.Optionv("content_data"), "data.user_list"), func(index int, value map[string]interface{}) {
					msg := m.Cmd(EMPLOYEE, value[OPEN_ID])
					m.PushImages(aaa.AVATAR, msg.Append("avatar_72"))
					m.Push(aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"))
					m.Push(kit.MDB_NAME, msg.Append(kit.MDB_NAME))
					m.Push(kit.MDB_TEXT, msg.Append("description"))
					m.Push(OPEN_ID, msg.Append(OPEN_ID))
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 { // 组织列表
				msg := _lark_get(m, "bot", "/open-apis/contact/v1/scope/get/")
				kit.Fetch(kit.Value(msg.Optionv("content_data"), "data.authed_departments"), func(index int, value string) {
					m.Push(SHIP_ID, value)
					msg := m.Cmd(COMPANY, "info", value)
					m.Push(kit.MDB_NAME, msg.Append(kit.MDB_NAME))
					m.Push(kit.MDB_COUNT, msg.Append("member_count"))
					m.Push(CHAT_ID, msg.Append(CHAT_ID))
				})
				m.Sort(kit.MDB_NAME)

			} else if len(arg) == 1 { // 员工列表
				m.Cmdy(COMPANY, "list", arg[0])

			} else if len(arg) == 2 { // 员工详情
				m.Cmdy(EMPLOYEE, arg[1])

			} else { // 员工通知
				m.Cmdy(SEND, OPEN_ID, arg[1], arg[2:])
			}
		}},
		EMPLOYEE: {Name: "employee open_id|mobile|email auto", Help: "员工", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				return
			}
			if strings.HasPrefix(arg[0], "ou_") {
				msg := _lark_get(m, "bot", "/open-apis/contact/v1/user/batch_get", "open_ids", arg[0])
				kit.Fetch(kit.Value(msg.Optionv("content_data"), "data.user_infos"), func(index int, value map[string]interface{}) {
					m.Push(mdb.DETAIL, value)
				})
				return
			}

			us := []string{}
			for i := 0; i < len(arg); i++ {
				us = append(us, kit.Select("mobiles", "emails", strings.Contains(arg[i], "@")), arg[i])
			}

			_lark_get(m, "bot", "/open-apis/user/v1/batch_get_id", us)
			for i := 0; i < len(arg); i++ {
				m.Echo(m.Append(kit.Keys("data.mobile_users", arg[i], "0.open_id")))
			}
		}},
		GROUP: {Name: "group chat_id open_id text auto", Help: "群组", Action: map[string]*ice.Action{
			"list": {Name: "list chat_id", Hand: func(m *ice.Message, arg ...string) {
				msg := _lark_get(m, "bot", "/open-apis/chat/v4/info", "chat_id", m.Option(CHAT_ID))
				kit.Fetch(kit.Value(msg.Optionv("content_data"), "data.members"), func(index int, value map[string]interface{}) {
					msg := m.Cmd(EMPLOYEE, value[OPEN_ID])
					m.PushImages(aaa.AVATAR, msg.Append("avatar_72"))
					m.Push(aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"))
					m.Push(kit.MDB_NAME, msg.Append(kit.MDB_NAME))
					m.Push(kit.MDB_TEXT, msg.Append("description"))
					m.Push(OPEN_ID, msg.Append(OPEN_ID))
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 { // 群组列表
				msg := _lark_get(m, "bot", "/open-apis/chat/v4/list")
				kit.Fetch(kit.Value(msg.Optionv("content_data"), "data.groups"), func(index int, value map[string]interface{}) {
					m.Push(CHAT_ID, value[CHAT_ID])
					m.PushImages(aaa.AVATAR, kit.Format(value[aaa.AVATAR]), "72")
					m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
					m.Push(kit.MDB_TEXT, value["description"])
					m.Push(OPEN_ID, value["owner_open_id"])
				})
				m.Sort(kit.MDB_NAME)

			} else if len(arg) == 1 { // 组员列表
				m.Cmdy(GROUP, "list", arg[0])

			} else if len(arg) == 2 { // 组员详情
				m.Cmdy(EMPLOYEE, arg[1])

			} else { // 组员通知
				m.Cmdy(SEND, CHAT_ID, arg[0], arg[2:])
			}
		}},

		SEND: {Name: "send [chat_id|open_id|user_id|email] target [title] text", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			var form = kit.Dict("content", kit.Dict())
			switch arg[0] {
			case CHAT_ID, OPEN_ID, "user_id", "email":
				form[arg[0]], arg = arg[1], arg[2:]
			default:
				form[CHAT_ID], arg = arg[0], arg[1:]
			}

			switch len(arg) {
			case 0:
			case 1:
				kit.Value(form, "msg_type", "text")
				kit.Value(form, "content.text", arg[0])
				if strings.TrimSpace(arg[0]) == "" {
					return
				}
			default:
				if len(arg) == 2 && strings.TrimSpace(arg[1]) == "" {
					return
				}
				content := []interface{}{}
				line := []interface{}{}
				for _, v := range arg[1:] {
					if v == "\n" {
						content, line = append(content, line), []interface{}{}
						continue
					}
					line = append(line, map[string]interface{}{"tag": "text", "text": v + " "})
				}
				content = append(content, line)

				kit.Value(form, "msg_type", "post")
				kit.Value(form, "content.post", map[string]interface{}{
					"zh_cn": map[string]interface{}{"title": arg[0], "content": content},
				})
			}

			m.Copy(_lark_post(m, "bot", "/open-apis/message/v4/send/", web.SPIDE_DATA, kit.Format(form)))
		}},
		DUTY: {Name: "duty [title] text", Help: "通告", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmdy(SEND, m.Conf(APP, kit.Keym(DUTY)), arg)
		}},
		HOME: {Name: "home river storm title", Help: "首页", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			name := kit.Select(m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERNICK))
			if len(name) > 10 {
				name = name[:10]
			}
			name += "的" + kit.Select("应用列表", arg, 2)
			text := ""

			link, list := m.Conf(web.SHARE, "meta.domain"), []string{}
			if len(arg) == 0 {
				m.Cmd("web.chat./river").Table(func(index int, val map[string]string, head []string) {
					m.Cmd("web.chat./river", val[kit.MDB_HASH], chat.TOOL).Table(func(index int, value map[string]string, head []string) {
						list = append(list, kit.Keys(val[kit.MDB_NAME], value[kit.MDB_NAME]),
							kit.SSH_CMD, kit.Format([]string{"home", val[kit.MDB_HASH], value[kit.MDB_HASH], val[kit.MDB_NAME] + "." + value[kit.MDB_NAME]}))
					})
				})
			} else {
				m.Option(ice.MSG_RIVER, arg[0])
				m.Option(ice.MSG_STORM, arg[1])
				link = kit.MergeURL(link, chat.RIVER, arg[0], chat.STORM, arg[1])
				m.Cmd("web.chat./river", arg[0], chat.TOOL, arg[1]).Table(func(index int, value map[string]string, head []string) {
					list = append(list, value[kit.SSH_CMD], kit.SSH_CMD, kit.Keys(value[kit.SSH_CTX], value[kit.SSH_CMD]))
				})
			}
			m.Cmd(FORM, CHAT_ID, m.Option(OPEN_CHAT_ID), name, text, "打开网页", "url", link, list)
		}},
		FORM: {Name: "form chat_id|open_id|user_id|email target title text [confirm|value|url arg...]...", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			var form = map[string]interface{}{"content": map[string]interface{}{}}
			switch arg[0] {
			case CHAT_ID, OPEN_ID, "user_id", "email":
				form[arg[0]], arg = arg[1], arg[2:]
			default:
				form[CHAT_ID], arg = arg[0], arg[1:]
			}

			elements := []interface{}{}
			elements = append(elements, map[string]interface{}{
				"tag": "div", "text": map[string]interface{}{
					"tag": "plain_text", "content": arg[1],
				},
			})

			actions := []interface{}{}
			for i := 2; i < len(arg); i++ {
				button := map[string]interface{}{
					"type": "default", "tag": "button", "text": map[string]interface{}{
						"tag": "plain_text", "content": arg[i],
					},
				}

				content := arg[i]
				switch arg[i+1] {
				case "confirm":
					button[arg[i+1]], i = map[string]interface{}{
						"title": map[string]interface{}{"tag": "lark_md", "content": arg[i+2]},
						"text":  map[string]interface{}{"tag": "lark_md", "content": arg[i+3]},
					}, i+3
				case "value":
					button[arg[i+1]], i = map[string]interface{}{arg[i+2]: arg[i+3]}, i+3
				case "url":
					button[arg[i+1]], i = arg[i+2], i+2
				default:
					button["value"], i = map[string]interface{}{
						arg[i+1]:      arg[i+2],
						ice.MSG_RIVER: m.Option(ice.MSG_RIVER),
						ice.MSG_STORM: m.Option(ice.MSG_STORM),
					}, i+2
				}
				kit.Value(button, "value.content", content)
				kit.Value(button, "value.open_chat_id", m.Option(OPEN_CHAT_ID))
				kit.Value(button, "value.description", arg[1])
				kit.Value(button, "value.title", arg[0])

				actions = append(actions, button)
			}
			elements = append(elements, map[string]interface{}{"tag": "action", "actions": actions})

			kit.Value(form, "msg_type", "interactive")
			kit.Value(form, "card", map[string]interface{}{
				"config": map[string]interface{}{"wide_screen_mode": true},
				"header": map[string]interface{}{
					"title": map[string]interface{}{"tag": "lark_md", "content": arg[0]},
				},
				"elements": elements,
			})

			_lark_post(m, "bot", "/open-apis/message/v4/send/", web.SPIDE_DATA, kit.Formats(form))
		}},
		TALK: {Name: "talk text", Help: "聊天", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option(ice.MSG_USERZONE, LARK)

			cmds := kit.Split(strings.Join(arg, " "))
			if aaa.UserLogin(m, m.Option(OPEN_ID), ""); !m.Right(cmds) {
				if aaa.UserLogin(m, m.Option(OPEN_CHAT_ID), ""); !m.Right(cmds) {
					m.Cmd(DUTY, m.Option(OPEN_CHAT_ID), m.Option("text_without_at_bot"))
					m.Cmd(HOME)
					return // 没有权限
				}
			}

			if cmds[0] == HOME {
				m.Cmd(HOME, cmds[1:])
				return // 没有命令
			}

			// 执行命令
			if msg := m.Cmd(cmds); len(msg.Appendv(ice.MSG_APPEND)) > 0 || len(msg.Resultv()) > 0 {
				if m.Copy(msg); len(m.Resultv()) == 0 {
					m.Table()
				}
			} else {
				m.Cmdy(cli.SYSTEM, cmds)
			}
		}},
		RAND: {Name: "rand", Help: "随机", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			msg := m.Cmd(GROUP, EMPLOYEE, m.Option(OPEN_CHAT_ID))
			list := msg.Appendv("name")
			if strings.Contains(m.Option("content"), "誰") {
				m.Echo(strings.Replace(m.Option("content"), "誰", list[rand.Intn(len(list))], 1))
				return
			}
			m.Echo(list[rand.Intn(len(list))])
		}},

		web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/msg": {Name: "/msg", Help: "聊天消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			data := m.Optionv(ice.MSG_USERDATA)
			if kit.Value(data, "action") != nil {
				m.Option(ice.MSG_USERUA, "")
				kit.Fetch(kit.Value(data, "action.value"), func(key string, value string) { m.Option(key, value) })

				m.Cmdy(TALK, kit.Parse(nil, "", kit.Split(m.Option(kit.SSH_CMD))...))
				m.Cmd(SEND, CHAT_ID, m.Option(OPEN_CHAT_ID), m.Option(wiki.TITLE)+" "+m.Option(kit.SSH_CMD), m.Result())
				return
			}

			switch _lark_parse(m); m.Option("msg.type") {
			case "url_verification": // 绑定验证
				m.Render(ice.RENDER_RESULT, kit.Format(kit.Dict("challenge", m.Option("msg.challenge"))))

			case "event_callback":
				switch m.Option("type") {
				case "message_read":
				case "chat_disband":
				case P2P_CHAT_CREATE, ADD_BOT:
					// 创建对话
					if m.Options(OPEN_CHAT_ID) {
						m.Cmdy(SEND, m.Option(OPEN_CHAT_ID), m.Conf(APP, kit.Keym(kit.MDB_TEMPLATE, m.Option("type"))))
					}
				default:
					switch m.Option("msg_type") {
					case "location":
					case "image":
						// m.Rich(META, nil, kit.Dict(
						// 	"url", m.Option("image_url"),
						// 	"width", m.Option("image_width"),
						// 	"height", m.Option("image_height"),
						// ))
					default:
						if m.Options(OPEN_CHAT_ID) {
							if m.Cmdy(TALK, strings.TrimSpace(m.Option("text_without_at_bot"))); len(m.Resultv()) > 0 {
								m.Cmd(SEND, m.Option(OPEN_CHAT_ID), m.Result())
							}
						} else {
							m.Cmd(DUTY, m.Option("type"), kit.Formats(data))
						}
					}
				}
			default:
				m.Cmd(DUTY, m.Option("msg.type"), kit.Formats(data))
			}
		}},
		"/sso": {Name: "/sso", Help: "网页", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if m.Options("code") {
				msg := m.Cmd(web.SPIDE, LARK, "/open-apis/authen/v1/access_token", "grant_type", "authorization_code",
					"code", m.Option("code"), "app_access_token", m.Cmdx(APP, "token", "bot"))

				m.Option(aaa.USERZONE, LARK)
				user := msg.Append("data.open_id")
				web.RenderCookie(m, aaa.SessCreate(m, user, aaa.UserRole(m, user)))
				m.Render("redirect", m.Conf(web.SHARE, "meta.domain"))

				msg = m.Cmd(EMPLOYEE, m.Option(aaa.USERNAME, user))
				m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERZONE, LARK, aaa.USERNICK, msg.Append(kit.MDB_NAME),
					aaa.AVATAR, msg.Append("avatar_url"), aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"),
					aaa.COUNTRY, msg.Append(aaa.COUNTRY), aaa.CITY, msg.Append(aaa.CITY),
					aaa.MOBILE, msg.Append(aaa.MOBILE),
				)
				return
			}

			m.Option(mdb.FIELDS, "time,appid,appmm,token,expire")
			m.Cmd(mdb.SELECT, m.Prefix(APP), "", mdb.HASH, kit.MDB_NAME, "bot").Table(func(index int, value map[string]string, head []string) {
				m.Render("redirect", kit.MergeURL2(m.Conf(APP, kit.Keym(LARK)), "/open-apis/authen/v1/index"),
					"app_id", value[APPID], "redirect_uri", kit.MergeURL2(m.Conf(web.SHARE, kit.Keym(kit.MDB_DOMAIN)), "/chat/lark/sso"),
				)
			})
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
