package wiki

import (
	"bytes"
	"encoding/csv"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DATA = "data"

func init() {
	Index.MergeCommands(ice.Commands{
		DATA: {Name: "data path type@key fields auto create push save draw", Help: "数据表格", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], "折线图", "比例图")
				}
			}},
			mdb.CREATE: {Name: "create path fields value", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", nfs.SAVE, m.Option(nfs.PATH), kit.Join(kit.Split(m.Option("fields")))+ice.NL+kit.Join(kit.Split(m.Option(mdb.VALUE)))+ice.NL)
			}},
			nfs.PUSH: {Name: "push path record", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.PUSH, path.Join(m.Config(nfs.PATH), arg[0]), kit.Join(arg[1:], ice.FS)+ice.NL)
			}},
			"draw": {Name: "draw", Help: "绘图"},
		}, WikiAction(ice.USR_LOCAL_EXPORT, nfs.CSV)), Hand: func(m *ice.Message, arg ...string) {
			if !_wiki_list(m, arg...) {
				CSV(m, m.Cmdx(nfs.CAT, arg[0])).StatusTimeCount()
			}
		}},
	})
}
func CSV(m *ice.Message, text string, head ...string) *ice.Message {
	bio := bytes.NewBufferString(text)
	r := csv.NewReader(bio)

	if len(head) == 0 {
		head, _ = r.Read()
	}
	for {
		line, e := r.Read()
		if e != nil {
			break
		}
		for i, k := range head {
			m.Push(k, kit.Select("", line, i))
		}
	}
	return m
}
