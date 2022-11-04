package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	kit "shylinux.com/x/toolkits"
)

func _spark_show(m *ice.Message, name, text string, arg ...string) *ice.Message {
	if _option(m, m.CommandKey(), name, text, arg...); name == "" {
		return _wiki_template(m, name, text, arg...)
	}

	m.Echo(`<div class="story" data-type="spark" data-name="%s" style="%s">`, name, m.Option("style"))
	defer m.Echo("</div>")

	switch name {
	case "inner", "field":
		return m.Echo(text)
	}

	prompt := kit.Select(name+"> ", m.Config(kit.Keys(ssh.PROMPT, name)))
	for _, l := range kit.SplitLine(text) {
		m.Echo(Format("div", Format("label", prompt), Format("span", l)))
	}
	return m
}

const (
	PROMPT = "prompt"
	BREAK  = "break"
	SHELL  = "shell"
)

const SPARK = "spark"

func init() {
	Index.MergeCommands(ice.Commands{
		SPARK: {Name: "spark [name] text auto field:text value:text", Help: "段落", Actions: ice.MergeActions(ice.Actions{
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
					for _, l := range kit.SplitLine(strings.Join(arg[1:], ice.NL)) {
						list = append(list, Format("div", Format("label", kit.Select("&gt; ", "$ ", arg[0] == SHELL)), Format("span", l)))
					}
					return strings.Join(append(list, "</div>"), "")
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
					code = append(code, line)
				})
				text()
			}},
		}, WordAction(`<p {{.OptionTemplate}}>{{.Option "text"}}</p>`, ssh.PROMPT, kit.Dict(SHELL, "$ "))), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Echo(`<br class="story" data-type="spark">`)
				return
			}
			if kit.Ext(arg[0]) == "md" {
				m.Cmdy(SPARK, "md", arg)
				return
			}
			arg = _name(m, arg)
			_spark_show(m, arg[0], strings.TrimSpace(arg[1]), arg[2:]...)
		}},
	})
}
func Format(tag string, arg ...ice.Any) string {
	return kit.Format("<%s>%s</%s>", tag, strings.Join(kit.Simple(arg), ""), tag)
}
