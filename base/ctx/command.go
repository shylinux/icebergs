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
		p.Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
			m.Push("key", s.Cap(ice.CTX_FOLLOW))
			m.Push("name", kit.Format(cmd.Name))
			m.Push("help", kit.Simple(cmd.Help)[0])
			m.Push("meta", kit.Formats(cmd.Meta))
			m.Push("list", kit.Formats(cmd.List))
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

	for _, k := range list {
		v := p.Target().Commands[k]
		m.Push("key", k)
		m.Push("name", kit.Format(v.Name))
		m.Push("help", kit.Simple(v.Help)[0])
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
