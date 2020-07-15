package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

const (
	INNER  = "inner"
	VEDIO  = "vedio"
	QRCODE = "qrcode"
)

const (
	LIST = "list"
	SAVE = "save"
	PLUG = "plug"
	SHOW = "show"
)

func _inner_protect(m *ice.Message, name string) bool {
	ls := strings.Split(name, "/")
	return !m.Right(ls) && m.Conf(INNER, kit.Keys("meta.protect", ls[0])) == "true"
}
func _inner_source(m *ice.Message, name string) bool {
	return m.Conf(INNER, kit.Keys("meta.source", _inner_ext(name))) == "true"
}
func _inner_ext(name string) string {
	return strings.ToLower(kit.Select(path.Base(name), strings.TrimPrefix(path.Ext(name), ".")))
}
func _inner_sub(m *ice.Message, action string, name string, arg ...string) bool {
	if _inner_protect(m, name) {
		m.Push("file", "../")
		return true
	}

	p := _inner_ext(name)
	if m.Cmdy(kit.Keys(p, action), name, arg); len(m.Resultv()) > 0 && m.Result(0) != "warn: " {
		return true
	}
	return false
}

func _inner_list(m *ice.Message, dir, file string) {
	if _inner_sub(m, LIST, path.Join(dir, file)) {
		return
	}

	if m.Set(ice.MSG_RESULT); file == "" || strings.HasSuffix(file, "/") || _inner_source(m, file) {
		m.Option(nfs.DIR_ROOT, dir)
		m.Option(nfs.DIR_DEEP, "true")
		m.Option(nfs.DIR_TYPE, nfs.TYPE_FILE)
		m.Cmdy(nfs.DIR, file, "path size time")
		return
	}
	m.Echo(path.Join(dir, file))
}
func _inner_save(m *ice.Message, name, text string) {
	if _inner_sub(m, SAVE, name) {
		return
	}

	if f, p, e := kit.Create(name); m.Assert(e) {
		defer f.Close()
		m.Cmd(web.FAVOR, "inner.save", "shell", name, text)
		if n, e := f.WriteString(text); m.Assert(e) {
			m.Log_EXPORT("file", name, "size", n)
		}
		m.Echo(p)
	}
}
func _inner_plug(m *ice.Message, name string) {
	if _inner_sub(m, PLUG, name) {
		return
	}

	p := _inner_ext(name)
	if ls := m.Confv(INNER, kit.Keys("meta.plug", p)); ls != nil {
		m.Echo(kit.Format(ls))
		return
	}

	m.Echo("{}")
}
func _inner_show(m *ice.Message, dir, file string) {
	name := path.Join(dir, file)
	if _inner_sub(m, SHOW, name) {
		return
	}

	p := _inner_ext(name)
	if ls := kit.Simple(m.Confv(INNER, kit.Keys("meta.show", p))); len(ls) > 0 {
		m.Cmdy(cli.SYSTEM, ls, name)
		m.Set(ice.MSG_APPEND)
		m.Cmd(web.FAVOR, "inner.run", "shell", name, m.Result())
		return
	}

	switch m.Set(ice.MSG_RESULT); p {
	case "go":
		m.Option(cli.CMD_DIR, dir)
		if strings.HasSuffix(name, "test.go") {
			m.Cmdy(cli.SYSTEM, "go", "test", "-v", "./"+file)
		} else {
			m.Cmdy(cli.SYSTEM, "go", "run", "./"+file)
		}

	case "csv":
		m.CSV(m.Cmdx("nfs.cat", name))
	case "md":
		m.Cmdy("web.wiki.md.note", name)
	case "shy":
		m.Echo(strings.ReplaceAll(strings.Join(m.Cmd("web.wiki.word", name).Resultv(), ""), "\n", " "))
	}
}
func _inner_main(m *ice.Message, dir, file string) {
	p := _inner_ext(file)
	key := strings.TrimSuffix(path.Base(file), "."+p)
	switch p {
	case "godoc":
		m.Option(cli.CMD_DIR, dir)
		m.Echo(m.Cmdx(cli.SYSTEM, "go", "doc", key))

	case "man8", "man3", "man2", "man1":
		p := m.Cmdx(cli.SYSTEM, "man", strings.TrimPrefix(p, "man"), key)
		p = strings.Replace(p, "_\x08", "", -1)
		res := make([]byte, 0, len(p))
		for i := 0; i < len(p); i++ {
			switch p[i] {
			case '\x08':
				i++
			default:
				res = append(res, p[i])
			}
		}

		m.Echo(string(res))
	default:
		_inner_list(m, dir, file)
	}
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			INNER: {Name: "inner", Help: "编辑器", Value: kit.Data(
				"protect", kit.Dict("etc", "true", "var", "true", "usr", "true"),
				"source", kit.Dict(
					"makefile", "true",
					"c", "true", "h", "true",
					"sh", "true", "shy", "true", "py", "true",
					"mod", "true", "sum", "true",
					"go", "true", "js", "true",

					"md", "true", "csv", "true",
					"txt", "true", "url", "true",
					"conf", "true", "json", "true",
					"ts", "true", "tsx", "true", "vue", "true", "sass", "true",
				),
				"plug", kit.Dict(
					"py", kit.Dict(
						"prefix", kit.Dict("#", "comment"),
						"keyword", kit.Dict("print", "keyword"),
					),
					"md", kit.Dict("display", true, "profile", true),
					"csv", kit.Dict("display", true),
					"ts", kit.Dict(
						"prefix", kit.Dict("//", "comment"),
						"split", kit.Dict(
							"space", " ",
							"operator", "{[(.:,;!|)]}",
						),
						"keyword", kit.Dict(
							"import", "keyword",
							"from", "keyword",
							"new", "keyword",
							"as", "keyword",
							"const", "keyword",
							"export", "keyword",
							"default", "keyword",

							"if", "keyword",
							"return", "keyword",

							"class", "keyword",
							"extends", "keyword",
							"interface", "keyword",
							"declare", "keyword",
							"async", "keyword",
							"await", "keyword",
							"try", "keyword",
							"catch", "keyword",

							"function", "function",
							"arguments", "function",
							"console", "function",
							"this", "function",

							"string", "datatype",
							"number", "datatype",

							"true", "string",
							"false", "string",
						),
					),
					"tsx", kit.Dict("link", "ts"),
					"vue", kit.Dict("link", "ts"),
					"sass", kit.Dict("link", "ts"),
				),
				"show", kit.Dict(
					"sh", []string{"sh"},
					"py", []string{"python"},
					"js", []string{"node"},
				),
			)},
		},
		Commands: map[string]*ice.Command{
			INNER: {Name: "inner path=usr/demo file=hi.qrc line=1 查看:button=auto", Help: "编辑器", Meta: kit.Dict(
				"display", "/plugin/local/code/inner.js", "style", "editor",
			), Action: map[string]*ice.Action{
				"cmd": {Name: "cmd arg", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if m.Cmdy(kit.Split(arg[0])); !m.Hand {
						m.Cmdy(cli.SYSTEM, kit.Split(arg[0]))
					}
				}},

				"favor": {Name: "favor", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(web.FAVOR, arg, "extra", "extra.poster").Table(func(index int, value map[string]string, header []string) {
						m.Push("image", kit.Format(`<a title="%s" href="%s" target="_blank"><img src="%s" width=200></a>`,
							value["name"], value["text"], value["extra.poster"]))
						m.Push("video", kit.Format(`<video src="%s" controls></video>`, value["text"]))
					})
				}},
				"find": {Name: "find word", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					web.FavorList(m, arg[0], arg[1], arg[2:]...)
				}},

				"history": {Name: "history path name", Help: "历史", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(web.STORY, web.HISTORY, path.Join("./", arg[0], arg[1]))
					m.Copy(msg, ice.MSG_APPEND, "time", "count", "key")

					if len(arg) > 2 && arg[2] != "" {
						m.Echo(m.Cmd(web.STORY, web.INDEX, arg[2]).Result())
					}
				}},
				"commit": {Name: "commit path name", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(web.STORY, web.CATCH, "", path.Join("./", arg[0], arg[1]))
					m.Copy(msg, ice.MSG_APPEND, "time", "count", "key")
				}},
				"recover": {Name: "recover", Help: "复盘", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(web.STORY, web.HISTORY, path.Join("./", arg[0], arg[1])+".display")
					m.Copy(msg, ice.MSG_APPEND, "time", "count", "key", "drama")

					if len(arg) > 2 && arg[2] != "" {
						m.Echo(m.Cmd(web.STORY, web.INDEX, arg[2]).Result())
					}
				}},
				"record": {Name: "record", Help: "记录", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(web.STORY, web.CATCH, "display", path.Join("./", m.Option("path"), m.Option("name"))+".display", m.Option("display"))
					m.Copy(msg, ice.MSG_APPEND, "time", "count", "key")
				}},

				"log": {Name: "log path name", Help: "日志", Hand: func(m *ice.Message, arg ...string) {
					web.FavorList(m, "inner.run", "", "time", "id", "type", "name", "text")
				}},
				"run": {Name: "run path name", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					_inner_show(m, arg[0], arg[1])
				}},
				PLUG: {Name: "plug path name", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					_inner_plug(m, path.Join("./", arg[0], arg[1]))
				}},
				SAVE: {Name: "save path name content", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_inner_save(m, path.Join("./", arg[0], arg[1]), kit.Select(m.Option("content"), arg, 2))
				}},

				web.UPLOAD: {Name: "upload path name", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.CACHE, web.UPLOAD)
					m.Cmdy(web.CACHE, web.WATCH, m.Option(web.DATA), path.Join(m.Option("path"), m.Option("name")))
				}},
				mdb.SEARCH: {Name: "search type name text arg...", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.SEARCH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_inner_main(m, arg[0], kit.Select("", arg, 1))
			}},
		},
	}, nil)
}
