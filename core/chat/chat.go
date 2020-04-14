package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "chat", Help: "聊天中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.CHAT_RIVER: {Name: "river", Help: "群组", Value: kit.Data(
			"template", kit.Dict("root", []interface{}{
				[]interface{}{"river", `{{.Option "user.nick"|Format}}@{{.Conf "runtime" "node.name"|Format}}`, "mall"},

				[]interface{}{"storm", "mall", "mall"},
				[]interface{}{"field", "asset", "web.mall"},
				[]interface{}{"field", "spend", "web.mall"},
				[]interface{}{"field", "trans", "web.mall"},
				[]interface{}{"field", "bonus", "web.mall"},
				[]interface{}{"field", "month", "web.mall"},

				[]interface{}{"storm", "team", "team"},
				[]interface{}{"field", "plan", "web.team"},
				[]interface{}{"field", "miss", "web.team"},
				[]interface{}{"field", "stat", "web.team"},
				[]interface{}{"field", "task", "web.team"},

				[]interface{}{"storm", "wiki", "wiki"},
				[]interface{}{"field", "draw", "web.wiki"},
				[]interface{}{"field", "data", "web.wiki"},
				[]interface{}{"field", "word", "web.wiki"},
				[]interface{}{"field", "walk", "web.wiki"},
				[]interface{}{"field", "feel", "web.wiki"},

				[]interface{}{"storm", "code", "code"},
				[]interface{}{"field", "login", "web.code"},
				[]interface{}{"field", "buffer", "web.code.tmux"},
				[]interface{}{"field", "session", "web.code.tmux"},
				[]interface{}{"field", "image", "web.code.docker"},
				[]interface{}{"field", "container", "web.code.docker"},
				[]interface{}{"field", "command", "web.code.docker"},
				[]interface{}{"field", "repos", "web.code.git"},
				[]interface{}{"field", "total", "web.code.git"},
				[]interface{}{"field", "status", "web.code.git"},

				[]interface{}{"storm", "root"},
				[]interface{}{"field", "spide"},
				[]interface{}{"field", "space"},
				[]interface{}{"field", "dream"},
				[]interface{}{"field", "favor"},
				[]interface{}{"field", "story"},
				[]interface{}{"field", "share"},
			}, "void", []interface{}{
				[]interface{}{"storm", "wiki", "wiki"},
				[]interface{}{"field", "note", "web.wiki"},
			}),
			"black", kit.Dict("void", []interface{}{
				[]interface{}{"/debug"},
				[]interface{}{"/river", "add"},
				[]interface{}{"/river", "share"},
				[]interface{}{"/river", "rename"},
				[]interface{}{"/river", "remove"},
				[]interface{}{"/storm", "remove"},
				[]interface{}{"/storm", "rename"},
				[]interface{}{"/storm", "share"},
				[]interface{}{"/storm", "add"},
			}),
			"white", kit.Dict("void", []interface{}{
				[]interface{}{"/toast"},
				[]interface{}{"/carte"},
				[]interface{}{"/tutor"},
				[]interface{}{"/login"},
				[]interface{}{"/river"},
				[]interface{}{"/storm"},
				[]interface{}{"/action"},
				[]interface{}{"web.wiki.note"},
			}),
			"fe", "volcanos",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Watch(ice.SYSTEM_INIT, m.Prefix("init"))
			m.Watch(ice.USER_CREATE, m.Prefix("auto"))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("river")
		}},

		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(m.Confm(ice.CHAT_RIVER, "hash")) == 0 {
				if m.Richs(ice.WEB_FAVOR, nil, "river.root", nil) == nil {
					// 系统群组
					kit.Fetch(m.Confv(ice.CHAT_RIVER, "meta.template.root"), func(index int, value interface{}) {
						m.Cmd(ice.WEB_FAVOR, "river.root", value)
					})
					// 默认群组
					kit.Fetch(m.Confv(ice.CHAT_RIVER, "meta.template.void"), func(index int, value interface{}) {
						m.Cmd(ice.WEB_FAVOR, "river.void", value)
					})
					// 黑名单
					kit.Fetch(m.Confv(ice.CHAT_RIVER, "meta.black.void"), func(index int, value interface{}) {
						m.Cmd(ice.AAA_ROLE, "black", ice.ROLE_VOID, "enable", value)
					})
					// 白名单
					kit.Fetch(m.Confv(ice.CHAT_RIVER, "meta.white.void"), func(index int, value interface{}) {
						m.Cmd(ice.AAA_ROLE, "white", ice.ROLE_VOID, "enable", value)
					})
				}
				// 超级用户
				m.Cmd(ice.AAA_USER, "first", m.Conf(ice.CLI_RUNTIME, "boot.username"))
			}

			// 前端框架
			m.Cmd("web.code.git.repos", m.Conf(ice.CHAT_RIVER, "meta.fe"))
			m.Cap(ice.CTX_STREAM, m.Conf(ice.CHAT_RIVER, "meta.fe"))
			m.Cap(ice.CTX_STATUS, "start")
		}},
		"auto": {Name: "auto user", Help: "自动化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(ice.AAA_USER, nil, arg[0], func(key string, value map[string]interface{}) {
				m.Option(ice.MSG_USERNICK, value["usernick"])
				m.Option(ice.MSG_USERNAME, value["username"])
				m.Option(ice.MSG_USERROLE, "root")

				// 创建应用
				storm, river := "", ""
				m.Option("cache.limit", -2)
				m.Richs(ice.WEB_FAVOR, nil, kit.Keys("river", m.Cmdx(ice.AAA_ROLE, "check", value["username"])), func(key string, value map[string]interface{}) {
					m.Grows(ice.WEB_FAVOR, kit.Keys("hash", key), "", "", func(index int, value map[string]interface{}) {
						switch value["type"] {
						case "river":
							name, _ := kit.Render(kit.Format(value["name"]), m)
							river = m.Option(ice.MSG_RIVER, m.Cmdx("/ocean", "spawn", string(name)))
						case "storm":
							storm = m.Option(ice.MSG_STORM, m.Cmdx("/steam", river, "spawn", value["name"]))
						case "field":
							m.Cmd("/storm", river, storm, "add", "", kit.Select("", value["text"]), value["name"], "")
						}
					})
				})
			})
		}},

		ice.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(ice.MSG_RIVER, "")
			m.Option(ice.MSG_STORM, "")

			if len(arg) > 0 {
				switch arg[0] {
				case "login":
					// 密码登录
					if len(arg) > 2 {
						m.Render("cookie", m.Option(ice.MSG_SESSID, m.Cmdx(ice.AAA_USER, "login", m.Option(ice.MSG_USERNAME, arg[1]), arg[2])))
					}

				default:
					// 群组检查
					m.Richs(ice.CHAT_RIVER, nil, arg[0], func(key string, value map[string]interface{}) {
						m.Richs(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "user"), m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
							if m.Option(ice.MSG_RIVER, arg[0]); len(arg) > 1 {
								// 应用检查
								m.Richs(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "tool"), arg[1], func(key string, value map[string]interface{}) {
									m.Option(ice.MSG_STORM, arg[1])
								})
							}
							m.Log(ice.LOG_LOGIN, "river: %s storm: %s", m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM))
						})
					})
				}
			}
			if m.Option(ice.MSG_USERURL) == "/login" {
				return
			}

			// 登录检查
			if m.Warn(!m.Options(ice.MSG_USERNAME), "not login") {
				m.Render("status", 401, "not login")
				m.Option(ice.MSG_USERURL, "")
				return
			}

			// 权限检查
			if m.Warn(!m.Right(m.Option(ice.MSG_USERURL), m.Optionv("cmds")), "not auth") {
				m.Render("status", 403, "not auth")
				m.Option(ice.MSG_USERURL, "")
				return
			}
		}},

		"search": {Name: "search label=some word=启动流程 auto", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy(ice.WEB_LABEL, arg)
				return
			}
			m.Cmdy(ice.WEB_LABEL, arg[0], "*", "favor", "search", arg[1:])
		}},
		"commend": {Name: "commend label=some word=请求响应 auto", Help: "推荐引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy(ice.WEB_LABEL, arg)
				return
			}
			m.Cmdy(ice.WEB_LABEL, arg[0], "*", "favor", "search", arg[1:])
		}},

		"/toast": {Name: "/toast", Help: "提示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/carte": {Name: "/carte", Help: "菜单", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/tutor": {Name: "/tutor", Help: "向导", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/debug": {Name: "/debug", Help: "调试", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/input": {Name: "/input", Help: "输入", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/login": {Name: "/login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "check":
				m.Richs(ice.AAA_USER, nil, m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
					m.Push("nickname", value["nickname"])
				})
				m.Echo(m.Option(ice.MSG_USERNAME))

			case "login":
				if len(arg) > 1 {
					m.Cmdy(ice.AAA_USER, "login", arg[1:])
					break
				}
				m.Echo(m.Option(ice.MSG_SESSID))

			case "share":
				switch arg[1] {
				case "river":
				case "storm":
				case "action":
					if m.Option("index") != "" {
						arg = append(arg, "tool.0.pod", m.Option("node"))
						arg = append(arg, "tool.0.ctx", m.Option("group"))
						arg = append(arg, "tool.0.cmd", m.Option("index"))
						arg = append(arg, "tool.0.args", m.Option("args"))
						arg = append(arg, "tool.0.single", "yes")
					} else {
						m.Option(ice.MSG_RIVER, arg[5])
						m.Option(ice.MSG_STORM, arg[7])
						m.Cmd("/action", arg[5], arg[7]).Table(func(index int, value map[string]string, head []string) {
							arg = append(arg, kit.Format("tool.%d.pod", index), value["node"])
							arg = append(arg, kit.Format("tool.%d.ctx", index), value["group"])
							arg = append(arg, kit.Format("tool.%d.cmd", index), value["index"])
							arg = append(arg, kit.Format("tool.%d.args", index), value["args"])
						})
					}
				default:
					return
				}
				m.Cmdy(ice.WEB_SHARE, "add", arg[1], arg[2], arg[3], arg[4:])
			}
		}},

		"/ocean": {Name: "/ocean", Help: "大海洋", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 用户列表
				m.Richs(ice.AAA_USER, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"username", "usernode"})
				})
				return
			}

			switch arg[0] {
			case "spawn":
				// 创建群组
				river := m.Rich(ice.CHAT_RIVER, nil, kit.Dict(
					kit.MDB_META, kit.Dict(kit.MDB_NAME, arg[1]),
					"user", kit.Data(kit.MDB_SHORT, "username"),
					"tool", kit.Data(),
				))
				m.Log(ice.LOG_CREATE, "river: %v name: %v", river, arg[1])
				// 添加用户
				m.Cmd("/river", river, "add", m.Option(ice.MSG_USERNAME), arg[2:])
				m.Echo(river)
			}
		}},
		"/river": {Name: "/river", Help: "小河流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				// 群组列表
				m.Richs(ice.CHAT_RIVER, nil, "*", func(key string, value map[string]interface{}) {
					m.Richs(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, key, "user"), m.Option(ice.MSG_USERNAME), func(k string, val map[string]interface{}) {
						m.Push(key, value["meta"], []string{kit.MDB_KEY, kit.MDB_NAME})
					})
				})
				m.Log(ice.LOG_SELECT, "%s", m.Format("append"))
				return
			}

			if !m.Right(cmd, arg[1]) {
				m.Render("status", 403, "not auth")
				return
			}

			switch arg[1] {
			case "add":
				m.Rich(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "user"), kit.Data("username", m.Conf(ice.CLI_RUNTIME, "boot.username")))
				// 添加用户
				for _, v := range arg[2:] {
					user := m.Rich(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "user"), kit.Data("username", v))
					m.Log(ice.LOG_INSERT, "river: %s user: %s name: %s", arg[0], user, v)
				}
			case "rename":
				// 重命名群组
				old := m.Conf(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], kit.MDB_META, kit.MDB_NAME))
				m.Log(ice.LOG_MODIFY, "river: %s %s->%s", arg[0], old, arg[2])
				m.Conf(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], kit.MDB_META, kit.MDB_NAME), arg[2])

			case "remove":
				// 删除群组
				m.Richs(ice.CHAT_RIVER, nil, arg[0], func(value map[string]interface{}) {
					m.Log(ice.LOG_REMOVE, "river: %s value: %s", arg[0], kit.Format(value))
				})
				m.Conf(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0]), "")
			}
		}},
		"/storm": {Name: "/storm", Help: "暴风雨", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Warn(m.Option(ice.MSG_RIVER) == "", "not join") {
				m.Render("status", 402, "not join")
				return
			}

			prefix := kit.Keys(kit.MDB_HASH, arg[0], "tool")
			if len(arg) < 3 {
				// 应用列表
				m.Richs(ice.CHAT_RIVER, prefix, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"], []string{kit.MDB_KEY, kit.MDB_NAME})
				})
				m.Log(ice.LOG_SELECT, "%s", m.Format("append"))
				m.Sort(kit.MDB_NAME)
				return
			}

			if !m.Right(cmd, arg[2]) {
				m.Render("status", 403, "not auth")
				return
			}

			switch arg[2] {
			case "add":
				// 添加命令
				for i := 3; i < len(arg)-3; i += 4 {
					id := m.Grow(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1]), kit.Data(
						"pod", arg[i], "ctx", arg[i+1], "cmd", arg[i+2], "help", arg[i+3],
					))
					m.Log(ice.LOG_INSERT, "storm: %s %d: %v", arg[1], id, arg[i:i+4])
				}

			case "rename":
				// 重命名应用
				old := m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1], kit.MDB_META, kit.MDB_NAME))
				m.Log(ice.LOG_MODIFY, "storm: %s %s->%s", arg[1], old, arg[3])
				m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1], kit.MDB_META, kit.MDB_NAME), arg[3])

			case "remove":
				// 删除应用
				m.Richs(ice.CHAT_RIVER, kit.Keys(prefix), arg[1], func(value map[string]interface{}) {
					m.Log(ice.LOG_REMOVE, "storm: %s value: %s", arg[1], kit.Format(value))
				})
				m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, kit.MDB_HASH, arg[1]), "")
			}
		}},
		"/steam": {Name: "/steam", Help: "大气层", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Warn(m.Option(ice.MSG_RIVER) == "", "not join") {
				m.Render("status", 402, "not join")
				return
			}

			if len(arg) < 2 {
				if list := []string{}; m.Option("pod") != "" {
					// 远程空间
					m.Cmdy(ice.WEB_SPACE, m.Option("pod"), "web.chat./steam").Table(func(index int, value map[string]string, head []string) {
						list = append(list, kit.Keys(m.Option("pod"), value["name"]))
					})
					m.Append("name", list)
				} else {
					// 本地空间
					m.Richs(ice.WEB_SPACE, nil, "*", func(key string, value map[string]interface{}) {
						switch value[kit.MDB_TYPE] {
						case ice.WEB_SERVER, ice.WEB_WORKER:
							m.Push(key, value, []string{"type", "name", "user"})
						}
					})
				}
				return
			}

			if !m.Right(cmd, arg[1]) {
				m.Render("status", 403, "not auth")
				return
			}

			switch arg[1] {
			case "spawn":
				// 创建应用
				storm := m.Rich(ice.CHAT_RIVER, kit.Keys(kit.MDB_HASH, arg[0], "tool"), kit.Dict(
					kit.MDB_META, kit.Dict(kit.MDB_NAME, arg[2]),
				))
				m.Log(ice.LOG_CREATE, "storm: %s name: %v", storm, arg[2])
				// 添加命令
				m.Cmd("/storm", arg[0], storm, "add", arg[3:])
				m.Echo(storm)

			case "append":
				// 追加命令
				m.Cmd("/storm", arg[0], arg[2], "add", arg[3:])

			default:
				// 命令列表
				m.Cmdy(ice.WEB_SPACE, arg[2], ice.CTX_COMMAND)
			}
		}},

		"/header": {Name: "/header", Help: "标题栏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(m.Conf(ice.WEB_SHARE, "meta.repos"))
		}},
		"/footer": {Name: "/footer", Help: "状态栏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(m.Conf(ice.WEB_SHARE, "meta.email"))
			m.Echo(m.Conf(ice.WEB_SHARE, "meta.legal"))
		}},

		"/target": {Name: "/target", Help: "对话框", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/source": {Name: "/source", Help: "输入框", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/action": {Name: "/action", Help: "工作台", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Warn(m.Option(ice.MSG_RIVER) == "" || m.Option(ice.MSG_STORM) == "", "not join") {
				m.Render("status", 402, "not join")
				return
			}

			prefix := kit.Keys(kit.MDB_HASH, arg[0], "tool", kit.MDB_HASH, arg[1])
			if len(arg) == 2 {
				if p := m.Option("pod"); p != "" {
					m.Option("pod", "")
					if m.Cmdy(ice.WEB_SPACE, p, "web.chat./action", arg); len(m.Appendv("river")) > 0 {
						// 远程查询
						return
					}
				}

				// 命令列表
				m.Grows(ice.CHAT_RIVER, prefix, "", "", func(index int, value map[string]interface{}) {
					if meta, ok := kit.Value(value, "meta").(map[string]interface{}); ok {
						m.Push("river", arg[0])
						m.Push("storm", arg[1])
						m.Push("action", index)

						m.Push("node", meta["pod"])
						m.Push("group", meta["ctx"])
						m.Push("index", meta["cmd"])
						m.Push("args", kit.Select("[]", kit.Format(meta["args"])))

						msg := m.Cmd(m.Space(meta["pod"]), ice.CTX_COMMAND, meta["ctx"], meta["cmd"])
						m.Push("name", meta["cmd"])
						m.Push("help", kit.Select(msg.Append("help"), kit.Format(meta["help"])))
						m.Push("feature", msg.Append("meta"))
						m.Push("inputs", msg.Append("list"))
					}
				})
				return
			}

			switch arg[2] {
			case "save":
				if p := m.Option("pod"); p != "" {
					// 远程保存
					m.Option("pod", "")
					m.Cmd(ice.WEB_SPACE, p, "web.chat./action", arg)
					return
				}

				// 保存应用
				m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, "list"), "")
				for i := 3; i < len(arg)-4; i += 5 {
					id := m.Grow(ice.CHAT_RIVER, kit.Keys(prefix), kit.Data(
						"pod", arg[i], "ctx", arg[i+1], "cmd", arg[i+2],
						"help", arg[i+3], "args", arg[i+4],
					))
					m.Log(ice.LOG_INSERT, "storm: %s %d: %v", arg[1], id, arg[i:i+5])
				}
			}

			if m.Option("_action") == "上传" {
				msg := m.Cmd(ice.WEB_CACHE, "upload")
				m.Option("_data", msg.Append("data"))
				m.Option("_name", msg.Append("name"))
				m.Cmd(ice.WEB_FAVOR, "upload", msg.Append("type"), msg.Append("name"), msg.Append("data"))
				m.Option("_option", m.Optionv("option"))
			}

			// 查询命令
			cmds := []string{}
			m.Grows(ice.CHAT_RIVER, prefix, kit.MDB_ID, kit.Format(kit.Int(arg[2])+1), func(index int, value map[string]interface{}) {
				if meta, ok := kit.Value(value, "meta").(map[string]interface{}); ok {
					if len(arg) > 3 && arg[3] == "action" {
						// 命令补全
						switch arg[4] {
						case "input":
							switch arg[5] {
							case "location":
								// 查询位置
								m.Copy(m.Cmd("aaa.location"), "append", "name")
								return
							}

						case "favor":
							m.Cmdy(ice.WEB_FAVOR, arg[5:])
							return

						case "device":
							// 记录位置
							m.Cmd(ice.WEB_FAVOR, kit.Select("device", m.Option("hot")), arg[5], arg[6],
								kit.Select("", arg, 7), kit.KeyValue(map[string]interface{}{}, "", kit.UnMarshal(kit.Select("{}", arg, 8))))
							return

						case "upload":
							m.Cmdy(ice.WEB_STORY, "upload")
							return

						case "share":
							list := []string{}
							for k, v := range meta {
								list = append(list, k, kit.Format(v))
							}
							// 共享命令
							m.Cmdy(ice.WEB_SHARE, "add", "action", arg[5], arg[6], list)
							return
						}
					}

					// 组装命令
					cmds = kit.Simple(m.Space(meta["pod"]), kit.Keys(meta["ctx"], meta["cmd"]), arg[3:])
				}
			})

			if len(cmds) == 0 {
				return
			}

			if !m.Right(cmds) {
				m.Render("status", 403, "not auth")
				return
			}

			// 代理命令
			proxy := []string{}
			if m.Option("pod") != "" {
				proxy = append(proxy, ice.WEB_PROXY, m.Option("pod"))
				m.Option("pod", "")
			}

			// 执行命令
			m.Cmdy(proxy, cmds).Option("cmds", cmds)
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
