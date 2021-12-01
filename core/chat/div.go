package chat

import (
	"encoding/json"
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
			kit.MDB_TEMPLATE, _div_template,
		)},
	}, Commands: map[string]*ice.Command{
		"/div/": {Name: "/div/", Help: "定制", Action: ice.MergeAction(ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch p := path.Join(arg...); kit.Ext(kit.Select("", p)) {
			case "html":
				m.RenderDownload(p)
			case "css":
				m.RenderResult(_div_template, m.Cmdx(nfs.CAT, p), m.Cmdx(nfs.CAT, strings.ReplaceAll(p, ".css", ".js")))
			case "js":
				m.RenderResult(_div_template, m.Cmdx(nfs.CAT, strings.ReplaceAll(p, ".js", ".css")), m.Cmdx(nfs.CAT, p))
			case "json":
				var res interface{}
				err := json.Unmarshal([]byte(m.Cmdx(nfs.CAT, p)), &res)
				m.Assert(err)
				m.RenderResult(_div_template2, kit.Format(res))
			default:
				m.RenderCmd(m.PrefixKey(), p)
			}
		}},
		DIV: {Name: "div hash auto", Help: "定制", Action: ice.MergeAction(map[string]*ice.Action{
			lex.SPLIT: {Name: "split name=hi text", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				h := m.Cmdx(DIV, mdb.CREATE, m.OptionSimple(kit.MDB_NAME), kit.MDB_TEXT, _div_parse(m, m.Option(kit.MDB_TEXT)))
				m.ProcessRewrite(kit.MDB_HASH, h)
			}},
			mdb.CREATE: {Name: "create type=page name=hi text", Help: "创建"},
		}, mdb.HashAction(), ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch kit.Ext(kit.Select("", arg, 0)) {
			case "shy":
				m.Fields(0)
				m.Option(ice.MSG_DISPLAY, "/plugin/local/chat/div.js")
				m.Push(kit.MDB_TEXT, _div_parse(m, m.Cmdx(nfs.CAT, arg[0])))
			default:
				if mdb.HashSelect(m, arg...); len(arg) > 0 {
					m.Option(ice.MSG_DISPLAY, "/plugin/local/chat/div.js")
					m.Action("添加", "保存", "预览")
				} else {
					m.Action(lex.SPLIT, mdb.CREATE)
				}
			}
		}},
	}})
}

var _div_template = `<!DOCTYPE html>
<head>
	<meta charset="utf-8">
	<title>volcanos</title>
	<link rel="shortcut icon" type="image/ico" href="/favicon.ico">
	<link rel="stylesheet" type="text/css" href="/page/cache.css">
	<link rel="stylesheet" type="text/css" href="/page/index.css">
    <style type="text/css">%s</style>
</head>
<body>
	<script src="/proto.js"></script>
	<script src="/page/cache.js"></script>
	<script>%s</script>
</body>
`
var _div_template2 = `<!DOCTYPE html>
<head>
	<meta charset="utf-8">
	<title>volcanos</title>
	<link rel="shortcut icon" type="image/ico" href="/favicon.ico">
	<link rel="stylesheet" type="text/css" href="/page/cache.css">
	<link rel="stylesheet" type="text/css" href="/page/index.css">
</head>
<body>
	<script src="/proto.js"></script>
	<script src="/page/cache.js"></script>
	<script>
Volcanos({name: "chat", panels: [
    {name: "Header", help: "标题栏", pos: chat.HEAD, state: ["time", "usernick", "avatar"]},
    {name: "River",  help: "群聊组", pos: chat.LEFT, action: ["create", "refresh"]},
    {name: "Action", help: "工作台", pos: chat.MAIN},
    {name: "Search", help: "搜索框", pos: chat.AUTO},
    {name: "Footer", help: "状态条", pos: chat.FOOT, state: ["ncmd"]},
], main: {name: "Header", list: ["/publish/order.js"]}, plugin: [
    "/plugin/state.js",
    "/plugin/input.js",
    "/plugin/table.js",
    "/plugin/input/key.js",
    "/plugin/input/date.js",
    "/plugin/story/spide.js",
    "/plugin/story/trend.js",
    "/plugin/local/code/inner.js",
    "/plugin/local/code/vimer.js",
    "/plugin/local/wiki/draw/path.js",
    "/plugin/local/wiki/draw.js",
    "/plugin/local/wiki/word.js",
    "/plugin/local/chat/div.js",
    "/plugin/local/team/plan.js",
    "/plugin/input/province.js",
], river: JSON.parse('%s')})
	</script>
</body>
`
