package wiki

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _spark_show(m *ice.Message, name, text string, arg ...string) {
	if name == "" {
		_wiki_template(m, SPARK, name, text, arg...)
		return
	}

	prompt := kit.Select(name+"> ", m.Conf(SPARK, kit.Keym("prompt", name)))
	m.Echo(`<div class="story" data-type="spark" data-name="%s">`, name)
	for _, l := range strings.Split(text, "\n") {
		m.Echo("<div>")
		m.Echo("<label>").Echo(prompt).Echo("</label>")
		m.Echo("<span>").Echo(l).Echo("</span>")
		m.Echo("</div>")
	}
	m.Echo("</div>")
}

const SPARK = "spark"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			SPARK: {Name: "spark [name] text", Help: "段落", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Echo(`<br class="story" data-type="spark">`)
					return
				}

				arg = _name(m, arg)
				_spark_show(m, arg[0], strings.TrimSpace(arg[1]), arg[2:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			SPARK: {Name: SPARK, Help: "段落", Value: kit.Data(
				kit.MDB_TEMPLATE, `<p {{.OptionTemplate}}>{{.Option "text"}}</p>`,
				"prompt", kit.Dict("shell", "$ "),
			)},
		},
	})
}
