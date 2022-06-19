package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _refer_show(m *ice.Message, text string, arg ...string) {
	list := [][]string{}
	for _, v := range kit.Split(strings.TrimSpace(text), ice.NL, ice.NL) {
		if ls := kit.Split(v, " ", " "); len(ls) == 1 {
			list = append(list, []string{path.Base(ls[0]), ls[0]})
		} else {
			list = append(list, ls)
		}
	}
	m.Optionv(mdb.LIST, list)
	_wiki_template(m, REFER, "", text, arg...)
}

const REFER = "refer"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		REFER: {Name: "refer `[[name] url\n]...`", Help: "参考", Hand: func(m *ice.Message, arg ...string) {
			_refer_show(m, arg[0], arg[1:]...)
		}},
	}, Configs: map[string]*ice.Config{
		REFER: {Name: REFER, Help: "参考", Value: kit.Data(
			nfs.TEMPLATE, `<ul {{.OptionTemplate}}>{{range $index, $value := .Optionv "list"}}<li>{{index $value 0}}: <a href="{{index $value 1}}" data-name="{{index $value 0}}" target="_blank">{{index $value 1}}</a></li>{{end}}</ul>`,
		)},
	}})
}
