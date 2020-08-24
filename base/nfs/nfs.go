package nfs

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"bufio"
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
		if m.Option(ice.MSG_USERROLE) == "root" {
			break
		}
		if m.Option(ice.MSG_USERROLE) == "tech" {
			break
		}
		return
	}

	if fs, e := ioutil.ReadDir(path.Join(root, name)); e != nil {
		if f, e := os.Open(path.Join(root, name)); e == nil {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e == nil {
				m.Echo(string(b))
				return
			}
		}
		m.Log(ice.LOG_WARN, "%s", e)
	} else {
		for _, f := range fs {
			if f.Name() == "." || f.Name() == ".." {
				continue
			}
			if strings.HasPrefix(f.Name(), ".") && dir_type != "all" {
				continue
			}

			p := path.Join(root, name, f.Name())
			// if f, e = os.Lstat(p); e != nil {
			// 	m.Log("info", "%s", e)
			// 	continue
			// } else if (f.Mode()&os.ModeSymlink) != 0 && f.IsDir() {
			// 	// continue
			// } else {
			// }
			//
			if !(dir_type == "file" && f.IsDir() || dir_type == "dir" && !f.IsDir()) && (dir_reg == nil || dir_reg.MatchString(f.Name())) {
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
							m.Push("link", m.Cmdx("mdb.render", "download",
								kit.MergeURL(path.Join("/share/local/", root, name, f.Name()), "pod", m.Option("user.pod")), f.Name()))
						}

					case "tree":
						if level == 0 {
							m.Push("tree", f.Name())
						} else {
							m.Push("tree", strings.Repeat("| ", level-1)+"|-"+f.Name())
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
							if d, e := ioutil.ReadDir(p); m.Assert(e) {
								count := 0
								for _, f := range d {
									if strings.HasPrefix(f.Name(), ".") {
										continue
									}
									count++
								}
								m.Push("line", count)
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
							m.Push("hash", hex.EncodeToString(h[:4]))
						}
					}
				}
			}
			if f.IsDir() && deep {
				_file_list(m, root, path.Join(name, f.Name()), level+1, deep, dir_type, dir_reg, fields)
			}
		}
	}
}
func _file_show(m *ice.Message, name string) {
	if f, e := os.OpenFile(path.Join(m.Option(DIR_ROOT), name), os.O_RDONLY, 0640); m.Assert(e) {
		defer f.Close()

		if s, e := f.Stat(); m.Assert(e) {
			buf := make([]byte, s.Size())
			if n, e := f.Read(buf); m.Assert(e) {
				m.Log_IMPORT(kit.MDB_FILE, name, kit.MDB_SIZE, n)
				m.Echo(string(buf[:n]))
			}
		}
	}
}
func _file_save(m *ice.Message, name string, text ...string) {
	if f, p, e := kit.Create(name); m.Assert(e) {
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
			if s, e := os.Open(v); !m.Warn(e != nil, "%s", e) {
				defer s.Close()
				if n, e := io.Copy(f, s); !m.Warn(e != nil, "%s", e) {
					m.Log_IMPORT(kit.MDB_FILE, p, kit.MDB_SIZE, n)
				}
			}
		}
	}
}
func _file_link(m *ice.Message, name string, from string) {
	os.MkdirAll(path.Dir(name), 0760)
	os.Remove(name)
	os.Link(from, name)
	m.Echo(name)
}
func _file_trash(m *ice.Message, name string) {
	if s, e := os.Stat(name); e == nil {
		if s.IsDir() {
			name := path.Base(name) + ".tar.gz"
			m.Cmd(cli.SYSTEM, "tar", "zcf", name, name)
		}

		if f, e := os.Open(name); m.Assert(e) {
			defer f.Close()

			h := kit.Hashs(f)
			p := path.Join(m.Conf(TRASH, "meta.path"), h[:2], h)
			os.MkdirAll(path.Dir(p), 0777)
			os.Rename(name, p)

			m.Cmd("web.favor", TRASH, "bin", name, p)
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
	_file_list(msg, "./", "", 0, true, "both", rg, []string{"time", "size", "type", "path"})
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

		m.Push("pod", m.Option("pod"))
		m.Push("ctx", NFS)
		m.Push("cmd", FILE)
		m.Push(kit.MDB_TIME, value["time"])
		m.Push(kit.MDB_SIZE, value["size"])
		m.Push(kit.MDB_TYPE, ext)
		m.Push(kit.MDB_NAME, value["path"])
		m.Push(kit.MDB_TEXT, "")
	})
}

