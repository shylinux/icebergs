package nfs

import (
	"bufio"
	"crypto/sha1"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _dir_list(m *ice.Message, root string, name string, level int, deep bool, dir_type string, dir_reg *regexp.Regexp, fields []string) *ice.Message {
	if !_cat_right(m, path.Join(root, name)) {
		return m // 没有权限
	}

	if len(ice.Info.Pack) > 0 && m.Option(DIR_PACK) == ice.TRUE {
		for k, b := range ice.Info.Pack {
			p := strings.TrimPrefix(k, root)
			if !strings.HasPrefix(p, name) {
				if p = strings.TrimPrefix(k, root+ice.PS); !strings.HasPrefix(p, name) {
					if p = strings.TrimPrefix(k, ice.PS); !strings.HasPrefix(p, name) {
						continue
					}
				}
			}

			m.Debug("dir binpack %s", p)
			for _, field := range fields {
				switch field {
				case PATH:
					m.Push(field, p)
				case SIZE:
					m.Push(field, len(b))
				default:
					m.Push(field, "")
				}
			}
		}
		return m
	}

	list, e := ioutil.ReadDir(path.Join(root, name))
	if e != nil { // 单个文件
		ls, _ := ioutil.ReadDir(path.Dir(path.Join(root, name)))
		for _, k := range ls {
			if k.Name() == path.Base(name) {
				list = append(list, k)
			}
		}
		name = path.Dir(name)
	}

	// 文件排序
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].Name() > list[j].Name() {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	for _, f := range list {
		if f.Name() == ice.PT || f.Name() == ".." {
			continue
		}
		if strings.HasPrefix(f.Name(), ice.PT) && dir_type != TYPE_ALL {
			continue
		}

		p := path.Join(root, name, f.Name())
		if !(dir_type == TYPE_CAT && f.IsDir() || dir_type == TYPE_DIR && !f.IsDir()) && (dir_reg == nil || dir_reg.MatchString(f.Name())) {
			switch cb := m.OptionCB(DIR).(type) {
			case func(f os.FileInfo, p string):
				cb(f, p)
				continue
			case func(p string):
				cb(p)
				continue
			}

			for _, field := range fields {
				switch field {
				case mdb.TIME:
					m.Push(field, f.ModTime().Format(ice.MOD_TIME))
				case mdb.TYPE:
					m.Push(field, kit.Select(CAT, DIR, f.IsDir()))

				case "tree":
					if level == 0 {
						m.Push(field, f.Name())
					} else {
						m.Push(field, strings.Repeat("| ", level-1)+"|-"+f.Name())
					}
				case "full":
					m.Push(field, path.Join(root, name, f.Name())+kit.Select("", ice.PS, f.IsDir()))
				case PATH:
					m.Push(field, path.Join(name, f.Name())+kit.Select("", ice.PS, f.IsDir()))
				case FILE:
					m.Push(field, f.Name()+kit.Select("", ice.PS, f.IsDir()))
				case mdb.NAME:
					m.Push(field, f.Name())

				case SIZE:
					if f.IsDir() {
						if ls, e := ioutil.ReadDir(path.Join(root, name, f.Name())); e == nil {
							m.Push(field, len(ls))
						} else {
							m.Push(field, 0)
						}
					} else {
						m.Push(field, kit.FmtSize(f.Size()))
					}
				case LINE:
					if f.IsDir() {
						if ls, e := ioutil.ReadDir(path.Join(root, name, f.Name())); e == nil {
							m.Push(field, len(ls))
						} else {
							m.Push(field, 0)
						}
					} else {
						nline := 0
						if f, e := os.Open(p); m.Assert(e) {
							defer f.Close()
							for bio := bufio.NewScanner(f); bio.Scan(); nline++ {
								bio.Text()
							}
						}
						m.Push(field, nline)
					}
				case mdb.HASH, "hashs":
					var h [20]byte
					if f.IsDir() {
						if d, e := ioutil.ReadDir(p); m.Assert(e) {
							meta := []string{}
							for _, v := range d {
								meta = append(meta, kit.Format("%s%d%s", v.Name(), v.Size(), v.ModTime()))
							}
							kit.Sort(meta)
							h = sha1.Sum([]byte(strings.Join(meta, "")))
						}
					} else {
						if f, e := ioutil.ReadFile(path.Join(name, f.Name())); m.Assert(e) {
							h = sha1.Sum(f)
						}
					}

					m.Push(mdb.HASH, kit.Select(kit.Format(h[:6]), kit.Format(h[:]), field == mdb.HASH))
				case mdb.LINK:
					m.PushDownload(mdb.LINK, kit.Select("", f.Name(), !f.IsDir()), path.Join(root, name, f.Name()))
				case mdb.SHOW:
					switch p := m.MergeURL2("/share/local/"+path.Join(name, f.Name()), ice.POD, m.Option(ice.MSG_USERPOD)); kit.Ext(f.Name()) {
					case "png", "jpg":
						m.PushImages(field, p)
					case "mp4":
						m.PushVideos(field, p)
					default:
						m.Push(field, "")
					}
				case "action":
					if m.IsCliUA() || m.Option(ice.MSG_USERROLE) == aaa.VOID {
						break
					}
					m.PushButton(kit.Select("", TRASH, !f.IsDir()))
				default:
					m.Push(field, "")
				}
			}
		}

		if f.IsDir() && deep {
			_dir_list(m, root, path.Join(name, f.Name()), level+1, deep, dir_type, dir_reg, fields)
		}
	}
	return m
}
func _dir_search(m *ice.Message, kind, name string) {
	msg := _dir_list(m.Spawn(), PWD, "", 0, true, TYPE_BOTH, nil, kit.Split("time,type,name"))
	msg.Table(func(index int, value map[string]string, head []string) {
		if !strings.Contains(value[mdb.NAME], name) {
			return
		}
		if value[mdb.TYPE] == CAT {
			value[mdb.TYPE] = kit.Ext(value[mdb.NAME])
		}
		m.PushSearch(value)
	})
}

func Dir(m *ice.Message, sort string) *ice.Message {
	m.Option(DIR_TYPE, TYPE_DIR)
	m.Copy(m.Cmd(DIR, PWD).Sort(sort))

	m.Option(DIR_TYPE, TYPE_CAT)
	m.Copy(m.Cmd(DIR, PWD).Sort(sort))
	return m
}
func MkdirAll(m *ice.Message, p string) error {
	m.Log_EXPORT("mkdir", "dir", p)
	return os.MkdirAll(p, ice.MOD_DIR)
}

const (
	TYPE_ALL  = "all"
	TYPE_CAT  = "cat"
	TYPE_DIR  = "dir"
	TYPE_BOTH = "both"
)
const (
	DIR_PACK = "dir_pack"
	DIR_ROOT = "dir_root"
	DIR_TYPE = "dir_type"
	DIR_DEEP = "dir_deep"
	DIR_REG  = "dir_reg"

	DIR_DEF_FIELDS = "time,path,size,action"
	DIR_WEB_FIELDS = "time,size,path,action,link"
	DIR_CLI_FIELDS = "path,size,time"
)
const DIR = "dir"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DIR: {Name: DIR, Help: "目录", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		DIR: {Name: "dir path field... auto upload", Help: "目录", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, m.CommandKey(), m.PrefixKey())
				m.Cmd(mdb.RENDER, mdb.CREATE, m.CommandKey(), m.PrefixKey())
			}},
			mdb.SEARCH: {Name: "search type name", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH {
					return
				}
				_dir_search(m, arg[0], arg[1])
			}},
			mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_dir_list(m, arg[2], arg[1], 0, m.Option(DIR_DEEP) == ice.TRUE, kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)),
					nil, kit.Split(kit.Select("time,size,type,path", m.OptionFields())))
			}},
			mdb.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				m.Upload(m.Option(PATH))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				os.Remove(m.Option(PATH))
			}},
			TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TRASH, m.Option(PATH))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option(DIR_ROOT) != "" {
				m.Info("dir_root: %v", m.Option(DIR_ROOT))
			}
			_dir_list(m, kit.Select(PWD, m.Option(DIR_ROOT)), kit.Select(PWD, arg, 0),
				0, m.Option(DIR_DEEP) == ice.TRUE, kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)), kit.Regexp(m.Option(DIR_REG)),
				kit.Split(kit.Select(kit.Select(DIR_DEF_FIELDS, m.OptionFields()), kit.Join(kit.Slice(arg, 1)))))
			m.SortTimeR(mdb.TIME)
			m.StatusTimeCount()
		}},
	}})
}
