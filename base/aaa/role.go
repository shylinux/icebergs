package aaa

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _role_chain(arg ...string) string {
	key := path.Join(strings.ReplaceAll(kit.Keys(arg), ice.PT, ice.PS))
	return strings.TrimPrefix(strings.TrimSuffix(strings.ReplaceAll(key, ice.PS, ice.PT), ice.PT), ice.PT)
}
func _role_set(m *ice.Message, userrole, zone, chain string, status bool) {
	m.Logs(mdb.INSERT, ROLE, userrole, zone, chain)
	mdb.HashSelectUpdate(m, userrole, func(value ice.Map) {
		black := value[zone].(ice.Map)
		black[chain] = status
	})
}
func _role_black(m *ice.Message, userrole, chain string) {
	_role_set(m, userrole, BLACK, chain, true)
}
func _role_white(m *ice.Message, userrole, chain string) {
	_role_set(m, userrole, WHITE, chain, true)
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
		ok = _role_check(value, keys, userrole == TECH)
	})
	return
}
func _role_list(m *ice.Message, userrole string) *ice.Message {
	mdb.HashSelectDetail(m, kit.Select(VOID, userrole), func(value ice.Map) {
		kit.Fetch(value[BLACK], func(k string, v ice.Any) {
			m.Push(ROLE, kit.Value(value, mdb.NAME))
			m.Push(mdb.ZONE, BLACK)
			m.Push(mdb.KEY, k)
			m.Push(mdb.VALUE, v)
		})
		kit.Fetch(value[WHITE], func(k string, v ice.Any) {
			m.Push(ROLE, kit.Value(value, mdb.NAME))
			m.Push(mdb.ZONE, WHITE)
			m.Push(mdb.KEY, k)
			m.Push(mdb.VALUE, v)
		})
	})
	return m.Sort(mdb.KEY).StatusTimeCount()
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
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd("", mdb.CREATE, TECH, VOID) }},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				for _, role := range arg {
					mdb.Rich(m, ROLE, nil, kit.Dict(mdb.NAME, role, BLACK, kit.Dict(), WHITE, kit.Dict()))
				}
			}},
			mdb.INSERT: {Name: "insert role=void,tech zone=white,black key", Hand: func(m *ice.Message, arg ...string) {
				_role_set(m, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY), true)
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				_role_set(m, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY), false)
			}},
			BLACK: {Name: "black role chain", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
				_role_black(m, arg[0], _role_chain(arg[1:]...))
			}},
			WHITE: {Name: "white role chain", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
				_role_white(m, arg[0], _role_chain(arg[1:]...))
			}},
			RIGHT: {Name: "right role chain", Help: "检查权限", Hand: func(m *ice.Message, arg ...string) {
				if _role_right(m, arg[0], kit.Split(_role_chain(arg[1:]...), ice.PT)...) {
					m.Echo(ice.OK)
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME)), Hand: func(m *ice.Message, arg ...string) {
			_role_list(m, kit.Select("", arg, 0)).PushAction(mdb.DELETE)
		}},
	})
}

func Right(m *ice.Message, arg ...ice.Any) bool {
	return m.Option(ice.MSG_USERROLE) == ROOT || !m.Warn(m.Cmdx(ROLE, RIGHT, m.Option(ice.MSG_USERROLE), arg) != ice.OK,
		ice.ErrNotRight, kit.Join(kit.Simple(arg), ice.PT), USERROLE, m.Option(ice.MSG_USERROLE), logs.FileLineMeta(logs.FileLine(2)))
}
func RoleRight(m *ice.Message, userrole string, arg ...string) bool {
	return m.Cmdx(ROLE, RIGHT, userrole, arg) == ice.OK
}
func RoleAction(cmds ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		if c, ok := ice.Info.Index[m.CommandKey()].(*ice.Context); ok && c == m.Target() {
			m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
			m.Cmd(ROLE, BLACK, VOID, m.CommandKey(), ice.ACTION)
		}
		m.Cmd(ROLE, WHITE, VOID, m.PrefixKey())
		m.Cmd(ROLE, BLACK, VOID, m.PrefixKey(), ice.ACTION)
		for _, cmd := range cmds {
			m.Cmd(ROLE, WHITE, VOID, m.PrefixKey(), ice.ACTION, cmd)
		}
	}}}
}
func WhiteAction(cmds ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
		m.Cmd(ROLE, BLACK, VOID, m.CommandKey(), ice.ACTION)
		for _, cmd := range cmds {
			m.Cmd(ROLE, WHITE, VOID, m.CommandKey(), ice.ACTION, cmd)
		}
	}}}
}
func BlackAction(cmds ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
		for _, cmd := range cmds {
			m.Cmd(ROLE, BLACK, VOID, m.CommandKey(), ice.ACTION, cmd)
		}
	}}}
}
