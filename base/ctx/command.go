package ctx

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"sort"
	"strings"
)

func _command_list(m *ice.Message, name string) {
	p := m.Spawn(m.Source())
	if name != "" {
		// 命令详情
		p.Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
			m.Push(kit.MDB_KEY, s.Cap(ice.CTX_FOLLOW))
			m.Push(kit.MDB_NAME, kit.Format(cmd.Name))
			m.Push(kit.MDB_HELP, kit.Simple(cmd.Help)[0])
			m.Push(kit.MDB_META, kit.Formats(cmd.Meta))
			m.Push(kit.MDB_LIST, kit.Formats(cmd.List))
		})
		return
	}

	list := []string{}
	for k := range p.Target().Commands {
		if k[0] == '/' || k[0] == '_' {
			continue // 内部命令
		}
		list = append(list, k)
	}
	sort.Strings(list)

	// 命令列表
	for _, k := range list {
		v := p.Target().Commands[k]
		m.Push(kit.MDB_KEY, k)
		m.Push(kit.MDB_NAME, kit.Format(v.Name))
		m.Push(kit.MDB_HELP, kit.Simple(v.Help)[0])
	}
}

const COMMAND = "command"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			COMMAND: {Name: "command key auto", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_command_list(m, strings.Join(arg, "."))
			}},
		},
	}, nil)
}