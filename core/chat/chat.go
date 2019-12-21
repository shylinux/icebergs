package chat

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "chat", Help: "聊天模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"group": {Name: "group", Help: "群组", Value: ice.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", "aaa.json")
			m.Cmd(ice.CTX_CONFIG, "load", "web.json")
			m.Cmd(ice.CTX_CONFIG, "load", "chat.json")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "chat.json", "web.chat.group")
			m.Cmd(ice.CTX_CONFIG, "save", "web.json", "web.story", "web.cache")
			m.Cmd(ice.CTX_CONFIG, "save", "aaa.json", "aaa.role", "aaa.user", "aaa.sess")
		}},

		ice.WEB_LOGIN: {Name: "login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if cmd != "/login" {
				if !m.Options(ice.WEB_SESS) || !m.Options("username") {
					// 检查失败
					m.Log(ice.LOG_WARN, "not login").Error("not login")
					m.Option("path", "")
					return
				}
			}

			// 查询群组
			if len(arg) > 0 && m.Confs("group", kit.Keys("hash", arg[0])) {
				m.Option("sess.river", arg[0])
				if len(arg) > 1 && m.Confs("group", kit.Keys("hash", arg[0], "tool", "hash", arg[1])) {
					m.Option("sess.storm", arg[1])
				}
			}
			m.Log("info", "river: %s storm: %s", m.Option("sess.river"), m.Option("sess.storm"))
		}},

		"/login": {Name: "/login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "check":
				m.Push("username", m.Option("username"))
				m.Push("userrole", m.Option("userrole"))
				m.Echo(m.Option("username"))
			case "login":
				m.Echo(web.Cookie(m, m.Cmdx("aaa.user", "login", arg[1], arg[2])))
			}
		}},
		"/favor": {Name: "/favor", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("cli.system", arg)
		}},

		"/ocean": {Name: "/ocean", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Confm("aaa.user", ice.MDB_HASH, func(key string, value map[string]interface{}) {
					m.Push("key", key)
					m.Push("user.route", m.Conf("cli.runtime", "node.name"))
				})
				return
			}
			switch arg[0] {
			case "spawn":
				if arg[1] == "" {
					arg[1] = kit.ShortKey(m.Confm("group", ice.MDB_HASH), 6)
				}
				user := map[string]interface{}{
					ice.MDB_META: map[string]interface{}{},
					ice.MDB_HASH: map[string]interface{}{},
					ice.MDB_LIST: []interface{}{},
				}
				tool := map[string]interface{}{
					ice.MDB_META: map[string]interface{}{},
					ice.MDB_HASH: map[string]interface{}{},
					ice.MDB_LIST: []interface{}{},
				}
				for _, v := range arg[3:] {
					kit.Value(user, "hash."+v, map[string]interface{}{
						"create_time": m.Time(),
					})
					kit.Value(user, "list."+v+".-2", map[string]interface{}{
						"create_time": m.Time(),
						"operation":   "add",
						"username":    v,
					})
				}

				m.Conf("group", "hash."+arg[1], map[string]interface{}{
					ice.MDB_META: map[string]interface{}{
						"create_time": m.Time(),
						"name":        arg[2],
					},
					"user": user,
					"tool": tool,
				})

				m.Log("info", "river create %v %v", arg[1], kit.Formats(m.Confv("group", "hash."+arg[1])))
				m.Echo(arg[1])
			}
		}},
		"/river": {Name: "/river", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Confm("group", ice.MDB_HASH, func(key string, value map[string]interface{}) {
					m.Push("key", key)
					m.Push("name", kit.Value(value[ice.MDB_META], "name"))
				})
				return
			}
		}},
		"/action": {Name: "/action", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 2 {
				m.Set(ice.MSG_OPTION)
				m.Confm("group", "hash."+arg[0]+".tool.hash."+arg[1]+".list", func(index int, value map[string]interface{}) {
					m.Push("river", arg[0])
					m.Push("storm", arg[1])
					m.Push("action", index)

					m.Push("node", value["pod"])
					m.Push("group", value["ctx"])
					m.Push("index", value["cmd"])

					msg := m.Cmd("web.space", value["pod"], "ctx.command", value["ctx"], value["cmd"])
					m.Push("name", value["cmd"])
					m.Push("help", msg.Append("help"))
					m.Push("inputs", msg.Append("list"))
					m.Push("feature", msg.Append("meta"))
				})
				return
			}

			m.Confm("group", "hash."+arg[0]+".tool.hash."+arg[1]+".list."+arg[2], func(value map[string]interface{}) {
				m.Cmdy("web.space", value["pod"], "ctx.command", value["ctx"], value["cmd"], "run", arg[3:])
			})
		}},
		"/storm": {Name: "/storm", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Confm("group", "hash."+arg[0]+".tool.hash", func(key string, value map[string]interface{}) {
					m.Push("key", key).Push("count", len(value[ice.MDB_LIST].([]interface{})))
				})
				return
			}
		}},
		"/steam": {Name: "/steam", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Push("user", m.Conf("cli.runtime", "boot.username"))
				m.Push("node", m.Conf("cli.runtime", "node.name"))
				m.Confm("web.space", ice.MDB_HASH, func(key string, value map[string]interface{}) {
					m.Push("user", m.Conf("cli.runtime", "boot.username"))
					m.Push("node", value["name"])
				})
				return
			}
			switch arg[1] {
			case "spawn":
				list := []interface{}{}
				for i := 3; i < len(arg)-3; i += 4 {
					if arg[i] == m.Conf("cli.runtime", "node.name") {
						arg[i] = ""
					}
					list = append(list, map[string]interface{}{
						"pod": arg[i],
						"ctx": arg[i+1],
						"cmd": arg[i+2],
						"key": arg[i+3],
					})
				}
				m.Conf("group", "hash."+arg[0]+".tool.hash."+arg[2], map[string]interface{}{
					ice.MDB_META: map[string]interface{}{
						"create_time": m.Time(),
					},
					ice.MDB_HASH: map[string]interface{}{},
					ice.MDB_LIST: list,
				})

				m.Log("info", "steam spawn %v %v %v", arg[0], arg[2], list)
			default:
				if arg[2] == m.Conf("cli.runtime", "node.name") {
					arg[2] = ""
				}
				m.Cmdy("web.space", arg[2], "ctx.command")
			}
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
