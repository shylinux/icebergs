package aaa

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"strings"
)

var Index = &ice.Context{Name: "aaa", Help: "认证模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.AAA_ROLE: {Name: "role", Help: "角色", Value: kit.Data(kit.MDB_SHORT, "chain")},
		ice.AAA_USER: {Name: "user", Help: "用户", Value: kit.Data(kit.MDB_SHORT, "username")},
		ice.AAA_SESS: {Name: "sess", Help: "会话", Value: kit.Data(kit.MDB_SHORT, "uniq", "expire", "720h")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			// 权限索引
			m.Conf(ice.AAA_ROLE, "black.tech.meta.short", "chain")
			m.Conf(ice.AAA_ROLE, "white.tech.meta.short", "chain")
			m.Conf(ice.AAA_ROLE, "black.void.meta.short", "chain")
			m.Conf(ice.AAA_ROLE, "white.void.meta.short", "chain")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(ice.AAA_ROLE, ice.AAA_USER, ice.AAA_SESS)
		}},

		ice.AAA_ROLE: {Name: "role check|black|white|right", Help: "角色", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				kit.Fetch(m.Confv("role", "meta.root"), func(key string, value string) {
					m.Push("userrole", "root")
					m.Push("username", key)
				})
				kit.Fetch(m.Confv("role", "meta.tech"), func(key string, value string) {
					m.Push("userrole", "tech")
					m.Push("username", key)
				})
				return
			}
			switch arg[0] {
			case "check":
				// 用户角色
				if len(arg) > 1 && arg[1] != "" {
					m.Echo(kit.Select(kit.Select("void",
						"tech", m.Confs(ice.AAA_ROLE, kit.Keys("meta.tech", arg[1]))),
						"root", m.Confs(ice.AAA_ROLE, kit.Keys("meta.root", arg[1]))))
				}

			case "black", "white":
				// 黑白名单
				m.Rich(ice.AAA_ROLE, kit.Keys(arg[0], arg[1]), kit.Dict(
					"status", arg[2], "chain", kit.Keys(arg[3:]),
				))
				m.Log(ice.LOG_ENABLE, "role: %s %s: %v", arg[1], arg[0], arg[3:])

			case "right":
				if m.Option(ice.MSG_USERROLE) == ice.ROLE_ROOT {
					// 超级用户
					m.Echo("ok")
					break
				}

				ok := true
				keys := strings.Split(kit.Keys(arg[2:]), ".")
				for i := 0; i < len(keys); i++ {
					if !ok {
						break
					}
					// 黑名单
					m.Richs(ice.AAA_ROLE, kit.Keys("black", arg[1]), kit.Keys(keys[:i+1]), func(key string, value map[string]interface{}) {
						ok = value["status"] != "enable"
					})
				}
				if m.Warn(!ok, "black right %s", keys) {
					break
				}
				if m.Option(ice.MSG_USERROLE) == ice.ROLE_TECH {
					// 管理用户
					m.Echo("ok")
					break
				}

				ok = false
				for i := 0; i < len(keys); i++ {
					if ok {
						break
					}
					// 白名单
					m.Richs(ice.AAA_ROLE, kit.Keys("white", arg[1]), kit.Keys(keys[:i+1]), func(key string, value map[string]interface{}) {
						ok = value["status"] == "enable"
					})
				}
				if m.Warn(!ok, "no white right %s", keys) {
					break
				}
				// 普通用户
				m.Echo("ok")

			default:
				m.Conf(ice.AAA_ROLE, kit.Keys("meta", arg[0], arg[1]), kit.Select("true", arg, 2))
			}
		}},
		ice.AAA_USER: {Name: "user first|login", Help: "用户", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 用户列表
				m.Richs(ice.AAA_USER, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "username", "usernode"})
				})
				return
			}

			switch arg[0] {
			case "first":
				// 超级用户
				if m.Richs(ice.AAA_USER, nil, "%", nil) == nil {
					m.Rich(ice.AAA_USER, nil, kit.Dict("username", arg[1],
						"usernode", m.Conf(ice.CLI_RUNTIME, "boot.hostname"),
					))
					user := m.Richs(ice.AAA_USER, nil, arg[1], nil)
					m.Info("create user: %s %s", arg[1], kit.Format(user))
					m.Event(ice.USER_CREATE, arg[1])
				}

			case "login":
				// 用户认证
				user := m.Richs(ice.AAA_USER, nil, arg[1], nil)
				if word := kit.Select("", arg, 2); user == nil {
					nick := arg[1]
					if len(nick) > 8 {
						nick = nick[:8]
					}
					// 创建用户
					m.Rich(ice.AAA_USER, nil, kit.Dict(
						"usernick", nick, "username", arg[1], "password", word,
						"usernode", m.Conf(ice.CLI_RUNTIME, "boot.hostname"),
					))
					user = m.Richs(ice.AAA_USER, nil, arg[1], nil)
					m.Log(ice.LOG_CREATE, "%s: %s", arg[1], kit.Format(user))
					m.Event(ice.USER_CREATE, arg[1])

				} else if word != "" {
					if kit.Format(user["password"]) == "" {
						// 设置密码
						user["password"] = word
					} else if kit.Format(user["password"]) != word {
						// 认证失败
						m.Info("login fail user: %s", arg[1])
						break
					}
				}

				if m.Options(ice.MSG_SESSID) && m.Cmdx(ice.AAA_SESS, "check", m.Option(ice.MSG_SESSID)) == arg[1] {
					// 复用会话
					m.Echo(m.Option(ice.MSG_SESSID))
					break
				}

				// 创建会话
				m.Echo(m.Cmdx(ice.AAA_SESS, "create", arg[1]))
			}
		}},
		ice.AAA_SESS: {Name: "sess check|login", Help: "会话", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 会话列表
				m.Richs(ice.AAA_SESS, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"key", "time", "username", "userrole"})
				})
				return
			}

			switch arg[0] {
			case "auth":
				m.Richs(ice.AAA_SESS, nil, arg[1], func(value map[string]interface{}) {
					value["username"], value["userrole"] = arg[2], m.Cmdx(ice.AAA_ROLE, "check", arg[2])
					m.Log(ice.LOG_AUTH, "sessid: %s username: %s userrole: %s", arg[1], arg[2], value["userrole"])
					m.Echo("%v", value["userrole"])
				})

			case "check":
				// 查看会话
				m.Richs(ice.AAA_SESS, nil, arg[1], func(value map[string]interface{}) {
					m.Push(arg[1], value, []string{"username", "userrole"})
					m.Echo("%s", value["username"])
				})

			case "create":
				// 创建会话
				h := m.Rich(ice.AAA_SESS, nil, kit.Dict(
					"username", arg[1], "userrole", m.Cmdx(ice.AAA_ROLE, "check", arg[1]),
					"from", m.Option(ice.MSG_SESSID),
				))
				m.Log(ice.LOG_CREATE, "sessid: %s username: %s", h, arg[1])
				m.Echo(h)
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
