package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _table_run(m *ice.Message, arg ...string) {
	list := [][]string{}
	m.Cmd(arg).Table(func(index int, value ice.Maps, head []string) {
		ls := []string{}
		kit.If(index == 0, func() { m.Options(HEAD, head) })
		kit.For(head, func(h string) { ls = append(ls, value[h]) })
		list = append(list, ls)
	})
	_wiki_template(m.Options(LIST, list), "", "", "")
}
func _table_show(m *ice.Message, text string, arg ...string) {
	head, list := []string{}, [][]string{}
	m.Cmd(nfs.CAT, "", kit.Dict(nfs.CAT_CONTENT, text), func(ls []string) {
		if len(head) == 0 {
			head = ls
			return
		}
		list = append(list, kit.Simple(ls, func(value string) string {
			if ls := kit.SplitWord(value); len(ls) > 1 {
				return kit.Format(`<span style="%s">%s</span>`, kit.JoinKV(":", ";", transArgKey(ls[1:])...), ls[0])
			}
			return value
		}))
	})
	_wiki_template(m.Options(HEAD, head, LIST, list), "", "", text, arg...)
}

const (
	HEAD = "head"
	LIST = "list"
)
const TABLE = "table"

func init() {
	Index.MergeCommands(ice.Commands{
		TABLE: {Name: "table text", Help: "表格", Actions: ice.MergeActions(ice.Actions{
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) { _table_run(m, arg...) }},
		}), Hand: func(m *ice.Message, arg ...string) { _table_show(m, arg[0], arg[1:]...) }},
	})
}

func transArgKey(arg []string) []string {
	for i := 0; i < len(arg)-1; i += 2 {
		switch arg[i] {
		case BG:
			arg[i] = html.BG_COLOR
		case FG:
			arg[i] = html.FG_COLOR
		}
	}
	return arg
}
