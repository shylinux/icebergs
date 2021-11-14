package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _div_parse(m *ice.Message, text string) string {
	m.Option(nfs.CAT_CONTENT, text)
	return m.Cmdx(lex.SPLIT, "", "index", "args", func(ls []string, meta map[string]interface{}) []string {
		if ls[0] == "_span" {
			ls = append([]string{"", "", "style", kit.Select("span", ls, 1)}, kit.Slice(ls, 2)...)
		}
		return ls
	})
}

const DIV = "div"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DIV: {Name: "div", Help: "定制", Value: kit.Data(
			kit.MDB_FIELD, "time,hash,type,name,text", kit.MDB_PATH, ice.USR_PUBLISH,
		)},
	}, Commands: map[string]*ice.Command{
		"/div/": {Name: "/div/", Help: "定制", Action: ice.MergeAction(ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.RenderCmd(m.PrefixKey(), path.Join(arg...))
		}},
		DIV: {Name: "div hash auto", Help: "定制", Action: ice.MergeAction(map[string]*ice.Action{
			lex.SPLIT: {Name: "split name=hi text", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				h := m.Cmdx(DIV, mdb.CREATE, m.OptionSimple(kit.MDB_NAME), kit.MDB_TEXT, _div_parse(m, m.Option(kit.MDB_TEXT)))
				m.ProcessRewrite(kit.MDB_HASH, h)
			}},
			mdb.CREATE: {Name: "create type=page name=hi text", Help: "创建"},
		}, mdb.HashAction(), ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && strings.HasSuffix(arg[0], ".shy") {
				m.Fields(0)
				m.Option(ice.MSG_DISPLAY, "/plugin/local/chat/div.js")
				m.Push(kit.MDB_TEXT, _div_parse(m, m.Cmdx(nfs.CAT, arg[0])))
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.Option(ice.MSG_DISPLAY, "/plugin/local/chat/div.js")
				m.Action("添加", "保存", "预览")
			} else {
				m.Action(lex.SPLIT, mdb.CREATE)
			}
		}},
	}})
}
