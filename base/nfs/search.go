package nfs

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"path"
	"strings"
)

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

func init() {
	Index.Register(&ice.Context{Name: "search", Help: "搜索",
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
							m.Push("time", s.ModTime().Format(ice.ICE_TIME))
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
		},
	}, nil)
}
