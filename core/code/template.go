package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const TEMPLATE = "template"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TEMPLATE: {Name: "template name auto", Help: "模板", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, _template := range _template_list {
					m.Cmd(TEMPLATE, mdb.CREATE, kit.SimpleKV(kit.Format(_template[0]), _template[1:]...))
				}
			}},
			mdb.CREATE: {Name: "create type name text args", Help: "创建"},
			nfs.DEFS: {Name: "defs file=hi/hi.js", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				m.Option("tags", "`"+m.Option("tags")+"`")
				if buf, err := kit.Render(m.Option(mdb.TEXT), m); !m.Warn(err) {
					switch m.Cmd(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), string(buf)); kit.Ext(m.Option(nfs.FILE)) {
					case GO:
						if m.Option(cli.MAIN) != "" && m.Option(mdb.ZONE) != "" {
							_autogen_import(m, path.Join(m.Option(nfs.PATH), m.Option(cli.MAIN)), m.Option(mdb.ZONE), _autogen_mod(m, ice.GO_MOD))
						}
					default:
						m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Option(mdb.TEXT))
					}
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text,args")), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.HashSelect(m, arg...).Sort(mdb.NAME); len(arg) == 0 {
				m.Cut("time,action,type,name,text,args")
				m.Action(mdb.CREATE)
			}
			m.PushAction(nfs.DEFS, mdb.REMOVE)
		}}},
	})
}

var _template_list = [][]interface{}{
	[]interface{}{"", "txt", "网站索引", `
hi
	hi
		cli.qrcode
		cli.system
		cli.runtime

`},
	[]interface{}{"", "js", "前端模块", `Volcanos("onimport", {help: "导入数据", list: [], _init: function(can, msg, cb, target) {
	can.onmotion.clear(can)
	can.onappend.table(can, msg)
	can.onappend.board(can, msg)
}})`},
	[]interface{}{"", "go", "后端模块", `package {{.Option "zone"}}

import (
	"shylinux.com/x/ice"
)

type {{.Option "name"}} struct {
	ice.{{.Option "type"}}

	list string {{.Option "tags"}}
}

func (h {{.Option "name"}}) List(m *ice.Message, arg ...string) {
	h.{{.Option "type"}}.List(m, arg...)
}

func init() { ice.Cmd("{{.Option "key"}}", {{.Option "name"}}{}) }
`, "args", `[
	{"name": "zone", "value": "hi"},
	{"name": "name", "value": "hi"},
	{"name": "key", "value": "web.code.hi.hi"},
	{"name": "type", "values": "Hash,Zone,List"},
	{"name": "tags", "value": "name:\"list hash id auto insert\" help:\"数据\""},
	{"name": "main", "value": "main.go"}
]`},
}
