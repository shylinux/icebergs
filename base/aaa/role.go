package aaa

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _role_user(m *ice.Message, userrole string, username ...string) {
	m.Richs(ROLE, nil, userrole, func(key string, value map[string]interface{}) {
		for _, user := range username {
			kit.Value(value, kit.Keys(USER, user), true)
		}
	})
}
func _role_list(m *ice.Message, userrole string) {
	m.Richs(ROLE, nil, kit.Select(kit.MDB_FOREACH, userrole), func(key string, value map[string]interface{}) {
		kit.Fetch(value[WHITE], func(k string, v interface{}) {
			m.Push(ROLE, kit.Value(value, kit.MDB_NAME))
			m.Push(kit.MDB_ZONE, WHITE)
			m.Push(kit.MDB_KEY, k)
		})
		kit.Fetch(value[BLACK], func(k string, v interface{}) {
			m.Push(ROLE, kit.Value(value, kit.MDB_NAME))
			m.Push(kit.MDB_ZONE, BLACK)
			m.Push(kit.MDB_KEY, k)
		})
	})
}
func _role_chain(arg ...string) string {
	return strings.ReplaceAll(kit.Keys(arg), "/", ice.PT)
}
func _role_black(m *ice.Message, userrole, chain string, status bool) {
	m.Richs(ROLE, nil, userrole, func(key string, value map[string]interface{}) {
		m.Log_MODIFY(ROLE, userrole, BLACK, chain)
		list := value[BLACK].(map[string]interface{})
		list[chain] = status
	})
}
func _role_white(m *ice.Message, userrole, chain string, status bool) {
	m.Richs(ROLE, nil, userrole, func(key string, value map[string]interface{}) {
		m.Log_MODIFY(ROLE, userrole, WHITE, chain)
		list := value[WHITE].(map[string]interface{})
		list[chain] = status
	})
}
func _role_right(m *ice.Message, userrole string, keys ...string) (ok bool) {
	if userrole == ROOT {
		return true // 超级用户
	}

	m.Richs(ROLE, nil, kit.Select(VOID, userrole), func(key string, value map[string]interface{}) {
		ok = true
		list := value[BLACK].(map[string]interface{})
		for i := 0; i < len(keys); i++ {
			if v, o := list[kit.Join(keys[:i+1], ice.PT)]; o && v == true {
				ok = false
			}
		}

		if m.Warn(!ok, ice.ErrNotRight, userrole, ice.OF, keys) {
			return
		}
		if userrole == TECH {
			return // 管理用户
		}

		ok = false
		list = value[WHITE].(map[string]interface{})
		for i := 0; i < len(keys); i++ {
			if v, o := list[kit.Join(keys[:i+1], ice.PT)]; o && v == true {
				ok = true
			}
		}

		if m.Warn(!ok, ice.ErrNotRight, userrole, ice.OF, keys) {
			return
		}
		// 普通用户
	})
	return ok
}

func RoleRight(m *ice.Message, userrole string, keys ...string) bool {
	return _role_right(m, userrole, kit.Split(kit.Keys(keys), ice.PT)...)
}

const ( // 用户角色
	ROOT = "root"
	TECH = "tech"
	VOID = "void"
)
const ( // 角色操作
	BLACK = "black"
	WHITE = "white"
	RIGHT = "right"
)
const ROLE = "role"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ROLE: {Name: ROLE, Help: "角色", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
	}, Commands: map[string]*ice.Command{
		ROLE: {Name: "role role auto create", Help: "角色", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create role=void,tech zone=white,black key=", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
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

			BLACK: {Name: "black role chain...", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
				_role_black(m, arg[0], _role_chain(arg[1:]...), true)
			}},
			WHITE: {Name: "white role chain...", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
				_role_white(m, arg[0], _role_chain(arg[1:]...), true)
			}},
			RIGHT: {Name: "right role chain...", Help: "查看权限", Hand: func(m *ice.Message, arg ...string) {
				if _role_right(m, arg[0], kit.Split(_role_chain(arg[1:]...), ice.PT)...) {
					m.Echo(ice.OK)
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 { // 角色列表
				_role_list(m, kit.Select("", arg, 0))
				m.PushAction(mdb.REMOVE)
				return
			}

			// 设置角色
			_role_user(m, arg[0], arg[1:]...)
		}},
	}})
}
