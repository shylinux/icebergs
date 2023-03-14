package nfs

import (
	"os"
	"path"
	"regexp"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _dir_size(m *ice.Message, dir string) int {
	if ls, e := ReadDir(m, dir); !m.Warn(e) {
		return len(ls)
	}
	return 0
}
func _dir_hash(m *ice.Message, dir string) string {
	if ls, e := ReadDir(m, dir); !m.Warn(e) {
		meta := []string{}
		for _, s := range ls {
			meta = append(meta, kit.Format("%s%d%s", s.Name(), s.Size(), s.ModTime()))
		}
		return kit.Hashs(meta)
	}
	return ""
}
func _dir_list(m *ice.Message, root string, dir string, level int, deep bool, dir_type string, dir_reg *regexp.Regexp, fields []string) *ice.Message {
	ls, _ := ReadDir(m, path.Join(root, dir))
	if len(ls) == 0 {
		if s, e := StatFile(m, path.Join(root, dir)); e == nil && !s.IsDir() {
			_ls, _ := ReadDir(m, path.Dir(path.Join(root, dir)))
			for _, s := range _ls {
				if s.Name() == path.Base(dir) {
					ls = append(ls, s)
				}
			}
			dir, deep = path.Dir(dir), false
		}
	}
	for _, f := range ls {
		if f.Name() == ice.PT || f.Name() == ".." {
			continue
		}
		if strings.HasPrefix(f.Name(), ice.PT) && dir_type != TYPE_ALL {
			continue
		}
		p, _dir := path.Join(root, dir, f.Name()), path.Join(dir, f.Name())
		isDir := f.IsDir() || kit.IsDir(p)
		if !(dir_type == TYPE_CAT && isDir || dir_type == TYPE_DIR && !isDir) && (dir_reg == nil || dir_reg.MatchString(f.Name())) {
			switch cb := m.OptionCB("").(type) {
			case func(f os.FileInfo, p string):
				cb(f, p)
				continue
			case func(p string):
				cb(p)
				continue
			case nil:
			default:
				m.ErrorNotImplement(cb)
			}
			for _, field := range fields {
				switch field {
				case mdb.TIME:
					m.Push(field, f.ModTime().Format(ice.MOD_TIME))
				case mdb.TYPE:
					m.Push(field, kit.Select(CAT, DIR, isDir))
				case TREE:
					if level == 0 {
						m.Push(field, f.Name())
					} else {
						m.Push(field, strings.Repeat("| ", level-1)+"|-"+f.Name())
					}
				case FULL:
					m.Push(field, p+kit.Select("", ice.PS, isDir))
				case PATH:
					m.Push(field, _dir+kit.Select("", ice.PS, isDir))
				case FILE:
					m.Push(field, f.Name()+kit.Select("", ice.PS, isDir))
				case NAME:
					m.Push(field, f.Name())
				case SIZE:
					if isDir {
						m.Push(field, _dir_size(m, p))
					} else {
						m.Push(field, kit.FmtSize(f.Size()))
					}
				case LINE:
					if isDir {
						m.Push(field, _dir_size(m, p))
					} else {
						m.Push(field, _cat_size(m, p))
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
					if strings.Contains(p, "ice.windows.") {
						m.PushDownload(mdb.LINK, "ice.exe", p)
					} else {
						m.PushDownload(mdb.LINK, kit.Select("", f.Name(), !isDir), p)
					}
				case mdb.SHOW:
					switch p := kit.MergeURL("/share/local/"+p, ice.POD, m.Option(ice.MSG_USERPOD)); kit.Ext(f.Name()) {
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
					m.PushButton(TRASH)
				default:
					m.Push(field, "")
				}
			}
		}
		if deep && isDir {
			switch f.Name() {
			case "node_modules", "pluged", "target", "trash":
				continue
			}
			_dir_list(m, root, _dir, level+1, deep, dir_type, dir_reg, fields)
		}
	}
	return m
}

const (
	TYPE_ALL  = "all"
	TYPE_CAT  = "cat"
	TYPE_DIR  = "dir"
	TYPE_BOTH = "both"

	SRC = "src/"
	USR = "usr/"
)
const (
	DIR_ROOT = "dir_root"
	DIR_TYPE = "dir_type"
	DIR_DEEP = "dir_deep"
	DIR_REG  = "dir_reg"

	DIR_DEF_FIELDS = "time,size,path,action"
	DIR_WEB_FIELDS = "time,size,path,link,action"
	DIR_CLI_FIELDS = "path,size,time"
)
const (
	ROOT = "root"
	TREE = "tree"
	FULL = "full"
	PATH = "path"
	FILE = "file"
	NAME = "name"
	SIZE = "size"
	LINE = "line"

	OPENS = "opens"
)
const DIR = "dir"

func init() {
	Index.MergeCommands(ice.Commands{
		DIR: {Name: "dir path field auto upload", Help: "目录", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				aaa.White(m, ice.SRC, ice.BIN, ice.USR)
				aaa.Black(m, ice.USR_LOCAL)
			}}, mdb.UPLOAD: {},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" && m.Cmdx("host", "islocal", m.Option(ice.MSG_USERIP)) == ice.OK {
					for _, p := range []string{"Desktop", "Documents", "Downloads", "Pictures"} {
						p := kit.HomePath(p)
						m.Cmd(DIR, PWD, mdb.NAME, mdb.TIME, kit.Dict(DIR_ROOT, p)).SortTimeR(mdb.TIME).TablesLimit(5, func(value ice.Maps) {
							name := value[mdb.NAME]
							if len(kit.TrimExt(name)) > 30 {
								name = name[:10] + ".." + name[len(name)-10:]
							}
							m.PushSearch(mdb.TYPE, OPENS, mdb.NAME, name, mdb.TEXT, path.Join(p, value[mdb.NAME]))
						})
					}
				}
			}},
			TRASH: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(TRASH, mdb.CREATE, m.Option(PATH)) }},
		}, Hand: func(m *ice.Message, arg ...string) {
			root, dir := kit.Select(PWD, m.Option(DIR_ROOT)), kit.Select(PWD, arg, 0)
			if strings.HasPrefix(dir, ice.PS) {
				root, dir = ice.PS, strings.TrimPrefix(dir, ice.PS)
			}
			if root == ice.PS && dir == ice.PS {
				root, dir = PWD, PWD
			}
			if !aaa.Right(m, path.Join(root, dir)) {
				return
			}
			fields := kit.Split(kit.Select(kit.Select(DIR_DEF_FIELDS, m.OptionFields()), kit.Join(kit.Slice(arg, 1))))
			if root != "" {
				m.Logs(mdb.SELECT, DIR_ROOT, root, dir)
			}
			_dir_list(m, root, dir, 0, m.Option(DIR_DEEP) == ice.TRUE, kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)), kit.Regexp(m.Option(DIR_REG)), fields)
			m.StatusTimeCount()
		}},
	})
}

