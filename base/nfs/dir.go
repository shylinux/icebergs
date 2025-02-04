package nfs

import (
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _dir_size(m *ice.Message, p string) (n int) {
	Open(m, p+PS, func(ls []os.FileInfo) { n = len(ls) })
	return
}
func _dir_hash(m *ice.Message, p string) (h string) {
	list := []string{}
	Open(m, p+PS, func(s os.FileInfo) { list = append(list, kit.Format("%s%d%s", s.Name(), s.Size(), s.ModTime())) })
	kit.If(len(list) > 0, func() { h = kit.Hashs(list) })
	return ""
}
func _dir_list(m *ice.Message, root string, dir string, level int, deep bool, dir_type string, dir_reg *regexp.Regexp, fields []string) (total int64, last time.Time) {
	ls, _ := ReadDir(m, path.Join(root, dir))
	if len(ls) == 0 {
		if s, e := StatFile(m, path.Join(root, dir)); e == nil && !s.IsDir() {
			Open(m, path.Dir(path.Join(root, dir))+PS, func(s os.FileInfo) { kit.If(s.Name() == path.Base(dir), func() { ls = append(ls, s) }) })
			dir, deep = path.Dir(dir), false
		}
	}
	for _, s := range ls {
		if s.Name() == PT || s.Name() == ".." || strings.HasPrefix(s.Name(), PT) && dir_type != TYPE_ALL {
			continue
		}
		p, pp := path.Join(root, dir, s.Name()), path.Join(dir, s.Name())
		isDir := s.IsDir() || kit.IsDir(p) && deep == false
		isBin := s.Mode().String()[3] == 'x' || kit.Ext(s.Name()) == "exe"
		if !(dir_type == TYPE_BIN && (!isBin || isDir) || dir_type == TYPE_CAT && isDir || dir_type == TYPE_DIR && !isDir) && (dir_reg == nil || dir_reg.MatchString(s.Name())) {
			switch cb := m.OptionCB("").(type) {
			case func(os.FileInfo, string):
				cb(s, p)
				continue
			case func(string):
				cb(p)
				continue
			case nil:
			default:
				m.ErrorNotImplement(cb)
			}
			kit.If(s.ModTime().After(last), func() { last = s.ModTime() })
			for _, field := range fields {
				switch field {
				case mdb.TIME:
					m.Push(field, s.ModTime().Format(ice.MOD_TIME))
				case mdb.TYPE:
					m.Push(field, kit.Select(CAT, DIR, isDir))
				case TREE:
					if level == 0 {
						m.Push(field, s.Name())
					} else {
						m.Push(field, strings.Repeat("| ", level-1)+"|-"+s.Name())
					}
				case FULL:
					m.Push(field, p+kit.Select("", PS, isDir))
				case PATH:
					m.Push(field, pp+kit.Select("", PS, isDir))
				case FILE:
					m.Push(field, s.Name()+kit.Select("", PS, isDir))
				case NAME:
					m.Push(field, s.Name())
				case SIZE:
					if isDir {
						m.Push(field, _dir_size(m, p))
					} else {
						m.Push(field, kit.FmtSize(s.Size()))
						total += s.Size()
					}
				case LINE:
					if isDir {
						m.Push(field, _dir_size(m, p))
					} else {
						m.Push(field, _cat_line(m, p))
					}
				case mdb.HASH, "hashs":
					h := ""
					if isDir {
						h = _dir_hash(m, p)
					} else {
						h = _cat_hash(m, p)
					}
					m.Push(mdb.HASH, kit.Select(h[:6], h[:], field == mdb.HASH))
				case mdb.LINK:
					if isDir {
						m.Push(mdb.LINK, "")
					} else {
						if strings.Contains(p, "ice.windows") {
							m.PushDownload(mdb.LINK, "ice.exe", p)
						} else {
							m.PushDownload(mdb.LINK, p)
						}
					}
				case mdb.SHOW:
					switch p := m.MergeLink(SHARE_LOCAL+p, ice.POD, m.Option(ice.MSG_USERPOD)); kit.Ext(s.Name()) {
					case PNG, JPG:
						m.PushImages(field, p)
					case MP4:
						m.PushVideos(field, p)
					default:
						m.Push(field, "")
					}
				case mdb.ACTION:
					if m.IsCliUA() || m.Option(ice.MSG_USERROLE) == aaa.VOID {
						break
					}
					m.PushButton(mdb.SHOW, "rename", TRASH)
				default:
					m.Push(field, "")
				}
			}
		}
		if deep && isDir {
			switch s.Name() {
			case "pluged", "node_modules":
				continue
			}
			_total, _last := _dir_list(m, root, pp, level+1, deep, dir_type, dir_reg, fields)
			if total += _total; _last.After(last) {
				last = _last
			}
		}
	}
	return
}

const (
	PWD = "./"
	SRC = "src/"
	ETC = "etc/"
	BIN = "bin/"
	VAR = "var/"
	USR = "usr/"

	SCAN   = "scan"
	GOWORK = "gowork"

	PORTAL_GO           = "portal.go"
	PORTAL_JSON         = "portal.json"
	ETC_LOCAL_SH        = "etc/local.sh"
	ETC_CERT_KEY        = "etc/cert/cert.key"
	ETC_CERT_PEM        = "etc/cert/cert.pem"
	SRC_DOCUMENT        = "src/document/"
	SRC_PRIVATE         = "src/private/"
	SRC_TEMPLATE        = ice.SRC_TEMPLATE
	USR_TOOLKITS        = ice.USR_TOOLKITS
	USR_ICEBERGS        = ice.USR_ICEBERGS
	USR_RELEASE         = ice.USR_RELEASE
	USR_PUBLISH         = ice.USR_PUBLISH
	USR_LOCAL           = ice.USR_LOCAL
	USR_LOCAL_WORK      = ice.USR_LOCAL_WORK
	USR_IMAGE           = "usr/image/"
	USR_MATERIAL        = "usr/material/"
	USR_LOCAL_IMAGE     = "usr/local/image/"
	USR_LEARNING_PORTAL = "usr/learning/portal/"
	USR_MODULES         = "usr/node_modules/"
	USR_PACKAGE         = "usr/package.json"

	VAR_LOG_BENCH_LOG  = "var/log/bench.log"
	USR_ICONS_AVATAR   = "usr/icons/avatar.jpg"
	USR_ICONS_CONTEXTS = "usr/icons/contexts.jpg"
	USR_ICONS_ICEBERGS = "usr/icons/icebergs.png"
	USR_ICONS_VOLCANOS = "usr/icons/volcanos.png"
	USR_ICONS          = "usr/icons/"

	V               = "/v/"
	M               = "/m/"
	P               = "/p/"
	X               = "/x/"
	S               = "/s/"
	C               = "/c/"
	INTSHELL        = "/intshell/"
	VOLCANOS        = "/volcanos/"
	VOLCANOS_PLUGIN = "/volcanos/plugin/"
	REQUIRE_MODULES = "/require/modules/"
	REQUIRE_USR     = "/require/usr/"
	REQUIRE_SRC     = "/require/src/"
	REQUIRE         = "/require/"
	PLUGIN          = "/plugin/"
	SHARE_LOCAL     = "/share/local/"
	PATHNAME        = "pathname"
	FILENAME        = "filename"

	TYPE_ALL  = "all"
	TYPE_BIN  = "bin"
	TYPE_CAT  = "cat"
	TYPE_DIR  = "dir"
	TYPE_BOTH = "both"

	DIR_ROOT = "dir_root"
	DIR_DEEP = "dir_deep"
	DIR_TYPE = "dir_type"
	DIR_REG  = "dir_reg"

	DIR_DEF_FIELDS = "time,path,size,action"
	DIR_WEB_FIELDS = "time,path,size,link,action"
	DIR_CLI_FIELDS = "path,size,time"

	ROOT = "root"
	TREE = "tree"
	FULL = "full"
	PATH = "path"
	FILE = "file"
	NAME = "name"
	SIZE = "size"
	LINE = "line"
)
const DIR = "dir"

func init() {
	Index.MergeCommands(ice.Commands{
		DIR: {Name: "dir path auto upload app", Icon: "dir.png", Help: "文件夹", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				aaa.White(m, ice.MAKEFILE, ice.README_MD, ice.LICENSE)
				aaa.White(m, ice.SRC, ice.BIN, ice.USR)
				aaa.Black(m, ice.SRC_PRIVATE)
				aaa.Black(m, ice.USR_LOCAL)
			}},
			ice.APP: {Help: "本机", Hand: func(m *ice.Message, arg ...string) {
				switch runtime.GOOS {
				case "darwin":
					m.System("open", kit.Path(m.Option(PATH)))
				}
			}},
			mdb.SHOW: {Help: "预览", Hand: func(m *ice.Message, arg ...string) {
				Show(m.ProcessInner(), path.Join(m.Option(DIR_ROOT), m.Option(PATH)))
			}}, mdb.UPLOAD: {},
			SIZE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.Select("", kit.Split(m.System("du", "-sh").Result()), 0))
			}},
			"rename": {Name: "rename to", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(MOVE, path.Join(path.Dir(m.Option(PATH)), m.Option(TO)), m.Option(PATH))
			}},
			TRASH: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(TRASH, mdb.CREATE, m.Option(PATH))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			root, dir := kit.Select(PWD, m.Option(DIR_ROOT)), kit.Select(PWD, arg, 0)
			kit.If(strings.HasPrefix(dir, PS), func() { root = "" })
			if !aaa.Right(m, path.Join(root, dir)) {
				return
			}
			m.Logs(FIND, DIR_ROOT, root, PATH, dir, m.OptionSimple(DIR_TYPE, DIR_REG))
			fields := kit.Split(kit.Select(kit.Select(DIR_DEF_FIELDS, m.OptionFields()), kit.Join(kit.Slice(arg, 1))))
			size, last := _dir_list(m, root, dir, 0, m.Option(DIR_DEEP) == ice.TRUE, kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)), regexp.MustCompile(m.Option(DIR_REG)), fields)
			kit.If(m.Option(DIR_ROOT), func() { m.Option(DIR_ROOT, path.Join(m.Option(DIR_ROOT))+PS) })
			m.StatusTimeCount(mdb.TIME, last, SIZE, kit.FmtSize(size), m.OptionSimple(DIR_ROOT))
		}},
	})
}

