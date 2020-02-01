package nfs

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

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

var Index = &ice.Context{Name: "nfs", Help: "存储模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"dir": {Name: "dir", Help: "目录", List: kit.List(
			kit.MDB_INPUT, "text", "name", "path", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			rg, _ := regexp.Compile(m.Option("dir_reg"))
			dir(m, kit.Select("./", m.Option("dir_root")), kit.Select("", arg, 0), 0, m.Options("dir_deep"), "both", rg,
				strings.Split(kit.Select("time size line path", arg, 1), " "), ice.ICE_TIME)
		}},
		"cat": {Name: "cat path", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, e := os.OpenFile(arg[0], os.O_RDONLY, 0777); m.Assert(e) {
				defer f.Close()
				buf := make([]byte, 1024)
				if n, e := f.Read(buf); m.Assert(e) {
					m.Log(ice.LOG_IMPORT, "%d: %s", n, arg[0])
					m.Echo(string(buf[:n]))
				}
			}
		}},

		"save": {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
		"echo": {Name: "echo path text", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, e := os.OpenFile(arg[0], os.O_WRONLY, 0777); m.Assert(e) {
				defer f.Close()
				for _, v := range arg[1:] {
					if n, e := f.WriteString(v); m.Assert(e) {
						m.Log(ice.LOG_EXPORT, "%d: %s", n, arg[0])
					}
				}
			}
		}},
		"copy": {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
	},
}

func init() { ice.Index.Register(Index, nil) }
