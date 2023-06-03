package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _spark_md(m *ice.Message, arg ...string) *ice.Message {
	block, code := "", []string{}
	text := func() {
		if len(code) > 0 {
			m.Cmdy(SPARK, kit.Join(code, lex.NL))
			code = []string{}
		}
	}
	m.Cmd(nfs.CAT, arg[0], func(line string) {
		for _, ls := range [][]string{[]string{"# ", TITLE}, []string{"## ", TITLE, CHAPTER}, []string{"### ", TITLE, SECTION}} {
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
				m.Cmdy(SPARK, SHELL, kit.Join(code, lex.NL))
				block, code = "", []string{}
			}
			return
		}
		code = append(code, line)
	})
	text()
	return m
}
func _spark_show(m *ice.Message, name, text string, arg ...string) *ice.Message {
	return _wiki_template(m.Options(mdb.LIST, kit.SplitLine(text)), name, name, text, arg...)
}
func _spark_tabs(m *ice.Message, arg ...string) {
	m.Echo(`<div class="story" data-type="spark_tabs">`)
	{
		m.Echo(`<div class="tabs">`)
		{
			kit.For(arg[1:], func(k, v string) { m.Echo(`<div class="item">%s</div>`, k) })
		}
		m.Echo(`</div>`)
		kit.For(arg[1:], func(k, v string) { m.Cmdy("", arg[0], v) })
	}
	m.Echo(`</div>`)
}

const (
	SHELL = "shell"
)

const SPARK = "spark"

func init() {
	Index.MergeCommands(ice.Commands{
		SPARK: {Name: "spark type=inner,shell,redis,mysql text", Help: "段落", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				ice.AddRender(ice.RENDER_SCRIPT, func(msg *ice.Message, args ...ice.Any) string { return m.Cmdx(SPARK, SHELL, args) })
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			if kit.Ext(arg[0]) == "md" {
				_spark_md(m, arg...)
			} else if arg[0] == SHELL && kit.IsIn(kit.Select("", arg, 1), cli.LINUX, cli.MACOS, cli.DARWIN, cli.WINDOWS) {
				_spark_tabs(m, arg...)
			} else {
				arg = _name(m, arg)
				_spark_show(m, arg[0], strings.TrimSpace(arg[1]), arg[2:]...)
			}
		}},
	})
}
