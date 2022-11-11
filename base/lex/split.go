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
			return
		}
	}
	return
}
func _split_deep(stack []int, text string) ([]int, int) {
	tab := _split_tab(text)
	for i := len(stack) - 1; i >= 0; i-- {
		if tab <= stack[i] {
			stack = stack[:len(stack)-1]
		}
	}
	stack = append(stack, tab)
	return stack, len(stack)
}
func _split_list(m *ice.Message, file string, arg ...string) ice.Map {
	const DEEP = "_deep"
	stack, deep := []int{}, 0
	list, line := kit.List(kit.Data(DEEP, -1)), ""
	m.Cmd(nfs.CAT, file, func(text string) {
		if strings.TrimSpace(text) == "" {
			return // 空行
		}
		if line += text; strings.Count(text, "`")%2 == 1 {
			return // 多行
		}
		if strings.HasPrefix(strings.TrimSpace(text), "# ") {
			return // 注释
		}

		stack, deep = _split_deep(stack, text)
		data := kit.Data(DEEP, deep)

		// 回调函数
		ls := kit.Split(text, m.Option(SPLIT_SPACE), m.Option(SPLIT_BLOCK), m.Option(SPLIT_QUOTE), m.Option(SPLIT_TRANS))
		switch cb := m.OptionCB(SPLIT).(type) {
		case func(int, []string):
			cb(deep, ls)
		case func(int, []string) []string:
			ls = cb(deep, ls)
		case func(int, []string, ice.Map, ice.List):

		case func(int, []string, ice.Map, ice.Map):
			root, _ := kit.Value(list[0], "list.0").(ice.Map)
			cb(deep, ls, data, root)
		case func(int, []string, ice.Map) []string:
			ls = cb(deep, ls, data)
		case func([]string, ice.Map) []string:
			ls = cb(ls, data)
		case func([]string):
			cb(ls)
		case func([]string) []string:
			ls = cb(ls)
		case nil:
		default:
			m.ErrorNotImplement(cb)
		}

		// 参数字段
		for _, k := range arg {
			if kit.Value(data, kit.Keym(k), kit.Select("", ls, 0)); len(ls) > 0 {
				ls = ls[1:]
			}
		}

		// 属性字段
		for i := 0; i < len(ls)-1; i += 2 {
			kit.Value(data, kit.Keym(ls[i]), ls[i+1])
		}

		// 查找节点
		for i := len(list) - 1; i >= 0; i-- {
			if deep > kit.Int(kit.Value(list[i], kit.Keym(DEEP))) {
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

func Split(m *ice.Message, arg ...string) ice.Map {
	return kit.Value(_split_list(m, arg[0], arg[1:]...), kit.Keys(mdb.LIST, "0")).(ice.Map)
}
