package aaa

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "aaa", Help: "认证模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.AAA_ROLE: {Name: "role", Help: "角色", Value: kit.Data()},
		ice.AAA_USER: {Name: "user", Help: "用户", Value: kit.Data(kit.MDB_SHORT, "username")},
		ice.AAA_SESS: {Name: "sess", Help: "会话", Value: kit.Data(kit.MDB_SHORT, "uniq", "expire", "720h")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", "aaa.json")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "aaa.json", ice.AAA_ROLE, ice.AAA_USER, ice.AAA_SESS)
		}},
		ice.AAA_ROLE: {Name: "role", Help: "角色", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "check":
				m.Echo(kit.Select("void", "root", arg[1] == m.Conf(ice.CLI_RUNTIME, "boot.username")))
			}
		}},
		ice.AAA_USER: {Name: "user", Help: "用户", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "login":
				// 用户认证
				user := m.Richs(ice.AAA_USER, nil, arg[1], nil)
				if user == nil {
					m.Rich(ice.AAA_USER, nil, kit.Dict(
						"username", arg[1], "password", arg[2],
						"usernode", m.Conf(ice.CLI_RUNTIME, "boot.hostname"),
					))
					user = m.Richs(ice.AAA_USER, nil, arg[1], nil)
					m.Info("create user: %s %s", arg[1], kit.Format(user))
				} else if kit.Format(user["password"]) != arg[2] {
					m.Info("login fail user: %s", arg[1])
					break
				}

				// 用户授权
				sessid := kit.Format(user[ice.WEB_SESS])
				if sessid == "" {
					role := m.Cmdx(ice.AAA_ROLE, "check", arg[1])
					sessid = m.Rich(ice.AAA_SESS, nil, kit.Dict(
						"username", arg[1], "userrole", role,
					))
					m.Info("user: %s role: %s sess: %s", arg[1], role, sessid)
				}
				m.Echo(sessid)
			}
		}},
		ice.AAA_SESS: {Name: "sess check|login", Help: "会话", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "check":
				m.Richs(ice.AAA_SESS, nil, arg[1], func(value map[string]interface{}) {
					m.Push(arg[1], value, []string{"username", "userrole"})
					m.Echo("%s", value["username"])
				})
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
