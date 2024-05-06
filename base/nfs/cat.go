package nfs

import (
	"bytes"
	"encoding/csv"
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
	Open(m, p, func(r io.Reader) { h = kit.Hashs(r) })
	return
}
func _cat_line(m *ice.Message, p string) (n int) {
	Open(m, p, func(r io.Reader) { kit.For(r, func(s string) { n++ }) })
	return
}
func _cat_list(m *ice.Message, p string) {
	if m.Option(CAT_CONTENT) == "" && !aaa.Right(m, path.Join(m.Option(DIR_ROOT), p)) {
		return
	}
	f, e := _cat_find(m, p)
	if m.WarnNotFound(e, FILE, p) {
		return
	}
	defer f.Close()
	switch cb := m.OptionCB("").(type) {
	case func(string, int) string:
		list := []string{}
		kit.For(f, func(s string, i int) { list = append(list, cb(s, i)) })
		m.Echo(strings.Join(list, ice.NL) + ice.NL)
	case func([]string, string) string:
		list := []string{}
		kit.For(f, func(s string, i int) { list = append(list, cb(kit.Split(s), s)) })
		m.Echo(strings.Join(list, ice.NL) + ice.NL)
	case func(string, int):
		kit.For(f, cb)
	case func(string):
		kit.For(f, cb)
	case func([]string, string):
		kit.For(f, cb)
	case func([]string):
		kit.For(f, cb)
	case nil:
		if b, e := ioutil.ReadAll(f); !m.WarnNotFound(e) {
			m.Echo(string(b)).StatusTime(FILE, p, SIZE, len(b))
		}
	default:
		m.ErrorNotImplement(cb)
	}
}

const (
	CAT_CONTENT = "cat_content"
	CONFIGURE   = "configure"
	STDIO       = "stdio"

	TAGS   = "tags"
	MODULE = "module"
	SOURCE = "source"
	TARGET = "target"
	BINARY = "binary"
	SCRIPT = "script"
	FORMAT = "format"
	TRANS  = "trans"

	CLONE   = "clone"
	REPOS   = "repos"
	REMOTE  = "remote"
	ORIGIN  = "origin"
	COMMIT  = "commit"
	BRANCH  = "branch"
	MASTER  = "master"
	VERSION = "version"
	COMPILE = "compile"
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
	MOD  = "mod"

	PROTO = "proto"
	YAML  = "yaml"
	CONF  = "conf"
	XML   = "xml"
	YML   = "yml"
	TXT   = "txt"
	MD    = "md"
	PY    = "py"

	IMAGE = "image"
	JPEG  = "jpeg"
	JPG   = "jpg"
	PNG   = "png"
	MP4   = "mp4"
	MOV   = "mov"
	PDF   = "pdf"

	DF = ice.DF
	PS = ice.PS
	PT = ice.PT
)

const CAT = "cat"

func init() {
	Index.MergeCommands(ice.Commands{
		CAT: {Name: "cat path auto", Help: "文件", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: mdb.AutoConfig(SOURCE, kit.DictList(
				HTML, CSS, JS, GO, SH, PY, SHY, CSV, JSON, CONFIGURE, PROTO, YAML, CONF, XML, YML, TXT, MD, strings.ToLower(ice.LICENSE), strings.ToLower(ice.MAKEFILE),
			)),
		}, DIR), Hand: func(m *ice.Message, arg ...string) {
			if !DirList(m, arg...) {
				if arg[0] != "" {
					m.Logs(FIND, m.OptionSimple(DIR_ROOT), FILE, arg[0])
				}
				_cat_list(m, arg[0])
			}
		}},
	})
}

func DirList(m *ice.Message, arg ...string) bool {
	if len(arg) == 0 || strings.HasSuffix(arg[0], PS) {
		m.Cmdy(DIR, kit.Slice(arg, 0, 1))
		return true
	} else {
		return false
	}
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
		kit.If(p == PS, func() { p = "" })
		if ls, e := ReadDir(m, p); !m.WarnNotFound(e) {
			switch cb := cb.(type) {
			case func([]os.FileInfo):
				cb(ls)
			case func(os.FileInfo):
				kit.For(ls, cb)
			case func(io.Reader, string):
				kit.For(ls, func(s os.FileInfo) { kit.If(!s.IsDir(), func() { Open(m, path.Join(p, s.Name()), cb) }) })
			default:
				m.ErrorNotImplement(cb)
			}
		}
	} else if f, e := OpenFile(m, p); !m.WarnNotFound(e, p) {
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
			if b, e := ioutil.ReadAll(f); !m.WarnNotFound(e) {
				cb(string(b))
			}
		default:
			m.ErrorNotImplement(cb)
		}
	}
}
func ReadAll(m *ice.Message, r io.Reader) []byte {
	if b, e := ioutil.ReadAll(r); !m.WarnNotFound(e) {
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
func ScanCSV(m *ice.Message, file string, cb func([]string), arg ...string) {
	f, e := OpenFile(m, file)
	if m.Warn(e) {
		return
	}
	r := csv.NewReader(f)
	head, err := r.Read()
	if err != nil {
		return
	}
	index := []int{}
	kit.If(len(arg) == 0, func() { arg = append(arg, head...) })
	kit.For(arg, func(h string) { index = append(index, kit.IndexOf(head, h)) })
	for {
		data, err := r.Read()
		if err != nil {
			break
		}
		res := []string{}
		kit.For(index, func(i int) { res = append(res, data[i]) })
		cb(res)
	}
}
