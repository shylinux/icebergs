package lark

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"

	"encoding/json"
	"strings"
	"time"
)

const (
	APP  = "app"
	USER = "user"
	DUTY = "duty"
	SEND = "send"
	LARK = "lark"
)

func post(m *ice.Message, bot string, arg ...interface{}) {
	m.Richs("app", nil, bot, func(key string, value map[string]interface{}) {
		m.Option("header", "Authorization", "Bearer "+m.Cmdx("app", "token", bot), "Content-Type", "application/json")
		m.Cmdy(ice.WEB_SPIDE, "lark", arg)
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

var Index = &ice.Context{Name: "lark", Help: "机器人",
	Configs: map[string]*ice.Config{
		APP: &ice.Config{Name: "app", Help: "服务配置", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
			LARK, "https://open.feishu.cn",
		)},
		USER: &ice.Config{Name: "user", Help: "用户配置", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(ice.WEB_SPIDE, "add", LARK, m.Conf(APP, "meta.lark"))
			m.Cmd(DUTY, "boot", m.Conf(ice.CLI_RUNTIME, "boot.hostname"), m.Time())
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(APP, USER)
		}},
		DUTY: {Name: "send [title] text", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmdy("send", m.Conf(APP, "meta.duty"), arg)
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

		ice.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"login": {Name: "login", Help: "应用", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmdy(ice.WEB_SHARE, "user", m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_SESSID))
		}},

		"app": {Name: "app login|token bot", Help: "应用", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				// 应用列表
				m.Richs("app", nil, "*", func(key string, value map[string]interface{}) {
					m.Push("key", value, []string{"time", "name", "id", "expire"})
				})
				return
			}

			switch arg[0] {
			case "login":
				m.Rich("app", nil, kit.Dict("name", arg[1], "id", arg[2], "mm", arg[3]))

			case "token":
				m.Richs("app", nil, arg[1], func(key string, value map[string]interface{}) {
					if now := time.Now().Unix(); kit.Format(value["token"]) == "" || kit.Int64(value["expire"]) < now {
						m.Cmdy(ice.WEB_SPIDE, "lark", "/open-apis/auth/v3/tenant_access_token/internal/", "app_id", value["id"], "app_secret", value["mm"])
						value["token"] = m.Append("tenant_access_token")
						value["expire"] = kit.Int64(m.Append("expire")) + now
						m.Set("result")
					}
					m.Echo("%s", value["token"])
				})

			case "watch":
				// 消息通知
				for _, v := range arg[3:] {
					m.Watch(v, "web.chat.lark.send", arg[1], arg[2], v)
				}
			}
			return
		}},
		"ship": {Name: "ship", Help: "组织", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			data := kit.UnMarshal(m.Cmdx(ice.WEB_SPIDE, "lark", "/open-apis/contact/v1/scope/get/",
				"headers", "Authorization", "Bearer "+m.Cmdx(".app", "token", "bot"),
			))
			kit.Fetch(kit.Value(data, "data.authed_departments"), func(index int, value string) {
				m.Push("type", "ship")
				m.Push("value", value)
			})
			kit.Fetch(kit.Value(data, "data.authed_open_ids"), func(index int, value string) {
				m.Push("type", "user")
				m.Push("value", value)
			})
			m.Table()
			return
		}},
		"group": {Name: "group", Help: "群组", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				kit.Fetch(kit.Value(kit.UnMarshal(m.Cmdx("web.get", "feishu", "chat/v4/list", "headers", "Authorization", "Bearer "+m.Cmdx(".app", "token", "bot"))),
					"data.groups"), func(index int, value map[string]interface{}) {
					m.Push("key", value["chat_id"])
					m.Push("name", value["name"])
					m.Push("avatar", value["avatar"])
				})
				m.Table()
			}
			kit.Fetch(kit.Value(kit.UnMarshal(m.Cmdx("web.get", "feishu", "chat/v4/info", "headers", "Authorization", "Bearer "+m.Cmdx(".app", "token", "bot"))),
				"data.members"), func(index int, value map[string]interface{}) {
				m.Push("key", value["open_id"])
			})
			m.Table()
			return
		}},
		"user": {Name: "user code|email|mobile", Help: "用户", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			switch arg[0] {
			case "code":
				m.Richs("app", nil, "bot", func(key string, value map[string]interface{}) {
					m.Cmd(ice.WEB_SPIDE, "lark", "/connect/qrconnect/oauth2/access_token/",
						"app_secret", value["mm"], "app_id", value["id"],
						"grant_type", "authorization_code", "code", arg[1],
					)

				})

				msg := m.Cmd(ice.WEB_SPIDE, "lark", "/connect/qrconnect/oauth2/user_info/",
					"headers", "Authorization", "Bearer "+m.Append("access_token"),
				)
				m.Confv("user", m.Append("open_id"), map[string]interface{}{
					"name":          m.Append("name"),
					"en_name":       m.Append("en_name"),
					"avatar_url":    m.Append("avatar_url"),
					"access_token":  m.Append("access_token"),
					"token_type":    m.Append("token_type"),
					"expire":        kit.Int64(m.Append("expire")) + time.Now().Unix(),
					"refresh_token": m.Append("refresh_token"),
					"tenant_key":    m.Append("tenant_key"),

					"email":    msg.Append("Email"),
					"mobile":   msg.Append("Mobile"),
					"status":   msg.Append("status"),
					"employee": msg.Append("EmployeeID"),
				})

			default:
				us := []string{}
				ps := []string{}
				for i := 0; i < len(arg); i++ {
					us = append(us, kit.Select("mobiles", "emails", strings.Contains(arg[i], "@")), arg[i])
					ps = append(ps, kit.Select("mobile", "email", strings.Contains(arg[i], "@"))+"_users")
				}

				post(m, "bot", "GET", "/open-apis/user/v1/batch_get_id", us)
			}
		}},
		"menu": {Name: "send chat_id|open_id|user_id|email [menu] [title] text", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			var form = map[string]interface{}{"content": map[string]interface{}{}}
			switch arg[0] {
			case "chat_id", "open_id", "user_id", "email":
				form[arg[0]], arg = arg[1], arg[2:]
			default:
				form["chat_id"], arg = arg[0], arg[1:]
			}

			switch arg[0] {
			case "menu":
				elements := []interface{}{}
				elements = append(elements, map[string]interface{}{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "plain_text",
						"content": arg[2],
					},
				})

				actions := []interface{}{}
				for i := 3; i < len(arg); i++ {
					button := map[string]interface{}{
						"tag": "button", "text": map[string]interface{}{
							"tag": "plain_text", "content": arg[i],
						},
						"type": "default",
					}

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
							arg[i+1]: arg[i+2],
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
							"tag": "lark_md", "content": arg[1],
						},
					},
					"elements": elements,
				})
			default:
				switch len(arg) {
				case 0:
				case 1:
					kit.Value(form, "msg_type", "text")
					kit.Value(form, "content.text", arg[0])
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
			}

			post(m, "bot", "/open-apis/message/v4/send/", "data", kit.Formats(form))
			return
		}},

		"/msg": {Name: "/msg", Help: "聊天消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			switch parse(m); m.Option("msg.type") {
			case "url_verification":
				// 绑定验证
				m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
				m.Echo(kit.Format(map[string]interface{}{"challenge": m.Option("msg.challenge")}))

			case "event_callback":
				switch m.Option("type") {
				case "chat_disband":
				case "p2p_chat_create":
					// 创建对话
					if m.Options("open_chat_id") {
						m.Cmdy("send", m.Option("open_chat_id"), "我们做朋友吧~")
					}
				case "add_bot":
					// 加入群聊
					if m.Options("open_chat_id") {
						m.Cmdy("send", m.Option("open_chat_id"), "我来也~")
					}
				default:
					switch m.Option("msg_type") {
					case "image":
					default:
						if m.Options("open_chat_id") {
							// 用户登录
							m.Option(ice.MSG_USERNAME, m.Option("open_id"))
							m.Option(ice.MSG_USERROLE, m.Cmdx(ice.AAA_ROLE, "check", m.Option(ice.MSG_USERNAME)))
							m.Info("%s: %s", m.Option(ice.MSG_USERROLE), m.Option(ice.MSG_USERNAME))

							if cmd := kit.Split(m.Option("text_without_at_bot")); cmd[0] == "login" || m.Right(cmd) {
								// 执行命令
								msg := m.Cmd(cmd)
								if m.Hand = false; !msg.Hand {
									msg = m.Cmd(ice.CLI_SYSTEM, cmd)
								}
								if msg.Result() == "" {
									msg.Table()
								}
								m.Echo(msg.Result())
							} else {
								m.Cmd("duty", m.Option("open_chat_id"), m.Option("text_without_at_bot"))
							}

							// 返回结果
							m.Cmd("send", m.Option("open_chat_id"), kit.Select("你好", m.Result()))
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
			return
		}},
		"/sso": {Name: "/sso", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if m.Options("code") {
				m.Option("username", m.Cmd(".user", "code", m.Option("code")).Append("open_id"))
				m.Option("sessid", m.Cmdx("aaa.user", "session", "select"))
				m.Cmd("ssh._check", "work", "create", m.Option("username"))

				// web.Cookie(m)
				// m.Append("redirect", m.Cmdx("web.spide", "serve", "merge", m.Option("index_path")), "code", "")
				return
			}

			if !m.Options("sessid") || !m.Options("username") {
				m.Append("redirect", m.Cmdx("web.spide", "feishu", "merge", "/connect/qrconnect/page/sso/",
					"redirect_uri", m.Cmdx("web.spide", "serve", "merge", m.Option("index_path")),
					"app_id", m.Conf("app", "bot.id"), "state", "ok"))
				return
			}
			m.Cmd("/render")
			return
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
