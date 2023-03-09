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

func _cat_find(m *ice.Message, file string) (io.ReadCloser, error) {
	if m.Option(CAT_CONTENT) != "" {
		return NewReadCloser(bytes.NewBufferString(m.Option(CAT_CONTENT))), nil
	}
	return OpenFile(m, path.Join(m.Option(DIR_ROOT), file))
}
func _cat_size(m *ice.Message, file string) (nline int) {
	if f, e := OpenFile(m, file); !m.Warn(e) {
		defer f.Close()
		for bio := bufio.NewScanner(f); bio.Scan(); nline++ {
			bio.Text()
		}
	}
	return nline
}
func _cat_hash(m *ice.Message, file string) string {
	if f, e := OpenFile(m, file); !m.Warn(e) {
		defer f.Close()
		return kit.Hashs(f)
	}
	return ""
}
func _cat_list(m *ice.Message, file string) {
	if m.Option(CAT_CONTENT) == "" && !aaa.Right(m, file) {
		return
	}
	f, e := _cat_find(m, file)
	if m.Warn(e, ice.ErrNotFound, file) {
		return
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
		buf, size := make([]byte, 10*ice.MOD_BUFS), 0
		for {
			if n, e := f.Read(buf[size:]); !m.Warn(e, ice.ErrNotValid, file) {
				m.Logs(LOAD, FILE, file, SIZE, n)
				if size += n; size < len(buf) {
					buf = buf[:size]
					break
				}
				buf = append(buf, make([]byte, ice.MOD_BUFS)...)
			} else {
				break
			}
		}
		m.Echo(string(buf)).StatusTime(FILE, file, SIZE, size)
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	CAT_CONTENT = "cat_content"

	STDIO  = "stdio"
	MODULE = "module"
	SOURCE = "source"
	SCRIPT = "script"
	BINARY = "binary"
	TARGET = "target"
	TAGS   = "tags"

	TEMPLATE = "template"
	VERSION  = "version"
	MASTER   = "master"
	BRANCH   = "branch"
	REMOTE   = "remote"
	ORIGIN   = "origin"
	REPOS    = "repos"
)
const (
	SVG  = ice.SVG
	HTML = ice.HTML
	CSS  = ice.CSS
	JS   = ice.JS
	GO   = ice.GO
	SH   = ice.SH
	SHY  = ice.SHY
	CSV  = ice.CSV
	JSON = ice.JSON

	PY  = "py"
	MD  = "md"
	TXT = "txt"
	XML = "xml"
	YML = "yml"
	ZML = "zml"
	IML = "iml"

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
	Index.MergeCommands(ice.Commands{
		CAT: {Name: "cat path auto", Help: "文件", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, ice.SRC_MAIN_JS, ice.SRC_MAIN_GO, ice.SRC_MAIN_SHY) }},
		}, ice.Actions{ice.CTX_INIT: mdb.AutoConfig(SOURCE, kit.DictList(
			HTML, CSS, JS, GO, SH, SHY, CSV, JSON, PY, MD, TXT, XML, YML, ZML, IML,
			"license", "makefile", "configure", "conf",
		))}), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
				m.Cmdy(DIR, arg)
				return
			}
			if m.Option(DIR_ROOT) != "" {
				m.Logs(mdb.SELECT, m.OptionSimple(DIR_ROOT))
			}
			_cat_list(m, arg[0])
		}},
	})
}
func IsSourceFile(m *ice.Message, ext string) bool {
	return m.Conf(CAT, kit.Keym(SOURCE, ext)) == ice.TRUE
}
func OptionLoad(m *ice.Message, file string) *ice.Message {
	if f, e := OpenFile(m, file); e == nil {
		defer f.Close()
		var data ice.Any
		m.Assert(json.NewDecoder(f).Decode(&data))
		kit.Fetch(data, func(key string, value ice.Any) { m.Option(key, kit.Simple(value)) })
	}
	return m
}

func Template(m *ice.Message, file string, arg ...ice.Any) string {
	return kit.Renders(kit.Format(TemplateText(m, file), arg...), m)
}
func TemplateText(m *ice.Message, file string) string {
	return m.Cmdx(CAT, path.Join(ice.SRC_TEMPLATE, m.PrefixKey(), path.Base(file)))
}
