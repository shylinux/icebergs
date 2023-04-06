package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _refer_show(m *ice.Message, text string, arg ...string) {
	list := [][]string{}
	for _, v := range kit.SplitLine(text) {
		if ls := kit.SplitWord(v); len(ls) == 1 {
			p := kit.QueryUnescape(ls[0])
			list = append(list, []string{kit.Select(ls[0], path.Base(p)), ls[0], p})
		} else {
			list = append(list, append(ls, kit.QueryUnescape(ls[1])))
		}
	}
	m.Optionv(mdb.LIST, list)
	_wiki_template(m, "", text, arg...)
}

const REFER = "refer"

func init() {
	Index.MergeCommands(ice.Commands{
		REFER: {Name: "refer text", Help: "参考", Actions: WordAction(
			`<ul {{.OptionTemplate}}>{{range $index, $value := .Optionv "list"}}<li>{{index $value 0}}: <a href="{{index $value 1}}" data-name="{{index $value 0}}" target="_blank">{{index $value 2}}</a></li>{{end}}</ul>`,
		), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 1 {
				_refer_show(m, arg[0], arg[1:]...)
			}
		}},
	})
}
