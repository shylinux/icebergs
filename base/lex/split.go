package lex

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _split_deep(m *ice.Message, text string) (deep int) {
	for _, c := range text {
		switch c {
		case '\t':
			deep += 4
		case ' ':
			deep++
		default:
			return
		}
	}
	return
}
func _split_list(m *ice.Message, file string, arg ...string) {
	const DEEP = "_deep"
	list := kit.List(kit.Data(DEEP, -1))
	m.Cmd(nfs.CAT, file, func(text string) {
		// if text = kit.Split(text, "#", "#")[0]; strings.TrimSpace(text) == "" {
		// 	return
		// }
		if strings.HasPrefix(strings.TrimSpace(text), "# ") {
			return
		}
		if strings.TrimSpace(text) == "" {
			return
		}

		deep := _split_deep(m, text)
		data := kit.Data(DEEP, deep)

		ls := kit.Split(text)
		switch cb := m.OptionCB(SPLIT).(type) {
		case func([]string, map[string]interface{}) []string:
			ls = cb(ls, data)
		case func(int, []string, map[string]interface{}) []string:
			ls = cb(deep, ls, data)
		}

		for _, k := range arg {
			if kit.Value(data, kit.Keym(k), kit.Select("", ls, 0)); len(ls) > 0 {
				ls = ls[1:]
			}
		}

		for i := 0; i < len(ls)-1; i += 2 {
			kit.Value(data, kit.Keym(ls[i]), ls[i+1])
		}

		for i := len(list) - 1; i >= 0; i-- {
			if deep > kit.Int(kit.Value(list[i], kit.Keym(DEEP))) {
				kit.Value(list[i], "list.-2", data)
				list = append(list, data)
				break
			}
			list = list[:len(list)-1]
		}
	})
	m.Echo(kit.Format(list[0]))
}

const SPLIT = "split"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SPLIT: {Name: "split", Help: "解析", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		SPLIT: {Name: "split path key auto", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
				m.Cmdy(nfs.DIR, arg)
				return
			}

			_split_list(m, arg[0], arg[1:]...)
			m.ProcessDisplay("/plugin/local/wiki/json.js")
		}},
	}})
}
