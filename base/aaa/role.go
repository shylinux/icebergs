package aaa

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _role_chain(arg ...string) string {
	key := path.Join(strings.ReplaceAll(kit.Keys(arg), ice.PT, ice.PS))
	return strings.TrimPrefix(strings.TrimSuffix(strings.ReplaceAll(key, ice.PS, ice.PT), ice.PT), ice.PT)
}
func _role_black(m *ice.Message, userrole, chain string) {
	mdb.HashSelectUpdate(m, userrole, func(value ice.Map) {
		m.Logs(mdb.INSERT, ROLE, userrole, BLACK, chain)
		black := value[BLACK].(ice.Map)
		black[chain] = true
	})
}
func _role_white(m *ice.Message, userrole, chain string) {
	mdb.HashSelectUpdate(m, userrole, func(value ice.Map) {
		m.Logs(mdb.INSERT, ROLE, userrole, WHITE, chain)
		white := value[WHITE].(ice.Map)
		white[chain] = true
	})
}
func _role_check(value ice.Map, keys []string, ok bool) bool {
	white := value[WHITE].(ice.Map)
	black := value[BLACK].(ice.Map)
	for i := 0; i < len(keys); i++ {
		if v, o := white[kit.Join(keys[:i+1], ice.PT)]; o && v == true {
			ok = true
		}
		if v, o := black[kit.Join(keys[:i+1], ice.PT)]; o && v == true {
			ok = false
		}
	}
	return ok
}
func _role_right(m *ice.Message, userrole string, keys ...string) (ok bool) {
	if userrole == ROOT {
		return true
	}
	mdb.HashSelectDetail(m, kit.Select(VOID, userrole), func(value ice.Map) {
		if userrole == TECH {
			ok = _role_check(value, keys, true)
		} else {
			ok = _role_check(value, keys, false)
		}
	})
	return ok
}
func _role_list(m *ice.Message, userrole string) *ice.Message {
	mdb.HashSelectDetail(m, kit.Select(VOID, userrole), func(value ice.Map) {
		kit.Fetch(value[BLACK], func(k string, v ice.Any) {
			m.Push(ROLE, kit.Value(value, mdb.NAME))
			m.Push(mdb.ZONE, BLACK)
			m.Push(mdb.KEY, k)
		})
		kit.Fetch(value[WHITE], func(k string, v ice.Any) {
			m.Push(ROLE, kit.Value(value, mdb.NAME))
			m.Push(mdb.ZONE, WHITE)
			m.Push(mdb.KEY, k)
		})
	})
	return m
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
	Index.MergeCommands(ice.Commands{
		ROLE: {Name: "role role auto insert", Help: "角色", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.Rich(m, ROLE, nil, kit.Dict(mdb.NAME, TECH, BLACK, kit.Dict(), WHITE, kit.Dict()))
				mdb.Rich(m, ROLE, nil, kit.Dict(mdb.NAME, VOID, WHITE, kit.Dict(), BLACK, kit.Dict()))
				m.Cmd(ROLE, WHITE, VOID, ice.SRC)
				m.Cmd(ROLE, WHITE, VOID, ice.BIN)
				m.Cmd(ROLE, WHITE, VOID, ice.USR)
				m.Cmd(ROLE, BLACK, VOID, ice.USR_LOCAL)
				m.Cmd(ROLE, WHITE, VOID, ice.USR_LOCAL_GO)
			}},
			mdb.INSERT: {Name: "insert role=void,tech zone=white,black key=", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectUpdate(m, m.Option(ROLE), func(key string, value ice.Map) {
					m.Logs(mdb.INSERT, ROLE, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY))
					list := value[m.Option(mdb.ZONE)].(ice.Map)
					list[_role_chain(m.Option(mdb.KEY))] = true
				})
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectUpdate(m, m.Option(ROLE), func(key string, value ice.Map) {
					m.Logs(mdb.DELETE, ROLE, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY))
					list := value[m.Option(mdb.ZONE)].(ice.Map)
					delete(list, _role_chain(m.Option(mdb.KEY)))
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
		}, mdb.HashAction(mdb.SHORT, mdb.NAME)), Hand: func(m *ice.Message, arg ...string) {
			_role_list(m, kit.Select("", arg, 0)).PushAction(mdb.DELETE)
		}},
	})
}

func RoleRight(m *ice.Message, userrole string, arg ...string) bool {
	return m.Cmdx(ROLE, RIGHT, userrole, arg) == ice.OK
}
func RoleAction(cmds ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
		m.Cmd(ROLE, WHITE, VOID, m.PrefixKey())
		m.Cmd(ROLE, BLACK, VOID, m.PrefixKey(), "action")
		for _, cmd := range cmds {
			m.Cmd(ROLE, WHITE, VOID, m.PrefixKey(), "action", cmd)
		}
	}}}
}
func WhiteAction(cmds ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
		m.Cmd(ROLE, BLACK, VOID, m.CommandKey(), "action")
		for _, cmd := range cmds {
			m.Cmd(ROLE, WHITE, VOID, m.CommandKey(), "action", cmd)
		}
	}}}
}
func BlackAction(cmds ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
		for _, cmd := range cmds {
			m.Cmd(ROLE, BLACK, VOID, m.CommandKey(), "action", cmd)
		}
	}}}
}
