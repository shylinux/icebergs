package wiki

import (
	"bytes"
	"encoding/csv"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DATA = "data"

func init() {
	Index.MergeCommands(ice.Commands{
		DATA: {Name: "data path type@key field auto", Help: "数据表格", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], "比例图", "折线图")
				}
			}},
			mdb.CREATE: {Name: "create path field value", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", nfs.SAVE, m.Option(nfs.PATH), kit.Join(kit.Split(m.Option(mdb.FIELD)))+lex.NL+kit.Join(kit.Split(m.Option(mdb.VALUE)))+lex.NL)
			}},
			nfs.PUSH: {Name: "push path record", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.PUSH, path.Join(mdb.Config(m, nfs.PATH), arg[0]), kit.Join(arg[1:], mdb.FS)+lex.NL)
			}},
		}, WikiAction(ice.USR_LOCAL_EXPORT, nfs.CSV, nfs.JSON)), Hand: func(m *ice.Message, arg ...string) {
			kit.If(!_wiki_list(m, arg...), func() {
				if kit.Ext(arg[0]) == nfs.JSON {
					ctx.DisplayStoryJSON(m.Cmdy(nfs.CAT, arg[0]))
				} else {
					CSV(m, m.Cmdx(nfs.CAT, arg[0])).StatusTimeCount()
				}
			})
		}},
	})
}
func CSV(m *ice.Message, text string, head ...string) *ice.Message {
	r := csv.NewReader(bytes.NewBufferString(text))
	kit.If(len(head) == 0, func() { head, _ = r.Read() })
	for {
		if line, e := r.Read(); e != nil {
			break
		} else {
			kit.For(head, func(i int, k string) { m.Push(k, kit.Select("", line, i)) })
		}
	}
	return m
}
