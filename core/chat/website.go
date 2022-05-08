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
	p := path.Join(WEBSITE, file)
	if m.Option(ice.MSG_USERPOD) != "" {
		p = path.Join(ice.POD, m.Option(ice.MSG_USERPOD), WEBSITE, file)
	}
	return strings.Split(kit.MergeURL2(m.Option(ice.MSG_USERWEB), path.Join("/chat", p)), "?")[0]
}
func _website_parse(m *ice.Message, text string, args ...string) (map[string]interface{}, bool) {
	if text == "" {
		return nil, false
	}

	m.Option(nfs.CAT_CONTENT, text)
	river, storm, last := kit.Dict(
		"Header", kit.Dict("menus", kit.List(), "style", kit.Dict("display", "none")),
		"River", kit.Dict("menus", kit.List(), "action", kit.List()),
		"Action", kit.Dict("menus", kit.List(), "action", kit.List(), "legend_event", "onclick"),
		"Footer", kit.Dict("style", kit.Dict("display", "none")), args,
	), kit.Dict(), kit.Dict()
	prefix := ""

	nriver := 0
	nstorm := 0
	m.Cmd(lex.SPLIT, "", mdb.KEY, mdb.NAME, func(deep int, ls []string, meta map[string]interface{}) []string {
		if deep == 1 {
			switch ls[0] {
			case "header":
				for i := 1; i < len(ls); i += 2 {
					kit.Value(river, kit.Keys("Header", ls[i]), ls[i+1])
				}
				return ls
			}
		}
		data := kit.Dict()
		switch display := ice.DisplayRequire(1, ls[0])[ctx.DISPLAY]; kit.Ext(ls[0]) {
		case nfs.JS:
			key := ice.GetFileCmd(display)
			if key == "" {
				if ls := strings.Split(display, ice.PS); len(ls) > 4 {
					ls[3] = ice.USR
					key = ice.GetFileCmd(path.Join(ls[3:]...))
				}
			}
			if key == "" {
				for p, k := range ice.Info.File {
					if strings.HasPrefix(p, path.Dir(display)) {
						key = k
					}
				}
			}
			ls[0] = kit.Select("can.code.inner.plugin", key)
			data[ctx.DISPLAY] = display
		case nfs.GO:
			key := ice.GetFileCmd(display)
			if key == "" {
				for k, v := range ice.Info.File {
					if strings.HasSuffix(k, ls[0]) {
						key = v
					}
				}
			}
			ls[0] = key
		case nfs.SH:
			key := ice.GetFileCmd(display)
			if key == "" {
				key = "cli.system"
			}
			data[ctx.ARGS] = kit.List(ls[0])
			ls[0] = key
		case nfs.SHY:
			data[ctx.ARGS] = kit.List(ls[0])
			data[mdb.NAME] = kit.TrimExt(ls[0], ".shy")
			if data[mdb.NAME] == "main" {
				data[mdb.NAME] = strings.TrimSuffix(strings.Split(ls[0], ice.PS)[1], "-story")
			}
			ls[0] = "web.wiki.word"
		case "~":
			prefix = ls[1]
			ls = ls[1:]
			fallthrough
		case "-":
			for _, v := range ls[1:] {
				last[mdb.LIST] = append(last[mdb.LIST].([]interface{}), kit.Dict(mdb.INDEX, kit.Keys(prefix, v), "order", len(last)))
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
				data[ls[i]] = ice.DisplayRequire(1, ls[i+1])[ctx.DISPLAY]

			case "title", "menus", "action", "style":
				data[ls[i]] = kit.UnMarshal(ls[i+1])
			default:
				data[ls[i]] = ls[i+1]
			}
		}

		switch deep {
		case 1:
			nriver++
			nstorm = 0
			storm = kit.Dict()
			if ls[0] == "auto" {
				ls[0] = kit.Format(nriver)
			}
			river[ls[0]] = kit.Dict(mdb.NAME, ls[1], STORM, storm, data, "order", len(river))
		case 2:
			nstorm++
			if ls[0] == "auto" {
				ls[0] = kit.Format(nstorm)
			}
			last = kit.Dict(mdb.NAME, ls[1], mdb.LIST, kit.List(), data, "order", len(storm))
			storm[ls[0]] = last
			prefix = ""
		default:
			last[mdb.LIST] = append(last[mdb.LIST].([]interface{}),
				kit.Dict(mdb.NAME, kit.Select(ls[0], data[mdb.NAME]), mdb.HELP, ls[1], mdb.INDEX, ls[0], "order", len(last), data))
		}
		return ls
	})
	return river, true
}
func _website_render(m *ice.Message, w http.ResponseWriter, r *http.Request, kind, text string) bool {
	msg := m.Spawn(w, r)
	switch kind {
	case nfs.SVG:
		msg.RenderResult(`<body style="background-color:cadetblue">%s</body>`, msg.Cmdx(nfs.CAT, text))
	case nfs.SHY:
		if r.Method == http.MethodGet {
			msg.RenderCmd(msg.Prefix(DIV), text)
		} else {
			r.URL.Path = "/chat/cmd/web.chat.div"
			return false
		}
	case nfs.TXT:
		msg.RenderCmd("can.parse", text)

	case nfs.IML:
		res, _ := _website_parse(msg, text)
		msg.RenderResult(_website_template2, kit.Format(res))
	case nfs.JSON:
		msg.RenderResult(_website_template2, kit.Format(kit.UnMarshal(text)))
	case nfs.JS:
		msg.RenderResult(_website_template, text)
	case nfs.HTML:
		msg.RenderResult(text)
	default:
		msg.RenderDownload(text)
	}
	web.Render(msg, msg.Option(ice.MSG_OUTPUT), msg.Optionv(ice.MSG_ARGS).([]interface{})...)
	return true
}
func _website_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Cmd(m.PrefixKey(), ice.OptionFields("")).Table(func(index int, value map[string]string, head []string) {
		m.PushSearch(value, mdb.TEXT, m.MergeWebsite(value[nfs.PATH]))
	})
}

