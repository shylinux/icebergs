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
	if _key := kit.Select("", strings.Split(key[0], ice.PT), -1); _key != "" {
		if c, ok := ice.Info.Index[_key].(*ice.Context); ok && kit.Keys(c.Prefix(), _key) == key[0] {
			key[0] = _key
		}
	}
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(strings.ReplaceAll(path.Join(strings.ReplaceAll(kit.Keys(key), ice.PT, ice.PS)), ice.PS, ice.PT), ice.PT), ice.PT), "web.")
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
		kit.If(white[kit.Join(key[:i+1], ice.PT)], func() { ok = true })
		kit.If(black[kit.Join(key[:i+1], ice.PT)], func() { ok = false })
	}
	return ok
}
func _role_right(m *ice.Message, role string, key ...string) (ok bool) {
	return role == ROOT || len(mdb.HashSelectDetails(m, kit.Select(VOID, role), func(value ice.Map) bool { return _role_check(value, key, role == TECH) })) > 0
}
func _role_list(m *ice.Message, role string, arg ...string) *ice.Message {
	mdb.HashSelectDetail(m, kit.Select(VOID, role), func(value ice.Map) {
		kit.For(value[WHITE], func(k string, v ice.Any) {
			if len(arg) == 0 || k == arg[0] {
				m.Push(ROLE, kit.Value(value, mdb.NAME)).Push(mdb.ZONE, WHITE).Push(mdb.KEY, k).Push(mdb.STATUS, v)
			}
		})
		kit.For(value[BLACK], func(k string, v ice.Any) {
			if len(arg) == 0 || k == arg[0] {
				m.Push(ROLE, kit.Value(value, mdb.NAME)).Push(mdb.ZONE, BLACK).Push(mdb.KEY, k).Push(mdb.STATUS, v)
			}
		})
	})
	return m.Sort(mdb.KEY)
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
const (
	AUTH    = "auth"
	ACCESS  = "access"
	PUBLIC  = "public"
	PRIVATE = "private"
	CONFIRM = "confirm"
)
const ROLE = "role"

func init() {
	Index.MergeCommands(ice.Commands{
		ROLE: {Name: "role name key auto", Help: "角色", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(ROLE, mdb.CREATE, VOID, TECH)
				has := map[string]bool{VOID: true, TECH: true}
				m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
					role := map[string][]string{}
					kit.For(kit.Split(cmd.Role), func(k string) { role[k] = []string{} })
					for sub, action := range cmd.Actions {
						kit.For(kit.Split(action.Role), func(k string) { role[k] = append(role[k], sub) })
					}
					kit.For(role, func(role string, list []string) {
						kit.If(!has[role], func() { m.Cmd(ROLE, mdb.CREATE, role); has[role] = true })
						roleHandle(m, role, list...)
					})
				})
				m.Cmd(ROLE, WHITE, VOID, ROLE, "action", RIGHT)
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.KEY {
					mdb.HashInputs(m, ice.INDEX).CutTo(ice.INDEX, arg[0])
				}
			}},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				kit.For(arg, func(role string) {
					mdb.Rich(m, ROLE, nil, kit.Dict(mdb.NAME, role, BLACK, kit.Dict(), WHITE, kit.Dict()))
				})
			}},
			mdb.INSERT: {Name: "insert role*=void zone*=white,black key*", Hand: func(m *ice.Message, arg ...string) {
				_role_set(m, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY), true)
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				_role_set(m, m.Option(ROLE), m.Option(mdb.ZONE), m.Option(mdb.KEY), false)
			}},
			WHITE: {Hand: func(m *ice.Message, arg ...string) { _role_white(m, arg[0], _role_keys(arg[1:]...)) }},
			BLACK: {Hand: func(m *ice.Message, arg ...string) { _role_black(m, arg[0], _role_keys(arg[1:]...)) }},
			RIGHT: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 2 {
					m.Search(arg[1], func(key string, cmd *ice.Command) {
						if _, ok := cmd.Actions[arg[2]]; ok {
							arg = kit.Simple(arg[0], arg[1], ice.ACTION, arg[2:])
						}
					})
				}
				for _, role := range kit.AddUniq(kit.Split(arg[0]), VOID) {
					if _role_right(m, role, kit.Split(_role_keys(arg[1:]...), ice.PT)...) {
						m.Echo(ice.OK)
						break
					}
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...)
			} else {
				_role_list(m, kit.Select("", arg, 0), kit.Slice(arg, 1)...)
			}
		}},
	})
}
func roleHandle(m *ice.Message, role string, key ...string) {
	cmd := m.ShortKey()
	role = kit.Select(VOID, role)
	m.Cmd(ROLE, WHITE, role, cmd)
	if cmd == "header" {
		return
	}
	m.Cmd(ROLE, BLACK, role, cmd, ice.ACTION)
	kit.For(key, func(key string) { m.Cmd(ROLE, WHITE, role, cmd, ice.ACTION, key) })
}
func WhiteAction(role string, key ...string) ice.Actions {
	role = kit.Select(VOID, role)
	return ice.Actions{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		cmd := m.CommandKey()
		m.Cmd(ROLE, WHITE, role, cmd)
		if cmd == "header" {
			return
		}
		m.Cmd(ROLE, BLACK, role, cmd, ice.ACTION)
		kit.For(key, func(key string) { m.Cmd(ROLE, WHITE, role, cmd, ice.ACTION, key) })
	}}}
}
func White(m *ice.Message, key ...string) {
	kit.For(key, func(key string) { m.Cmd(ROLE, WHITE, VOID, key) })
}
func Black(m *ice.Message, key ...string) {
	kit.For(key, func(key string) { m.Cmd(ROLE, BLACK, VOID, key) })
}
func Right(m *ice.Message, key ...ice.Any) bool {
	if key := kit.Simple(key); len(key) > 2 && key[1] == ice.ACTION && kit.IsIn(kit.Format(key[2]), ice.RUN, ice.COMMAND) {
		return true
	} else if len(key) > 0 && key[0] == ice.ETC_PATH {
		return true
	}
	// m.Option(ice.MSG_TITLE, kit.Keys(m.Option(ice.MSG_USERPOD), m.CommandKey(), m.ActionKey())+" "+logs.FileLine(-2))
	return !ice.Info.Important || m.Option(ice.MSG_USERROLE) == ROOT || !m.WarnNotRight(m.Cmdx(ROLE, RIGHT, m.Option(ice.MSG_USERROLE), key, logs.FileLineMeta(-1)) != ice.OK,
		kit.Keys(key...), USERROLE, m.Option(ice.MSG_USERROLE))
}
func IsTechOrRoot(m *ice.Message) bool { return kit.IsIn(m.Option(ice.MSG_USERROLE), TECH, ROOT) }
