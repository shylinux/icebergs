package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func _pack_file(m *ice.Message, file string) string {
	list := ""
	if f, e := os.Open(file); e == nil {
		defer f.Close()
		if b, e := ioutil.ReadAll(f); e == nil {
			list = fmt.Sprintf("%v", b)
		}
	}
	list = strings.ReplaceAll(list, " ", ",")
	if len(list) > 0 {
		return fmt.Sprintf(`[]byte{%v}`, list[1:len(list)-1])
	}
	return "[]byte{}"
}
func _pack_dir(m *ice.Message, pack *os.File, dir string) {
	m.Option(nfs.DIR_ROOT, dir)
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		switch strings.Split(value[kit.MDB_PATH], "/")[0] {
		case "pluged", "trash":
			return
		}

		pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n",
			path.Join(dir, value[kit.MDB_PATH]), _pack_file(m, path.Join(dir, value[kit.MDB_PATH]))))
	})
	pack.WriteString("\n")
}

func _pack_volcanos(m *ice.Message, pack *os.File, dir string) {
	m.Option(nfs.DIR_ROOT, dir)
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	for _, k := range []string{"favicon.ico", "index.html", "index.css", "index.js", "proto.js", "frame.js", "cache.js", "cache.css"} {
		pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n",
			kit.Select("", k, k != "index.html"), _pack_file(m, path.Join(dir, k))))
	}
	for _, k := range []string{"lib", "pane", "plugin"} {
		m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
			pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n",
				value[kit.MDB_PATH], _pack_file(m, path.Join(dir, value[kit.MDB_PATH]))))
		})
	}
	pack.WriteString("\n")
}
func _pack_contexts(m *ice.Message, pack *os.File) {
	for _, k := range []string{"src/main.go", "src/main.shy", "src/main.svg"} {
		pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n",
			k, _pack_file(m, k)))
	}
	pack.WriteString("\n")
}

