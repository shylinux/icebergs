package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _div_parse(m *ice.Message, text string) string {
	return m.Cmdx(lex.SPLIT, "", ctx.INDEX, ctx.ARGS, kit.Dict(nfs.CAT_CONTENT, text), func(deep int, ls []string) []string {
		if ls[0] == DIV {
			ls = append([]string{"", "", ctx.STYLE, kit.Select(DIV, ls, 1)}, kit.Slice(ls, 2)...)
		}
		return ls
	})
}

const DIV = "div"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		DIV: {Name: "div", Help: "定制", Value: kit.Data(
			mdb.FIELD, "time,hash,type,name,text", nfs.PATH, ice.USR_PUBLISH,
			nfs.TEMPLATE, _div_template,
		)},
	}, Commands: ice.Commands{
		"/div/": {Name: "/div/", Help: "定制", Actions: ice.MergeAction(ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			switch p := path.Join(arg...); kit.Ext(kit.Select("", p)) {
			case nfs.HTML:
				m.RenderDownload(p)
			case nfs.CSS:
				m.RenderResult(_div_template, m.Cmdx(nfs.CAT, p), m.Cmdx(nfs.CAT, strings.ReplaceAll(p, ".css", ".js")))
			case nfs.JS:
				m.RenderResult(_div_template, m.Cmdx(nfs.CAT, strings.ReplaceAll(p, ".js", ".css")), m.Cmdx(nfs.CAT, p))
			case nfs.JSON:
				m.RenderResult(_div_template2, kit.Format(kit.UnMarshal(m.Cmdx(nfs.CAT, p))))
			default:
				web.RenderCmd(m, m.PrefixKey(), p)
			}
		}},
		DIV: {Name: "div hash auto import", Help: "定制", Actions: ice.MergeAction(ice.Actions{
			lex.SPLIT: {Name: "split name=hi text", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessRewrite(mdb.HASH, m.Cmdx(DIV, mdb.CREATE, m.OptionSimple(mdb.NAME), mdb.TEXT, _div_parse(m, m.Option(mdb.TEXT))))
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.PATH:
					m.Cmdy(nfs.DIR, arg[1:]).ProcessAgain()
				case ctx.INDEX:
					m.OptionFields(mdb.INDEX)
					m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "")
				case ctx.STYLE:
					m.Push(arg[0], "div")
					m.Push(arg[0], "span")
					m.Push(arg[0], "output")
				}
			}},
			mdb.CREATE: {Name: "create type=page name=hi text", Help: "创建"},
			mdb.IMPORT: {Name: "import path=src/", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, kit.Dict(nfs.DIR_ROOT, m.Option(nfs.PATH)), func(p string) {
					switch kit.Ext(p) {
					case nfs.SHY:
						m.Cmd(m.PrefixKey(), lex.SPLIT, mdb.NAME, p, mdb.TEXT, m.Cmdx(nfs.CAT, p))
					}
				})
			}},
		}, mdb.HashAction(), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			switch kit.Ext(kit.Select("", arg, 0)) {
			case nfs.SHY:
				m.Fields(0)
				m.Push(mdb.TEXT, _div_parse(m, m.Cmdx(nfs.CAT, arg[0])))
				ctx.DisplayLocal(m, "")
			default:
				if mdb.HashSelect(m, arg...); len(arg) > 0 {
					m.Action("添加", "保存")
					ctx.DisplayLocal(m, "")
					m.StatusTime(mdb.LINK, web.MergeURL2(m, "/chat/div/"+arg[0]))
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
	<script>Volcanos({name: "chat", river: JSON.parse('%s')})</script>
</body>
`
