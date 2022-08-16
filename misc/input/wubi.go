package input

import (
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

type wubi struct {
	input
	short string `data:"zone"`
	field string `data:"time,zone,id,text,code,weight"`
	store string `data:"usr/local/export/"`
	fsize string `data:"100000"`
	limit string `data:"10000"`
	least string `data:"1000"`

	load string `name:"load file=usr/wubi-dict/wubi86 zone=wubi86" help:"加载"`
	save string `name:"save file=usr/wubi-dict/person zone=person" help:"保存"`
	list string `name:"list method=word,line code auto" help:"五笔"`
}

func (w wubi) Input(m *ice.Message, arg ...string) {
	if arg[0] = strings.TrimSpace(arg[0]); strings.HasPrefix(arg[0], "ice") {
		switch list := kit.Split(arg[0]); list[1] {
		case "add": // ice add 想你 shwq [9999 [person]]
			m.Cmd(w, mdb.INSERT, mdb.TEXT, list[2], cli.CODE, list[3], mdb.VALUE, kit.Select("999999", list, 4), mdb.ZONE, kit.Select("person", list, 5))
			m.Echo(list[3] + ice.NL)
		}
		return
	}

	m.Cmd(w, WORD, arg[0], func(value ice.Maps) {
		m.Echo(value[mdb.TEXT] + ice.NL)
	})
}

func init() { ice.CodeCtxCmd(wubi{}) }
