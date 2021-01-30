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

	for _, k := range []string{"favicon.ico", "proto.js", "frame.js", "index.html"} {
		pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n",
			kit.Select("", k, k != "index.html"), _pack_file(m, path.Join(dir, k))))
	}
	for _, k := range []string{"lib", "page", "pane", "plugin"} {
		m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
			pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n",
				value[kit.MDB_PATH], _pack_file(m, path.Join(dir, value[kit.MDB_PATH]))))
		})
	}
	pack.WriteString("\n")
}
func _pack_contexts(m *ice.Message, pack *os.File) {
	_pack_dir(m, pack, "src")
	pack.WriteString("\n")
}

const BINPACK = "binpack"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			BINPACK: {Name: BINPACK, Help: "binpack", Value: kit.Data()},
		},
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
					m.Cmd(COMPILE, name)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(PUBLISH, kit.META_PATH)))
				m.Option(nfs.DIR_TYPE, nfs.FILE)

				m.Cmdy(nfs.DIR, BINPACK).Table(func(index int, value map[string]string, head []string) {
					m.PushDownload(value[kit.MDB_PATH])
				})
			}},
		},
	})
}
