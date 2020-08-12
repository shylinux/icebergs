package code

import (
	"bufio"
	"bytes"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func _pack_volcanos(m *ice.Message, pack *os.File) {
	m.Option(nfs.DIR_ROOT, "usr/volcanos")
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	for _, k := range []string{"favicon.ico", "index.html", "index.css", "index.js", "proto.js", "frame.js", "cache.js", "cache.css"} {
		what := ""
		if f, e := os.Open("usr/volcanos/" + k); e == nil {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e == nil {
				what = fmt.Sprintf("%v", b)
			}
		}
		if k == "index.html" {
			k = ""
		}
		what = strings.ReplaceAll(what, " ", ",")
		pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "/"+k, what[1:len(what)-1]))
	}
	for _, k := range []string{"lib", "pane", "plugin"} {
		m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
			what := ""
			if f, e := os.Open("usr/volcanos/" + value["path"]); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					what = fmt.Sprintf("%v", b)
				}
			}
			what = strings.ReplaceAll(what, " ", ",")
			pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "/"+value["path"], what[1:len(what)-1]))
		})
	}
	pack.WriteString("\n")
}
func _pack_learning(m *ice.Message, pack *os.File) {
	m.Option(nfs.DIR_ROOT, "usr/learning")
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		what := ""
		if f, e := os.Open("usr/learning/" + value["path"]); e == nil {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e == nil {
				what = fmt.Sprintf("%v", b)
			}
		}
		what = strings.ReplaceAll(what, " ", ",")
		pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "usr/learning/"+value["path"], what[1:len(what)-1]))
	})
	pack.WriteString("\n")
}
func _pack_icebergs(m *ice.Message, pack *os.File) {
	m.Option(nfs.DIR_ROOT, "usr/icebergs")
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		what := ""
		if strings.HasPrefix(value["path"], "pack") {
			return
		}
		if f, e := os.Open("usr/icebergs/" + value["path"]); e == nil {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e == nil {
				what = fmt.Sprintf("%v", b)
			}
		}
		if len(what) > 0 {
			what = strings.ReplaceAll(what, " ", ",")
			pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "usr/icebergs/"+value["path"], what[1:len(what)-1]))
		}
	})
	pack.WriteString("\n")
}
func _pack_intshell(m *ice.Message, pack *os.File) {
	m.Option(nfs.DIR_ROOT, "usr/intshell")
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		if strings.HasPrefix(value["path"], "pluged") {
			return
		}
		what := ""
		if f, e := os.Open("usr/intshell/" + value["path"]); e != nil {
			return
		} else {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e != nil {
				return
			} else {
				what = fmt.Sprintf("%v", b)
			}
		}
		what = strings.ReplaceAll(what, " ", ",")
		pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "usr/intshell/"+value["path"], what[1:len(what)-1]))
	})
}

