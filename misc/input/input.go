package input

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"

	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strings"
)

var Index = &ice.Context{Name: "input", Help: "输入法",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"input": {Name: "input", Help: "输入法", Value: kit.Data(
			"store", "var/input/", "fsize", "100000", "limit", "2000", "least", "1000",
			"repos", "wubi-dict", "local", "some",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd("web.code.git.repos", m.Conf("input", "meta.repos"))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("input")
		}},

		"load": {Name: "load file [name]", Help: "加载词库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 默认词库
				if m.Cmd("web.code.git.check", m.Conf("input", "meta.repos")); m.Confs("input", "wubi86") {
					m.Echo("wubi86: %v", m.Conf("input", "wubi86.meta.count"))
					return
				}
				arg = append(arg, path.Join("usr", m.Conf("input", "meta.repos"), "wubi86"))
			}
			lib := kit.Select(path.Base(arg[0]), arg, 1)

			// 缓存配置
			m.Option("cache.least", m.Conf("input", "meta.least"))
			m.Option("cache.limit", m.Conf("input", "meta.limit"))
			m.Option("cache.fsize", m.Conf("input", "meta.fsize"))
			m.Assert(os.RemoveAll(m.Option("cache.store", path.Join(m.Conf("input", "meta.store"), lib))))
			m.Conf("input", lib, "")

			if f, e := os.Open(arg[0]); m.Assert(e) {
				bio := bufio.NewScanner(f)
				// 加载词库
				for bio.Scan() {
					if strings.HasPrefix(bio.Text(), "#") {
						continue
					}
					line := kit.Split(bio.Text(), " \t")
					if line[2] == "0" {
						continue
					}
					m.Grow("input", lib, kit.Dict("text", line[0], "code", line[1], "weight", line[2]))
				}
				// 保存词库
				m.Option("cache.least", 0)
				m.Option("cache.limit", 0)
				m.Echo("%s: %d", lib, m.Grow("input", lib, kit.Dict("text", "成功", "code", "z", "weight", "0")))
			}
		}},
		"push": {Name: "push text code [weight [lib]]", Help: "添加词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			lib := kit.Select("person", arg, 2)
			m.Option("cache.least", 0)
			m.Option("cache.limit", 0)
			m.Option("cache.store", path.Join(m.Conf("input", "meta.store"), lib))
			m.Echo("%s: %d", lib, m.Grow("input", lib, kit.Dict("text", arg[0], "code", arg[1], "weight", kit.Select("99990000", arg, 3))))
		}},
		"list": {Name: "list [lib [offend [limit]]]", Help: "查看词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			lib := kit.Select("person", arg, 0)
			m.Option("cache.offend", kit.Select("0", arg, 1))
			m.Option("cache.limit", kit.Select("10", arg, 2))
			m.Grows("input", lib, "", "", func(index int, value map[string]interface{}) {
				m.Push("", value, []string{"id", "code", "text", "weight"})
			})
		}},
		"save": {Name: "save lib [filename]", Help: "导出词库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			lib := kit.Select("person", arg, 0)
			m.Option("cache.limit", 1000000)
			m.Option("cache.offend", 0)
			if f, p, e := kit.Create(path.Join("usr", m.Conf("input", "meta.repos"), lib)); m.Assert(e) {
				defer f.Close()
				n := 0
				m.Grows("input", lib, "", "", func(index int, value map[string]interface{}) {
					n++
					fmt.Fprintf(f, "%s %s %s\n", value["text"], value["code"], value["weight"])
				})
				m.Log(ice.LOG_EXPORT, "%s: %d", p, n)
				m.Echo("%s: %d", p, n)
			}
		}},

		"find": {Name: "find key [word|line [limit]]", Help: "五笔字码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 收藏列表
				m.Cmdy(ice.WEB_FAVOR, "input.word")
				return
			}

			// 搜索方法
			method := "word"
			if len(arg) > 1 {
				method = arg[1]
			}
			switch method {
			case "line":
			case "word":
				arg[0] = "^" + arg[0] + ","
			}

			// 搜索词汇
			bio := csv.NewReader(bytes.NewBufferString(strings.Replace(m.Cmdx(ice.CLI_SYSTEM, "grep", "-rn", arg[0], m.Conf("input", "meta.store")), ":", ",", -1)))
			for i := 0; i < kit.Int(kit.Select("100", arg, 2)); i++ {
				if line, e := bio.Read(); e != nil {
					break
				} else if len(line) < 3 {
				} else {
					if method == "word" && i == 0 {
						// 添加收藏
						m.Cmd(ice.WEB_FAVOR, "input.word", "input", line[2], line[4], "id", line[3], "weight", line[6])
					}

					// 输出词汇
					m.Push("file", path.Base(line[0]))
					m.Push("id", line[3])
					m.Push("code", line[2])
					m.Push("text", line[4])
					m.Push("weight", line[6])
				}
			}
			m.Sort("weight", "int_r")
		}},
	},
}

func init() { code.Index.Register(Index, nil) }