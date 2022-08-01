package nfs

import (
	"bufio"
	"bytes"
	"io"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _cat_find(m *ice.Message, name string) (io.ReadCloser, error) {
	if m.Option(CAT_CONTENT) != "" {
		return NewReadCloser(bytes.NewBufferString(m.Option(CAT_CONTENT))), nil
	}
	return OpenFile(m, path.Join(m.Option(DIR_ROOT), name))
}
func _cat_size(m *ice.Message, p string) (nline int) {
	if f, e := OpenFile(m, p); m.Warn(e) {
		defer f.Close()
		for bio := bufio.NewScanner(f); bio.Scan(); nline++ {
			bio.Text()
		}
	}
	return nline
}
func _cat_hash(m *ice.Message, p string) string {
	if f, e := OpenFile(m, p); !m.Warn(e) {
		defer f.Close()
		return kit.Hashs(f)
	}
	return ""
}
func _cat_list(m *ice.Message, name string) {
	if !m.Right(m, name) {
		return // 没有权限
	}
	f, e := _cat_find(m, name)
	if m.Warn(e, ice.ErrNotFound, name) {
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
	case nil:
		buf, begin := make([]byte, ice.MOD_BUFS), 0
		for {
			if n, e := f.Read(buf[begin:]); !m.Warn(e, ice.ErrNotValid, name) {
				m.Log_IMPORT(FILE, name, SIZE, n)
				if begin += n; begin < len(buf) {
					buf = buf[:begin]
					break
				}
				buf = append(buf, make([]byte, ice.MOD_BUFS)...)
			} else {
				break
			}
		}
		m.Echo(string(buf))
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	CAT_CONTENT = "cat_content"
	TEMPLATE    = "template"
	WEBSITE     = "website"

	STDIO  = "stdio"
	SOURCE = "source"
	SCRIPT = "script"
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
	SHY  = ice.SHY
	SVG  = ice.SVG
	CSV  = ice.CSV
	JSON = ice.JSON

	PY  = "py"
	TXT = "txt"
	IML = "iml"
	XML = "xml"
	YML = "yml"
	ZML = "zml"

	PNG = "png"
	JPG = "jpg"
	MP4 = "mp4"

	PWD = "./"
	PS  = ice.PS
	PT  = ice.PT
)

const CAT = "cat"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		CAT: {Name: CAT, Help: "文件", Value: kit.Data(
			SOURCE, kit.Dict(
				HTML, ice.TRUE, CSS, ice.TRUE, JS, ice.TRUE, GO, ice.TRUE, SH, ice.TRUE, CSV, ice.TRUE, JSON, ice.TRUE,
				"md", ice.TRUE, "shy", ice.TRUE, "makefile", ice.TRUE, "license", ice.TRUE,
				"conf", ice.TRUE, YML, ice.TRUE, ZML, ice.TRUE, IML, ice.TRUE, "txt", ice.TRUE,
				"py", ice.TRUE,
			),
		)},
	}, Commands: ice.Commands{
		CAT: {Name: "cat path auto", Help: "文件", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
				m.Cmdy(DIR, arg)
				return
			}
			if m.Option(DIR_ROOT) != "" {
				m.Log_SELECT(DIR_ROOT, m.Option(DIR_ROOT))
			}
			_cat_list(m, arg[0])
		}},
	}})
}
