package nfs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
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
	if m.Option(CAT_CONTENT) == "" && !aaa.Right(m, name) {
		return // 没有权限
	}
	f, e := _cat_find(m, name)
	if m.Warn(e, ice.ErrNotFound) {
		return // 没有文件
	}
	defer f.Close()

	switch cb := m.OptionCB("").(type) {
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
				m.Logs(mdb.IMPORT, FILE, name, SIZE, n)
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

	STDIO  = "stdio"
	SOURCE = "source"
	MODULE = "module"
	SCRIPT = "script"
	BINARY = "binary"
	TARGET = "target"
	TAGS   = "tags"

	TEMPLATE = "template"
	MASTER   = "master"
	BRANCH   = "branch"
	REPOS    = "repos"
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
	MD  = "md"
	TXT = "txt"
	IML = "iml"
	XML = "xml"
	YML = "yml"
	ZML = "zml"

	PNG = "png"
	JPG = "jpg"
	MP4 = "mp4"
	PDF = "pdf"

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
				SHY, ice.TRUE, "conf", ice.TRUE, "makefile", ice.TRUE, "license", ice.TRUE,
				PY, ice.TRUE, MD, ice.TRUE, TXT, ice.TRUE, IML, ice.TRUE, XML, ice.TRUE, YML, ice.TRUE, ZML, ice.TRUE,
				"configure", ice.TRUE,
			),
		)},
	}, Commands: ice.Commands{
		CAT: {Name: "cat path auto", Help: "文件", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, ice.SRC_MAIN_SHY)
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, ice.SRC_MAIN_GO)
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, ice.USR_PUBLISH)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
				m.Cmdy(DIR, arg)
				return
			}
			if m.Option(DIR_ROOT) != "" {
				m.Logs(mdb.SELECT, DIR_ROOT, m.Option(DIR_ROOT))
			}
			_cat_list(m, arg[0])
		}},
	}})
}
func IsSourceFile(m *ice.Message, ext string) bool {
	return m.Conf(CAT, kit.Keym(SOURCE, ext)) == ice.TRUE
}
func OptionLoad(m *ice.Message, file string) *ice.Message {
	if f, e := OpenFile(m, file); e == nil {
		defer f.Close()

		var data ice.Any
		m.Assert(json.NewDecoder(f).Decode(&data))

		kit.Fetch(data, func(key string, value ice.Any) {
			m.Option(key, kit.Simple(value))
		})
	}
	return m
}
