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

var rewriteList = []interface{}{}

func AddRewrite(cb interface{}) { rewriteList = append(rewriteList, cb) }

func _cat_right(m *ice.Message, name string) bool {
	return aaa.RoleRight(m, m.Option(ice.MSG_USERROLE), strings.Split(name, ice.PS)...)
}
func _cat_find(m *ice.Message, name string) io.ReadCloser {
	if m.Option(CAT_CONTENT) != "" {
		return NewReadCloser(bytes.NewBufferString(m.Option(CAT_CONTENT)))
	}

	// 模块回调
	for _, h := range rewriteList {
		switch h := h.(type) {
		case func(m *ice.Message, name string) io.ReadCloser:
			if r := h(m, name); r != nil {
				return r
			}
		case func(m *ice.Message, name string) []byte:
			if b := h(m, name); b != nil {
				return NewReadCloser(bytes.NewBuffer(b))
			}
		case func(m *ice.Message, name string) string:
			if s := h(m, name); s != "" {
				return NewReadCloser(bytes.NewBufferString(s))
			}
		}
	}

	// 本地文件
	if f, e := os.Open(path.Join(m.Option(DIR_ROOT), name)); e == nil {
		return f
	}
	return nil
}
func _cat_list(m *ice.Message, name string) {
	if m.Warn(!_cat_right(m, name), ice.ErrNotRight) {
		return // 没有权限
	}

	f := _cat_find(m, name)
	if m.Warn(f == nil, ice.ErrNotFound, name) {
		return // 没有文件
	}
	defer f.Close()

	switch cb := m.OptionCB(CAT).(type) {
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
	case func([]string, string):
		for bio := bufio.NewScanner(f); bio.Scan(); {
			cb(kit.Split(bio.Text()), bio.Text())
		}

	default:
		buf := make([]byte, ice.MOD_BUFS)
		for begin := 0; true; {
			if n, e := f.Read(buf[begin:]); !m.Warn(e, ice.ErrNotFound, name) {
				m.Log_IMPORT(FILE, name, SIZE, n)
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
	CAT_CONTENT = "cat_content"
	TEMPLATE    = "template"

	SOURCE = "source"
	BINARY = "binary"
	TARGET = "target"

	MASTER = "master"
	BRANCH = "branch"
	REPOS  = "repos"

	LOAD = "load"
	TAGS = "tags"
)
const (
	HTML = ice.HTML
	CSS  = ice.CSS
	JS   = ice.JS
	GO   = ice.GO
	SH   = ice.SH
	CSV  = ice.CSV
	JSON = ice.JSON
	TXT  = "txt"
	SHY  = "shy"
	SVG  = "svg"

	PWD = "./"
)

const (
	PATH = "path"
	FILE = "file"
	LINE = "line"
	SIZE = "size"
)
const CAT = "cat"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CAT: {Name: CAT, Help: "文件", Value: kit.Data(
			SOURCE, kit.Dict(
				HTML, ice.TRUE, CSS, ice.TRUE, JS, ice.TRUE, GO, ice.TRUE, SH, ice.TRUE, CSV, ice.TRUE, JSON, ice.TRUE,
				"md", ice.TRUE, "shy", ice.TRUE, "makefile", ice.TRUE, "license", ice.TRUE,
				"conf", ice.TRUE, "yaml", ice.TRUE, "yml", ice.TRUE,
				"py", ice.TRUE, "txt", ice.TRUE,
			),
		)},
	}, Commands: map[string]*ice.Command{
		CAT: {Name: "cat path auto", Help: "文件", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, m.CommandKey(), m.PrefixKey())
			}},
			mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_cat_list(m, path.Join(arg[2], arg[1]))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
				m.Cmdy(DIR, arg)
				return
			}
			if m.Option(DIR_ROOT) != "" {
				m.Info("dir_root: %v", m.Option(DIR_ROOT))
			}
			_cat_list(m, arg[0])
		}},
	}})
}
