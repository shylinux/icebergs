package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _local_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(INPUT, m.Cmdx(nfs.CAT, text))
	_wiki_template(m, LOCAL, name, text, arg...)
}

const LOCAL = "local"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		LOCAL: {Name: "local [name] file", Help: "文件", Hand: func(m *ice.Message, arg ...string) {
			arg = _name(m, arg)
			_local_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}, Configs: map[string]*ice.Config{
		LOCAL: {Name: LOCAL, Help: "文件", Value: kit.Data(
			nfs.TEMPLATE, `<code {{.OptionTemplate}}>{{range $index, $value := .Optionv "input"}}{{$value}}{{end}}</code>`,
		)},
	}})
}
