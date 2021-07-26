package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

func _local_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(INPUT, m.Cmdx(nfs.CAT, text))
	_wiki_template(m, LOCAL, name, text, arg...)
}

const LOCAL = "local"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			LOCAL: {Name: "local [name] file", Help: "文件", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_local_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			LOCAL: {Name: LOCAL, Help: "文件", Value: kit.Data(
				kit.MDB_TEMPLATE, `<code {{.OptionTemplate}}>{{range $index, $value := .Optionv "input"}}{{$value}}{{end}}</code>`,
			)},
		},
	})
}