func Relative(m *ice.Message, p string) string {
	if _p := kit.ExtChange(p, JS); Exists(m, _p) {
		return _p
	} else if _p := kit.ExtChange(path.Join(ice.USR_VOLCANOS, ice.PLUGIN_LOCAL, path.Join(kit.Slice(kit.Split(p, PS), -2)...)), JS); Exists(m, kit.Split(_p, "?")[0]) {
		return _p
	} else {
		return p
	}
}
func SplitPath(m *ice.Message, p string) []string {
	if kit.HasPrefix(p, REQUIRE_SRC, REQUIRE_USR) {
		p = strings.TrimPrefix(p, REQUIRE)
	} else if kit.HasPrefix(p, REQUIRE) {
		ls := kit.Split(p, PS)
		return []string{ice.USR_REQUIRE + path.Join(ls[1:4]...) + PS, path.Join(ls[4:]...)}
	} else if kit.HasPrefix(p, P) {
		p = strings.TrimPrefix(p, P)
	}
	line := kit.Select("1", strings.Split(p, DF), 1)
	p = strings.Split(p, DF)[0]
	p = strings.Split(p, "?")[0]
	if ls := kit.Split(kit.Select(ice.SRC_MAIN_GO, p), PS); len(ls) == 1 {
		return []string{PWD, ls[0], line}
	} else if ls[0] == ice.USR {
		return []string{strings.Join(ls[:2], PS) + PS, strings.Join(ls[2:], PS), line}
	} else {
		return []string{strings.Join(ls[:1], PS) + PS, strings.Join(ls[1:], PS), line}
	}
}
func Dir(m *ice.Message, field string) *ice.Message {
	m.Copy(m.Cmd(DIR, PWD, kit.Dict(DIR_TYPE, TYPE_DIR)).Sort(field))
	m.Copy(m.Cmd(DIR, PWD, kit.Dict(DIR_TYPE, TYPE_CAT)).Sort(field))
	return m
}
func DirDeepAll(m *ice.Message, root, dir string, cb func(ice.Maps), arg ...string) *ice.Message {
	m.Options(DIR_TYPE, CAT, DIR_ROOT, root, DIR_DEEP, ice.TRUE)
	defer m.Options(DIR_TYPE, "", DIR_ROOT, "", DIR_DEEP, "")
	if msg := m.Cmd(DIR, dir, arg); cb == nil {
		return m.Copy(msg)
	} else {
		return msg.Table(cb)
	}
}
func Show(m *ice.Message, file string) bool {
	p := SHARE_LOCAL + file
	kit.If(m.Option(ice.MSG_USERPOD), func(pod string) { p = kit.MergeURL(p, ice.POD, pod) })
	switch strings.ToLower(kit.Ext(file)) {
	case PNG, JPG, JPEG, "gif":
		m.EchoImages(p)
	case MP4, MOV:
		m.EchoVideos(p)
	default:
		if IsSourceFile(m, kit.Ext(file)) {
			m.Cmdy(CAT, file)
		} else {
			m.ProcessOpen(p)
			return false
		}
	}
	return true
}