const (
	CAT   = "cat"
	SAVE  = "save"
	COPY  = "copy"
	LINK  = "link"
	TRASH = "trash"

	DIR  = "dir"
	FILE = "file"
	NFS  = "nfs"
)
const (
	DIR_ROOT = "dir_root"
	DIR_TYPE = "dir_type"
	DIR_DEEP = "dir_deep"
	DIR_REG  = "dir_reg"
)
const (
	TYPE_ALL  = "all"
	TYPE_DIR  = "dir"
	TYPE_FILE = "file"
)

var Index = &ice.Context{Name: "nfs", Help: "存储模块",
	Configs: map[string]*ice.Config{
		TRASH: {Name: "trash", Help: "删除", Value: kit.Data("path", "var/trash")},
		FILE: {Name: "file", Help: "文件", Value: kit.Data(
			"source", kit.Dict(
				"makefile", "true",
				"sh", "true",
				"shy", "true",
				"py", "true",
				"js", "true",
				"go", "true",
				"vim", "true",
				"conf", "true",
				"json", "true",
			),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(mdb.SEARCH, mdb.CREATE, FILE)
			m.Cmd(mdb.SEARCH, mdb.CREATE, DIR)
			m.Cmd(mdb.RENDER, mdb.CREATE, FILE, FILE, NFS)
			m.Cmd(mdb.RENDER, mdb.CREATE, DIR)

			m.Cmd(mdb.RENDER, mdb.CREATE, "bin", m.AddCmd(&ice.Command{Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo("hello world")
			}}))
			m.Cmd(mdb.RENDER, mdb.CREATE, "m4v", m.AddCmd(&ice.Command{Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if !strings.HasPrefix(arg[1], "http") {
					arg[1] = path.Join("/share/local", kit.Path(arg[1]))
				}
				m.Echo(`<video src="%s" height=400 controls autoplay loop></video>`, arg[1])
			}}))
		}},

		FILE: {Name: "file", Help: "文件", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_file_search(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_file_show(m, path.Join(arg[2], arg[1]))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		DIR: {Name: "dir path=auto field... 查看:button=auto 返回 上传", Help: "目录", Action: map[string]*ice.Action{
			"upload": {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					m.Cmdy("spide", "dev", "cache", "GET", kit.MergeURL(arg[2], "path", arg[1], "name", arg[0]))
					m.Option("name", arg[0])
				} else {
					m.Cmdy("cache", "upload")
				}
				m.Cmdy("cache", "watch", m.Option("data"), path.Join(m.Option("path"), m.Option("name")))
			}},
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_file_search(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_file_list(m, arg[2], arg[1], 0, m.Option(DIR_DEEP) == "true", kit.Select("both", m.Option(DIR_TYPE)),
					nil, []string{"time", "size", "type", "path"})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			rg, _ := regexp.Compile(m.Option(DIR_REG))
			_file_list(m, kit.Select("./", m.Option(DIR_ROOT)), kit.Select("", arg, 0),
				0, m.Options(DIR_DEEP), kit.Select("both", m.Option(DIR_TYPE)), rg,
				strings.Split(kit.Select("time size line path", strings.Join(arg[1:], " ")), " "))
		}},
		CAT: {Name: "cat file", Help: "查看", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_show(m, arg[0])
		}},
		SAVE: {Name: "save file text...", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
	},
}

func init() { ice.Index.Register(Index, nil, DIR, CAT, SAVE, COPY, LINK, TRASH) }
