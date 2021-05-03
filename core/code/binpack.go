package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func _pack_write(o *os.File, arg ...string) {
	for _, v := range arg {
		o.WriteString(v)
	}
	o.WriteString("\n")
}
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
	m.Option(nfs.DIR_TYPE, nfs.CAT)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		switch strings.Split(value[kit.MDB_PATH], "/")[0] {
		case "pluged", "trash":
			return
		}

		pack.WriteString(fmt.Sprintf("        \"%s\": %s,\n",
			path.Join(dir, value[kit.MDB_PATH]), _pack_file(m, path.Join(dir, value[kit.MDB_PATH]))))
	})
	pack.WriteString("\n")
}

func _pack_volcanos(m *ice.Message, pack *os.File, dir string) {
	m.Option(nfs.DIR_ROOT, dir)
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.CAT)

	for _, k := range []string{"favicon.ico", "proto.js", "frame.js"} {
		pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n", k, _pack_file(m, path.Join(dir, k))))
	}
	for _, k := range []string{"lib", "page", "panel", "plugin"} {
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
				mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					if pack, p, e := kit.Create("src/binpack.go"); m.Assert(e) {
						defer pack.Close()

						_pack_write(pack, `package main`)
						_pack_write(pack, "")
						_pack_write(pack, `import (`)
						_pack_write(pack, `	ice "github.com/shylinux/icebergs"`)
						_pack_write(pack, `)`)
						_pack_write(pack, "")

						_pack_write(pack, `func init() {`)
						_pack_write(pack, `    ice.BinPack = map[string][]byte{`)

						_pack_volcanos(m, pack, "usr/volcanos")
						_pack_dir(m, pack, "usr/learning")
						// _pack_dir(m, pack, "usr/icebergs")
						// _pack_dir(m, pack, "usr/toolkits")
						_pack_dir(m, pack, "usr/intshell")
						// _pack_contexts(m, pack)

						_pack_write(pack, `    }`)
						_pack_write(pack, `}`)
						m.Echo(p)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				for k, v := range ice.BinPack {
					m.Push(kit.MDB_NAME, k)
					m.Push(kit.MDB_SIZE, len(v))
				}
				m.Sort(kit.MDB_NAME)
			}},
		},
	})
}
