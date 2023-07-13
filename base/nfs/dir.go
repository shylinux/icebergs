package nfs

import (
	"os"
	"path"
	"regexp"
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
			if s.ModTime().After(last) {
				last = s.ModTime()
			}
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
					} else if strings.Contains(p, "ice.windows.") {
						m.PushDownload(mdb.LINK, "ice.exe", p)
					} else {
						m.PushDownload(mdb.LINK, p)
					}
				case mdb.SHOW:
					switch p := kit.MergeURL("/share/local/"+p, ice.POD, m.Option(ice.MSG_USERPOD)); kit.Ext(s.Name()) {
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
					m.PushButton(mdb.SHOW, TRASH)
				default:
					m.Push(field, "")
				}
			}
		}
		if deep && isDir {
			switch s.Name() {
			case "pluged":
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
	PWD          = "./"
	SRC          = "src/"
	BIN          = "bin/"
	USR          = "usr/"
	USR_PORTAL   = ice.USR_PORTAL
	SRC_DOCUMENT = ice.SRC_DOCUMENT
	REQUIRE      = "/require/"

	TYPE_ALL  = "all"
	TYPE_BIN  = "bin"
	TYPE_CAT  = "cat"
	TYPE_DIR  = "dir"
	TYPE_BOTH = "both"

	DIR_ROOT = "dir_root"
	DIR_TYPE = "dir_type"
	DIR_DEEP = "dir_deep"
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
		DIR: {Name: "dir path auto upload finder", Help: "目录", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				aaa.White(m, ice.SRC, ice.BIN, ice.USR)
				aaa.Black(m, ice.USR_LOCAL)
			}}, mdb.UPLOAD: {},
			"finder": {Help: "本机", Hand: func(m *ice.Message, arg ...string) { m.Cmd("cli.system", "opens", "Finder.app") }},
			TRASH:    {Hand: func(m *ice.Message, arg ...string) { m.Cmd(TRASH, mdb.CREATE, m.Option(PATH)) }},
			mdb.SHOW: {Hand: func(m *ice.Message, arg ...string) {
				Show(m.ProcessInner(), path.Join(m.Option(DIR_ROOT), m.Option(PATH)))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			root, dir := kit.Select(PWD, m.Option(DIR_ROOT)), kit.Select(PWD, arg, 0)
			kit.If(strings.HasPrefix(dir, PS), func() { root = "" })
			if !aaa.Right(m, path.Join(root, dir)) {
				return
			}
			m.Logs(FIND, DIR_ROOT, root, PATH, dir, DIR_TYPE, m.Option(DIR_TYPE))
			fields := kit.Split(kit.Select(kit.Select(DIR_DEF_FIELDS, m.OptionFields()), kit.Join(kit.Slice(arg, 1))))
			size, last := _dir_list(m, root, dir, 0, m.Option(DIR_DEEP) == ice.TRUE, kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)), regexp.MustCompile(m.Option(DIR_REG)), fields)
			m.Status(mdb.TIME, last, mdb.COUNT, kit.Split(m.FormatSize())[0], SIZE, kit.FmtSize(size), kit.MDB_COST, m.FormatCost())
		}},
	})
}

var bind = []string{
	"usr/icebergs/core/chat/", "usr/volcanos/panel/",
	"usr/icebergs/core/", "usr/volcanos/plugin/local/",
}

func Relative(m *ice.Message, p string) string {
	for i := 0; i < len(bind); i += 2 {
		if strings.HasPrefix(p, bind[i]) {
			return strings.Replace(p, bind[i], bind[i+1], 1)
		}
	}
	return p
}
func SplitPath(m *ice.Message, p string) []string {
	if kit.HasPrefix(p, ice.REQUIRE_SRC, ice.REQUIRE_USR) {
		p = strings.TrimPrefix(p, REQUIRE)
	} else if kit.HasPrefix(p, REQUIRE) {
		ls := kit.Split(p, PS)
		return []string{ice.USR_REQUIRE + path.Join(ls[1:4]...) + PS, path.Join(ls[4:]...)}
	}
	line := kit.Select("1", strings.Split(p, DF), 1)
	p = strings.TrimPrefix(p, kit.Path("")+PS)
	p = strings.Split(p, DF)[0]
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
func Show(m *ice.Message, file string) bool {
	switch strings.ToLower(kit.Ext(file)) {
	case "png", "jpg":
		m.EchoImages("/share/local/" + file)
	case "mp4", "mov":
		m.EchoVideos("/share/local/" + file)
	default:
		return false
	}
	return true
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
