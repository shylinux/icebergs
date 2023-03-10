package aaa

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _role_keys(key ...string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.Replace(path.Join(strings.Replace(kit.Keys(key), ice.PT, ice.PS, -1)), ice.PS, ice.PT, -1), ice.PT), ice.PT)
}
func _role_set(m *ice.Message, role, zone, key string, status bool) {
	m.Logs(mdb.INSERT, mdb.KEY, ROLE, ROLE, role, zone, key)
	mdb.HashSelectUpdate(m, role, func(value ice.Map) { value[zone].(ice.Map)[key] = status })
}
func _role_white(m *ice.Message, role, key string) { _role_set(m, role, WHITE, key, true) }
func _role_black(m *ice.Message, role, key string) { _role_set(m, role, BLACK, key, true) }
func _role_check(value ice.Map, key []string, ok bool) bool {
	white, black := value[WHITE].(ice.Map), value[BLACK].(ice.Map)
	for i := 0; i < len(key); i++ {
		if v, o := white[kit.Join(key[:i+1], ice.PT)]; o && v == true {
			ok = true
		}
		if v, o := black[kit.Join(key[:i+1], ice.PT)]; o && v == true {
			ok = false
		}
	}
	return ok
}
func _role_right(m *ice.Message, role string, key ...string) (ok bool) {
	if role == ROOT {
		return true
	}
	mdb.HashSelectDetail(m, kit.Select(VOID, role), func(value ice.Map) {
		ok = _role_check(value, key, role == TECH)
	})
	return
}
func _role_list(m *ice.Message, role string) *ice.Message {
	mdb.HashSelectDetail(m, kit.Select(VOID, role), func(value ice.Map) {
		kit.For(value[WHITE], func(k string, v ice.Any) {
			m.Push(ROLE, kit.Value(value, mdb.NAME))
			m.Push(mdb.ZONE, WHITE).Push(mdb.KEY, k).Push(mdb.STATUS, v)
		})
		kit.For(value[BLACK], func(k string, v ice.Any) {
			m.Push(ROLE, kit.Value(value, mdb.NAME))
			m.Push(mdb.ZONE, BLACK).Push(mdb.KEY, k).Push(mdb.STATUS, v)
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
	WHITE = "white"
	BLACK = "black"
	RIGHT = "right"
)
const ROLE = "role"

func init() {
	Index.MergeCommands(ice.Commands{
		ROLE: {Name: "role role auto insert", Help: "角色", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd("", mdb.CREATE, TECH, VOID) }},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				kit.For(arg, func(role string) {
					mdb.Rich(m, ROLE, nil, kit.Dict(mdb.NAME, role, BLACK, kit.Dict(), WHITE, kit.Dict()))
				})
			}},
			mdb.INSERT: {Name: "insert role*=void,tech zone*=white,black key*", Hand: func(m *ice.Message, arg ...string) {
				_role_set(m, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY), true)
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				_role_set(m, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY), false)
			}},
			WHITE: {Hand: func(m *ice.Message, arg ...string) { _role_white(m, arg[0], _role_keys(arg[1:]...)) }},
			BLACK: {Hand: func(m *ice.Message, arg ...string) { _role_black(m, arg[0], _role_keys(arg[1:]...)) }},
			RIGHT: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(_role_right(m, arg[0], kit.Split(_role_keys(arg[1:]...), ice.PT)...), func() { m.Echo(ice.OK) })
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME)), Hand: func(m *ice.Message, arg ...string) {
			_role_list(m, kit.Select("", arg, 0)).PushAction(mdb.DELETE)
		}},
	})
}

func RoleAction(key ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		if c, ok := ice.Info.Index[m.CommandKey()].(*ice.Context); ok && c == m.Target() {
			m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
			m.Cmd(ROLE, BLACK, VOID, m.CommandKey(), ice.ACTION)
		}
		m.Cmd(ROLE, WHITE, VOID, m.PrefixKey())
		m.Cmd(ROLE, BLACK, VOID, m.PrefixKey(), ice.ACTION)
		m.Cmd(ROLE, WHITE, VOID, strings.TrimPrefix(m.PrefixKey(), "web."))
		m.Cmd(ROLE, BLACK, VOID, strings.TrimPrefix(m.PrefixKey(), "web."), ice.ACTION)
		kit.For(key, func(key string) { m.Cmd(ROLE, WHITE, VOID, m.PrefixKey(), ice.ACTION, key) })
	}}}
}
func WhiteAction(key ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
		m.Cmd(ROLE, BLACK, VOID, m.CommandKey(), ice.ACTION)
		kit.For(key, func(key string) { m.Cmd(ROLE, WHITE, VOID, m.CommandKey(), ice.ACTION, key) })
	}}}
}
func BlackAction(key ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Cmd(ROLE, WHITE, VOID, m.CommandKey())
		kit.For(key, func(key string) { m.Cmd(ROLE, BLACK, VOID, m.CommandKey(), ice.ACTION, key) })
	}}}
}
func RoleRight(m *ice.Message, role string, key ...string) bool {
	return m.Cmdx(ROLE, RIGHT, role, key) == ice.OK
}
func Right(m *ice.Message, key ...ice.Any) bool {
	return m.Option(ice.MSG_USERROLE) == ROOT || !m.Warn(m.Cmdx(ROLE, RIGHT, m.Option(ice.MSG_USERROLE), key, logs.FileLineMeta(-1)) != ice.OK,
		ice.ErrNotRight, kit.Keys(key...), USERROLE, m.Option(ice.MSG_USERROLE), logs.FileLineMeta(-1))
}
func White(m *ice.Message, key ...string) {
	kit.For(key, func(key string) { m.Cmd(ROLE, WHITE, VOID, key) })
}
func Black(m *ice.Message, key ...string) {
	kit.For(key, func(key string) { m.Cmd(ROLE, BLACK, VOID, key) })
}
