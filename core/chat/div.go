package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _div_deep(str string) int {
	for i, c := range str {
		if c != ' ' {
			return i
		}
	}
	return 0
}
func _div_parse(m *ice.Message, root map[string]interface{}, list []string) int {
	var last map[string]interface{}
	deep := _div_deep(list[0])
	for i := 0; i < len(list); i++ {
		if d := _div_deep(list[i]); d < deep {
			return i
		} else if d > deep {
			i += _div_parse(m, last, list[i:]) - 1
			continue
		}

		ls := kit.Split(list[i])
		if ls[0] == "_span" {
			ls = append([]string{"", "", "style", "span"}, ls[1:]...)
		}
		meta := kit.Dict(
			"index", kit.Select("", ls, 0),
			"args", kit.Select("", ls, 1),
		)
		for i := 2; i < len(ls); i += 2 {
			meta[ls[i]] = ls[i+1]
		}
		last = kit.Dict("meta", meta, "list", kit.List())
		kit.Value(root, "list.-2", last)
	}
	return len(list)
}

const DIV = "div"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DIV: {Name: "div", Help: "定制", Value: kit.Data(
			kit.MDB_SHORT, "", kit.MDB_FIELD, "time,hash,type,name,text",
			kit.MDB_PATH, ice.USR_PUBLISH,
		)},
	}, Commands: map[string]*ice.Command{
		DIV: {Name: "div hash auto", Help: "定制", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create type=page name=hi.html text", Help: "创建"},
			cli.MAKE: {Name: "make name=some line:textarea", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				m.Fields(0)
				node := kit.Dict("meta", kit.Dict("name", m.Option("name")), "list", []interface{}{})
				_div_parse(m, node, kit.Split(m.Option("line"), "\n", "\n", "\n"))
				m.ProcessDisplay("/plugin/local/chat/div.js")
				m.Push("text", kit.Formats(node))
			}},
		}, mdb.HashAction(), ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), m.Conf(DIV, kit.META_FIELD))
			m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, kit.MDB_HASH, arg)
			m.Table(func(index int, value map[string]string, head []string) {
				m.PushAnchor("/" + path.Join(ice.PUBLISH, value[kit.MDB_NAME]))
			})
			if m.PushAction(cli.MAKE, mdb.REMOVE); len(arg) > 0 {
				m.Option(ice.MSG_DISPLAY, "/plugin/local/chat/div.js")
				m.Action("添加", "保存", "预览")
			} else {
				m.Action(mdb.CREATE, cli.MAKE)
			}
		}},
		"/div": {Name: "/div", Help: "定制", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.RenderIndex(web.SERVE, ice.VOLCANOS, "page/div.html")
		}},
	}})

}
