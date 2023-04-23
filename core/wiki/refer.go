package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _refer_show(m *ice.Message, text string, arg ...string) {
	list := [][]string{}
	for _, v := range kit.SplitLine(text) {
		if ls := kit.SplitWord(v); len(ls) == 1 {
			p := kit.QueryUnescape(ls[0])
			list = append(list, []string{kit.Select(ls[0], path.Base(strings.Split(p, mdb.QS)[0])), ls[0], p})
		} else {
			list = append(list, append(ls, kit.QueryUnescape(ls[1])))
		}
	}
	_wiki_template(m.Options(mdb.LIST, list), "", "", text, arg...)
}

const REFER = "refer"

func init() {
	Index.MergeCommands(ice.Commands{
		REFER: {Name: "refer text", Help: "参考", Hand: func(m *ice.Message, arg ...string) {
			_refer_show(m, arg[0], arg[1:]...)
		}},
	})
}
