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

const (
	FILE  = "file"
	TRASH = "trash"
)

func _file_list(m *ice.Message, root string, name string, level int, deep bool, dir_type string, dir_reg *regexp.Regexp, fields []string) {
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
						m.Push("time", f.ModTime().Format(ice.ICE_TIME))
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
				_file_list(m, root, p, level+1, deep, dir_type, dir_reg, fields)
			}
		}
	}
}
func _file_show(m *ice.Message, name string) {
	if f, e := os.OpenFile(name, os.O_RDONLY, 0777); m.Assert(e) {
		defer f.Close()
		if s, e := f.Stat(); m.Assert(e) {
			buf := make([]byte, s.Size())
			if n, e := f.Read(buf); m.Assert(e) {
				m.Log_IMPORT("file", name, "size", n)
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
				m.Log_EXPORT("file", p, "size", n)
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
					m.Log_IMPORT("file", p, "size", n)
				}
			}
		}
	}
}
func _file_link(m *ice.Message, name string, from string) {
	m.Cmd("nfs.trash", name)
	os.MkdirAll(path.Dir(name), 0777)
	os.Link(from, name)
}
func _file_trash(m *ice.Message, name string, from ...string) {
	if s, e := os.Stat(name); e == nil {
		if s.IsDir() {
			name := path.Base(name) + ".tar.gz"
			m.Cmd(ice.CLI_SYSTEM, "tar", "zcf", name, name)
		}

		if f, e := os.Open(name); m.Assert(e) {
			defer f.Close()

			h := kit.Hashs(f)
			p := path.Join(m.Conf("trash", "meta.path"), h[:2], h)
			os.MkdirAll(path.Dir(p), 0777)
			os.Rename(name, p)

			m.Cmd(ice.WEB_FAVOR, "trash", "bin", name, p)
		}
	}
}

func FileSave(m *ice.Message, file string, text ...string) {
	_file_save(m, file, text...)
}

var Index = &ice.Context{Name: "nfs", Help: "存储模块",
	Configs: map[string]*ice.Config{
		TRASH: {Name: "trash", Help: "删除", Value: kit.Data("path", "var/trash")},
	},
	Commands: map[string]*ice.Command{
		"dir": {Name: "dir path field...", Help: "目录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			rg, _ := regexp.Compile(m.Option("dir_reg"))
			_file_list(m, kit.Select("./", m.Option("dir_root")), kit.Select("", arg, 0),
				0, m.Options("dir_deep"), kit.Select("both", m.Option("dir_type")), rg,
				strings.Split(kit.Select("time size line path", strings.Join(arg[1:], " ")), " "))
		}},
		"cat": {Name: "cat file", Help: "查看", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_show(m, arg[0])
		}},
		"save": {Name: "save file text...", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_save(m, arg[0], arg[1:]...)
		}},
		"copy": {Name: "copy file from...", Help: "复制", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_copy(m, arg[0], arg[1:]...)
		}},
		"link": {Name: "link file from", Help: "链接", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_link(m, arg[0], arg[1])
		}},

		TRASH: {Name: "trash file", Help: "删除", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_file_trash(m, arg[0])
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
