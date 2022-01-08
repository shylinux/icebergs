package chat

import (
	"net/http"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _website_parse(m *ice.Message, text string) map[string]interface{} {
	m.Option(nfs.CAT_CONTENT, text)
	river, storm, last := kit.Dict(), kit.Dict(), kit.Dict()
	m.Cmd(lex.SPLIT, "", mdb.KEY, mdb.NAME, func(deep int, ls []string, meta map[string]interface{}) []string {
		if len(ls) == 1 {
			ls = append(ls, ls[0])
		}
		data := kit.Dict()
		for i := 2; i < len(ls); i += 2 {
			switch ls[i] {
			case ctx.ARGS:
				data[ls[i]] = kit.Split(ls[i+1])
				// data[ls[i]] = kit.UnMarshal(ls[i+1])
			default:
				data[ls[i]] = ls[i+1]
			}
		}
		switch deep {
		case 1:
			storm = kit.Dict()
			river[ls[0]] = kit.Dict(mdb.NAME, ls[1], "storm", storm, data)
		case 2:
			last = kit.Dict(mdb.NAME, ls[1], mdb.LIST, kit.List(), data)
			storm[ls[0]] = last
		default:
			last[mdb.LIST] = append(last[mdb.LIST].([]interface{}),
				kit.Dict(mdb.NAME, ls[0], mdb.HELP, ls[1], mdb.INDEX, ls[0], data))
		}
		return ls
	})
	return river
}

const WEBSITE = "website"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		WEBSITE: {Name: "website", Help: "网站", Value: kit.Data(
			mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		WEBSITE: {Name: "website path auto create import", Help: "网站", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
					if ok := true; m.Richs(WEBSITE, nil, r.URL.Path, func(key string, value map[string]interface{}) {
						msg, value := m.Spawn(w, r), kit.GetMeta(value)
						switch text := kit.Format(value[mdb.TEXT]); value[mdb.TYPE] {
						case "svg":
							msg.RenderResult(`<body style="background-color:cadetblue">%s</body>`, m.Cmdx(nfs.CAT, text))
						case "shy":
							if r.Method == http.MethodGet {
								msg.RenderCmd(msg.Prefix(DIV), text)
							} else {
								r.URL.Path, ok = "/chat/cmd/web.chat.div", false
								return
							}
						case "txt":
							res := _website_parse(msg, kit.Format(value[mdb.TEXT]))
							msg.RenderResult(_website_template2, kit.Format(res))
						case "json":
							msg.RenderResult(_website_template2, kit.Format(kit.UnMarshal(text)))
						case "js":
							msg.RenderResult(_website_template, text)
						case "html":
							msg.RenderResult(text)
						default:
							msg.RenderDownload(text)
						}
						web.Render(msg, msg.Option(ice.MSG_OUTPUT), msg.Optionv(ice.MSG_ARGS).([]interface{})...)
					}) != nil && ok {
						return true
					}
					return false
				})
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case mdb.CREATE:
					m.Cmdy(mdb.INPUTS, m.PrefixKey(), "", mdb.HASH, arg)
				default:
					switch arg[0] {
					case nfs.PATH:
						m.Cmdy(nfs.DIR, arg[1:]).ProcessAgain()
					}
				}
			}},
			mdb.CREATE: {Name: "create path type=txt,json,js,html name text", Help: "创建"},
			mdb.IMPORT: {Name: "import path=src/", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, kit.Dict(nfs.DIR_ROOT, m.Option(nfs.PATH)), func(p string) {
					switch name := strings.TrimPrefix(p, m.Option(nfs.PATH)); kit.Ext(p) {
					case "html", "js", "json", "txt":
						m.Cmd(m.PrefixKey(), mdb.CREATE, nfs.PATH, ice.PS+name,
							mdb.TYPE, kit.Ext(p), mdb.NAME, name, mdb.TEXT, m.Cmdx(nfs.CAT, p))
					default:
						m.Cmd(m.PrefixKey(), mdb.CREATE, nfs.PATH, ice.PS+name,
							mdb.TYPE, kit.Ext(p), mdb.NAME, name, mdb.TEXT, p)
					}
				})
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(index int, value map[string]string, head []string) {
				m.PushAnchor(m.MergeURL2(value[nfs.PATH]))
			})
			if m.Sort(nfs.PATH); m.FieldsIsDetail() {
				m.PushQRCode(mdb.SCAN, m.MergeURL2(m.Append(nfs.PATH)))
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
