package nfs

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _dir_show(m *ice.Message, root string, name string, level int, deep bool, dir_type string, dir_reg *regexp.Regexp, fields []string) *ice.Message {
	if !_cat_right(m, name) {
		return m // 没有权限
	}

	fs, e := ioutil.ReadDir(path.Join(root, name))
	if e != nil { // 单个文件
		ls, _ := ioutil.ReadDir(path.Dir(path.Join(root, name)))
		for _, k := range ls {
			if k.Name() == path.Base(name) {
				fs = append(fs, k)
			}
		}
		name = path.Dir(name)
	}

	for _, f := range fs {
		if f.Name() == "." || f.Name() == ".." {
			continue
		}
		if strings.HasPrefix(f.Name(), ".") && dir_type != TYPE_ALL {
			continue
		}

		p := path.Join(root, name, f.Name())
		if !(dir_type == TYPE_CAT && f.IsDir() || dir_type == TYPE_DIR && !f.IsDir()) && (dir_reg == nil || dir_reg.MatchString(f.Name())) {
			for _, field := range fields {
				switch field {
				case kit.MDB_TIME:
					m.Push(field, f.ModTime().Format(ice.MOD_TIME))
				case kit.MDB_TYPE:
					m.Push(field, kit.Select(CAT, DIR, f.IsDir()))

				case "tree":
					if level == 0 {
						m.Push(field, f.Name())
					} else {
						m.Push(field, strings.Repeat("| ", level-1)+"|-"+f.Name())
					}
				case "full":
					m.Push(field, path.Join(root, name, f.Name())+kit.Select("", "/", f.IsDir()))

				case kit.MDB_PATH:
					m.Push(field, path.Join(name, f.Name())+kit.Select("", "/", f.IsDir()))
				case kit.MDB_FILE:
					m.Push(field, f.Name()+kit.Select("", "/", f.IsDir()))
				case kit.MDB_NAME:
					m.Push(field, f.Name())

				case kit.MDB_LINK:
					m.PushDownload(kit.MDB_LINK, kit.Select("", f.Name(), !f.IsDir()), path.Join(root, name, f.Name()))

				case kit.MDB_SIZE:
					if f.IsDir() {
						if ls, e := ioutil.ReadDir(path.Join(root, name, f.Name())); e == nil {
							m.Push(field, len(ls))
						} else {
							m.Push(field, 0)
						}
					} else {
						m.Push(field, kit.FmtSize(f.Size()))
					}
				case kit.MDB_LINE:
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
				case kit.MDB_HASH, "hashs":
					var h [20]byte
					if f.IsDir() {
						if d, e := ioutil.ReadDir(p); m.Assert(e) {
							meta := []string{}
							for _, v := range d {
								meta = append(meta, fmt.Sprintf("%s%d%s", v.Name(), v.Size(), v.ModTime()))
							}
							sort.Strings(meta)
							h = sha1.Sum([]byte(strings.Join(meta, "")))
						}
					} else {
						if f, e := ioutil.ReadFile(path.Join(name, f.Name())); m.Assert(e) {
							h = sha1.Sum(f)
						}
					}

					m.Push(kit.MDB_HASH, kit.Select(hex.EncodeToString(h[:6]), hex.EncodeToString(h[:]), field == kit.MDB_HASH))
				default:
					m.Push(field, "")
				}
			}
		}

		if f.IsDir() && deep {
			_dir_show(m, root, path.Join(name, f.Name()), level+1, deep, dir_type, dir_reg, fields)
		}
	}
	return m
}
func _dir_search(m *ice.Message, kind, name string) {
	if kind == kit.MDB_FOREACH {
		return
	}

	msg := _dir_show(m.Spawn(), "./", "", 0, true, TYPE_BOTH, nil, strings.Split("time,type,name,text", ","))
	msg.Table(func(index int, value map[string]string, head []string) {
		if !strings.Contains(value[kit.MDB_NAME], name) {
			return
		}

		if value[kit.MDB_TYPE] == CAT {
			value[kit.MDB_TYPE] = _cat_ext(value[kit.MDB_NAME])
		}

		m.PushSearch(kit.SSH_CMD, CAT, value)
	})
}

func Dir(m *ice.Message, sort string) {
	m.Option(DIR_TYPE, TYPE_DIR)
	m.Copy(m.Cmd(DIR, "./").Sort(sort))

	m.Option(DIR_TYPE, TYPE_CAT)
	m.Copy(m.Cmd(DIR, "./").Sort(sort))
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
)
const DIR = "dir"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DIR: {Name: DIR, Help: "目录", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			DIR: {Name: "dir path field... auto upload", Help: "目录", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_dir_search(m, arg[0], arg[1])
				}},
				mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_dir_show(m, arg[2], arg[1], 0, m.Option(DIR_DEEP) == "true", kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)),
						nil, kit.Split("time,size,type,path"))
				}},

				mdb.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
					if p := path.Join(m.Option(kit.MDB_PATH), up[1]); m.Option(ice.MSG_USERPOD) == "" {
						m.Cmdy("web.cache", "watch", up[0], p)
					} else {
						m.Cmdy("web.spide", "dev", "save", p, "GET", kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/cache/"+up[0]))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					arg = append(arg, "")
				}
				_dir_show(m, kit.Select("./", m.Option(DIR_ROOT)), arg[0],
					0, m.Options(DIR_DEEP), kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)), kit.Regexp(m.Option(DIR_REG)),
					kit.Split(kit.Select("time,size,path", strings.Join(arg[1:], ","))))
				m.SortTimeR(kit.MDB_TIME)
			}},
		},
	})
}
