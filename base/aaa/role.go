package aaa

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const ( // 用户角色
	ROOT = "root"
	TECH = "tech"
	VOID = "void"
)
const ( // 角色操作
	White = "white"
	Black = "black"
	Right = "right"

	WHITE = "white"
	BLACK = "black"
)
const ( // 返回结果
	OK = "ok"
)

func _role_list(m *ice.Message, userrole string) {
	m.Richs(ROLE, nil, kit.Select(kit.MDB_FOREACH, userrole), func(key string, value map[string]interface{}) {
		for k := range value[White].(map[string]interface{}) {
			m.Push(ROLE, kit.Value(value, kit.MDB_NAME))
			m.Push(kit.MDB_ZONE, White)
			m.Push(kit.MDB_KEY, k)
		}
		for k := range value[Black].(map[string]interface{}) {
			m.Push(ROLE, kit.Value(value, kit.MDB_NAME))
			m.Push(kit.MDB_ZONE, Black)
			m.Push(kit.MDB_KEY, k)
		}
	})
}
func _role_user(m *ice.Message, userrole string, username ...string) {
	m.Richs(ROLE, nil, userrole, func(key string, value map[string]interface{}) {
		for _, user := range username {
			kit.Value(value, kit.Keys(USER, user), true)
		}
	})
}
func _role_right(m *ice.Message, userrole string, keys ...string) (ok bool) {
	if userrole == ROOT {
		// 超级用户
		return true
	}

	m.Richs(ROLE, nil, kit.Select(VOID, userrole), func(key string, value map[string]interface{}) {
		ok = true
		list := value[Black].(map[string]interface{})
		for i := 0; i < len(keys); i++ {
			if v, o := list[kit.Join(keys[:i+1], ".")]; o && v == true {
				ok = false
			}
		}

		if m.Warn(!ok, "%s black right %s", userrole, keys) {
			return
		}
		if userrole == TECH {
			// 管理用户
			return
		}

		ok = false
		list = value[White].(map[string]interface{})
		for i := 0; i < len(keys); i++ {
			if v, o := list[kit.Join(keys[:i+1], ".")]; o && v == true {
				ok = true
			}
		}
		if m.Warn(!ok, ErrNotAuth, userrole, " of ", keys) {
			return
		}
		// 普通用户
	})
	return ok
}
func _role_black(m *ice.Message, userrole, chain string, status bool) {
	m.Richs(ROLE, nil, userrole, func(key string, value map[string]interface{}) {
		m.Log_MODIFY(ROLE, userrole, Black, chain)
		list := value[Black].(map[string]interface{})
		list[chain] = status
	})
}
func _role_white(m *ice.Message, userrole, chain string, status bool) {
	m.Richs(ROLE, nil, userrole, func(key string, value map[string]interface{}) {
		m.Log_MODIFY(ROLE, userrole, White, chain)
		list := value[White].(map[string]interface{})
		list[chain] = status
	})
}

func RoleRight(m *ice.Message, userrole string, keys ...string) bool {
	return _role_right(m, userrole, kit.Split(kit.Keys(keys), ".")...)
}

const ROLE = "role"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ROLE: {Name: ROLE, Help: "角色", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			ROLE: {Name: "role role auto 添加", Help: "角色", Action: map[string]*ice.Action{
				White: {Name: "white role chain...", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
					_role_white(m, arg[0], strings.ReplaceAll(kit.Keys(arg[1:]), "/", "."), true)
				}},
				Black: {Name: "black role chain...", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
					_role_black(m, arg[0], strings.ReplaceAll(kit.Keys(arg[1:]), "/", "."), true)
				}},
				Right: {Name: "right role chain...", Help: "查看权限", Hand: func(m *ice.Message, arg ...string) {
					if _role_right(m, arg[0], kit.Split(strings.ReplaceAll(kit.Keys(arg[1:]), "/", "."), ".")...) {
						m.Echo(OK)
					}
				}},
				mdb.CREATE: {Name: "create zone=white,black role=void@key key=", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(ROLE, nil, m.Option(ROLE), func(key string, value map[string]interface{}) {
						list := value[m.Option(kit.MDB_ZONE)].(map[string]interface{})
						m.Log_CREATE(ROLE, m.Option(ROLE), list[m.Option(kit.MDB_KEY)])
						list[m.Option(kit.MDB_KEY)] = true
					})
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(ROLE, nil, m.Option(ROLE), func(key string, value map[string]interface{}) {
						list := value[m.Option(kit.MDB_ZONE)].(map[string]interface{})
						m.Log_REMOVE(ROLE, m.Option(ROLE), list[m.Option(kit.MDB_KEY)])
						delete(list, m.Option(kit.MDB_KEY))
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 {
					_role_user(m, arg[0], arg[1:]...)
				}

				_role_list(m, kit.Select("", arg, 0))
				m.PushAction(mdb.REMOVE)
			}},
		},
	}, nil)
}
