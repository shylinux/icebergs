package ctx

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _command_list(m *ice.Message, name string) {
	if name == "" { // 命令列表
		for k, v := range m.Source().Commands {
			if k[0] == '/' || k[0] == '_' {
				continue // 内部命令
			}

			m.Push(kit.MDB_KEY, k)
			m.Push(kit.MDB_NAME, v.Name)
			m.Push(kit.MDB_HELP, v.Help)
		}
		m.Sort(kit.MDB_KEY)
		return
	}

	// 命令详情
	m.Spawn(m.Source()).Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		m.Push(kit.MDB_KEY, s.Cap(ice.CTX_FOLLOW))
		m.Push(kit.MDB_NAME, kit.Format(cmd.Name))
		m.Push(kit.MDB_HELP, kit.Format(cmd.Help))
		m.Push(kit.MDB_META, kit.Formats(cmd.Meta))
		m.Push(kit.MDB_LIST, kit.Formats(cmd.List))
	})
}
func _command_search(m *ice.Message, kind, name, text string) {
	ice.Pulse.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if key[0] == '/' || key[0] == '_' {
			return // 内部命令
		}
		if name != "" && !strings.HasPrefix(key, name) && !strings.Contains(s.Name, name) {
			return
		}

		m.PushSearch("cmd", COMMAND, CONTEXT, s.Cap(ice.CTX_FOLLOW), COMMAND, key,
			kit.MDB_TYPE, kind, kit.MDB_NAME, key, kit.MDB_TEXT, s.Cap(ice.CTX_FOLLOW),
		)
	})
}

const (
	ACTION = "action"
)
const COMMAND = "command"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		COMMAND: {Name: "command key auto", Help: "命令", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == COMMAND || arg[1] != "" {
					_command_search(m, arg[0], arg[1], arg[2])
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_command_list(m, kit.Keys(arg))
		}},
	}})
}
