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

type ReadCloser struct {
	r io.Reader
}

func (r *ReadCloser) Read(buf []byte) (int, error) {
	return r.r.Read(buf)
}
func (r *ReadCloser) Close() error {
	if c, ok := r.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
func NewReadCloser(r io.Reader) *ReadCloser {
	return &ReadCloser{r: r}
}

func _cat_right(m *ice.Message, name string) bool {
	switch ls := strings.Split(name, "/"); ls[0] {
	case ice.USR:
		switch kit.Select("", ls, 1) {
		case "local":
			if m.Warn(m.Option(ice.MSG_USERROLE) == aaa.VOID, ice.ErrNotRight, "of", name) {
				return false
			}
		}
	case ice.ETC, ice.VAR:
		if m.Warn(m.Option(ice.MSG_USERROLE) == aaa.VOID, ice.ErrNotRight, "of", name) {
			return false
		}
	}
	return true
}
func _cat_find(m *ice.Message, name string) io.ReadCloser {
	// if m.Option("content") != "" {
	// 	return NewReadCloser(bytes.NewBufferString(m.Option("content")))
	// }

	if f, e := os.Open(path.Join(m.Option(DIR_ROOT), name)); e == nil {
		return f
	}

	if b, ok := ice.Info.Pack[name]; ok {
		m.Logs("binpack", len(b), name)
		return NewReadCloser(bytes.NewBuffer(b))
	}

	msg := m.Cmd("web.spide", ice.DEV, "raw", "GET", path.Join("/share/local/", name))
	if msg.Result(0) == ice.ErrWarn {
		return NewReadCloser(bytes.NewBufferString(""))
	}
	return NewReadCloser(bytes.NewBufferString(msg.Result()))
}
func _cat_show(m *ice.Message, name string) {
	if !_cat_right(m, name) {
		return // 没有权限
	}

	f := _cat_find(m, name)
	defer f.Close()

	switch cb := m.Optionv(kit.Keycb(CAT)).(type) {
	case func(string, int) string:
		list := []string{}
		for bio, i := bufio.NewScanner(f), 0; bio.Scan(); i++ {
			list = append(list, cb(bio.Text(), i))
		}
		m.Echo(strings.Join(list, ice.NL) + ice.NL)

	case func(string, int):
		for bio, i := bufio.NewScanner(f), 0; bio.Scan(); i++ {
			cb(bio.Text(), i)
		}

	case func(string):
		for bio := bufio.NewScanner(f); bio.Scan(); {
			cb(bio.Text())
		}

	default:
		buf := make([]byte, ice.MOD_BUFS)
		for begin := 0; true; {
			if n, e := f.Read(buf[begin:]); !m.Warn(e != nil && e != io.EOF, e) {
				m.Log_IMPORT(kit.MDB_FILE, name, kit.MDB_SIZE, n)
				if begin += n; begin < len(buf) {
					buf = buf[:begin]
					break
				}
				buf = append(buf, make([]byte, ice.MOD_BUFS)...)
			}
		}
		m.Echo(string(buf))
	}
}

const (
	PATH = "path"
	FILE = "file"
	SIZE = "size"
)
const CAT = "cat"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CAT: {Name: CAT, Help: "文件", Value: kit.Data(
			kit.SSH_SOURCE, kit.Dict(
				"sh", "true", "shy", "true", "py", "true",
				"go", "true", "vim", "true", "js", "true",
				"json", "true", "conf", "true", "yml", "true",
				"makefile", "true",
			),
		)},
	}, Commands: map[string]*ice.Command{
		CAT: {Name: "cat path auto", Help: "文件", Action: map[string]*ice.Action{
			mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_cat_show(m, path.Join(arg[2], arg[1]))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				m.Cmdy(DIR, arg)
			} else {
				_cat_show(m, arg[0])
			}
		}},
	}})
}
