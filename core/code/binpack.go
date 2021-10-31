package code

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _pack_write(o io.Writer, arg ...string) {
	for _, v := range arg {
		fmt.Fprint(o, v)
	}
	fmt.Fprintln(o)
}
func _pack_file(m *ice.Message, file string) string {
	list := ""
	if f, e := os.Open(file); e == nil {
		defer f.Close()

		if b, e := ioutil.ReadAll(f); e == nil {
			list = fmt.Sprintf("%v", b)
		}
	}

	if list = strings.ReplaceAll(list, ice.SP, ","); len(list) > 0 {
		return fmt.Sprintf(`[]byte{%v}`, list[1:len(list)-1])
	}
	return "[]byte{}"
}
func _pack_dir(m *ice.Message, pack *os.File, dir string) {
	m.Option(nfs.DIR_DEEP, ice.TRUE)
	m.Option(nfs.DIR_TYPE, nfs.CAT)
	m.Option(nfs.DIR_ROOT, dir)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		if path.Base(value[kit.MDB_PATH]) == "binpack.go" {
			return
		}
		switch strings.Split(value[kit.MDB_PATH], ice.PS)[0] {
		case "pluged", "trash":
			return
		}

		pack.WriteString(fmt.Sprintf("        \"%s\": %s,\n",
			path.Join(dir, value[kit.MDB_PATH]), _pack_file(m, path.Join(dir, value[kit.MDB_PATH]))))
	})
	pack.WriteString(ice.NL)
}

func _pack_volcanos(m *ice.Message, pack *os.File, dir string) {
	m.Option(nfs.DIR_DEEP, ice.TRUE)
	m.Option(nfs.DIR_TYPE, nfs.CAT)
	m.Option(nfs.DIR_ROOT, dir)

	for _, k := range []string{ice.FAVICON, ice.PROTO_JS, ice.FRAME_JS} {
		pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n", k, _pack_file(m, path.Join(dir, k))))
	}
	for _, k := range []string{"lib", "page", "panel", "plugin"} {
		m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
			pack.WriteString(fmt.Sprintf("        \"/%s\": %s,\n",
				value[kit.MDB_PATH], _pack_file(m, path.Join(dir, value[kit.MDB_PATH]))))
		})
	}
	pack.WriteString(ice.NL)
}
func _pack_ctx(m *ice.Message, pack *os.File) {
	_pack_dir(m, pack, ice.SRC_HELP)
	_pack_dir(m, pack, ice.SRC)
	_pack_write(pack)
}

const BINPACK = "binpack"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		BINPACK: {Name: "binpack path auto create remove export", Help: "打包", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if pack, p, e := kit.Create(ice.SRC_BINPACK_GO); m.Assert(e) {
					defer pack.Close()
					defer m.Echo(p)

					_pack_write(pack, `package main`)
					_pack_write(pack)
					_pack_write(pack, `import (`)
					_pack_write(pack, `	ice "shylinux.com/x/icebergs"`)
					_pack_write(pack, `)`)
					_pack_write(pack)

					_pack_write(pack, `func init() {`)
					_pack_write(pack, `	ice.Info.Pack = map[string][]byte{`)

					_pack_volcanos(m, pack, ice.USR_VOLCANOS)
					_pack_dir(m, pack, ice.USR_LEARNING)
					_pack_dir(m, pack, ice.USR_INTSHELL)
					_pack_ctx(m, pack)

					_pack_write(pack, `	}`)
					_pack_write(pack, `}`)
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				ice.Info.Pack = map[string][]byte{}
			}},
			mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				for key, value := range ice.Info.Pack {
					if strings.HasPrefix(key, ice.PS) {
						key = ice.USR_VOLCANOS + key
					}
					m.Warn(os.MkdirAll(path.Dir(key), ice.MOD_DIR), "mkdir", key)
					m.Warn(ioutil.WriteFile(key, value, ice.MOD_FILE), "write", key)
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for k, v := range ice.Info.Pack {
				m.Push(kit.MDB_NAME, k)
				m.Push(kit.MDB_SIZE, len(v))
			}
			m.Sort(kit.MDB_NAME)
		}},
	}})
}
