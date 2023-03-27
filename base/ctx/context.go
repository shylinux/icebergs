package ctx

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _context_list(m *ice.Message, sub *ice.Context, name string) {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if name != "" && name != ice.ICE && !strings.HasPrefix(s.Prefix(), name+ice.PT) {
			return
		}
		m.Push(mdb.NAME, s.Prefix()).Push(mdb.HELP, s.Help)
	})
}

const CONTEXT = "context"

func init() {
	Index.MergeCommands(ice.Commands{
		CONTEXT: {Name: "context name=web action=context,command,config key auto", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 0, func() { arg = append(arg, m.Source().Prefix()) })
			m.Search(arg[0]+ice.PT, func(p *ice.Context, s *ice.Context) {
				msg := m.Spawn(s)
				defer m.Copy(msg)
				switch kit.Select(CONTEXT, arg, 1) {
				case CONTEXT:
					_context_list(msg, s, arg[0])
				case COMMAND:
					msg.Cmdy(COMMAND, arg[2:])
				case CONFIG:
					msg.Cmdy(CONFIG, arg[2:])
				}
			})
		}},
	})
}
