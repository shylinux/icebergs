package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
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

	switch name {
	case "inner", "field":
		m.Echo(text)
		return
	}

	for _, l := range strings.Split(text, ice.NL) {
		m.Echo(web.Format("div", web.Format("label", prompt), web.Format("span", l)))
	}
}

const (
	PROMPT = "prompt"
	BREAK  = "break"
)

const SPARK = "spark"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		SPARK: {Name: "spark [name] text auto field:text value:text", Help: "段落", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				ice.AddRender(ice.RENDER_SCRIPT, func(m *ice.Message, cmd string, args ...ice.Any) string {
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
						list = append(list, "<div>")
						switch arg[0] {
						case SHELL:
							list = append(list, web.Format("label", "$ "))
						default:
							list = append(list, web.Format("label", "&gt; "))
						}
						list = append(list, web.Format("span", l))
						list = append(list, "</div>")
					}
					list = append(list, "</div>")
					return strings.Join(list, "")
				})
			}},
			"md": {Name: "md file", Help: "md", Hand: func(m *ice.Message, arg ...string) {
				block, code := "", []string{}
				text := func() {
					if len(code) > 0 {
						m.Cmdy(SPARK, kit.Join(code, ice.NL))
						code = []string{}
					}
				}
				m.Cmd(nfs.CAT, m.Option(nfs.FILE), func(line string) {
					for _, ls := range [][]string{
						[]string{"# ", TITLE}, []string{"## ", TITLE, CHAPTER}, []string{"### ", TITLE, SECTION},
					} {
						if strings.HasPrefix(line, ls[0]) {
							text()
							m.Cmdy(ls[1:], strings.TrimPrefix(line, ls[0]))
							return
						}
					}

					if strings.HasPrefix(line, "```") {
						if block == "" {
							text()
							block = "```"
						} else {
							m.Cmdy(SPARK, SHELL, kit.Join(code, ice.NL))
							block, code = "", []string{}
						}
						return
					}

					switch block {
					case "":
						code = append(code, line)
					default:
						code = append(code, line)
					}
				})
				text()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Echo(`<br class="story" data-type="spark">`)
				return
			}
			switch kit.Ext(arg[0]) {
			case "md":
				m.Cmdy(SPARK, "md", arg)
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
