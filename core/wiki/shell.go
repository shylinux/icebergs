package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	kit "shylinux.com/x/toolkits"
)

func _shell_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(OUTPUT, m.Cmdx(cli.SYSTEM, "sh", "-c", m.Option(INPUT, text)))
	_wiki_template(m, SHELL, name, text, arg...)
}

const (
	INPUT  = "input"
	OUTPUT = "output"
)
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
