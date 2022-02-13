package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const TEMPLATE = "template"
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TEMPLATE: {Name: "template name auto create", Help: "模板", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(TEMPLATE, mdb.CREATE, kit.SimpleKV("", "txt", "网站索引", `
hi
	hi
		cli.qrcode
		cli.system
		cli.runtime

`))
				m.Cmd(TEMPLATE, mdb.CREATE, kit.SimpleKV("", "js", "前端模块", `Volcanos("onimport", {help: "导入数据", list: [], _init: function(can, msg, cb, target) {
	can.onmotion.clear(can)
	can.onappend.table(can, msg)
	can.onappend.board(can, msg)
}})`))
				m.Cmd(TEMPLATE, mdb.CREATE, kit.SimpleKV("", "go", "后端模块", `package {{.Option "zone"}}

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
`), "args", `[
	{"name": "zone", "value": "hi"},
	{"name": "name", "value": "hi"},
	{"name": "key", "value": "web.code.hi.hi"},
	{"name": "type", "values": "Hash,Zone,List"},
	{"name": "tags", "value": "name:\"list hash id auto insert\" help:\"数据\""},
	{"name": "main", "value": "main.go"}
]`)
			}},
			mdb.CREATE: {Name: "create type name text args", Help: "创建"},
			nfs.DEFS: {Name: "defs file=hi/hi.go", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				m.Option("tags", "`"+m.Option("tags")+"`")
				if buf, err := kit.Render(m.Option(mdb.TEXT), m); !m.Warn(err) {
					m.Cmd(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), string(buf))
					switch kit.Ext(m.Option(nfs.FILE)) {
					case GO:
						if m.Option(MAIN) != "" && m.Option(mdb.ZONE) != "" {
							_autogen_import(m, path.Join(m.Option(nfs.PATH), m.Option(MAIN)), m.Option(mdb.ZONE), _autogen_mod(m, ice.GO_MOD))
						}
					}
				}
			}},
		}, mdb.HashAction(mdb.SHORT, "name", mdb.FIELD, "time,type,name,text,args")), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.PushAction(nfs.DEFS, mdb.REMOVE)
			m.Sort("name")
			m.Cut("time,action,type,name,text,args")
		}}},
	})
}
