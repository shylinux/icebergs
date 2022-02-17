package code

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _binpack_write(o io.Writer, arg ...string) {
	for _, v := range arg {
		fmt.Fprint(o, v)
	}
	fmt.Fprintln(o)
}
func _binpack_file(m *ice.Message, name, file string) string {
	text := ""
	if f, e := os.Open(file); e == nil {
		defer f.Close()

		if b, e := ioutil.ReadAll(f); e == nil && len(b) > 0 {
			if list := strings.ReplaceAll(fmt.Sprintf("%v", b), ice.SP, ice.FS); len(list) > 0 {
				text = list[1 : len(list)-1]
			}
		}
	}
	return fmt.Sprintf("        \"%s\": []byte{%s},\n", name, text)
}
func _binpack_dir(m *ice.Message, pack *os.File, dir string) {
	m.Option(nfs.DIR_ROOT, dir)
	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_TYPE, nfs.CAT)

	m.Cmd(nfs.DIR, nfs.PWD).Sort(nfs.PATH).Table(func(index int, value map[string]string, head []string) {
		if path.Base(value[nfs.PATH]) == "binpack.go" {
			return
		}
		switch strings.Split(value[nfs.PATH], ice.PS)[0] {
		case "pluged", "trash":
			return
		}

		pack.WriteString(_binpack_file(m, path.Join(dir, value[nfs.PATH]), path.Join(dir, value[nfs.PATH])))
	})
	pack.WriteString(ice.NL)
}

func _binpack_can(m *ice.Message, pack *os.File, dir string) {
	m.Option(nfs.DIR_ROOT, dir)
	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_TYPE, nfs.CAT)

	for _, k := range []string{ice.FAVICON, ice.PROTO_JS, ice.FRAME_JS} {
		pack.WriteString(_binpack_file(m, ice.PS+k, path.Join(dir, k)))
	}
	for _, k := range []string{LIB, PAGE, PANEL, PLUGIN} {
		m.Cmd(nfs.DIR, k).Sort(nfs.PATH).Table(func(index int, value map[string]string, head []string) {
			pack.WriteString(_binpack_file(m, ice.PS+value[nfs.PATH], path.Join(dir, value[nfs.PATH])))
		})
	}
	pack.WriteString(ice.NL)
}
func _binpack_ctx(m *ice.Message, pack *os.File) {
	_binpack_dir(m, pack, ice.SRC_HELP)
	_binpack_dir(m, pack, ice.SRC)
}

const BINPACK = "binpack"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		BINPACK: {Name: "binpack path auto create remove export", Help: "打包", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
					if len(ice.Info.Pack) == 0 {
						return false
					}
					if ice.Dump(w, r.URL.Path, func(name string) { web.RenderType(w, name, "") }) {
						return true // 打包文件
					}
					return false
				})
				nfs.AddRewrite(func(msg *ice.Message, name string) []byte {
					if len(ice.Info.Pack) == 0 {
						return nil
					}
					if strings.HasPrefix(name, ice.SRC) && kit.FileExists(name) {
						return nil
					}
					if b, ok := ice.Info.Pack[name]; ok {
						m.Logs(BINPACK, len(b), name)
						return b // 打包文件
					}
					if b, ok := ice.Info.Pack[strings.TrimPrefix(name, ice.USR_VOLCANOS)]; ok && len(b) > 0 {
						m.Logs(BINPACK, len(b), name)
						return b // 打包文件
					}
					if b, ok := ice.Info.Pack[path.Join(m.Option(nfs.DIR_ROOT), name)]; ok && len(b) > 0 {
						m.Logs(BINPACK, len(b), name)
						return b // 打包文件
					}
					if b, ok := ice.Info.Pack[path.Join(ice.PS, name)]; ok && len(b) > 0 {
						m.Logs(BINPACK, len(b), name)
						return b // 打包文件
					}
					return nil
				})
			}},
			mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if f, p, e := kit.Create(ice.SRC_BINPACK_GO); m.Assert(e) {
					defer f.Close()
					defer m.Echo(p)

					_binpack_write(f, `package main`)
					_binpack_write(f)
					_binpack_write(f, `import (`)
					_binpack_write(f, `	ice "shylinux.com/x/icebergs"`)
					_binpack_write(f, `)`)
					_binpack_write(f)

					_binpack_write(f, `func init() {`)
					_binpack_write(f, `	ice.Info.Pack = map[string][]byte{`)

					_binpack_dir(m, f, ice.USR_LEARNING)
					_binpack_can(m, f, ice.USR_VOLCANOS)
					_binpack_dir(m, f, ice.USR_INTSHELL)
					_binpack_ctx(m, f)

					_binpack_write(f, `	}`)
					_binpack_write(f, `}`)
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
					m.Log_EXPORT(nfs.FILE, key, nfs.SIZE, len(value))
					m.Warn(nfs.MkdirAll(m, path.Dir(key)), "mkdir", key)
					m.Warn(ioutil.WriteFile(key, value, ice.MOD_FILE), "write", key)
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for k, v := range ice.Info.Pack {
				m.Push(mdb.NAME, k).Push(nfs.SIZE, len(v))
			}
			m.Sort(mdb.NAME)
		}},
	}})
}
