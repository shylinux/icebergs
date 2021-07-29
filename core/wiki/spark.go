package wiki

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ssh"
	kit "github.com/shylinux/toolkits"
)

func _spark_show(m *ice.Message, name, text string, arg ...string) {
	if name == "" {
		_wiki_template(m, SPARK, name, text, arg...)
		return
	}

	prompt := kit.Select(name+"> ", m.Conf(SPARK, kit.Keym(ssh.PROMPT, name)))
	m.Echo(`<div class="story" data-type="spark" data-name="%s">`, name)
	for _, l := range strings.Split(text, "\n") {
		m.Echo("<div>")
		m.Echo("<label>").Echo(prompt).Echo("</label>")
		m.Echo("<span>").Echo(l).Echo("</span>")
		m.Echo("</div>")
	}
	m.Echo("</div>")
}

const (
	PROMPT = "prompt"
)

const SPARK = "spark"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				ice.AddRender(ice.RENDER_SCRIPT, func(m *ice.Message, cmd string, args ...interface{}) string {
					arg := kit.Simple(args...)
					if len(arg) == 1 && arg[0] != kit.SSH_BREAK {
						arg = []string{kit.SSH_SHELL, arg[0]}
					}
					list := []string{kit.Format(`<div class="story" data-type="spark" data-name="%s">`, arg[0])}
					for _, l := range strings.Split(strings.Join(arg[1:], "\n"), "\n") {
						switch list = append(list, "<div>"); arg[0] {
						case kit.SSH_SHELL:
							list = append(list, "<label>", "$ ", "</label>")
						default:
							list = append(list, "<label>", "&gt; ", "</label>")
						}
						list = append(list, "<span>", l, "</span>")
						list = append(list, "</div>")
					}
					list = append(list, "</div>")
					return strings.Join(list, "")
				})
			}},
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
				ssh.PROMPT, kit.Dict("shell", "$ "),
			)},
		},
	})
}
