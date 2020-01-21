package input

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"
	"os"
	"path"
	"strings"
)

var Index = &ice.Context{Name: "input", Help: "输入法",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"input": {Name: "input", Help: "输入法", Value: kit.Data(
			"store", "var/input/", "limit", "2000", "least", "1000", "fsize", "100000",
			"field", kit.Dict("file", 0, "line", 1, "code", 2, "id", 3, "text", 4, "time", 5, "weight", 6),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", "input.json")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "input.json", "cli.input.input")
		}},

		"input": {Name: "input load|list|push|save", Help: "输入法", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "list")
			}

			switch arg[0] {
			case "load":
				// 加载词库
				lib := kit.Select(path.Base(arg[1]), arg, 2)
				m.Option("cache.fsize", m.Conf("input", "meta.fsize"))
				m.Option("cache.limit", m.Conf("input", "meta.limit"))
				m.Option("cache.least", m.Conf("input", "meta.least"))
				m.Assert(os.RemoveAll(m.Option("cache.store", path.Join(m.Conf("input", "meta.store"), lib))))
				m.Conf("input", lib, "")
				if f, e := os.Open(arg[1]); m.Assert(e) {
					bio := bufio.NewScanner(f)
					for bio.Scan() {
						if strings.HasPrefix(bio.Text(), "#") {
							continue
						}
						line := kit.Split(bio.Text(), " \t")
						if line[2] == "0" {
							continue
						}
						m.Grow("input", lib, kit.Dict(
							"text", line[0], "code", line[1], "weight", line[2],
						))
					}
					m.Option("cache.limit", 0)
					m.Option("cache.least", 0)
					n := m.Grow("input", lib, kit.Dict(
						"text", "成功", "code", "z", "weight", "0",
					))
					m.Echo("%s: %d", lib, n)
				}
			case "push":
				// 添加词汇
				lib := kit.Select("person", arg, 3)
				m.Option("cache.limit", 0)
				m.Option("cache.least", 0)
				m.Option("cache.store", path.Join(m.Conf("input", "meta.store"), lib))
				n := m.Grow("input", lib, kit.Dict(
					"text", arg[1], "code", arg[2], "weight", kit.Select("99990000", arg, 4),
				))
				m.Echo("%s: %d", lib, n)
			case "list":
				// 词汇列表
				lib := kit.Select("person", arg, 1)
				m.Option("cache.offend", kit.Select("0", arg, 2))
				m.Option("cache.limit", kit.Select("10", arg, 3))
				m.Grows("input", lib, "", "", func(index int, value map[string]interface{}) {
					m.Push("", value, []string{"id", "code", "text", "weight"})
				})
			case "save":
				// 导出词库
				m.Option("cache.offend", 0)
				m.Option("cache.limit", 1000000)
				if f, p, e := kit.Create(kit.Select(arg[1], arg, 2)); m.Assert(e) {
					defer f.Close()
					n := 0
					m.Grows("input", arg[1], "", "", func(index int, value map[string]interface{}) {
						n++
						fmt.Fprintf(f, "%s %s %s\n", value["text"], value["code"], value["weight"])
					})
					m.Log(ice.LOG_EXPORT, "%s: %d", p, n)
					m.Echo("%s: %d", p, n)
				}
			}
		}},
		"match": {Name: "match [word [method]]", Help: "五笔字码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
			field := m.Confm("input", "meta.field")
			bio := csv.NewReader(bytes.NewBufferString(strings.Replace(m.Cmdx(ice.CLI_SYSTEM, "grep", "-rn", arg[0], m.Conf("input", "meta.store")), ":", ",", -1)))
			for i := 0; i < kit.Int(kit.Select("100", arg, 2)); i++ {
				if line, e := bio.Read(); e != nil {
					break
				} else if len(line) < 3 {
				} else {
					if method == "word" && i == 0 {
						// 添加收藏
						m.Cmd(ice.WEB_FAVOR, "input.word", "input",
							line[kit.Int(field["code"])], line[kit.Int(field["text"])],
							"id", line[kit.Int(field["id"])], "weight", line[kit.Int(field["weight"])],
						)
					}
					// 输出词汇
					m.Push("file", path.Base(line[kit.Int(field["file"])]))
					for _, k := range []string{"id", "code", "text", "weight"} {
						m.Push(k, line[kit.Int(field[k])])
					}
				}
			}
			m.Sort("weight", "int_r")
		}},
	},
}

func init() { cli.Index.Register(Index, nil) }
