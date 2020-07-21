package lark

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"

	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func post(m *ice.Message, bot string, arg ...interface{}) {
	m.Richs(APP, nil, bot, func(key string, value map[string]interface{}) {
		m.Option("header", "Authorization", "Bearer "+m.Cmdx(APP, "token", bot), "Content-Type", "application/json")
		m.Cmdy(web.SPIDE, LARK, arg)
	})
}
func parse(m *ice.Message) {
	data := m.Optionv("content_data")
	if data == nil {
		json.NewDecoder(m.R.Body).Decode(&data)
		m.Optionv("content_data", data)

		switch d := data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				switch d := v.(type) {
				case map[string]interface{}:
					for k, v := range d {
						m.Add("option", k, kit.Format(v))
					}
				default:
					for _, v := range kit.Simple(v) {
						m.Add("option", "msg."+k, kit.Format(v))
					}
				}
			}
		}
	}
	if kit.Fetch(kit.Value(data, "action.value"), func(key string, value string) {
		m.Add("option", key, value)
	}) != nil {
		m.Option("msg.type", "event_click")
	}
	m.Info("msg: %v", kit.Formats(data))
}

const (
	P2P_CHAT_CREATE = "p2p_chat_create"
	ADD_BOT         = "add_bot"
)
const (
	OPEN_CHAT_ID = "open_chat_id"
	USER_OPEN_ID = "user_open_id"
)
const (
	APP   = "app"
	SHIP  = "ship"
	USER  = "user"
	GROUP = "group"
	SEND  = "send"
	FORM  = "form"
	DUTY  = "duty"
	TALK  = "talk"

	RAND = "rand"

	LARK = "lark"
)

