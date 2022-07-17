package chat

import (
	"net/http"
	"os"
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

func _website_url(m *ice.Message, file string) string {
	return strings.Split(m.MergeWebsite(file), "?")[0]
}
func _website_parse(m *ice.Message, text string, args ...string) (ice.Map, bool) {
	if text == "" {
		return nil, false
	}

	const (
		HEADER = "Header"
		RIVER  = "River"
		FOOTER = "Footer"

		ORDER = "order"
		TITLE = "title"
		MENUS = "menus"
	)

	river, storm, last := kit.Dict(
		HEADER, kit.Dict(MENUS, kit.List(), ctx.STYLE, kit.Dict(ctx.DISPLAY, "none")),
		RIVER, kit.Dict(MENUS, kit.List(), ctx.ACTION, kit.List("")),
		FOOTER, kit.Dict(MENUS, kit.List(), ctx.STYLE, kit.Dict(ctx.DISPLAY, "none")),
		args,
	), kit.Dict(), kit.Dict()

	nriver, nstorm, prefix := 0, 0, ""
	m.Cmd(lex.SPLIT, "", mdb.KEY, mdb.NAME, kit.Dict(nfs.CAT_CONTENT, text), func(deep int, ls []string) []string {
		if deep == 1 {
			switch ls[0] {
			case HEADER, RIVER, FOOTER:
				for i := 1; i < len(ls); i += 2 {
					kit.Value(river, kit.Keys(ls[0], ls[i]), ls[i+1])
				}
				return ls
			}
		}

		data := kit.Dict()
		switch kit.Ext(ls[0]) {
		case nfs.JS:
			ls[0], data[ctx.DISPLAY] = kit.Select(ctx.CAN_PLUGIN, ice.GetFileCmd(ls[0])), ice.FileURI(ls[0])
		case nfs.GO:
			ls[0] = ice.GetFileCmd(ls[0])
		case nfs.SH:
			ls[0], data[ctx.ARGS] = "web.code.sh.sh", ls[0]
		case nfs.SHY:
			ls[0], data[ctx.ARGS] = "web.wiki.word", ls[0]
		case nfs.PY:
			ls[0], data[ctx.ARGS] = "web.code.sh.py", ls[0]
		case "~":
			prefix, ls = ls[1], ls[1:]
			fallthrough
		case "-":
			for _, v := range ls[1:] {
				last[mdb.LIST] = append(last[mdb.LIST].([]ice.Any), kit.Dict(mdb.INDEX, kit.Keys(prefix, v)))
			}
			return ls
		}

		if ls[0] == "" {
			return ls
		} else if len(ls) == 1 && deep > 2 {
			ls = append(ls, m.Cmd(ctx.COMMAND, ls[0]).Append(mdb.HELP))
		} else if len(ls) == 1 {
			ls = append(ls, ls[0])
		} else if ls[1] == "" {
			ls[1] = ls[0]
		}

		for i := 2; i < len(ls); i += 2 {
			switch ls[i] {
			case ctx.ARGS:
				data[ls[i]] = kit.Split(ls[i+1])
			case ctx.DISPLAY:
				data[ls[i]] = ice.Display(ls[i+1])[ctx.DISPLAY]
			case ctx.STYLE, ctx.ACTION, TITLE, MENUS:
				data[ls[i]] = kit.UnMarshal(ls[i+1])
			default:
				data[ls[i]] = ls[i+1]
			}
		}

		switch deep {
		case 1:
			if nriver++; ls[0] == ice.AUTO {
				ls[0] = kit.Format(nriver)
			}
			nstorm, storm = 0, kit.Dict()
			river[ls[0]] = kit.Dict(mdb.NAME, ls[1], STORM, storm, data, ORDER, len(river))
		case 2:
			if nstorm++; ls[0] == ice.AUTO {
				ls[0] = kit.Format(nstorm)
			}
			last = kit.Dict(mdb.NAME, ls[1], mdb.LIST, kit.List(), data, ORDER, len(storm))
			storm[ls[0]] = last
			prefix = ""
		default:
			last[mdb.LIST] = append(last[mdb.LIST].([]ice.Any), kit.Dict(mdb.NAME, kit.Select(ls[0], data[mdb.NAME]), mdb.HELP, ls[1], mdb.INDEX, ls[0], data))
		}
		return ls
	})
	return river, true
}
func _website_render(m *ice.Message, w http.ResponseWriter, r *http.Request, kind, text string) bool {
	msg := m.Spawn(w, r)
	switch kind {
	case nfs.ZML:
		msg.RenderCmd("can.parse", text)
	case nfs.IML:
		res, _ := _website_parse(msg, text)
		msg.RenderResult(_website_template2, kit.Format(res))
	case nfs.SHY:
		if r.Method == http.MethodGet {
			msg.RenderCmd(msg.Prefix(DIV), text)
		} else {
			r.URL.Path = "/chat/cmd/web.chat.div"
			return false
		}
	case nfs.JSON:
		msg.RenderResult(_website_template2, kit.Format(kit.UnMarshal(text)))
	case nfs.JS:
		msg.RenderResult(_website_template, text)
	case nfs.HTML:
		msg.RenderResult(text)
	case nfs.SVG:
		msg.RenderResult(`<body style="background-color:cadetblue">%s</body>`, msg.Cmdx(nfs.CAT, text))
	default:
		msg.RenderDownload(text)
	}
	web.Render(msg, msg.Option(ice.MSG_OUTPUT), msg.Optionv(ice.MSG_ARGS).([]ice.Any)...)
	return true
}
func _website_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Cmd(m.PrefixKey(), ice.OptionFields("")).Table(func(index int, value ice.Maps, head []string) {
		m.PushSearch(value, mdb.TEXT, m.MergeWebsite(value[nfs.PATH]))
	})
}

