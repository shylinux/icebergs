package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"sync"
)

const (
	RIVER = "river"
	STORM = "storm"
)

var Index = &ice.Context{Name: "chat", Help: "聊天中心",
	Configs: map[string]*ice.Config{
		RIVER: {Name: "river", Help: "群组", Value: kit.Data(
			"template", kit.Dict("root", []interface{}{
				[]interface{}{"river", `{{.Option "user.nick"|Format}}@{{.Conf "runtime" "node.name"|Format}}`, "mall"},

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

				[]interface{}{"storm", "wiki", "wiki"},
				[]interface{}{"field", "draw", "web.wiki"},
				[]interface{}{"field", "data", "web.wiki"},
				[]interface{}{"field", "word", "web.wiki"},
				[]interface{}{"field", "walk", "web.wiki"},
				[]interface{}{"field", "feel", "web.wiki"},

				[]interface{}{"storm", "root"},
				[]interface{}{"field", "spide"},
				[]interface{}{"field", "space"},
				[]interface{}{"field", "dream"},
				[]interface{}{"field", "favor"},
				[]interface{}{"field", "story"},
				[]interface{}{"field", "share"},

				[]interface{}{"storm", "miss"},
				[]interface{}{"field", "route"},
				[]interface{}{"field", "group"},
				[]interface{}{"field", "label"},
				[]interface{}{"field", "search"},
				[]interface{}{"field", "commend"},

				[]interface{}{"storm", "team", "team"},
				[]interface{}{"field", "plan", "web.team"},
				[]interface{}{"field", "miss", "web.team"},
				[]interface{}{"field", "stat", "web.team"},
				[]interface{}{"field", "task", "web.team"},

				[]interface{}{"storm", "mall", "mall"},
				[]interface{}{"field", "asset", "web.mall"},
				[]interface{}{"field", "spend", "web.mall"},
				[]interface{}{"field", "trans", "web.mall"},
				[]interface{}{"field", "bonus", "web.mall"},
				[]interface{}{"field", "month", "web.mall"},
			}, "void", []interface{}{
				[]interface{}{"storm", "wiki", "wiki"},
				[]interface{}{"field", "note", "web.wiki"},
			}),
			"black", kit.Dict("tech", []interface{}{
				"/debug",
				"/river.add",
				"/river.share",
				"/river.rename",
				"/river.remove",
				"/storm.remove",
				"/storm.rename",
				"/storm.share",
				"/storm.add",
			}),
			"white", kit.Dict("void", []interface{}{
				"/header",
				"/river",
				"/storm",
				"/action",
				"/footer",
			}),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Watch(gdb.SYSTEM_INIT, m.Prefix("init"))
			m.Watch(gdb.USER_CREATE, m.Prefix("auto"))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(RIVER)
		}},

		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(m.Confm(RIVER, kit.MDB_HASH)) == 0 {
				// 默认群组
				kit.Fetch(m.Confv(RIVER, "meta.template"), func(key string, val map[string]interface{}) {
					if favor := kit.Keys(c.Cap(ice.CTX_FOLLOW), key); m.Richs(web.FAVOR, nil, favor, nil) == nil {
						kit.Fetch(val, func(index int, value interface{}) {
							v := kit.Simple(value)
							web.FavorInsert(m, favor, v[0], v[1], v[2])
						})
					}
				})

				// 黑名单
				kit.Fetch(m.Confv(RIVER, "meta.black.tech"), func(index int, value interface{}) {
					m.Cmd(aaa.ROLE, aaa.Black, aaa.TECH, value)
				})
				// 白名单
				kit.Fetch(m.Confv(RIVER, "meta.white.void"), func(index int, value interface{}) {
					m.Cmd(aaa.ROLE, aaa.White, aaa.VOID, value)
				})
			}
			m.Cap(ice.CTX_STATUS, "start")
		}},
		"auto": {Name: "auto user", Help: "自动化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(aaa.USER, nil, arg[0], func(key string, value map[string]interface{}) {
				m.Option(ice.MSG_USERNICK, value[aaa.USERNAME])
				m.Option(ice.MSG_USERNAME, value[aaa.USERNAME])

				// 创建应用
				storm, river := "", ""
				m.Option("cache.limit", -2)
				web.FavorList(m, kit.Keys(c.Cap(ice.CTX_FOLLOW), aaa.UserRole(m, value[aaa.USERNAME])), "").Table(func(index int, value map[string]string, head []string) {
					switch value[kit.MDB_TYPE] {
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
		}},

		web.LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(ice.MSG_RIVER, "")
			m.Option(ice.MSG_STORM, "")

			if len(arg) > 0 {
				switch arg[0] {
				case "login":
					// 密码登录
					if len(arg) > 2 && aaa.UserLogin(m, arg[1], arg[2]) {
						m.Option(ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERROLE)))
						web.Render(m, "cookie", m.Option(ice.MSG_SESSID))
					}

				default:
					// 群组检查
					m.Richs(RIVER, nil, arg[0], func(key string, value map[string]interface{}) {
						m.Richs(RIVER, kit.Keys(kit.MDB_HASH, arg[0], "user"), m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
							if m.Option(ice.MSG_RIVER, arg[0]); len(arg) > 1 {
								// 应用检查
								m.Richs(RIVER, kit.Keys(kit.MDB_HASH, arg[0], "tool"), arg[1], func(key string, value map[string]interface{}) {
									m.Option(ice.MSG_STORM, arg[1])
								})
							}
							m.Log_AUTH(RIVER, m.Option(ice.MSG_RIVER), STORM, m.Option(ice.MSG_STORM))
						})
					})
				}
			}

			if m.Right(m.Option(ice.MSG_USERURL), m.Optionv("cmds")) {
				return
			}
			// 登录检查
			if m.Warn(!m.Options(ice.MSG_USERNAME), "not login") {
				if m.Option("share") == "" {
					m.Render("status", 401, "not login")
					m.Option(ice.MSG_USERURL, "")
					return
				}
				m.Option(ice.MSG_USERROLE, aaa.VOID)
			}

			// 权限检查
			if m.Warn(!m.Right(m.Option(ice.MSG_USERURL), m.Optionv("cmds")), "not auth") {
				m.Render("status", 403, "not auth")
				m.Option(ice.MSG_USERURL, "")
				return
			}
		}},

		"/toast": {Name: "/toast", Help: "提示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/carte": {Name: "/carte", Help: "菜单", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/tutor": {Name: "/tutor", Help: "向导", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/debug": {Name: "/debug", Help: "调试", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/input": {Name: "/input", Help: "输入", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/login": {Name: "/login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "check":
				m.Richs(aaa.USER, nil, m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
					m.Push("nickname", value["nickname"])
				})
				m.Render(m.Option(ice.MSG_USERNAME))

			case "login":
				m.Render(m.Option(ice.MSG_SESSID))

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
				m.Cmdy(web.SHARE, arg[1], arg[2], arg[3], arg[4:])
			}
		}},

		"/ocean": {Name: "/ocean", Help: "大海洋", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 用户列表
				m.Richs(aaa.USER, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"username", "usernode"})
				})
				return
			}

			switch arg[0] {
			case "spawn":
				// 创建群组
				river := m.Rich(RIVER, nil, kit.Dict(
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
		"/steam": {Name: "/steam", Help: "大气层", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Warn(m.Option(ice.MSG_RIVER) == "", "not join") {
				m.Render("status", 402, "not join")
				return
			}

			if len(arg) < 2 {
				if list := []string{}; m.Option("pod") != "" {
					// 远程空间
					m.Cmdy(web.SPACE, m.Option("pod"), "web.chat./steam").Table(func(index int, value map[string]string, head []string) {
						list = append(list, kit.Keys(m.Option("pod"), value["name"]))
					})
					m.Append("name", list)
				} else {
					// 本地空间
					m.Richs(web.SPACE, nil, "*", func(key string, value map[string]interface{}) {
						switch value[kit.MDB_TYPE] {
						case web.SERVER, web.WORKER:
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
				storm := m.Rich(RIVER, kit.Keys(kit.MDB_HASH, arg[0], "tool"), kit.Dict(
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
				m.Cmdy(web.SPACE, arg[2], ctx.COMMAND)
			}
		}},

		"/target": {Name: "/target", Help: "对话框", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		"/source": {Name: "/source", Help: "输入框", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"commend": {Name: "commend label pod engine work auto", Help: "推荐引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy(web.LABEL, arg)
				return
			}

			switch arg[0] {
			case "add":
				if m.Richs(cmd, nil, arg[1], nil) == nil {
					m.Rich(cmd, nil, kit.Data(kit.MDB_NAME, arg[1]))
				}
				m.Richs(cmd, nil, arg[1], func(key string, value map[string]interface{}) {
					m.Grow(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(
						kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3:],
					))
				})
			case "get":
				wg := &sync.WaitGroup{}
				m.Richs(cmd, nil, arg[1], func(key string, value map[string]interface{}) {
					wg.Add(1)
					m.Gos(m, func(m *ice.Message) {
						m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							m.Cmdy(value[kit.MDB_TEXT], arg[2:])
						})
						wg.Done()
					})
				})
				m.Sort("time", "time_r")
				wg.Wait()
			case "set":
				if arg[1] != "" {
					m.Cmdy(web.SPACE, arg[1], "web.chat.commend", "set", "", arg[2:])
					break
				}

				if m.Richs(cmd, "meta.user", m.Option(ice.MSG_USERNAME), nil) == nil {
					m.Rich(cmd, "meta.user", kit.Dict(
						kit.MDB_NAME, m.Option(ice.MSG_USERNAME),
					))
				}
				switch m.Option("_action") {
				case "喜欢":
					m.Richs(cmd, "meta.user", m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
						m.Grow(cmd, kit.Keys("meta.user", kit.MDB_HASH, key, "like"), kit.Dict(
							kit.MDB_EXTRA, kit.Dict("engine", arg[2], "favor", arg[3], "id", arg[4]),
							kit.MDB_TYPE, arg[5], kit.MDB_NAME, arg[6], kit.MDB_TEXT, arg[7],
						))
					})
				case "讨厌":
					m.Richs(cmd, "meta.user", m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
						m.Grow(cmd, kit.Keys("meta.user", kit.MDB_HASH, key, "hate"), kit.Dict(
							kit.MDB_EXTRA, kit.Dict("engine", arg[2], "favor", arg[3], "id", arg[4]),
							kit.MDB_TYPE, arg[5], kit.MDB_NAME, arg[6], kit.MDB_TEXT, arg[7],
						))
					})
				}

				m.Richs(cmd, nil, arg[2], func(key string, value map[string]interface{}) {
					m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
						m.Cmdy(value[kit.MDB_TEXT], "set", arg[3:])
					})
				})
			default:
				if len(arg) == 2 {
					m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
						m.Push(key, value)
					})
					break
				}
				if len(arg) == 3 {
					m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
						m.Push(key, value)
					})
					break
				}
				m.Cmdy(web.LABEL, arg[0], arg[1], "web.chat.commend", "get", arg[2:])
				// m.Cmdy("web.chat.commend", "get", arg[2:])
			}
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}, RIVER, STORM) }
