package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _role_list(m *ice.Message, userrole string) {
	m.Richs(ROLE, nil, kit.Select(mdb.FOREACH, userrole), func(key string, value map[string]interface{}) {
		kit.Fetch(value[BLACK], func(k string, v interface{}) {
			m.Push(ROLE, kit.Value(value, mdb.NAME))
			m.Push(mdb.ZONE, BLACK)
			m.Push(mdb.KEY, k)
		})
		kit.Fetch(value[WHITE], func(k string, v interface{}) {
			m.Push(ROLE, kit.Value(value, mdb.NAME))
			m.Push(mdb.ZONE, WHITE)
			m.Push(mdb.KEY, k)
		})
	})
}
func _role_chain(arg ...string) string {
	return kit.ReplaceAll(kit.Keys(arg), ice.PS, ice.PT)
}
func _role_black(m *ice.Message, userrole, chain string) {
	m.Richs(ROLE, nil, userrole, func(key string, value map[string]interface{}) {
		list := value[BLACK].(map[string]interface{})
		m.Log_CREATE(ROLE, userrole, BLACK, chain)
		list[chain] = true
	})
}
func _role_white(m *ice.Message, userrole, chain string) {
	m.Richs(ROLE, nil, userrole, func(key string, value map[string]interface{}) {
		list := value[WHITE].(map[string]interface{})
		m.Log_CREATE(ROLE, userrole, WHITE, chain)
		list[chain] = true
	})
}
func _role_right(m *ice.Message, userrole string, keys ...string) (ok bool) {
	if userrole == ROOT {
		return true // 超级权限
	}

	m.Richs(ROLE, nil, kit.Select(VOID, userrole), func(key string, value map[string]interface{}) {
		ok = true
		list := value[BLACK].(map[string]interface{})
		for i := 0; i < len(keys); i++ {
			if v, o := list[kit.Join(keys[:i+1], ice.PT)]; o && v == true {
				ok = false // 在黑名单
			}
		}
		if m.Warn(!ok, ice.ErrNotRight, keys, USERROLE, userrole) {
			return // 没有权限
		}
		if userrole == TECH {
			return // 管理权限
		}

		ok = false
		list = value[WHITE].(map[string]interface{})
		for i := 0; i < len(keys); i++ {
			if v, o := list[kit.Join(keys[:i+1], ice.PT)]; o && v == true {
				ok = true // 在白名单
			}
		}
		if m.Warn(!ok, ice.ErrNotRight, keys, USERROLE, userrole) {
			return // 没有权限
		}
		if userrole == VOID {
			return // 用户权限
		}
	})
	return ok
}

func RoleRight(m *ice.Message, userrole string, keys ...string) bool {
	return _role_right(m, userrole, kit.Split(kit.Keys(keys), ice.PT)...)
}

const (
	ROOT = "root"
	TECH = "tech"
	VOID = "void"
)
const (
	BLACK = "black"
	WHITE = "white"
	RIGHT = "right"
)
const ROLE = "role"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ROLE: {Name: ROLE, Help: "角色", Value: kit.Data(mdb.SHORT, mdb.NAME)},
	}, Commands: map[string]*ice.Command{
		ROLE: {Name: "role role auto insert", Help: "角色", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Rich(ROLE, nil, kit.Dict(mdb.NAME, TECH, BLACK, kit.Dict(), WHITE, kit.Dict()))
				m.Rich(ROLE, nil, kit.Dict(mdb.NAME, VOID, WHITE, kit.Dict(), BLACK, kit.Dict()))
				m.Cmd(ROLE, WHITE, VOID, ice.SRC)
				m.Cmd(ROLE, WHITE, VOID, ice.BIN)
				m.Cmd(ROLE, WHITE, VOID, ice.USR)
				m.Cmd(ROLE, BLACK, VOID, ice.USR_LOCAL)
				m.Cmd(ROLE, WHITE, VOID, ice.USR_LOCAL_GO)
			}},
			mdb.INSERT: {Name: "insert role=void,tech zone=white,black key=", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(ROLE, nil, m.Option(ROLE), func(key string, value map[string]interface{}) {
					m.Log_CREATE(ROLE, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY))
					list := value[m.Option(mdb.ZONE)].(map[string]interface{})
					list[m.Option(mdb.KEY)] = true
				})
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(ROLE, nil, m.Option(ROLE), func(key string, value map[string]interface{}) {
					m.Log_REMOVE(ROLE, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY))
					list := value[m.Option(mdb.ZONE)].(map[string]interface{})
					delete(list, m.Option(mdb.KEY))
				})
			}},

			BLACK: {Name: "black role chain", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
				_role_black(m, arg[0], _role_chain(arg[1:]...))
			}},
			WHITE: {Name: "white role chain", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
				_role_white(m, arg[0], _role_chain(arg[1:]...))
			}},
			RIGHT: {Name: "right role chain", Help: "查看权限", Hand: func(m *ice.Message, arg ...string) {
				if _role_right(m, arg[0], kit.Split(_role_chain(arg[1:]...), ice.PT)...) {
					m.Echo(ice.OK)
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_role_list(m, kit.Select("", arg, 0))
			m.PushAction(mdb.DELETE)
		}},
	}})
}