const (
	WEBPACK = "webpack"
	BINPACK = "binpack"
	MODPACK = "modpack"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			WEBPACK: {Name: "webpack", Help: "打包", Action: map[string]*ice.Action{
				"pack": {Name: "pack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					m.Option(nfs.DIR_ROOT, "usr/volcanos")
					m.Option(nfs.DIR_DEEP, "true")
					m.Option(nfs.DIR_TYPE, nfs.FILE)

					js, p, e := kit.Create("usr/volcanos/cache.js")
					m.Assert(e)
					defer js.Close()
					m.Echo(p)

					css, _, e := kit.Create("usr/volcanos/cache.css")
					m.Assert(e)
					defer css.Close()

					for _, k := range []string{"lib", "pane", "plugin"} {
						m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
							if strings.HasSuffix(value["path"], ".css") {
								js.WriteString(`Volcanos.meta.cache["` + path.Join("/", value["path"]) + "\"] = []\n")
								css.WriteString(m.Cmdx(nfs.CAT, "usr/volcanos/"+value["path"]))
							}
							if strings.HasSuffix(value["path"], ".js") {
								js.WriteString(`_can_name = "` + path.Join("/", value["path"]) + "\"\n")
								js.WriteString(m.Cmdx(nfs.CAT, "usr/volcanos/"+value["path"]))
							}
						})
					}
					for _, k := range []string{"frame.js"} {
						js.WriteString(`_can_name = "` + path.Join("/", k) + "\"\n")
						js.WriteString(m.Cmdx(nfs.CAT, "usr/volcanos/"+k))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, "usr/volcanos")
				m.Option(nfs.DIR_TYPE, nfs.FILE)
				m.Option(nfs.DIR_DEEP, "true")
				m.Cmdy(nfs.DIR, "pack")
				m.Table(func(index int, value map[string]string, head []string) {
					m.Push("link", m.Cmdx(mdb.RENDER, web.RENDER.Download, "/"+value["path"]))
				})
			}},
			BINPACK: {Name: "binpack", Help: "打包", Action: map[string]*ice.Action{
				"pack": {Name: "pack", Help: "pack", Hand: func(m *ice.Message, arg ...string) {
					pack, p, e := kit.Create("usr/icebergs/pack/binpack.go")
					m.Assert(e)
					defer pack.Close()

					pack.WriteString(`package pack` + "\n\n")
					pack.WriteString(`import "github.com/shylinux/icebergs"` + "\n\n")
					pack.WriteString(`func init() {` + "\n")
					pack.WriteString(`    ice.BinPack = map[string][]byte{` + "\n")

					_pack_volcanos(m, pack)
					_pack_learning(m, pack)
					_pack_icebergs(m, pack)
					_pack_intshell(m, pack)

					pack.WriteString(`    }` + "\n")
					pack.WriteString(`}` + "\n")
					m.Echo(p)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, "usr/icebergs")
				m.Option(nfs.DIR_TYPE, nfs.FILE)
				m.Option(nfs.DIR_DEEP, "true")
				m.Cmdy(nfs.DIR, "pack")
				m.Table(func(index int, value map[string]string, head []string) {
					m.Push("link", m.Cmdx(mdb.RENDER, web.RENDER.Download, value["path"], "/share/local/usr/icebergs/"+value["path"]))
				})
			}},
			MODPACK: {Name: "modpack path=auto 查看:button 返回:button 创建:button", Help: "打包", Meta: kit.Dict(
				"style", "editor", "创建", kit.List("_input", "text", "name", "name"),
			), Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Option("name", "hi")
					m.Option("help", "hello")
					for i := 0; i < len(arg)-1; i += 2 {
						m.Option(arg[i], arg[i+1])
					}

					name := m.Option("name")
					os.Mkdir(path.Join("src/", name), ice.MOD_DIR)
					kit.Fetch(m.Confv(MODPACK, "meta.base"), func(key string, value string) {
						p := path.Join("src/", name, name+"."+key)
						if _, e := os.Stat(p); e != nil && os.IsNotExist(e) {
							if f, p, e := kit.Create(p); m.Assert(e) {
								if b, e := kit.Render(value, m); m.Assert(e) {
									if n, e := f.Write(b); m.Assert(e) {
										m.Log_EXPORT("file", p, "size", n)
									}
								}
							}
						}
					})
					defer m.Cmdy(nfs.DIR, "src/"+name)

					mod := ""
					if f, e := os.Open("go.mod"); e == nil {
						defer f.Close()
						for bio := bufio.NewScanner(f); bio.Scan(); {
							if strings.HasPrefix(bio.Text(), "module") {
								mod = strings.Split(bio.Text(), " ")[1]
								break
							}
						}
					}

					begin, has := false, false
					if f, e := os.Open("src/main.go"); e == nil {
						for bio := bufio.NewScanner(f); bio.Scan(); {
							if strings.HasPrefix(strings.TrimSpace(bio.Text()), "//") {
								continue
							}
							if strings.HasPrefix(bio.Text(), "import") {
								if strings.Contains(bio.Text(), mod+"/src/"+name) {
									has = true
								}
								continue
							}
							if strings.HasPrefix(bio.Text(), "import (") {
								begin = true
								continue
							}
							if strings.HasPrefix(bio.Text(), ")") {
								begin = false
								continue
							}
							if begin {
								if strings.Contains(bio.Text(), mod+"/src/"+name) {
									has = true
								}
							}
						}
						f.Close()
					}
					if has {
						return
					}

					if f, e := os.Open("src/main.go"); e == nil {
						if b, e := ioutil.ReadAll(f); e == nil {
							f.Close()

							if f, e := os.Create("src/main.go"); e == nil {
								defer f.Close()
								for bio := bufio.NewScanner(bytes.NewBuffer(b)); bio.Scan(); {
									f.WriteString(bio.Text())
									f.WriteString("\n")
									if strings.HasPrefix(bio.Text(), "import (") {
										m.Info("src/main.go import: %v", mod+"/src/"+name)
										f.WriteString("\t_ \"" + mod + "/src/" + name + `"`)
										f.WriteString("\n")
										f.WriteString("\n")
									}
								}
							}
						}
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, "src")
				if m.Cmdy(nfs.DIR, kit.Select("", arg, 0)); len(m.Resultv()) > 0 {
					m.Option("_display", "/plugin/local/code/inner.js")
				}
			}},
		},
		Configs: map[string]*ice.Config{
			MODPACK: {Name: MODPACK, Help: "modpack", Value: kit.Data(
				"base", kit.Dict(
					"shy", `title {{.Option "name"}}
`,
					"go", `package {{.Option "name"}}

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	kit "github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "{{.Option "name"}}", Help: "{{.Option "help"}}",
	Configs: map[string]*ice.Config{
		"{{.Option "name"}}": {Name: "{{.Option "name"}}", Help: "{{.Option "name"}}", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		"{{.Option "name"}}": {Name: "{{.Option "name"}}", Help: "{{.Option "name"}}", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello {{.Option "name"}} world")
		}},
	},
}

func init() { chat.Index.Register(Index, &web.Frame{}) }
`,
				),
			)},
		},
	}, nil)
}
