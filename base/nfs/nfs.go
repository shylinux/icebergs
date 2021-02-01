package nfs

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

func _file_ext(name string) string {
	return strings.ToLower(kit.Select(path.Base(name), strings.TrimPrefix(path.Ext(name), ".")))
}

func _file_list(m *ice.Message, root string, name string, level int, deep bool, dir_type string, dir_reg *regexp.Regexp, fields []string) {
	switch strings.Split(name, "/")[0] {
	case "etc", "var":
		if m.Option(ice.MSG_USERROLE) == "void" {
			return // 保护目录
		}
	}

	fs, e := ioutil.ReadDir(path.Join(root, name))
	if e != nil {
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

		if !(dir_type == TYPE_FILE && f.IsDir() || dir_type == TYPE_DIR && !f.IsDir()) && (dir_reg == nil || dir_reg.MatchString(f.Name())) {
			for _, field := range fields {
				switch field {
				case "time":
					m.Push("time", f.ModTime().Format(ice.MOD_TIME))
				case "type":
					if m.Assert(e) && f.IsDir() {
						m.Push("type", DIR)
					} else {
						m.Push("type", FILE)
					}
				case "tree":
					if level == 0 {
						m.Push("tree", f.Name())
					} else {
						m.Push("tree", strings.Repeat("| ", level-1)+"|-"+f.Name())
					}
				case "full":
					if f.IsDir() {
						m.Push("full", path.Join(root, name, f.Name())+"/")
					} else {
						m.Push("full", path.Join(root, name, f.Name()))
					}
				case "path":
					if f.IsDir() {
						m.Push("path", path.Join(name, f.Name())+"/")
					} else {
						m.Push("path", path.Join(name, f.Name()))
					}
				case "file":
					if f.IsDir() {
						m.Push("file", f.Name()+"/")
					} else {
						m.Push("file", f.Name())
					}
				case "name":
					m.Push("name", f.Name())
				case "link":
					if f.IsDir() {
						m.Push("link", "")
					} else {
						m.PushDownload(f.Name(), path.Join(root, name, f.Name()))
					}

				case "size":
					if f.IsDir() {
						if ls, e := ioutil.ReadDir(path.Join(root, name, f.Name())); e == nil {
							m.Push("size", len(ls))
						} else {
							m.Push("size", 0)
						}
					} else {
						m.Push("size", kit.FmtSize(f.Size()))
					}
				case "line":
					if f.IsDir() {
						if ls, e := ioutil.ReadDir(path.Join(root, name, f.Name())); e == nil {
							m.Push("size", len(ls))
						} else {
							m.Push("size", 0)
						}
					} else {
						nline := 0
						if f, e := os.Open(p); m.Assert(e) {
							defer f.Close()
							for bio := bufio.NewScanner(f); bio.Scan(); nline++ {
								bio.Text()
							}
						}
						m.Push("line", nline)
					}
				case "hash", "hashs":
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
					if field == "hash" {
						m.Push("hash", hex.EncodeToString(h[:]))
					} else {
						m.Push("hash", hex.EncodeToString(h[:6]))
					}
				}
			}
		}

		if f.IsDir() && deep {
			_file_list(m, root, path.Join(name, f.Name()), level+1, deep, dir_type, dir_reg, fields)
		}
	}
}
func _file_show(m *ice.Message, name string) {
	if f, e := os.Open(path.Join(m.Option(DIR_ROOT), name)); e == nil {
		defer f.Close()

		switch cb := m.Optionv(CAT_CB).(type) {
		case func(string, int):
			bio := bufio.NewScanner(f)
			for i := 0; bio.Scan(); i++ {
				cb(bio.Text(), i)
			}

		default:
			if s, e := f.Stat(); m.Assert(e) {
				buf := make([]byte, s.Size())
				if n, e := f.Read(buf); m.Assert(e) {
					m.Log_IMPORT(kit.MDB_FILE, name, kit.MDB_SIZE, n)
					m.Echo(string(buf[:n]))
				}
			}
		}
		return
	}

	if b, ok := ice.BinPack["/"+name]; ok {
		m.Info("binpack %v %v", len(b), name)
		m.Echo(string(b))
		return // 打包文件
	}

	// 远程文件
	switch cb := m.Optionv(CAT_CB).(type) {
	case func(string, int):
		bio := bufio.NewScanner(bytes.NewBufferString(m.Cmdx("web.spide", "dev", "raw", "GET", path.Join("/share/local/", name))))
		for i := 0; bio.Scan(); i++ {
			cb(bio.Text(), i)
		}
	default:
		m.Cmdy("web.spide", "dev", "raw", "GET", path.Join("/share/local/", name))
	}
}
func _file_save(m *ice.Message, name string, text ...string) {
	if f, p, e := kit.Create(path.Join(m.Option(DIR_ROOT), name)); m.Assert(e) {
		defer f.Close()

		for _, v := range text {
			if n, e := f.WriteString(v); m.Assert(e) {
				m.Log_EXPORT(kit.MDB_FILE, p, kit.MDB_SIZE, n)
			}
		}
		m.Echo(p)
	}
}
func _file_copy(m *ice.Message, name string, from ...string) {
	if f, p, e := kit.Create(name); m.Assert(e) {
		defer f.Close()

		for _, v := range from {
			if s, e := os.Open(v); !m.Warn(e != nil, e) {
				defer s.Close()

				if n, e := io.Copy(f, s); !m.Warn(e != nil, e) {
					m.Log_EXPORT(kit.MDB_FILE, p, kit.MDB_SIZE, n)
				}
			}
		}
	}
}
func _file_link(m *ice.Message, name string, from string) {
	if from == "" {
		return
	}
	os.Remove(name)
	os.MkdirAll(path.Dir(name), 0750)
	os.Link(from, name)
	m.Echo(name)
}
func _file_trash(m *ice.Message, name string) {
	if s, e := os.Stat(name); e == nil {
		if s.IsDir() {
			tar := path.Base(name) + ".tar.gz"
			m.Cmd(cli.SYSTEM, "tar", "zcf", tar, name)
			name = tar
		}

		if f, e := os.Open(name); m.Assert(e) {
			defer f.Close()

			h := kit.Hashs(f)
			p := path.Join(m.Conf(TRASH, "meta.path"), h[:2], h)
			os.MkdirAll(path.Dir(p), 0777)
			os.Rename(name, p)
			m.Cmdy("mdb.insert", m.Prefix(TRASH), "", "list", "file", p, "from", name)
		}
	}
}
func _file_search(m *ice.Message, kind, name, text string, arg ...string) {
	if kind == kit.MDB_FOREACH {
		return
	}
	rg, e := regexp.Compile("")
	m.Assert(e)

	msg := m.Spawn()
	_file_list(msg, "./", "", 0, true, TYPE_ALL, rg, []string{"time", "size", "type", "path"})

	msg.Table(func(index int, value map[string]string, head []string) {
		if !strings.Contains(value["path"], name) {
			return
		}

		ext := _file_ext(value["path"])
		if value["type"] == DIR {
			ext = DIR
		} else if m.Richs(mdb.RENDER, nil, ext, nil) == nil {
			ext = value["type"]
		}

		m.PushSearch("cmd", FILE, "time", value["time"], "size", value["size"], "type", ext, "name", value["path"], "text", "")
	})
}

