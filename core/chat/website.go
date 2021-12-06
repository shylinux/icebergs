package chat

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _website_parse(m *ice.Message, text string) map[string]interface{} {
	m.Option(nfs.CAT_CONTENT, text)
	river, storm, last := kit.Dict(), kit.Dict(), kit.Dict()
	m.Cmd(lex.SPLIT, "", kit.MDB_KEY, kit.MDB_NAME, func(deep int, ls []string, meta map[string]interface{}) []string {
		data := kit.Dict()
		for i := 2; i < len(ls); i += 2 {
			switch ls[i] {
			case kit.MDB_ARGS:
				data[ls[i]] = kit.UnMarshal(ls[i+1])
			default:
				data[ls[i]] = ls[i+1]
			}
		}
		switch deep {
		case 0:
			storm = kit.Dict()
			river[ls[0]] = kit.Dict(kit.MDB_NAME, ls[1], "storm", storm, data)
		case 4:
			last = kit.Dict(kit.MDB_NAME, ls[1], kit.MDB_LIST, kit.List(), data)
			storm[ls[0]] = last
		case 8:
			last[kit.MDB_LIST] = append(last[kit.MDB_LIST].([]interface{}),
				kit.Dict(kit.MDB_NAME, ls[0], kit.MDB_HELP, ls[1], kit.MDB_INDEX, ls[0], data))
		}
		return ls
	})
	return river
}

const WEBSITE = "website"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		WEBSITE: {Name: "website", Help: "网站", Value: kit.Data(
			kit.MDB_SHORT, nfs.PATH, kit.MDB_FIELD, "time,path,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		WEBSITE: {Name: "website path auto create import", Help: "网站", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
					if m.Richs(WEBSITE, nil, r.URL.Path, func(key string, value map[string]interface{}) {
						msg, value := m.Spawn(w, r), kit.GetMeta(value)
						switch text := kit.Format(value[kit.MDB_TEXT]); value[kit.MDB_TYPE] {
						case "txt":
							res := _website_parse(msg, kit.Format(value[kit.MDB_TEXT]))
							web.RenderResult(msg, _website_template2, kit.Format(res))
						case "json":
							web.RenderResult(msg, _website_template2, kit.Format(kit.UnMarshal(text)))
						case "js":
							web.RenderResult(msg, _website_template, text)
						default:
							web.RenderResult(msg, text)
						}
					}) != nil {
						return true
					}
					return false
				})
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.PATH:
					m.Cmdy(nfs.DIR, arg[1:]).ProcessAgain()
				}
			}},
			mdb.CREATE: {Name: "create path type=html,js,json name text", Help: "创建"},
			mdb.IMPORT: {Name: "import path=src/", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, kit.Dict(nfs.DIR_ROOT, m.Option(nfs.PATH)), func(p string) {
					switch kit.Ext(p) {
					case "html", "js", "json", "txt":
						m.Cmd(m.PrefixKey(), mdb.CREATE, nfs.PATH, ice.PS+p,
							kit.MDB_TYPE, kit.Ext(p), kit.MDB_NAME, p, kit.MDB_TEXT, m.Cmdx(nfs.CAT, p))
					}
				})
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(index int, value map[string]string, head []string) {
				m.PushAnchor(m.MergeURL2(value[nfs.PATH]))
			})
			if m.Sort(nfs.PATH); m.FieldsIsDetail() {
				m.EchoIFrame(m.Append(nfs.PATH))
			}
		}},
	}})
}

var _website_template = `<!DOCTYPE html>
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
	<script>%s</script>
</body>
`

var _website_template2 = `<!DOCTYPE html>
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