const (
	SRC_WEBSITE  = "src/website/"
	CHAT_WEBSITE = "/chat/website/"
)
const WEBSITE = "website"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		WEBSITE: {Name: "website", Help: "网站", Value: kit.Data(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path,type,name,text")},
	}, Commands: map[string]*ice.Command{
		"/website/": {Name: "/website/", Help: "网站", Action: ice.MergeAction(map[string]*ice.Action{}, ctx.CmdAction())},
		WEBSITE: {Name: "website path auto create import", Help: "网站", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, nfs.IML, m.PrefixKey())
				m.Cmd(mdb.ENGINE, mdb.CREATE, nfs.IML, m.PrefixKey())
				m.Cmd(mdb.RENDER, mdb.CREATE, nfs.TXT, m.PrefixKey())
				m.Cmd(mdb.ENGINE, mdb.CREATE, nfs.TXT, m.PrefixKey())

				web.AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
					if ok := true; m.Richs(WEBSITE, nil, r.URL.Path, func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)
						ok = _website_render(m, w, r, kit.Format(value[mdb.TYPE]), kit.Format(value[mdb.TEXT]))
					}) != nil && ok {
						return true
					}
					if strings.HasPrefix(r.URL.Path, CHAT_WEBSITE) {
						if r.Method == http.MethodGet {
							_website_render(m, w, r, kit.Ext(r.URL.Path), m.Cmdx(nfs.CAT, strings.Replace(r.URL.Path, CHAT_WEBSITE, SRC_WEBSITE, 1)))
							return true
						}
					}
					return false
				})
			}},
			"show": {Hand: func(m *ice.Message, arg ...string) {
				if text := m.Cmd(m.PrefixKey(), ice.PS+arg[0]).Append(mdb.TEXT); text != "" {
					if res, ok := _website_parse(m, text, arg[1:]...); ok {
						m.Echo(_website_template2, kit.Format(res))
						return
					}
				}
				if res, ok := _website_parse(m, m.Cmdx(nfs.CAT, path.Join(SRC_WEBSITE, arg[0])), arg[1:]...); ok {
					m.Echo(_website_template2, kit.Format(res))
				}
			}},
			"inner": {Hand: func(m *ice.Message, arg ...string) {}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					_website_search(m, arg[0], arg[1], kit.Select("", arg, 2))
				}
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
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
			mdb.CREATE: {Name: "create path type=iml,json,js,html name text", Help: "创建"},
			mdb.IMPORT: {Name: "import path=src/website/", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, kit.Dict(nfs.DIR_ROOT, m.Option(nfs.PATH)), func(p string) {
					switch name := strings.TrimPrefix(p, m.Option(nfs.PATH)); kit.Ext(p) {
					case nfs.HTML, nfs.JS, nfs.JSON, nfs.IML, nfs.TXT:
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
				m.PushAnchor(m.MergeWebsite(value[nfs.PATH]))
			})
			if len(arg) == 0 {
				m.Cmd(nfs.DIR, SRC_WEBSITE, func(f os.FileInfo, p string) {
					m.Push("", kit.Dict(
						mdb.TIME, f.ModTime().Format(ice.MOD_TIME),
						nfs.PATH, ice.PS+strings.TrimPrefix(p, SRC_WEBSITE),
						mdb.TYPE, kit.Ext(p), mdb.NAME, path.Base(p), mdb.TEXT, m.Cmdx(nfs.CAT, p),
					), kit.Split(m.Config(mdb.FIELD))).PushButton("")
					m.PushAnchor(m.MergeLink(path.Join(CHAT_WEBSITE, strings.TrimPrefix(p, SRC_WEBSITE))))
				}).Sort(nfs.PATH)
			}

			if m.Length() == 0 && len(arg) > 0 {
				m.Push(mdb.TEXT, m.Cmdx(nfs.CAT, path.Join(SRC_WEBSITE, path.Join(arg...))))
				m.Push(nfs.PATH, path.Join(CHAT_WEBSITE, path.Join(arg...)))
				m.PushAnchor(m.MergeLink(m.Append(nfs.PATH)))
			}

			if m.FieldsIsDetail() {
				m.PushQRCode(mdb.SCAN, m.MergeURL2(m.Append(nfs.PATH)))
				m.EchoIFrame(m.Append(nfs.PATH))
			}
		}},
	}})
}

var _website_template = `<!DOCTYPE html>
<head>
	<meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=0">
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
	<meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=0">
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
