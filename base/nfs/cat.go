package nfs

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _cat_ext(name string) string {
	return strings.ToLower(kit.Select(path.Base(name), strings.TrimPrefix(path.Ext(name), ".")))
}
func _cat_right(m *ice.Message, name string) bool {
	switch strings.Split(name, "/")[0] {
	case kit.SSH_ETC, kit.SSH_VAR:
		if m.Warn(m.Option(ice.MSG_USERROLE) == aaa.VOID, ice.ErrNotRight, "of", name) {
			return false
		}
	}
	return true
}
func _cat_find(m *ice.Message, name string) io.ReadCloser {
	if f, e := os.Open(path.Join(m.Option(DIR_ROOT), name)); e == nil {
		return f
	}

	if b, ok := ice.Info.BinPack[name]; ok {
		m.Logs("binpack", len(b), name)
		return kit.NewReadCloser(bytes.NewBuffer(b))
	}

	msg := m.Cmd("web.spide", "dev", "raw", "GET", path.Join("/share/local/", name))
	if msg.Result(0) == ice.ErrWarn {
		return kit.NewReadCloser(bytes.NewBufferString(""))
	}
	return kit.NewReadCloser(bytes.NewBufferString(msg.Result()))
}
func _cat_show(m *ice.Message, name string) {
	if !_cat_right(m, name) {
		return // 没有权限
	}

	// 本地文件
	f := _cat_find(m, name)
	defer f.Close()

	switch cb := m.Optionv(kit.Keycb(CAT)).(type) {
	case func(string, int) string:
		list := []string{}
		bio := bufio.NewScanner(f)
		for i := 0; bio.Scan(); i++ {
			list = append(list, cb(bio.Text(), i))
		}
		m.Echo(strings.Join(list, "\n") + "\n")

	case func(string, int):
		bio := bufio.NewScanner(f)
		for i := 0; bio.Scan(); i++ {
			cb(bio.Text(), i)
		}

	default:
		buf := make([]byte, ice.MOD_BUFS)
		for begin := 0; true; {
			n, e := f.Read(buf[begin:])
			m.Warn(e != nil && e != io.EOF, e)
			m.Log_IMPORT(kit.MDB_FILE, name, kit.MDB_SIZE, n)
			if begin += n; begin < len(buf) {
				buf = buf[:begin]
				break
			}
			buf = append(buf, make([]byte, ice.MOD_BUFS)...)
		}
		m.Echo(string(buf))
	}
}

const (
	PATH = "path"
	SIZE = "size"
)
const CAT = "cat"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CAT: {Name: CAT, Help: "文件", Value: kit.Data(
				kit.SSH_SOURCE, kit.Dict(
					"sh", "true", "shy", "true", "py", "true",
					"go", "true", "vim", "true", "js", "true",
					"conf", "true", "json", "true",
					"makefile", "true",
					"yml", "true",
				),
			)},
		},
		Commands: map[string]*ice.Command{
			CAT: {Name: "cat path auto", Help: "文件", Action: map[string]*ice.Action{
				mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_cat_show(m, path.Join(arg[2], arg[1]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
					m.Cmdy(DIR, arg)
					return
				}
				_cat_show(m, arg[0])
			}},
		},
	})
}
