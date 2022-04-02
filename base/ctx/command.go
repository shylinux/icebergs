package ctx

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _command_list(m *ice.Message, name string) {
	if strings.HasPrefix(name, "can.") {
		m.Push(mdb.INDEX, name)
		return
	}
	if name == "" { // 命令列表
		for k, v := range m.Source().Commands {
			if k[0] == '/' || k[0] == '_' {
				continue // 内部命令
			}

			m.Push(mdb.KEY, k)
			m.Push(mdb.NAME, v.Name)
			m.Push(mdb.HELP, v.Help)
		}
		m.Sort(mdb.KEY)
		return
	}

	// 命令详情
	m.Spawn(m.Source()).Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		m.Push(mdb.INDEX, kit.Keys(s.Cap(ice.CTX_FOLLOW), key))
		m.Push(mdb.NAME, kit.Format(cmd.Name))
		m.Push(mdb.HELP, kit.Format(cmd.Help))
		m.Push(mdb.META, kit.Format(cmd.Meta))
		m.Push(mdb.LIST, kit.Format(cmd.List))
	})
}
func _command_search(m *ice.Message, kind, name, text string) {
	m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if key[0] == '/' || key[0] == '_' {
			return // 内部命令
		}
		if name != "" && !strings.HasPrefix(key, name) && !strings.Contains(s.Name, name) {
			return
		}

		m.PushSearch(ice.CTX, kit.PathName(1), ice.CMD, kit.FileName(1),
			kit.SimpleKV("", s.Cap(ice.CTX_FOLLOW), cmd.Name, cmd.Help),
			CONTEXT, s.Cap(ice.CTX_FOLLOW), COMMAND, key,
			INDEX, kit.Keys(s.Cap(ice.CTX_FOLLOW), key),
		)
	})
}

func CmdAction(fields ...string) map[string]*ice.Action {
	return ice.SelectAction(map[string]*ice.Action{
		COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
			if !m.PodCmd(COMMAND, arg) {
				m.Cmdy(COMMAND, arg)
			}
		}},
		ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
			if m.Right(arg) && !m.PodCmd(arg) {
				m.Cmdy(arg)
			}
		}},
	}, fields...)
}

const (
	ACTION  = "action"
	INDEX   = "index"
	ARGS    = "args"
	STYLE   = "style"
	DISPLAY = "display"
)
const COMMAND = "command"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		COMMAND: {Name: "command key auto", Help: "命令", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, m.CommandKey(), m.PrefixKey())
			}},
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == m.CommandKey() || len(arg) > 1 && arg[1] != "" {
					_command_search(m, arg[0], kit.Select("", arg, 1), kit.Select("", arg, 2))
				}
			}},
			INDEX: {Name: "index", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "")
			}
			for _, key := range arg {
				_command_list(m, key)
			}
		}},
	}})
}
