package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

const (
	INNER  = "inner"
	QRCODE = "qrcode"
	VEDIO  = "vedio"
)

func _inner_protect(m *ice.Message, name string) bool {
	ls := strings.Split(name, "/")
	return !m.Right(ls) && m.Conf(INNER, kit.Keys("meta.protect", ls[0])) == "true"
}
func _inner_binary(m *ice.Message, name string) bool {
	return m.Conf(INNER, kit.Keys("meta.binary", _inner_ext(name))) == "true"
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

func _inner_list(m *ice.Message, name string) {
	if _inner_sub(m, "list", name) {
		return
	}

	if m.Set(ice.MSG_RESULT); strings.HasSuffix(name, "/") || !_inner_binary(m, name) {
		m.Cmdy("nfs.dir", name, "file size time")
		return
	}
	m.Echo(name)
}
func _inner_save(m *ice.Message, name, text string) {
	if _inner_sub(m, "save", name) {
		return
	}

	if f, e := os.Create(name); m.Assert(e) {
		defer f.Close()
		m.Cmd(web.FAVOR, "inner.save", "shell", name, text)
		if n, e := f.WriteString(text); m.Assert(e) {
			m.Log_EXPORT("file", name, "size", n)
		}
	}
}
func _inner_plug(m *ice.Message, name string) {
	if _inner_sub(m, "plug", name) {
		return
	}

	p := _inner_ext(name)
	if ls := m.Confv(INNER, kit.Keys("meta.plug", p)); ls != nil {
		m.Echo(kit.Format(ls))
		return
	}

	m.Echo("{}")
}
func _inner_show(m *ice.Message, name string) {
	if _inner_sub(m, "show", name) {
		return
	}

	p := _inner_ext(name)
	if ls := kit.Simple(m.Confv(INNER, kit.Keys("meta.show", p))); len(ls) > 0 {
		m.Cmdy(ice.CLI_SYSTEM, ls, name)
		m.Set(ice.MSG_APPEND)
		m.Cmd(web.FAVOR, "inner.run", "shell", name, m.Result())
		return
	}

	switch m.Set(ice.MSG_RESULT); p {
	case "csv":
		m.CSV(m.Cmdx("nfs.cat", name))
	case "md":
		m.Cmdy("web.wiki.md.note", name)
	case "shy":
		m.Echo(strings.ReplaceAll(strings.Join(m.Cmd("web.wiki.word", name).Resultv(), ""), "\n", " "))
	}
}
func _inner_main(m *ice.Message, arg ...string) {
	if len(arg) > 2 && arg[2] != "" {
		web.StoryIndex(m, arg[2])
		return
	}
	_inner_list(m, path.Join(arg...))
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			INNER: {Name: "inner", Help: "编辑器", Value: kit.Data(
				"protect", kit.Dict("etc", "true", "var", "true", "usr", "true"),
				"binary", kit.Dict("bin", "true", "gz", "true"),
				"plug", kit.Dict(
					"py", kit.Dict(
						"prefix", kit.Dict("#", "comment"),
						"keyword", kit.Dict("print", "keyword"),
					),
					"md", kit.Dict("display", true, "profile", true),
					"csv", kit.Dict("display", true),
				),
				"show", kit.Dict(
					"sh", []string{"bash"},
					"py", []string{"python"},
					"go", []string{"go", "run"},
					"js", []string{"node"},
				),
			)},
		},
		Commands: map[string]*ice.Command{
			INNER: {Name: "inner path=auto name=auto auto", Help: "编辑器", Meta: map[string]interface{}{
				"display": "/plugin/inner.js", "style": "editor",
			}, Action: map[string]*ice.Action{
				"cmd": {Name: "cmd arg", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if m.Cmdy(kit.Split(arg[0])); !m.Hand {
						m.Cmdy(ice.CLI_SYSTEM, kit.Split(arg[0]))
					}
				}},

				"favor": {Name: "favor", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(ice.WEB_FAVOR, arg, "extra", "extra.poster").Table(func(index int, value map[string]string, header []string) {
						m.Push("image", kit.Format(`<a title="%s" href="%s" target="_blank"><img src="%s" width=200></a>`,
							value["name"], value["text"], value["extra.poster"]))
						m.Push("video", kit.Format(`<video src="%s" controls></video>`, value["text"]))
					})
				}},
				"find": {Name: "find word", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					web.FavorList(m, arg[0], arg[1], arg[2:]...)
				}},
				"upload": {Name: "upload path name", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					web.StoryWatch(m, m.Option("data"), path.Join(m.Option("path"), m.Option("name")))
				}},
				"project": {Name: "project path", Help: "项目", Hand: func(m *ice.Message, arg ...string) {
					_inner_list(m, path.Join("./", kit.Select("", arg, 0))+"/")
				}},

				"history": {Name: "history path name", Help: "历史", Hand: func(m *ice.Message, arg ...string) {
					msg := web.StoryHistory(m.Spawn(), path.Join("./", arg[0], arg[1]))
					m.Copy(msg, ice.MSG_APPEND, "time", "count", "key")

					if len(arg) > 2 && arg[2] != "" {
						m.Echo(web.StoryIndex(m.Spawn(), arg[2]).Result())
					}
				}},
				"commit": {Name: "commit path name", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
					web.StoryCatch(m, "", path.Join("./", arg[0], arg[1]))
				}},
				"recover": {Name: "recover", Help: "复盘", Hand: func(m *ice.Message, arg ...string) {
					msg := web.StoryHistory(m.Spawn(), path.Join("./", arg[0], arg[1])+".display")
					m.Copy(msg, ice.MSG_APPEND, "time", "count", "key", "drama")

					if len(arg) > 2 && arg[2] != "" {
						m.Echo(web.StoryIndex(m.Spawn(), arg[2]).Result())
					}
				}},
				"record": {Name: "record", Help: "记录", Hand: func(m *ice.Message, arg ...string) {
					web.StoryAdd(m, "display", path.Join("./", m.Option("path"), m.Option("name"))+".display", m.Option("display"))
				}},

				"log": {Name: "log path name", Help: "日志", Hand: func(m *ice.Message, arg ...string) {
					web.FavorList(m, "inner.run", "", "time", "id", "type", "name", "text")
				}},
				"run": {Name: "run path name", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					_inner_show(m, path.Join("./", arg[0], arg[1]))
				}},
				"plug": {Name: "plug path name", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					_inner_plug(m, path.Join("./", arg[0], arg[1]))
				}},
				"save": {Name: "save path name content", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_inner_save(m, path.Join("./", arg[0], arg[1]), kit.Select(m.Option("content"), arg, 2))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { _inner_main(m, arg...) }},
		},
	}, nil)
}
