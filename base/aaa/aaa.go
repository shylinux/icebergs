package aaa

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "aaa", Help: "认证模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"role": {Name: "role", Help: "角色", Value: map[string]interface{}{
			ice.MDB_META: map[string]interface{}{},
			ice.MDB_HASH: map[string]interface{}{
				"root": map[string]interface{}{},
				"tech": map[string]interface{}{},
				"void": map[string]interface{}{},
			},
			ice.MDB_LIST: map[string]interface{}{},
		}},
		"user": {Name: "user", Help: "用户", Value: map[string]interface{}{
			ice.MDB_META: map[string]interface{}{},
			ice.MDB_HASH: map[string]interface{}{},
			ice.MDB_LIST: map[string]interface{}{},
		}},
		"sess": {Name: "sess", Help: "会话", Value: map[string]interface{}{
			ice.MDB_META: map[string]interface{}{"expire": "720h"},
			ice.MDB_HASH: map[string]interface{}{},
			ice.MDB_LIST: map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"role": {Name: "role", Help: "角色", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"user": {Name: "user", Help: "用户", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "login":
				user := m.Confm("user", "hash."+arg[1])
				if user == nil {
					m.Confv("user", "hash."+arg[1], map[string]interface{}{
						"create_time": m.Time(),
						"password":    arg[2],
						"userrole":    kit.Select("void", "root", arg[1] == m.Conf("cli.runtime", "boot.username")),
					})
					m.Log("info", "create user %s %s", arg[1], m.Conf("user", "hash."+arg[1]))
				} else if kit.Format(user["password"]) != arg[2] {
					m.Log("warn", "login fail %s", arg[1])
					// 登录失败
					break
				}

				sessid := kit.Format(user[ice.WEB_SESS])
				if sessid == "" {
					sessid = m.Cmdx("aaa.sess", "login", arg[1])
				}

				m.Grow("user", nil, map[string]interface{}{
					"create_time": m.Time(),
					"remote_ip":   m.Option("remote_ip"),
					"username":    arg[1],
					ice.WEB_SESS:  sessid,
				})
				// 登录成功
				m.Echo(sessid)
			}
		}},
		"sess": {Name: "sess check|login", Help: "会话", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "check":
				user := m.Conf("sess", "hash."+arg[1]+".username")
				if user != "" {
					m.Confm("user", "hash."+user, func(value map[string]interface{}) {
						m.Push("username", user)
						m.Push("userrole", value["userrole"])
					})
				}

				m.Echo(user)
			case "login":
				sessid := kit.Hashs("uniq")
				m.Conf("sess", "hash."+sessid, map[string]interface{}{
					"create_time": m.Time(),
					"username":    arg[1],
				})
				m.Log("info", "create sess %s %s", arg[1], sessid)
				m.Echo(sessid)
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