var Index = &ice.Context{Name: "lark", Help: "机器人",
	Configs: map[string]*ice.Config{
		APP: &ice.Config{Name: APP, Help: "服务配置", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
			LARK, "https://open.feishu.cn", DUTY, "", "welcome", kit.Dict(
				P2P_CHAT_CREATE, "让我们做好朋友吧~", ADD_BOT, "我来也~",
			),
		)},
		USER: &ice.Config{Name: USER, Help: "用户配置", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(web.SPIDE, mdb.CREATE, LARK, m.Conf(APP, "meta.lark"))
			m.Cmd(DUTY, "boot", m.Conf(cli.RUNTIME, "boot.hostname"), m.Time())
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(APP, USER)
		}},

		APP: {Name: "app [name] auto", Help: "应用", Action: map[string]*ice.Action{
			"login": {Name: "login name id mm", Hand: func(m *ice.Message, arg ...string) {
				m.Rich(APP, nil, kit.Dict("name", arg[0], "id", arg[1], "mm", arg[2]))
			}},
			"token": {Name: "token name", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(APP, nil, arg[0], func(key string, value map[string]interface{}) {
					if now := time.Now().Unix(); kit.Format(value["token"]) == "" || kit.Int64(value["expire"]) < now {
						m.Cmdy(web.SPIDE, LARK, "/open-apis/auth/v3/tenant_access_token/internal/", "app_id", value["id"], "app_secret", value["mm"])
						value["token"] = m.Append("tenant_access_token")
						value["expire"] = kit.Int64(m.Append("expire")) + now
						m.Set("result")
					}
					m.Echo("%s", value["token"])
				})
			}},
			"watch": {Name: "watch user title text event...", Hand: func(m *ice.Message, arg ...string) {
				for _, v := range arg[2:] {
					m.Watch(v, m.Prefix("send"), arg[0], arg[1], v)
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Richs(APP, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
				if len(arg) == 0 || arg[0] == "" {
					m.Push(key, value, []string{"time", "name", "id", "expire"})
					return
				}
				m.Push("detail", value)
			})
		}},
		SHIP: {Name: "ship ship_id open_id text", Help: "组织", Action: map[string]*ice.Action{
			"user": {Name: "user id", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(APP, nil, "bot", func(key string, value map[string]interface{}) {
					m.Option("header", "Authorization", "Bearer "+m.Cmdx(APP, "token", "bot"), "Content-Type", "application/json")
					data := kit.UnMarshal(m.Cmdx(web.SPIDE, LARK, "raw", http.MethodGet, "/open-apis/contact/v1/department/user/list",
						"department_id", arg[0], "page_size", "100", "fetch_child", "true"))

					m.Debug(kit.Formats(data))
					kit.Fetch(kit.Value(data, "data.user_list"), func(index int, value map[string]interface{}) {
						msg := m.Cmd(m.Prefix("user"), value["open_id"])
						m.Push("name", msg.Append("name"))
						m.Push("open_id", msg.Append("open_id"))
					})
				})
			}},
			"info": {Name: "info id", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(APP, nil, "bot", func(key string, value map[string]interface{}) {
					m.Option("header", "Authorization", "Bearer "+m.Cmdx(APP, "token", "bot"), "Content-Type", "application/json")
					data := kit.UnMarshal(m.Cmdx(web.SPIDE, LARK, "raw", http.MethodGet, "/open-apis/contact/v1/department/detail/batch_get", "department_ids", arg[0]))

					m.Debug(kit.Formats(data))
					kit.Fetch(kit.Value(data, "data.department_infos"), func(index int, value map[string]interface{}) {
						m.Push("name", value)
					})
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				m.Option("header", "Authorization", "Bearer "+m.Cmdx(APP, "token", "bot"), "Content-Type", "application/json")
				data := kit.UnMarshal(m.Cmdx(web.SPIDE, LARK, "raw", http.MethodGet, "/open-apis/contact/v1/scope/get/"))
				kit.Fetch(kit.Value(data, "data.authed_departments"), func(index int, value string) {
					m.Push("ship_id", value)
					msg := m.Cmd(m.Prefix("ship"), "info", value)
					m.Push("name", msg.Append("name"))
					m.Push("chat_id", msg.Append("chat_id"))
				})
				return
			}
			if len(arg) == 1 {
				m.Cmdy(m.Prefix(SHIP), USER, arg[0])
				return
			}

			m.Cmdy(m.Prefix(SEND), "open_id", arg[1], arg[2:])
		}},
		USER: {Name: "user code|email|mobile", Help: "用户", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if !strings.HasPrefix(arg[0], "ou_") {
				us := []string{}
				for i := 0; i < len(arg); i++ {
					us = append(us, kit.Select("mobiles", "emails", strings.Contains(arg[i], "@")), arg[i])
				}
				post(m, "bot", "GET", "/open-apis/user/v1/batch_get_id", us)
				for i := 0; i < len(arg); i++ {
					m.Echo(m.Append(kit.Keys("data.mobile_users", arg[i], "0.open_id")))
				}
				return
			}

			m.Richs(APP, nil, "bot", func(key string, value map[string]interface{}) {
				m.Option("header", "Authorization", "Bearer "+m.Cmdx(APP, "token", "bot"), "Content-Type", "application/json")
				data := kit.UnMarshal(m.Cmdx(web.SPIDE, LARK, "raw", http.MethodGet, "/open-apis/contact/v1/user/batch_get", "open_ids", arg[0]))

				m.Debug(kit.Formats(data))
				kit.Fetch(kit.Value(data, "data.user_infos"), func(index int, value map[string]interface{}) {
					m.Push("name", value["name"])
					m.Push("mobile", value["mobile"])
					m.Push("open_id", value["open_id"])
				})
			})
		}},
		GROUP: {Name: "group chat_id open_id text", Help: "群组", Action: map[string]*ice.Action{
			"user": {Name: "user id", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(APP, nil, "bot", func(key string, value map[string]interface{}) {
					m.Option("header", "Authorization", "Bearer "+m.Cmdx(APP, "token", "bot"), "Content-Type", "application/json")
					data := kit.UnMarshal(m.Cmdx(web.SPIDE, LARK, "raw", http.MethodGet, "/open-apis/chat/v4/info", "chat_id", arg[0]))

					m.Debug(kit.Formats(data))
					kit.Fetch(kit.Value(data, "data.members"), func(index int, value map[string]interface{}) {
						msg := m.Cmd(m.Prefix("user"), value["open_id"])
						m.Push("name", msg.Append("name"))
						m.Push("open_id", msg.Append("open_id"))
					})
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				m.Richs(APP, nil, "bot", func(key string, value map[string]interface{}) {
					m.Option("header", "Authorization", "Bearer "+m.Cmdx(APP, "token", "bot"), "Content-Type", "application/json")
					data := kit.UnMarshal(m.Cmdx(web.SPIDE, LARK, "raw", http.MethodGet, "/open-apis/chat/v4/list"))
					kit.Fetch(kit.Value(data, "data.groups"), func(index int, value map[string]interface{}) {
						m.Push("chat_id", value["chat_id"])
						m.Push("name", value["name"])
						m.Push("description", value["description"])
						m.Push("open_id", value["owner_open_id"])
					})
				})
				return
			}
			if len(arg) == 1 {
				m.Cmdy(m.Prefix(GROUP), USER, arg[0])
				return
			}
			m.Cmdy(m.Prefix(SEND), "chat_id", arg[0], arg[2:])
		}},
		RAND: {Name: "rand", Help: "随机", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			msg := m.Cmd(GROUP, "user", m.Option(OPEN_CHAT_ID))
			list := msg.Appendv("name")
			m.Echo(list[rand.Intn(len(list))])
		}},
		SEND: {Name: "send [chat_id|open_id|user_id|email] user [title] text", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			var form = kit.Dict("content", kit.Dict())

			switch arg[0] {
			case "chat_id", "open_id", "user_id", "email":
				form[arg[0]], arg = arg[1], arg[2:]
			default:
				form["chat_id"], arg = arg[0], arg[1:]
			}

			switch len(arg) {
			case 0:
			case 1:
				kit.Value(form, "msg_type", "text")
				kit.Value(form, "content.text", arg[0])
				if arg[0] == "" {
					return
				}
			default:
				content := []interface{}{}
				line := []interface{}{}
				for _, v := range arg[1:] {
					if v == "\n" {
						content, line = append(content, line), []interface{}{}
						continue
					}
					line = append(line, map[string]interface{}{
						"tag": "text", "text": v + " ",
					})
				}
				content = append(content, line)

				kit.Value(form, "msg_type", "post")
				kit.Value(form, "content.post", map[string]interface{}{
					"zh_cn": map[string]interface{}{
						"title":   arg[0],
						"content": content,
					},
				})
			}

			post(m, "bot", "/open-apis/message/v4/send/", "data", kit.Formats(form))
		}},
		DUTY: {Name: "send [title] text", Help: "通告", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmdy(SEND, m.Conf(APP, "meta.duty"), arg)
		}},
		TALK: {Name: "talk text", Help: "聊天", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			cmd := kit.Split(arg[0])

			// 用户权限
			m.Option(ice.MSG_USERNAME, m.Option("open_id"))
			m.Option(ice.MSG_USERROLE, aaa.UserRole(m, m.Option(ice.MSG_USERNAME)))
			m.Info("%s: %s", m.Option(ice.MSG_USERROLE), m.Option(ice.MSG_USERNAME))

			if !m.Right(cmd) {
				// 群组权限
				m.Option(ice.MSG_USERNAME, m.Option(OPEN_CHAT_ID))
				m.Option(ice.MSG_USERROLE, aaa.UserRole(m, m.Option(ice.MSG_USERNAME)))
				m.Info("%s: %s", m.Option(ice.MSG_USERROLE), m.Option(ice.MSG_USERNAME))

				if !m.Right(cmd) {
					// 没有权限
					m.Cmd(DUTY, m.Option(OPEN_CHAT_ID), m.Option("text_without_at_bot"))
					m.Cmd("home")
					return
				}
			}

			if len(cmd) == 0 {
				m.Cmd("home")
				return
			}
			// 执行命令
			msg := m.Cmd(cmd)
			if m.Hand = false; !msg.Hand {
				msg = m.Cmd(cli.SYSTEM, cmd)
			}
			if m.Hand = true; msg.Result() == "" {
				msg.Table()
			}
			if m.Hand = true; msg.Result() == "" {
				m.Cmd("home")
				return
			}
			m.Echo(msg.Result())
		}},

		web.LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/msg": {Name: "/msg", Help: "聊天消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			data := m.Optionv(ice.MSG_USERDATA)
			if kit.Value(data, "action.tag") == "button" {
				kit.Fetch(kit.Value(data, "action.value"), func(key string, value string) {
					m.Option(key, value)
				})
				cmd := kit.Split(m.Option("cmd"))
				if len(cmd) == 0 {
					return
				}

				msg := m.Cmd(cmd)
				if m.Hand = false; !msg.Hand {
					msg = m.Cmd(cli.SYSTEM, cmd)
				}
				if m.Hand = true; msg.Result() == "" {
					msg.Table()
				}
				m.Cmd(SEND, "chat_id", m.Option(OPEN_CHAT_ID), msg.Result())
				return
			}

			switch parse(m); m.Option("msg.type") {
			case "url_verification":
				// 绑定验证
				m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
				m.Echo(kit.Format(map[string]interface{}{"challenge": m.Option("msg.challenge")}))

			case "event_callback":
				switch m.Option("type") {
				case "message_read":
				case "chat_disband":
				case P2P_CHAT_CREATE, ADD_BOT:
					// 创建对话
					if m.Options(OPEN_CHAT_ID) {
						m.Cmdy(SEND, m.Option(OPEN_CHAT_ID), m.Conf(APP, kit.Dict(kit.MDB_META, "welcome", m.Option("type"))))
					}
				default:
					switch m.Option("msg_type") {
					case "image":
					default:
						if m.Options(OPEN_CHAT_ID) {
							if m.Cmdy(TALK, strings.TrimSpace(m.Option("text_without_at_bot"))); len(m.Resultv()) > 0 {
								m.Cmd(SEND, m.Option(OPEN_CHAT_ID), m.Result())
							}
						}
					}
				}
			case "event_click":
				// 消息卡片
				m.Echo(kit.Format(map[string]interface{}{
					"header": map[string]interface{}{
						"title": map[string]interface{}{
							"tag": "lark_md", "content": "haha",
						},
					},
					"elements": []interface{}{
						map[string]interface{}{
							"tag": "action",
							"actions": []interface{}{
								map[string]interface{}{
									"tag":  "button",
									"type": "default",
									"text": map[string]interface{}{
										"tag":     "plain_text",
										"content": m.Time(),
									},
									"value": map[string]interface{}{
										"hi": "hello",
									},
								},
							},
						},
					},
				}))
			}
		}},

		"home": {Name: "form chat_id|open_id|user_id|email user title [text [confirm|value|url arg...]]...", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmd("form", "chat_id", m.Option(OPEN_CHAT_ID), "home", "list", "pwd", "cmd", "pwd")
		}},
		FORM: {Name: "form chat_id|open_id|user_id|email user title [text [confirm|value|url arg...]]...", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			var form = map[string]interface{}{"content": map[string]interface{}{}}
			switch arg[0] {
			case "chat_id", "open_id", "user_id", "email":
				form[arg[0]], arg = arg[1], arg[2:]
			default:
				form["chat_id"], arg = arg[0], arg[1:]
			}

			elements := []interface{}{}
			elements = append(elements, map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": arg[1],
				},
			})

			actions := []interface{}{}
			for i := 2; i < len(arg); i++ {
				button := map[string]interface{}{
					"tag": "button", "text": map[string]interface{}{
						"tag": "plain_text", "content": arg[i],
					},
					"type": "default",
				}

				m.Debug("fuck %v", arg[i:])
				switch arg[i+1] {
				case "confirm":
					button[arg[i+1]], i = map[string]interface{}{
						"title": map[string]interface{}{"tag": "lark_md", "content": arg[i+2]},
						"text":  map[string]interface{}{"tag": "lark_md", "content": arg[i+3]},
					}, i+3
				case "value":
					button[arg[i+1]], i = map[string]interface{}{
						arg[i+2]: arg[i+3],
					}, i+3
				case "url":
					button[arg[i+1]], i = arg[i+2], i+2
				default:
					button["value"], i = map[string]interface{}{
						arg[i+1]:     arg[i+2],
						OPEN_CHAT_ID: m.Option(OPEN_CHAT_ID),
					}, i+2
				}

				actions = append(actions, button)
			}
			elements = append(elements, map[string]interface{}{
				"tag": "action", "actions": actions,
			})

			kit.Value(form, "msg_type", "interactive")
			kit.Value(form, "card", map[string]interface{}{
				"config": map[string]interface{}{
					"wide_screen_mode": true,
					// "title": map[string]interface{}{
					// 	"tag": "lark_md", "content": arg[1],
					// },
				},
				"header": map[string]interface{}{
					"title": map[string]interface{}{
						"tag": "lark_md", "content": arg[0],
					},
				},
				"elements": elements,
			})

			post(m, "bot", "/open-apis/message/v4/send/", "data", kit.Formats(form))
		}},

		"/sso": {Name: "/sso", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if m.Options("code") {
				m.Richs(APP, nil, "bot", func(key string, value map[string]interface{}) {
					data := kit.UnMarshal(m.Cmdx(web.SPIDE, LARK, "raw", "/open-apis/authen/v1/access_token",
						"code", m.Option("code"), "grant_type", "authorization_code",
						"app_access_token", m.Cmdx(APP, "token", "bot"),
					))

					m.Option(ice.MSG_USERROLE, aaa.ROOT)
					user := kit.Format(kit.Value(data, "data.open_id"))
					web.RenderCookie(m, aaa.SessCreate(m, user, aaa.UserRole(m, user)))
					m.Render("redirect", m.Conf(web.SHARE, "meta.domain"))

					m.Debug("data %v", kit.Format(data))
					m.Cmd(aaa.USER, mdb.MODIFY, user,
						aaa.USERNICK, kit.Value(data, "data.name"),
					)
				})
				return
			}

			m.Richs(APP, nil, "bot", func(key string, value map[string]interface{}) {
				m.Render("redirect", kit.MergeURL2(m.Conf(APP, "meta.lark"), "/open-apis/authen/v1/index"),
					"app_id", value["id"], "redirect_uri", kit.MergeURL2(m.Conf(web.SHARE, "meta.domain"), "/chat/lark/sso"),
				)
			})
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
