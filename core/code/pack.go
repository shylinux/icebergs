package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

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

	if list = strings.ReplaceAll(list, " ", ","); len(list) > 0 {
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

const WEBPACK = "webpack"
const BINPACK = "binpack"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			BINPACK: {Name: "binpack path auto create", Help: "打包", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name=demo from=src/main.go", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					name := kit.Keys(m.Option(kit.MDB_NAME), "go")
					if pack, p, e := kit.Create(path.Join(m.Conf(PUBLISH, kit.META_PATH), BINPACK, name)); m.Assert(e) {
						defer pack.Close()

						pack.WriteString(m.Cmdx(nfs.CAT, m.Option("from")))

						pack.WriteString("\n")
						pack.WriteString(`func init() {` + "\n")
						pack.WriteString(`    ice.BinPack = map[string][]byte{` + "\n")

						_pack_volcanos(m, pack, "usr/volcanos")
						_pack_dir(m, pack, "usr/learning")
						_pack_dir(m, pack, "usr/icebergs")
						_pack_dir(m, pack, "usr/toolkits")
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
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(PUBLISH, kit.META_PATH)))
				m.Option(nfs.DIR_TYPE, nfs.FILE)

				m.Cmdy(nfs.DIR, BINPACK).Table(func(index int, value map[string]string, head []string) {
					m.PushDownload(value[kit.MDB_PATH])
				})
			}},

			WEBPACK: {Name: "webpack path auto create", Help: "打包", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name=demo", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					css, _, e := kit.Create(path.Join(m.Conf(WEBPACK, kit.META_PATH), "cache.css"))
					m.Assert(e)
					defer css.Close()

					js, _, e := kit.Create(path.Join(m.Conf(WEBPACK, kit.META_PATH), "cache.js"))
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
						})
					}

					for _, k := range []string{"lib", "pane", "plugin"} {
						m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
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

					if f, _, e := kit.Create("usr/publish/webpack/" + m.Option("name") + ".js"); m.Assert(e) {
						defer f.Close()

						f.WriteString(`Volcanos.meta.pack = ` + kit.Formats(kit.UnMarshal(kit.Select("{}", m.Option("content")))))
					}

					m.Option(nfs.DIR_ROOT, "")
					if f, p, e := kit.Create("usr/publish/webpack/" + m.Option("name") + ".html"); m.Assert(e) {
						f.WriteString(fmt.Sprintf(_pack,
							m.Cmdx(nfs.CAT, "usr/volcanos/cache.css"),
							m.Cmdx(nfs.CAT, "usr/volcanos/index.css"),

							m.Cmdx(nfs.CAT, "usr/volcanos/proto.js"),
							m.Cmdx(nfs.CAT, "usr/volcanos/cache.js"),
							m.Cmdx(nfs.CAT, "usr/publish/webpack/"+m.Option("name")+".js"),
							m.Cmdx(nfs.CAT, "usr/volcanos/index.js"),
						))
						m.Echo(p)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
				m.Option(nfs.DIR_TYPE, nfs.FILE)
				m.Option(nfs.DIR_DEEP, true)

				m.Cmdy(nfs.DIR, WEBPACK).Table(func(index int, value map[string]string, head []string) {
					m.PushDownload(path.Join(m.Option(nfs.DIR_ROOT), value[kit.MDB_PATH]))
				})
			}},
		},
		Configs: map[string]*ice.Config{
			WEBPACK: {Name: WEBPACK, Help: "webpack", Value: kit.Data(kit.MDB_PATH, "usr/volcanos")},
			BINPACK: {Name: BINPACK, Help: "binpack", Value: kit.Data()},
		},
	})
}

const _pack = `
<!DOCTYPE html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=0.7,user-scalable=no">
    <title>volcanos</title>
    <link rel="shortcut icon" type="image/ico" href="favicon.ico">
    <style type="text/css">%s</style>
    <style type="text/css">%s</style>
</head>
<body>
<script>%s</script>
<script>%s</script>
<script>%s</script>
<script>%s</script>
<script>Volcanos.meta.webpack = true</script>
</body>
`
