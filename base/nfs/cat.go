package nfs

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _cat_find(m *ice.Message, p string) (io.ReadCloser, error) {
	if m.Option(CAT_CONTENT) != "" {
		return NewReadCloser(bytes.NewBufferString(m.Option(CAT_CONTENT))), nil
	}
	return OpenFile(m, path.Join(m.Option(DIR_ROOT), p))
}
func _cat_hash(m *ice.Message, p string) (h string) {
	Open(m, p, func(r io.Reader) {
		h = kit.Hashs(r)
		m.Debug("what %v %v", p, h)
	})
	return
}
func _cat_line(m *ice.Message, p string) (n int) {
	Open(m, p, func(r io.Reader) { kit.For(r, func(s string) { n++ }) })
	return
}
func _cat_list(m *ice.Message, p string) {
	if m.Option(CAT_CONTENT) == "" && !aaa.Right(m, p) {
		return
	}
	f, e := _cat_find(m, p)
	if m.Warn(e, ice.ErrNotFound, p) {
		return
	}
	defer f.Close()
	switch cb := m.OptionCB("").(type) {
	case func(string, int) string:
		list := []string{}
		kit.For(f, func(s string, i int) { list = append(list, cb(s, i)) })
		m.Echo(strings.Join(list, ice.NL) + ice.NL)
	case func(string, int):
		kit.For(f, cb)
	case func(string):
		kit.For(f, cb)
	case func([]string, string):
		kit.For(f, cb)
	case nil:
		if b, e := ioutil.ReadAll(f); !m.Warn(e) {
			m.Echo(string(b)).StatusTime(FILE, p, SIZE, len(b))
		}
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	CAT_CONTENT = "cat_content"
	CONFIGURE   = "configure"
	TEMPLATE    = "template"
	STDIO       = "stdio"

	TAGS   = "tags"
	MODULE = "module"
	SOURCE = "source"
	TARGET = "target"
	BINARY = "binary"
	SCRIPT = "script"

	REPOS   = "repos"
	REMOTE  = "remote"
	ORIGIN  = "origin"
	BRANCH  = "branch"
	MASTER  = "master"
	VERSION = "version"
)
const (
	HTML = ice.HTML
	CSS  = ice.CSS
	SVG  = ice.SVG
	JS   = ice.JS
	GO   = ice.GO
	SH   = ice.SH
	SHY  = ice.SHY
	CSV  = ice.CSV
	JSON = ice.JSON

	PROTO = "proto"
	YAML  = "yaml"
	CONF  = "conf"
	XML   = "xml"
	YML   = "yml"
	TXT   = "txt"
	MD    = "md"
	PY    = "py"

	PNG = "png"
	JPG = "jpg"
	MP4 = "mp4"
	PDF = "pdf"

	DF = ice.DF
	PS = ice.PS
	PT = ice.PT
)

const CAT = "cat"

func init() {
	Index.MergeCommands(ice.Commands{
		CAT: {Name: "cat path auto", Help: "文件", Actions: ice.MergeActions(ice.Actions{ice.CTX_INIT: mdb.AutoConfig(SOURCE, kit.DictList(
			HTML, CSS, JS, GO, SH, PY, SHY, CSV, JSON, CONFIGURE, PROTO, YAML, CONF, XML, YML, TXT, MD, strings.ToLower(ice.LICENSE), strings.ToLower(ice.MAKEFILE),
		))}), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], PS) {
				m.Cmdy(DIR, arg)
			} else {
				_cat_list(m.Logs(FIND, m.OptionSimple(DIR_ROOT), FILE, arg[0]), arg[0])
			}
		}},
	})
}

type templateMessage interface {
	PrefixKey(...string) string
	Cmdx(...ice.Any) string
}

func Template(m templateMessage, p string, arg ...ice.Any) string {
	return kit.Renders(kit.Format(TemplateText(m, p), arg...), m)
}
func TemplateText(m templateMessage, p string) string {
	return m.Cmdx(CAT, path.Join(m.PrefixKey(), path.Base(p)), kit.Dict(DIR_ROOT, ice.SRC_TEMPLATE))
}
func IsSourceFile(m *ice.Message, ext string) bool {
	return mdb.Conf(m, Prefix(CAT), kit.Keym(SOURCE, ext)) == ice.TRUE
}
func OptionLoad(m *ice.Message, p string) *ice.Message {
	Open(m, p, func(r io.Reader) {
		var data ice.Any
		m.Assert(json.NewDecoder(r).Decode(&data))
		kit.For(data, func(k string, v ice.Any) { m.Optionv(k, v) })
	})
	return m
}
func Open(m *ice.Message, p string, cb ice.Any) {
	if p == "" {
		return
	} else if strings.HasSuffix(p, PS) {
		if p == PS {
			p = ""
		}
		if ls, e := ReadDir(m, p); !m.Warn(e) {
			switch cb := cb.(type) {
			case func([]os.FileInfo):
				cb(ls)
			case func(os.FileInfo):
				kit.For(ls, cb)
			case func(io.Reader, string):
				kit.For(ls, func(s os.FileInfo) {
					kit.If(!s.IsDir(), func() { Open(m, path.Join(p, s.Name()), cb) })
				})
			default:
				m.ErrorNotImplement(cb)
			}
		}
	} else if f, e := OpenFile(m, p); !m.Warn(e, ice.ErrNotFound, p) {
		defer f.Close()
		switch cb := cb.(type) {
		case func(io.Reader, os.FileInfo):
			s, _ := StatFile(m, p)
			cb(f, s)
		case func(io.Reader, string):
			cb(f, p)
		case func(io.Reader):
			cb(f)
		case func(string):
			if b, e := ioutil.ReadAll(f); !m.Warn(e) {
				cb(string(b))
			}
		default:
			m.ErrorNotImplement(cb)
		}
	}
}
func ReadAll(m *ice.Message, r io.Reader) []byte {
	if b, e := ioutil.ReadAll(r); !m.Warn(e) {
		return b
	}
	return nil
}
func ReadFile(m *ice.Message, p string) (b []byte, e error) {
	Open(m, p, func(r io.Reader) { b, e = ioutil.ReadAll(r) })
	return
}
func Rewrite(m *ice.Message, p string, cb func(string) string) {
	m.Cmd(SAVE, p, m.Cmdx(CAT, p, func(s string, i int) string { return cb(s) }))
}
