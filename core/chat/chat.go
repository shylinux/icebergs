package chat

import (
	"github.com/shylinux/toolkits"

	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
)

var Index = &ice.Context{Name: "chat", Help: "聊天模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"group": {Name: "group", Value: map[string]interface{}{
			"meta": map[string]interface{}{},
			"list": map[string]interface{}{},
			"hash": map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		"_init": {Name: "_init", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("ctx.config", "load", "var/conf/aaa.json")
			m.Cmd("ctx.config", "load", "var/conf/chat.json")
		}},
		"_login": {Name: "_login", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if cmd != "/login" {
				if !m.Options("sessid") || !m.Options("username") {
					m.Option("path", "")
					m.Log("warn", "not login")
					m.Echo("error").Echo("not login")
					return
				}
			}
			if len(arg) > 0 && m.Confs("group", "hash."+arg[0]) {
				m.Option("sess.river", arg[0])
				if len(arg) > 1 && m.Confs("group", "hash."+arg[0]+".tool.hash."+arg[1]) {
					m.Option("sess.storm", arg[1])
				}
			}
			m.Log("info", "river: %s storm: %s", m.Option("sess.river"), m.Option("sess.storm"))
		}},
		"_exit": {Name: "_init", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("ctx.config", "save", "var/conf/chat.json", "web.chat.group")
			m.Cmd("ctx.config", "save", "var/conf/aaa.json", "aaa.role", "aaa.user", "aaa.sess")
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
				m.Confm("aaa.user", "hash", func(key string, value map[string]interface{}) {
					m.Push("key", key)
					m.Push("user.route", m.Conf("cli.runtime", "node.name"))
				})
				return
			}
			switch arg[0] {
			case "spawn":
				if arg[1] == "" {
					arg[1] = kit.ShortKey(m.Confm("group", "hash"), 6)
				}
				user := map[string]interface{}{
					"meta": map[string]interface{}{},
					"hash": map[string]interface{}{},
					"list": []interface{}{},
				}
				tool := map[string]interface{}{
					"meta": map[string]interface{}{},
					"hash": map[string]interface{}{},
					"list": []interface{}{},
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
					"meta": map[string]interface{}{
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
				m.Confm("group", "hash", func(key string, value map[string]interface{}) {
					m.Push("key", key)
					m.Push("name", kit.Value(value["meta"], "name"))
				})
				return
			}
		}},
		"/action": {Name: "/action", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 2 {
				m.Set("option")
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
					m.Push("key", key).Push("count", len(value["list"].([]interface{})))
				})
				return
			}
		}},
		"/steam": {Name: "/steam", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Push("user", m.Conf("cli.runtime", "boot.username"))
				m.Push("node", m.Conf("cli.runtime", "node.name"))
				m.Confm("web.space", "hash", func(key string, value map[string]interface{}) {
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
					"meta": map[string]interface{}{
						"create_time": m.Time(),
					},
					"hash": map[string]interface{}{},
					"list": list,
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

func init() { web.Index.Register(Index, &web.WEB{}) }
