package aaa

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
	"math"
	"strings"
)

func distance(lat1, long1, lat2, long2 float64) float64 {
	lat1 = lat1 * math.Pi / 180
	long1 = long1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	long2 = long2 * math.Pi / 180
	return 2 * 6371 * math.Asin(math.Sqrt(math.Pow(math.Sin(math.Abs(lat1-lat2)/2), 2)+math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(math.Abs(long1-long2)/2), 2)))
}

var Index = &ice.Context{Name: "aaa", Help: "认证模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.AAA_ROLE: {Name: "role", Help: "角色", Value: kit.Data(kit.MDB_SHORT, "chain")},
		ice.AAA_USER: {Name: "user", Help: "用户", Value: kit.Data(kit.MDB_SHORT, "username")},
		ice.AAA_SESS: {Name: "sess", Help: "会话", Value: kit.Data(kit.MDB_SHORT, "uniq", "expire", "720h")},

		"location": {Name: "location", Help: "定位", Value: kit.Data(kit.MDB_SHORT, "name")},
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
			m.Save(ice.AAA_ROLE, ice.AAA_USER, ice.AAA_SESS, m.Prefix("location"))
		}},

		"location": {Name: "location", Help: "location", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Grows("location", nil, "", "", func(index int, value map[string]interface{}) {
					m.Push("", value)
				})
				return
			}
			if len(arg) == 1 {
				m.Richs("location", nil, arg[0], func(key string, value map[string]interface{}) {
					m.Info("what %v", value)
					m.Push("detail", value)
				})
				return
			}
			if len(arg) == 2 {
				m.Richs("aaa.location", nil, "*", func(key string, value map[string]interface{}) {
					m.Push("name", value["name"])
					m.Push("distance", kit.Int(distance(
						float64(kit.Int(arg[0]))/100000,
						float64(kit.Int(arg[1]))/100000,
						float64(kit.Int(value["latitude"]))/100000,
						float64(kit.Int(value["longitude"]))/100000,
					)*1000))
				})
				m.Sort("distance", "int")
				return
			}

			data := m.Richs("location", nil, arg[0], nil)
			if data != nil {
				data["count"] = kit.Int(data["count"]) + 1
			} else {
				data = kit.Dict("name", arg[0], "address", arg[1], "latitude", arg[2], "longitude", arg[3], "count", 1)
				m.Rich("location", nil, data)
			}
			m.Grow("location", nil, data)
		}},

		ice.AAA_ROLE: {Name: "role", Help: "角色", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "check":
				// 用户角色
				m.Echo(kit.Select(kit.Select("void", "tech", m.Confs(ice.AAA_ROLE, kit.Keys("meta.tech", arg[1]))), "root", m.Confs(ice.AAA_ROLE, kit.Keys("meta.root", arg[1]))))

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
				m.Conf(ice.AAA_ROLE, kit.Keys("meta", arg[0], arg[1]), "true")
			}
		}},
		ice.AAA_USER: {Name: "user", Help: "用户", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "first":
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
				if user == nil {
					// 创建用户
					m.Rich(ice.AAA_USER, nil, kit.Dict(
						"username", arg[1], "password", arg[2],
						"usernode", m.Conf(ice.CLI_RUNTIME, "boot.hostname"),
					))
					user = m.Richs(ice.AAA_USER, nil, arg[1], nil)
					m.Info("create user: %s %s", arg[1], kit.Format(user))
					m.Event(ice.USER_CREATE, arg[1])

				} else if kit.Format(user["password"]) == "" {
					user["password"] = arg[2]

				} else if kit.Format(user["password"]) != arg[2] {
					// 认证失败
					m.Info("login fail user: %s", arg[1])
					break
				}

				if m.Options(ice.MSG_SESSID) && m.Cmdx(ice.AAA_SESS, "check", m.Option(ice.MSG_SESSID)) == arg[1] {
					// 复用会话
					m.Echo(m.Option(ice.MSG_SESSID))
					break
				}

				// 创建会话
				role := m.Cmdx(ice.AAA_ROLE, "check", arg[1])
				sessid := m.Rich(ice.AAA_SESS, nil, kit.Dict(
					"username", arg[1], "userrole", role,
				))
				m.Info("user: %s role: %s sess: %s", arg[1], role, sessid)
				m.Echo(sessid)
			}
		}},
		ice.AAA_SESS: {Name: "sess check|login", Help: "会话", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Richs(ice.AAA_SESS, nil, "", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"key", "time", "username", "userrole"})
				})
				return
			}

			switch arg[0] {
			case "check":
				m.Richs(ice.AAA_SESS, nil, arg[1], func(value map[string]interface{}) {
					m.Push(arg[1], value, []string{"username", "userrole"})
					m.Echo("%s", value["username"])
				})
			case "create":
				h := m.Rich(ice.AAA_SESS, nil, kit.Dict(
					"username", arg[1], "userrole", kit.Select("", arg, 2),
				))
				m.Log(ice.LOG_CREATE, "%s: %s", h, arg[1])
				m.Echo(h)
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
