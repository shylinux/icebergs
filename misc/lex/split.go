package lex

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _split_tab(text string) (tab int) {
	for _, c := range text {
		switch c {
		case '\t':
			tab += 4
		case ' ':
			tab++
		default:
			break
		}
	}
	return
}
func _split_deep(stack []int, text string) ([]int, int) {
	tab := _split_tab(text)
	for i := len(stack) - 1; i >= 0; i-- {
		kit.If(tab <= stack[i], func() { stack = stack[:len(stack)-1] })
	}
	stack = append(stack, tab)
	return stack, len(stack)
}
func _split_list(m *ice.Message, file string, arg ...string) ice.Map {
	const INDENT = "_indent"
	stack, indent := []int{}, 0
	list, line := kit.List(kit.Data(INDENT, -1)), ""
	m.Cmd(nfs.CAT, file, func(text string) {
		if strings.TrimSpace(text) == "" {
			return
		}
		if line += text; strings.Count(text, "`")%2 == 1 {
			return
		}
		if strings.HasPrefix(strings.TrimSpace(text), "# ") {
			return
		}
		stack, indent = _split_deep(stack, text)
		data := kit.Data(INDENT, indent)
		ls := kit.Split(text, m.Option(SPLIT_SPACE), m.Option(SPLIT_BLOCK), m.Option(SPLIT_QUOTE), m.Option(SPLIT_TRANS))
		switch cb := m.OptionCB(SPLIT).(type) {
		case func(int, []string):
			cb(indent, ls)
		case func(int, []string) []string:
			ls = cb(indent, ls)
		case func(int, []string, ice.Map, ice.Map):
			root, _ := kit.Value(list[0], "list.0").(ice.Map)
			cb(indent, ls, data, root)
		case func(int, []string, ice.Map) []string:
			ls = cb(indent, ls, data)
		case func([]string, ice.Map) []string:
			ls = cb(ls, data)
		case func([]string) []string:
			ls = cb(ls)
		case func([]string):
			cb(ls)
		case nil:
		default:
			m.ErrorNotImplement(cb)
		}
		kit.For(arg, func(k string) {
			kit.Value(data, kit.Keym(k), kit.Select("", ls, 0))
			kit.If(len(ls) > 0, func() { ls = ls[1:] })
		})
		kit.For(ls, func(k, v string) { kit.Value(data, kit.Keym(k), v) })
		for i := len(list) - 1; i >= 0; i-- {
			if indent > kit.Int(kit.Value(list[i], kit.Keym(INDENT))) {
				kit.Value(list[i], kit.Keys(mdb.LIST, "-2"), data)
				list = append(list, data)
				break
			}
			list = list[:len(list)-1]
		}
		line = ""
	})
	return list[0].(ice.Map)
}

const (
	SPLIT_SPACE = "split.space"
	SPLIT_BLOCK = "split.block"
	SPLIT_QUOTE = "split.quote"
	SPLIT_TRANS = "split.trans"
)
const SPLIT = "split"

func init() {
	Index.MergeCommands(ice.Commands{
		SPLIT: {Name: "split path key auto", Help: "分词", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
				m.Cmdy(nfs.DIR, arg)
				return
			}
			m.Echo(kit.Format(_split_list(m, arg[0], kit.Split(kit.Join(arg[1:]))...)))
		}},
	})
}