const (
	WEBPACK = "webpack"
	BINPACK = "binpack"
	MODPACK = "modpack"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			WEBPACK: {Name: "webpack path auto 打包", Help: "打包", Action: map[string]*ice.Action{
				"pack": {Name: "pack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					css, _, e := kit.Create(path.Join(m.Conf(WEBPACK, kit.META_PATH), "cache.css"))
					m.Assert(e)
					defer css.Close()

					js, p, e := kit.Create(path.Join(m.Conf(WEBPACK, kit.META_PATH), "cache.js"))
					m.Assert(e)
					defer js.Close()

					m.Option(nfs.DIR_ROOT, m.Conf(WEBPACK, kit.META_PATH))
					m.Option(nfs.DIR_TYPE, nfs.FILE)
					m.Option(nfs.DIR_DEEP, true)

					for _, k := range []string{"lib", "pane", "plugin"} {
						m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
							if strings.HasSuffix(value[kit.MDB_PATH], ".css") {
								js.WriteString(`Volcanos.meta.cache["` + path.Join("/", value[kit.MDB_PATH]) + "\"] = []\n")
								css.WriteString(m.Cmdx(nfs.CAT, value[kit.MDB_PATH]))
							}

							if strings.HasSuffix(value[kit.MDB_PATH], ".js") {
								js.WriteString(`_can_name = "` + path.Join("/", value[kit.MDB_PATH]) + "\"\n")
								js.WriteString(m.Cmdx(nfs.CAT, value[kit.MDB_PATH]))
							}
						})
					}

					for _, k := range []string{"frame.js"} {
						js.WriteString(`_can_name = "` + path.Join("/", k) + "\"\n")
						js.WriteString(m.Cmdx(nfs.CAT, k))
					}
					m.Echo(p)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
				m.Option(nfs.DIR_TYPE, nfs.FILE)
				m.Option(nfs.DIR_DEEP, true)

				if len(arg) == 0 {
					m.Cmdy(nfs.DIR, WEBPACK).Table(func(index int, value map[string]string, head []string) {
						m.Push(kit.MDB_LINK, m.Cmdx(mdb.RENDER, web.RENDER.Download, "/publish/"+value[kit.MDB_PATH]))
					})
					return
				}

				m.Cmdy(nfs.CAT, arg[0])
			}},
			BINPACK: {Name: "binpack path auto 打包", Help: "打包", Action: map[string]*ice.Action{
				"pack": {Name: "pack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					name := kit.Keys(kit.Select(m.Option(kit.MDB_NAME), "demo"), "go")
					if pack, p, e := kit.Create(path.Join(m.Conf(PUBLISH, kit.META_PATH), BINPACK, name)); m.Assert(e) {
						defer pack.Close()

						pack.WriteString(m.Cmdx(nfs.CAT, "src/main.go"))

						pack.WriteString("\n")
						pack.WriteString(`func init() {` + "\n")
						pack.WriteString(`    ice.BinPack = map[string][]byte{` + "\n")

						_pack_volcanos(m, pack, "usr/volcanos")
						_pack_dir(m, pack, "usr/learning")
						_pack_dir(m, pack, "usr/icebergs")
						_pack_dir(m, pack, "usr/intshell")
						_pack_contexts(m, pack)

						pack.WriteString(`    }` + "\n")
						pack.WriteString(`}` + "\n")
						m.Echo(p)
					}

					m.Option(cli.CMD_DIR, path.Join(m.Conf(PUBLISH, kit.META_PATH), BINPACK))
					m.Cmd(COMPILE, "windows", "amd64", name)
					m.Cmd(COMPILE, "darwin", "amd64", name)
					m.Cmd(COMPILE, "linux", "amd64", name)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
				m.Option(nfs.DIR_TYPE, nfs.FILE)
				m.Option(nfs.DIR_DEEP, true)

				m.Cmdy(nfs.DIR, BINPACK).Table(func(index int, value map[string]string, head []string) {
					m.Push(kit.MDB_LINK, m.Cmdx(mdb.RENDER, web.RENDER.Download, "/publish/"+value[kit.MDB_PATH]))
				})
			}},
			MODPACK: {Name: "modpack path=auto auto 创建", Help: "打包", Meta: kit.Dict(
				"style", "editor", "创建", kit.List("_input", "text", "name", "name"),
			), Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Option("name", "hi")
					m.Option("help", "hello")
					for i := 0; i < len(arg)-1; i += 2 {
						m.Option(arg[i], arg[i+1])
					}

					// 生成文件
					name := m.Option(kit.MDB_NAME)
					os.Mkdir(path.Join("src/", name), ice.MOD_DIR)
					kit.Fetch(m.Confv(MODPACK, "meta.base"), func(key string, value string) {
						p := path.Join("src/", name, name+"."+key)
						if _, e := os.Stat(p); e != nil && os.IsNotExist(e) {
							if f, p, e := kit.Create(p); m.Assert(e) {
								if b, e := kit.Render(value, m); m.Assert(e) {
									if n, e := f.Write(b); m.Assert(e) {
										m.Log_EXPORT(kit.MDB_FILE, p, kit.MDB_SIZE, n)
									}
								}
							}
						}
					})
					defer m.Cmdy(nfs.DIR, "src/"+name)

					// 模块名称
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

					// 检查模块
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

					// 插入模块
					if f, e := os.Open("src/main.go"); m.Assert(e) {
						defer f.Close()
						if b, e := ioutil.ReadAll(f); m.Assert(e) {
							if f, e := os.Create("src/main.go"); m.Assert(e) {
								defer f.Close()

								for bio := bufio.NewScanner(bytes.NewBuffer(b)); bio.Scan(); {
									f.WriteString(bio.Text())
									f.WriteString("\n")
									if strings.HasPrefix(bio.Text(), "import (") {
										m.Info("src/main.go import: %v", mod+"/src/"+name)
										f.WriteString("\t_ \"" + mod + "/src/" + name + `"`)
										f.WriteString("\n\n")
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
			WEBPACK: {Name: WEBPACK, Help: "webpack", Value: kit.Data(
				kit.MDB_PATH, "usr/volcanos",
			)},
			BINPACK: {Name: BINPACK, Help: "binpack", Value: kit.Data()},
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
		"{{.Option "name"}}": {Name: "{{.Option "name"}}", Help: "{{.Option "help"}}", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		"{{.Option "name"}}": {Name: "{{.Option "name"}}", Help: "{{.Option "help"}}", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
