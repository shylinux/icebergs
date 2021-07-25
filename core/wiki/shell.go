package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	kit "github.com/shylinux/toolkits"
)

func _shell_show(m *ice.Message, name, text string, arg ...string) {
	m.Option("output", m.Cmdx(cli.SYSTEM, "sh", "-c", m.Option("input", text)))
	_option(m, SHELL, name, text, arg...)
	m.RenderTemplate(m.Conf(SHELL, kit.Keym(kit.MDB_TEMPLATE)))
}

const SHELL = "shell"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			SHELL: {Name: "shell [name] cmd", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_shell_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			SHELL: {Name: SHELL, Help: "命令", Value: kit.Data(
				kit.MDB_TEMPLATE, `<code {{.OptionTemplate}}>$ {{.Option "input"}} # {{.Option "name"}}
{{.Option "output"}}</code>`,
			)},
		},
	})
}