const (
	SRC_WEBSITE  = "src/website/"
	CHAT_WEBSITE = "/chat/website/"
)
const WEBSITE = "website"

func init() {
	Index.MergeCommands(ice.Commands{"/website/": {Name: "/website/", Help: "网站", Actions: ctx.CmdAction()},
		WEBSITE: {Name: "website path auto create import", Help: "网站", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, nfs.TXT, m.PrefixKey())
				m.Cmd(mdb.ENGINE, mdb.CREATE, nfs.TXT, m.PrefixKey())
				m.Cmd(mdb.RENDER, mdb.CREATE, nfs.IML, m.PrefixKey())
				m.Cmd(mdb.ENGINE, mdb.CREATE, nfs.IML, m.PrefixKey())

				web.AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
					if r.Method != http.MethodGet {
						return false
					}
					if ok := true; m.Richs(WEBSITE, nil, r.URL.Path, func(key string, value ice.Map) {
						value = kit.GetMeta(value)
						ok = _website_render(m, w, r, kit.Format(value[mdb.TYPE]), kit.Format(value[mdb.TEXT]))
					}) != nil && ok {
						return true
					}
					if strings.HasPrefix(r.URL.Path, CHAT_WEBSITE) {
						_website_render(m, w, r, kit.Ext(r.URL.Path), m.Cmdx(nfs.CAT, strings.Replace(r.URL.Path, CHAT_WEBSITE, SRC_WEBSITE, 1)))
						return true
					}
					return false
				})
			}},
			lex.PARSE: {Hand: func(m *ice.Message, arg ...string) {
				switch kit.Ext(arg[0]) {
				case nfs.ZML:
					m.RenderCmd("can.parse", m.Cmdx(nfs.CAT, path.Join(SRC_WEBSITE, arg[0])))

				case nfs.IML:
					if res, ok := _website_parse(m, m.Cmdx(nfs.CAT, path.Join(SRC_WEBSITE, arg[0])), arg[1:]...); ok {
						m.Echo(_website_template2, kit.Format(res))
					}

				default:
					if text := m.Cmd(m.PrefixKey(), ice.PS+arg[0]).Append(mdb.TEXT); text != "" {
						if res, ok := _website_parse(m, text, arg[1:]...); ok {
							m.Echo(_website_template2, kit.Format(res))
						}
					}
				}
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					_website_search(m, arg[0], arg[1], kit.Select("", arg, 2))
				}
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				if res, ok := _website_parse(m, m.Cmdx(nfs.CAT, path.Join(arg[2], arg[1]))); ok {
					m.Echo(kit.Formats(res)).DisplayStoryJSON()
				} else {
					m.Echo(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
				}
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
			mdb.CREATE: {Name: "create path type=iml,zml,json,js,html name text", Help: "创建"},
			mdb.IMPORT: {Name: "import path=src/website/", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, kit.Dict(nfs.DIR_ROOT, m.Option(nfs.PATH)), func(p string) {
					switch name := strings.TrimPrefix(p, m.Option(nfs.PATH)); kit.Ext(p) {
					case nfs.HTML, nfs.JS, nfs.JSON, nfs.ZML, nfs.IML, nfs.TXT:
						m.Cmd(m.PrefixKey(), mdb.CREATE, nfs.PATH, ice.PS+name, mdb.TYPE, kit.Ext(p), mdb.NAME, name, mdb.TEXT, m.Cmdx(nfs.CAT, p))
					default:
						m.Cmd(m.PrefixKey(), mdb.CREATE, nfs.PATH, ice.PS+name, mdb.TYPE, kit.Ext(p), mdb.NAME, name, mdb.TEXT, p)
					}
				})
			}},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Tables(func(value ice.Maps) { m.PushAnchor(m.MergeWebsite(value[nfs.PATH])) })

			if len(arg) == 0 { // 文件列表
				m.Cmd(nfs.DIR, SRC_WEBSITE, func(f os.FileInfo, p string) {
					m.Push("", kit.Dict(
						mdb.TIME, f.ModTime().Format(ice.MOD_TIME),
						nfs.PATH, ice.PS+strings.TrimPrefix(p, SRC_WEBSITE),
						mdb.TYPE, kit.Ext(p), mdb.NAME, path.Base(p), mdb.TEXT, m.Cmdx(nfs.CAT, p),
					), kit.Split(m.Config(mdb.FIELD))).PushButton("")
					m.PushAnchor(m.MergeLink(path.Join(CHAT_WEBSITE, strings.TrimPrefix(p, SRC_WEBSITE))))
				}).Sort(nfs.PATH)
			}
			if m.Length() == 0 && len(arg) > 0 { // 文件详情
				m.Push(mdb.TEXT, m.Cmdx(nfs.CAT, path.Join(SRC_WEBSITE, path.Join(arg...))))
				m.Push(nfs.PATH, path.Join(CHAT_WEBSITE, path.Join(arg...)))
				m.PushAnchor(m.MergeLink(m.Append(nfs.PATH)))
			}

			if len(arg) > 0 { // 文件预览
				m.PushQRCode(mdb.SCAN, m.MergeURL2(m.Append(nfs.PATH)))
				m.EchoIFrame(m.Append(nfs.PATH))
			}
		}},
	})
}

var _website_template = `<!DOCTYPE html>
<head>
	<meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=no"/>
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
	<meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=no"/>
	<meta charset="utf-8">
	<title>volcanos</title>
	<link rel="shortcut icon" type="image/ico" href="/favicon.ico">
	<link rel="stylesheet" type="text/css" href="/page/cache.css">
	<link rel="stylesheet" type="text/css" href="/page/index.css">
</head>
<body>
	<script src="/proto.js"></script>
	<script src="/page/cache.js"></script>
	<script>Volcanos({river: JSON.parse('%s')})</script>
</body>
`