func SplitPath(m *ice.Message, p string) []string {
	if ls := kit.Split(p, ice.PS); len(ls) == 1 {
		return []string{PWD, ls[0]}
	} else if ls[0] == ice.USR {
		return []string{strings.Join(ls[:2], ice.PS) + ice.PS, strings.Join(ls[2:], ice.PS)}
	} else {
		return []string{strings.Join(ls[:1], ice.PS) + ice.PS, strings.Join(ls[1:], ice.PS)}
	}
}
func Dir(m *ice.Message, sort string) *ice.Message {
	m.Copy(m.Cmd(DIR, PWD, kit.Dict(DIR_TYPE, TYPE_DIR)).Sort(sort))
	m.Copy(m.Cmd(DIR, PWD, kit.Dict(DIR_TYPE, TYPE_CAT)).Sort(sort))
	return m
}
func DirDeepAll(m *ice.Message, root, dir string, cb func(ice.Maps), arg ...string) *ice.Message {
	m.Options(DIR_TYPE, CAT, DIR_ROOT, root, DIR_DEEP, ice.TRUE)
	defer m.Options(DIR_TYPE, "", DIR_ROOT, "", DIR_DEEP, "")
	if msg := m.Cmd(DIR, dir, arg).Tables(cb); cb == nil {
		return m.Copy(msg)
	} else {
		return msg
	}
}
