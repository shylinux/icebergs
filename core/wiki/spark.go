package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	kit "shylinux.com/x/toolkits"
)

func _spark_show(m *ice.Message, name, text string, arg ...string) {
	for i := 0; i < len(arg); i += 2 {
		m.Option(arg[i], arg[i+1])
	}

	if name == "" {
		_wiki_template(m, SPARK, name, text, arg...)
		return
	}

	prompt := kit.Select(name+"> ", m.Config(kit.Keys(ssh.PROMPT, name)))
	m.Echo(`<div class="story" data-type="spark" data-name="%s" style="%s">`, name, m.Option("style"))
	defer m.Echo("</div>")

	if name == "inner" {
		m.Echo(text)
		return
	}

	for _, l := range strings.Split(text, ice.NL) {
		m.Echo("<div>")
		m.Echo("<label>").Echo(prompt).Echo("</label>")
		m.Echo("<span>").Echo(l).Echo("</span>")
		m.Echo("</div>")
	}
}

const (
	PROMPT = "prompt"
	BREAK  = "break"
)

const SPARK = "spark"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			ice.AddRender(ice.RENDER_SCRIPT, func(m *ice.Message, cmd string, args ...interface{}) string {
				arg := kit.Simple(args...)
				if m.IsCliUA() {
					if len(arg) > 1 {
						arg = arg[1:]
					}
					return strings.Join(arg, ice.NL)
				}
				if len(arg) == 1 && arg[0] != BREAK {
					arg = []string{SHELL, arg[0]}
				}
				list := []string{kit.Format(`<div class="story" data-type="spark" data-name="%s">`, arg[0])}
				for _, l := range strings.Split(strings.Join(arg[1:], ice.NL), ice.NL) {
					switch list = append(list, "<div>"); arg[0] {
					case SHELL:
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
		SPARK: {Name: "spark [name] text auto field:text value:text", Help: "段落", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Echo(`<br class="story" data-type="spark">`)
				return
			}

			arg = _name(m, arg)
			_spark_show(m, arg[0], strings.TrimSpace(arg[1]), arg[2:]...)
		}},
	}, Configs: map[string]*ice.Config{
		SPARK: {Name: SPARK, Help: "段落", Value: kit.Data(
			nfs.TEMPLATE, `<p {{.OptionTemplate}}>{{.Option "text"}}</p>`,
			ssh.PROMPT, kit.Dict(SHELL, "$ "),
		)},
	}})
}
