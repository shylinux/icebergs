package ctx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"sort"
	"strings"
)

func _command_list(m *ice.Message, all bool, name string) {
	p := m.Spawn(m.Source())
	if all {
		p = ice.Pulse
	}

	if name != "" {
		p.Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
			m.Push("key", s.Cap(ice.CTX_FOLLOW))
			m.Push("name", kit.Format(cmd.Name))
			m.Push("help", kit.Simple(cmd.Help)[0])
			m.Push("meta", kit.Format(cmd.Meta))
			if len(cmd.List) == 0 {
				_command_make(m, cmd)
			}
			m.Push("list", kit.Format(cmd.List))
		})
		return
	}

	// 命令列表
	p.Travel(func(p *ice.Context, s *ice.Context) {
		list := []string{}
		for k := range s.Commands {
			if k[0] == '/' || k[0] == '_' {
				// 内部命令
				continue
			}
			list = append(list, k)
		}
		sort.Strings(list)

		for _, k := range list {
			v := s.Commands[k]
			m.Push("key", s.Cap(ice.CTX_FOLLOW))
			m.Push("name", kit.Format(v.Name))
			m.Push("help", kit.Simple(v.Help)[0])
			m.Push("meta", kit.Format(v.Meta))
			if len(v.List) == 0 {
				_command_make(m, v)
			}
			m.Push("list", kit.Format(v.List))
		}
	})
}
func _command_make(m *ice.Message, cmd *ice.Command) {
	var list []string
	switch name := cmd.Name.(type) {
	case []string, []interface{}:
		list = kit.Split(kit.Simple(name)[0])
	default:
		list = kit.Split(strings.Split(kit.Format(name), ";")[0])
	}

	button := false
	for i, v := range list {
		if i == 0 {
			continue
		}
		switch ls := kit.Split(v, ":="); ls[0] {
		case "[", "]":
		case "auto":
			cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "查看", "value", "auto")...)
			cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "返回", "value", "Last")...)
			button = true
		default:
			kind, value := "text", ""
			if len(ls) == 3 {
				kind, value = ls[1], ls[2]
			} else if len(ls) == 2 {
				if strings.Contains(v, "=") {
					value = ls[1]
				} else {
					kind = ls[1]
				}
			}
			if kind == "button" {
				button = true
			}
			cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, kind, "name", ls[0], "value", value)...)
		}
	}
	if len(cmd.List) == 0 {
		cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "text", "name", "name")...)
	}
	if !button {
		cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "查看")...)
		cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "返回")...)
	}
}

const COMMAND = "command"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			COMMAND: {Name: "command [all] command", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				all, arg := _parse_arg_all(m, arg...)
				_command_list(m, all, kit.Select("", arg, 0))
			}},
		},
	}, nil)
}
