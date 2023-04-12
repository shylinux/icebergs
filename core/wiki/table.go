package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _table_run(m *ice.Message, arg ...string) {
	list := [][]string{}
	m.Cmd(arg).Table(func(index int, value ice.Maps, head []string) {
		kit.If(index == 0, func() { m.Optionv("head", head) })
		line := []string{}
		kit.For(head, func(h string) { line = append(line, value[h]) })
		list = append(list, line)
	})
	_wiki_template(m.Options("list", list), "", "", "")
}
func _table_show(m *ice.Message, text string, arg ...string) {
	head, list := []string{}, [][]string{}
	for i, line := range kit.SplitLine(text) {
		if line = strings.Replace(line, "%", "%%", -1); i == 0 {
			head = kit.SplitWord(line)
			continue
		}
		list = append(list, kit.Simple(kit.SplitWord(line), func(value string) string {
			if ls := kit.SplitWord(value); len(ls) > 1 {
				return kit.Format(`<span style="%s">%s</span>`, kit.JoinKV(":", ";", transArgKey(ls[1:])...), ls[0])
			}
			return value
		}))
	}
	_wiki_template(m.Options("head", head, "list", list), "", "", text, arg...)
}
func transArgKey(arg []string) []string {
	for i := 0; i < len(arg)-1; i += 2 {
		switch arg[i] {
		case BG:
			arg[i] = "background-color"
		case FG:
			arg[i] = "color"
		}
	}
	return arg
}

const TABLE = "table"

func init() {
	Index.MergeCommands(ice.Commands{
		TABLE: {Name: "table text", Help: "表格", Actions: ice.MergeActions(ice.Actions{
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) { _table_run(m, arg...) }},
		}), Hand: func(m *ice.Message, arg ...string) { _table_show(m, arg[0], arg[1:]...) }},
	})
}
