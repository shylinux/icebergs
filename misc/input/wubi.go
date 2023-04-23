package input

import (
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

type wubi struct {
	input
	store string `data:"usr/local/export/"`
	field string `data:"time,id,text,code,weight"`
	fsize string `data:"100000"`
	limit string `data:"10000"`
	least string `data:"1000"`
	load  string `name:"load file=usr/wubi-dict/wubi86 zone=wubi86"`
	save  string `name:"save file=usr/wubi-dict/person zone=person"`
	list  string `name:"list method=word,line code auto load" help:"五笔"`
}

func (w wubi) Input(m *ice.Message, arg ...string) {
	if arg[0] = strings.TrimSpace(arg[0]); strings.HasPrefix(arg[0], "ice") {
		switch list := kit.Split(arg[0]); list[1] {
		case "add": // ice add 想你 shwq [9999 [person]]
			w.Insert(m, mdb.ZONE, kit.Select("person", list, 5), mdb.TEXT, list[2], cli.CODE, list[3], mdb.VALUE, kit.Select("999999", list, 4))
			m.Echo(list[3] + lex.NL)
		}
		return
	}
	m.Cmd(w, WORD, arg[0], func(value ice.Maps) { m.Echo(value[mdb.TEXT] + lex.NL) })
}

func init() { ice.CodeCtxCmd(wubi{}) }
