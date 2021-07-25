package wiki

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _other_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, OTHER, name, text, arg...)
	m.RenderTemplate(m.Conf(OTHER, kit.Keym(kit.MDB_TEMPLATE)))
}

const OTHER = "other"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			OTHER: {Name: "other [name] url", Help: "网页", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_other_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			OTHER: {Name: FIELD, Help: "网页", Value: kit.Data(
				kit.MDB_TEMPLATE, ``,
			)},
		},
	})
}
