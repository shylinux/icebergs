package input

import (
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

type wubi struct {
	input

	short string `data:"zone"`
	store string `data:"usr/local/export/input/wubi"`
	fsize string `data:"300000"`
	limit string `data:"50000"`
	least string `data:"1000"`

	insert string `name:"insert zone=person text code weight" help:"添加"`
	load   string `name:"load file=usr/wubi-dict/wubi86 zone=wubi86" help:"加载"`
	save   string `name:"save file=usr/wubi-dict/person zone=person" help:"保存"`
	list   string `name:"list method=word,line code auto" help:"五笔"`
}

func (w wubi) Input(m *ice.Message, arg ...string) {
	if arg[0] = strings.TrimSpace(arg[0]); strings.HasPrefix(arg[0], "ice") {
		switch list := kit.Split(arg[0]); list[1] {
		case "add": // ice add 想你 shwq [person [9999]]
			m.Cmd(w, ctx.ACTION, mdb.INSERT, mdb.TEXT, list[2], cli.CODE, list[3],
				mdb.ZONE, kit.Select("person", list, 4), mdb.VALUE, kit.Select("999999", list, 5),
			)
			m.Echo(list[3] + ice.NL)
		}
		return
	}

	m.Option(mdb.CACHE_LIMIT, "10")
	m.Cmd(w, "word", arg[0]).Tables(func(value ice.Maps) {
		m.Echo(value[mdb.TEXT] + ice.NL)
	})
}

func init() { ice.CodeCtxCmd(wubi{}) }