const (
	OPEN  = "open"
	CLOSE = "close"
)
const (
	CAT_CB = "cat_cb"
	CAT    = "cat"
	SAVE   = "save"
	COPY   = "copy"
	LINK   = "link"
	TAIL   = "tail"
)
const (
	DIR_ROOT = "dir_root"
	DIR_TYPE = "dir_type"
	DIR_DEEP = "dir_deep"
	DIR_REG  = "dir_reg"
)
const (
	TYPE_ALL  = "all"
	TYPE_BOTH = "both"
	TYPE_FILE = "file"
	TYPE_DIR  = "dir"
)

const TRASH = "trash"
const FILE = "file"
const DIR = "dir"

var Index = &ice.Context{Name: "nfs", Help: "存储模块",
	Configs: map[string]*ice.Config{
		DIR: {Name: DIR, Help: "目录", Value: kit.Data()},
		FILE: {Name: FILE, Help: "文件", Value: kit.Data(
			"source", kit.Dict(
				"sh", "true", "shy", "true", "py", "true",
				"go", "true", "vim", "true", "js", "true",
				"conf", "true", "json", "true",
				"makefile", "true",
			),
		)},
		TRASH: {Name: TRASH, Help: "删除", Value: kit.Data("path", "var/trash")},

		TAIL: {Name: TAIL, Help: "跟踪", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(mdb.SEARCH, mdb.CREATE, DIR)
			m.Cmd(mdb.RENDER, mdb.CREATE, DIR)

			m.Cmd(mdb.SEARCH, mdb.CREATE, FILE, m.Prefix(FILE))
			m.Cmd(mdb.RENDER, mdb.CREATE, FILE, m.Prefix(FILE))
			m.Richs(TAIL, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				value = kit.GetMeta(value)
				m.Option(kit.MDB_HASH, key)
				m.Cmd(TAIL, mdb.CREATE,
					kit.MDB_FILE, kit.Format(value[kit.MDB_FILE]),
					kit.MDB_NAME, kit.Format(value[kit.MDB_NAME]),
				)
			})
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},
		DIR: {Name: "dir path field... auto upload", Help: "目录", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_file_search(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_file_list(m, arg[2], arg[1], 0, m.Option(DIR_DEEP) == "true", kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)),
					nil, []string{"time", "size", "type", "path"})
			}},

			"upload": {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))

				if p := path.Join(m.Option("path"), up[1]); m.Option(ice.MSG_USERPOD) == "" {
					m.Cmdy("cache", "watch", up[0], p)
				} else {
					m.Cmdy("spide", "dev", "save", p, "GET", kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/cache/"+up[0]))
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "")
			}
			reg, _ := regexp.Compile(m.Option(DIR_REG))
			_file_list(m, kit.Select("./", m.Option(DIR_ROOT)), arg[0],
				0, m.Options(DIR_DEEP), kit.Select(TYPE_BOTH, m.Option(DIR_TYPE)), reg,
				kit.Split(kit.Select("time size path", strings.Join(arg[1:], " "))))
			m.SortTimeR(kit.MDB_TIME)
		}},
		FILE: {Name: "file path auto", Help: "文件", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_file_search(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_file_show(m, path.Join(arg[2], arg[1]))
			}},

			"append": {Name: "append", Help: "追加", Hand: func(m *ice.Message, arg ...string) {
				if strings.Contains(arg[0], "/") {
					os.MkdirAll(path.Dir(arg[0]), ice.MOD_DIR)
				}

				if f, e := os.OpenFile(arg[0], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664); m.Assert(e) {
					defer f.Close()

					for _, k := range arg[1:] {
						if n, e := f.WriteString(k); m.Assert(e) {
							m.Log_EXPORT(kit.MDB_FILE, arg[0], kit.MDB_SIZE, n)
						}
					}
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				m.Cmdy(DIR, arg)
				return
			}
			_file_show(m, arg[0])
		}},

		CAT: {Name: "cat file", Help: "查看", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_show(m, arg[0])
		}},
		SAVE: {Name: "save file text...", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 1 {
				arg = append(arg, m.Option("content"))
			}
			_file_save(m, arg[0], arg[1:]...)
		}},
		COPY: {Name: "copy file from...", Help: "复制", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_copy(m, arg[0], arg[1:]...)
		}},
		LINK: {Name: "link file from", Help: "链接", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_link(m, arg[0], arg[1])
		}},
		TRASH: {Name: "trash file", Help: "删除", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_trash(m, arg[0])
		}},

		TAIL: {Name: "tail hash id auto create", Help: "文件", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create file name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_HASH) == "" {
					m.Cmdy(mdb.INSERT, TAIL, "", mdb.HASH, arg)
					m.Option(kit.MDB_HASH, m.Result())
				}
				h := m.Option(kit.MDB_HASH)
				for _, file := range kit.Split(m.Option(kit.MDB_FILE), ",") {
					func(file string) {
						r, w := io.Pipe()
						m.Go(func() {
							for bio := bufio.NewScanner(r); bio.Scan(); {
								m.Grow(TAIL, kit.Keys(kit.MDB_HASH, h), kit.Dict(
									kit.MDB_FILE, file, kit.MDB_TEXT, bio.Text(),
								))
							}
						})
						m.Option(cli.CMD_STDOUT, w)
						m.Option(cli.CMD_STDERR, w)
						m.Option(mdb.CACHE_CLEAR_ON_EXIT, "true")
						m.Cmd(cli.DAEMON, "tail", "-n", "0", "-f", file)
					}(file)
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, TAIL, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(mdb.FIELDS, kit.Select("time,hash,count,name,file", kit.Select("time,id,file,text", mdb.DETAIL, len(arg) > 1), len(arg) > 0))
			m.Cmdy(mdb.SELECT, TAIL, "", mdb.ZONE, arg)
			m.PushAction(kit.Select("", mdb.REMOVE, len(arg) == 0))
			if len(arg) == 1 {
				m.Option(ice.MSG_CONTROL, "_page")
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil, DIR, CAT, SAVE, COPY, LINK, TRASH, TAIL) }
