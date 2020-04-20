package nfs

import (
	ice "github.com/shylinux/icebergs"
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

func dir(m *ice.Message, root string, name string, level int, deep bool, dir_type string, dir_reg *regexp.Regexp, fields []string, format string) {

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
			if f, e = os.Lstat(p); e != nil {
				m.Log("info", "%s", e)
				continue
			} else if (f.Mode()&os.ModeSymlink) != 0 && f.IsDir() {
				continue
			}

			if !(dir_type == "file" && f.IsDir() || dir_type == "dir" && !f.IsDir()) && (dir_reg == nil || dir_reg.MatchString(f.Name())) {
				for _, field := range fields {
					switch field {
					case "time":
						m.Push("time", f.ModTime().Format(format))
					case "type":
						if m.Assert(e) && f.IsDir() {
							m.Push("type", "dir")
						} else {
							m.Push("type", "file")
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
					case "tree":
						if level == 0 {
							m.Push("tree", f.Name())
						} else {
							m.Push("tree", strings.Repeat("| ", level-1)+"|-"+f.Name())
						}
					case "size":
						m.Push("size", kit.FmtSize(f.Size()))
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
				dir(m, root, p, level+1, deep, dir_type, dir_reg, fields, format)
			}
		}
	}
}
func travel(m *ice.Message, root string, name string, cb func(name string)) {
	if fs, e := ioutil.ReadDir(path.Join(root, name)); e != nil {
		cb(name)
	} else {
		for _, f := range fs {
			if f.Name() == "." || f.Name() == ".." {
				continue
			}
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}

			p := path.Join(root, name, f.Name())
			if f, e = os.Lstat(p); e != nil {
				m.Log("info", "%s", e)
				continue
			} else if (f.Mode()&os.ModeSymlink) != 0 && f.IsDir() {
				continue
			}
			if f.IsDir() {
				travel(m, root, path.Join(name, f.Name()), cb)
				cb(path.Join(name, f.Name()))
			} else {
				cb(path.Join(name, f.Name()))
			}
		}
	}
}

var Index = &ice.Context{Name: "nfs", Help: "存储模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"trash": {Name: "trash", Help: "trash", Value: kit.Data("path", "var/trash")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.APP_SEARCH, "add", "dir", "base", m.AddCmd(&ice.Command{Name: "search word", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "set":
					m.Cmdy("nfs.dir", arg[5])
					return
				}

				travel(m, "./", "", func(name string) {
					if strings.Contains(name, arg[0]) {
						s, e := os.Stat(name)
						m.Assert(e)
						m.Push("pod", m.Option(ice.MSG_USERPOD))
						m.Push("engine", "dir")
						m.Push("favor", "file")
						m.Push("id", kit.FmtSize(s.Size()))
						m.Push("type", strings.TrimPrefix(path.Ext(name), "."))
						m.Push("name", path.Base(name))
						m.Push("text", name)
					}
				})
			}}))
			m.Cmd(ice.APP_COMMEND, "add", "dir", "base", m.AddCmd(&ice.Command{Name: "commend word", Help: "推荐引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "set":
					m.Cmdy("nfs.dir", arg[5])
					return
				}

				travel(m, "./", "", func(name string) {
					score := 0
					m.Richs(ice.APP_COMMEND, "meta.user", m.Option(ice.MSG_USERNAME), func(key string, value map[string]interface{}) {
						m.Grows(ice.APP_COMMEND, kit.Keys("meta.user", kit.MDB_HASH, key, "like"), "", "", func(index int, value map[string]interface{}) {
							switch kit.Value(value, "extra.engine") {
							case "dir":
								if value["type"] == strings.TrimPrefix(path.Ext(name), ".") {
									score += 1
								}
								if value["name"] == path.Base(name) {
									score += 2
								}
								if value["text"] == name {
									score += 3
								}
							default:
							}
						})
						m.Grows(cmd, kit.Keys("meta.user", kit.MDB_HASH, key, "hate"), "", "", func(index int, value map[string]interface{}) {
							switch kit.Value(value, "extra.engine") {
							case "dir":
								if value["type"] == strings.TrimPrefix(path.Ext(name), ".") {
									score -= 1
								}
								if value["name"] == path.Base(name) {
									score -= 2
								}
								if value["text"] == name {
									score -= 3
								}
							default:
							}
						})
					})

					if s, e := os.Stat(name); e == nil {
						m.Push("pod", m.Option(ice.MSG_USERPOD))
						m.Push("engine", "dir")
						m.Push("favor", "file")
						m.Push("id", kit.FmtSize(s.Size()))
						m.Push("score", score)
						m.Push("type", strings.TrimPrefix(path.Ext(name), "."))
						m.Push("name", path.Base(name))
						m.Push("text", name)
					}
				})
			}}))
		}},

		"dir": {Name: "dir name field auto", Help: "目录", List: kit.List(
			kit.MDB_INPUT, "text", "name", "path", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			rg, _ := regexp.Compile(m.Option("dir_reg"))
			dir(m, kit.Select("./", m.Option("dir_root")), kit.Select("", arg, 0), 0, m.Options("dir_deep"), kit.Select("both", m.Option("dir_type")), rg,
				strings.Split(kit.Select("time size line path", arg, 1), " "), ice.ICE_TIME)
		}},
		"cat": {Name: "cat path", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, e := os.OpenFile(arg[0], os.O_RDONLY, 0777); m.Assert(e) {
				defer f.Close()
				buf := make([]byte, 4096000)
				if n, e := f.Read(buf); m.Assert(e) {
					m.Log(ice.LOG_IMPORT, "%d: %s", n, arg[0])
					m.Echo(string(buf[:n]))
				}
			}
		}},

		"save": {Name: "save path text...", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, p, e := kit.Create(arg[0]); m.Assert(e) {
				defer f.Close()
				for _, v := range arg[1:] {
					if n, e := f.WriteString(v); m.Assert(e) {
						m.Log("export", "%v: %v", n, p)
						m.Echo(p)
					}
				}
			}
		}},
		"copy": {Name: "copy path file...", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, _, e := kit.Create(arg[0]); m.Assert(e) {
				defer f.Close()
				for _, v := range arg[1:] {
					if s, e := os.Open(v); !m.Warn(e != nil, "%s", e) {
						if n, e := io.Copy(f, s); m.Assert(e) {
							m.Log(ice.LOG_IMPORT, "%d: %v", n, v)
						}
					}
				}
			}
		}},
		"link": {Name: "link path file", Help: "链接", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("nfs.trash", arg[0])
			os.MkdirAll(path.Dir(arg[0]), 0777)
			os.Link(arg[1], arg[0])
		}},

		"trash": {Name: "trash file", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if s, e := os.Stat(arg[0]); e == nil {
				if s.IsDir() {
					name := path.Base(arg[0]) + ".tar.gz"
					m.Cmd(ice.CLI_SYSTEM, "tar", "zcf", name, arg[0])
				} else {
				}

				if f, e := os.Open(arg[0]); m.Assert(e) {
					defer f.Close()

					h := kit.Hashs(f)
					p := path.Join(m.Conf("trash", "meta.path"), h[:2], h)
					os.MkdirAll(path.Dir(p), 0777)
					os.Rename(arg[0], p)

					m.Cmd(ice.WEB_FAVOR, "trash", "bin", arg[0], p)
				}
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
