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

func _dir_size(m *ice.Message, p string) int {
	if ls, e := ReadDir(m, p); !m.Warn(e) {
		return len(ls)
	}
	return 0
}
func _dir_hash(m *ice.Message, p string) string {
	if ls, e := ReadDir(m, p); !m.Warn(e) {
		meta := []string{}
		for _, s := range ls {
			meta = append(meta, kit.Format("%s%d%s", s.Name(), s.Size(), s.ModTime()))
		}
		return kit.Hashs(meta)
	}
	return ""
}
func _dir_list(m *ice.Message, root string, name string, level int, deep bool, dir_type string, dir_reg *regexp.Regexp, fields []string) *ice.Message {
	if !aaa.Right(m, path.Join(root, name)) {
		return m // 没有权限
	}

	list, e := ReadDir(m, path.Join(root, name))
	if e != nil { // 单个文件
		ls, _ := ReadDir(m, path.Dir(path.Join(root, name)))
		for _, s := range ls {
			if s.Name() == path.Base(name) {
				list = append(list, s)
			}
		}
		name = path.Dir(name)
	}

	for _, f := range list {
		if f.Name() == ice.PT || f.Name() == ".." {
			continue
		}
		if strings.HasPrefix(f.Name(), ice.PT) && dir_type != TYPE_ALL {
			continue
		}

		p, pp := path.Join(root, name, f.Name()), path.Join(name, f.Name())
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
					m.Push(field, pp+kit.Select("", ice.PS, isDir))
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
					m.PushDownload(mdb.LINK, kit.Select("", f.Name(), !isDir), p)
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
					if m.IsCliUA() || m.Option(ice.MSG_USERROLE) == "void" {
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
				continue // 禁用递归
			}

			_dir_list(m, root, pp, level+1, deep, dir_type, dir_reg, fields)
		}
	}
	return m
}

const (
	TYPE_ALL  = "all"
	TYPE_CAT  = "cat"
	TYPE_DIR  = "dir"
	TYPE_BOTH = "both"
)
const (
	DIR_ROOT = "dir_root"
	DIR_TYPE = "dir_type"
	DIR_DEEP = "dir_deep"
	DIR_REG  = "dir_reg"

	DIR_DEF_FIELDS = "time,path,size,action"
	DIR_WEB_FIELDS = "time,size,path,action,link"
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
)
const DIR = "dir"

func init() {
	Index.MergeCommands(ice.Commands{
		DIR: {Name: "dir path field auto upload", Help: "目录", Actions: ice.Actions{
			mdb.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("web.cache", "upload_watch", m.Option(PATH))
			}},
			TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TRASH, m.Option(PATH))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(DIR_ROOT) != "" {
				m.Logs(mdb.SELECT, DIR_ROOT, m.Option(DIR_ROOT))
			}
			_dir_list(m, kit.Select(PWD, m.Option(DIR_ROOT)), kit.Select(PWD, arg, 0),
				0, m.Option(DIR_DEEP) == ice.TRUE, kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)), kit.Regexp(m.Option(DIR_REG)),
				kit.Split(kit.Select(kit.Select(DIR_DEF_FIELDS, m.OptionFields()), kit.Join(kit.Slice(arg, 1)))))
			m.SortTimeR(mdb.TIME)
			m.StatusTimeCount()
		}},
	})
}

func Dir(m *ice.Message, sort string) *ice.Message {
	m.Option(DIR_TYPE, TYPE_DIR)
	m.Copy(m.Cmd(DIR, PWD).Sort(sort))
	m.Option(DIR_TYPE, TYPE_CAT)
	m.Copy(m.Cmd(DIR, PWD).Sort(sort))
	return m
}
