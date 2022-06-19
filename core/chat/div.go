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
	return m.Cmdx(lex.SPLIT, "", "index", "args", func(ls []string, meta ice.Map) []string {
		if ls[0] == "div" {
			ls = append([]string{"", "", "style", kit.Select("div", ls, 1)}, kit.Slice(ls, 2)...)
		}
		return ls
	})
}

const DIV = "div"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DIV: {Name: "div", Help: "定制", Value: kit.Data(
			mdb.FIELD, "time,hash,type,name,text", nfs.PATH, ice.USR_PUBLISH,
			nfs.TEMPLATE, _div_template,
		)},
	}, Commands: map[string]*ice.Command{
		"/div/": {Name: "/div/", Help: "定制", Action: ice.MergeAction(ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
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
				m.RenderCmd(m.PrefixKey(), p)
			}
		}},
		DIV: {Name: "div hash auto import", Help: "定制", Action: ice.MergeAction(map[string]*ice.Action{
			lex.SPLIT: {Name: "split name=hi text", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				h := m.Cmdx(DIV, mdb.CREATE, m.OptionSimple(mdb.NAME), mdb.TEXT, _div_parse(m, m.Option(mdb.TEXT)))
				m.ProcessRewrite(mdb.HASH, h)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.PATH:
					m.Cmdy(nfs.DIR, arg[1:]).ProcessAgain()
				}
			}},
			mdb.CREATE: {Name: "create type=page name=hi text", Help: "创建"},
			mdb.IMPORT: {Name: "import path=src/", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, kit.Dict(nfs.DIR_ROOT, m.Option(nfs.PATH)), func(p string) {
					switch kit.Ext(p) {
					case "shy":
						m.Cmd(m.PrefixKey(), lex.SPLIT, mdb.NAME, p, mdb.TEXT, m.Cmdx(nfs.CAT, p))
					}
				})
			}},
		}, mdb.HashAction(), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			switch kit.Ext(kit.Select("", arg, 0)) {
			case "shy":
				m.Fields(0)
				m.Push(mdb.TEXT, _div_parse(m, m.Cmdx(nfs.CAT, arg[0])))
				m.DisplayLocal("")
			default:
				if mdb.HashSelect(m, arg...); len(arg) > 0 {
					m.Action("添加", "保存", "预览")
					m.DisplayLocal("")
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
