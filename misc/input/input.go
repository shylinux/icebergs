package input

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"
	"math/rand"
	"os"
	"path"
	"strings"
)

var Index = &ice.Context{Name: "input", Help: "输入法",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"input": {Name: "input", Help: "输入法", Value: kit.Data(
			"store", "var/input/", "limit", "2000", "least", "1000",
			kit.MDB_SHORT, "code",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", "input.json")
		}},

		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "input.json", "cli.input.input")
		}},

		"input": {Name: "input load|list", Help: "输入法", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			count := kit.Int(m.Conf("input", "meta.count"))
			if len(arg) == 0 {
				arg = append(arg, "list", kit.Format(count-rand.Intn(count)))
			}

			switch arg[0] {
			case "load":
				// 加载词库
				if f, e := os.Open(arg[1]); m.Assert(e) {
					bio := bufio.NewScanner(f)
					for bio.Scan() {
						if strings.HasPrefix(bio.Text(), "#") {
							continue
						}
						line := kit.Split(bio.Text(), " \t")
						m.Grow("input", nil, kit.Dict(
							"text", line[0], "code", line[1], "weight", line[2],
						))
					}
				}
			case "push":
				m.Rich("input", nil, kit.Dict(
					"id", "0", "text", arg[1], "code", arg[2], "weight", kit.Select("99990000", arg, 3),
				))
			case "list":
				// 词汇列表
				m.Option("cache.offend", kit.Select("0", arg, 1))
				m.Option("cache.limit", kit.Select("10", arg, 2))
				m.Grows("input", nil, "", "", func(index int, value map[string]interface{}) {
					m.Push("", value, []string{"id", "code", "text", "weight"})
				})
			}
		}},
		"match": {Name: "match [word [method]]", Help: "五笔字码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 收藏列表
				m.Cmdy(ice.WEB_FAVOR, "input.word")
				return
			}

			m.Info("add %s: %s", arg[0], kit.Hashs(arg[0]))
			m.Richs("input", nil, arg[0], func(key string, value map[string]interface{}) {
				m.Push(key, value, []string{"id", "code", "text", "weight"})
			})

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

			// 字段列表
			field := m.Confm("input", "meta.field")
			if field == nil {
				field = map[string]interface{}{}
				head := []string{}
				if f, e := os.Open(path.Join(m.Conf("input", "meta.store"), "input.csv")); m.Assert(e) {
					defer f.Close()

					bio := csv.NewReader(f)
					head, e = bio.Read()
				}
				for i, k := range head {
					field[k] = i
				}
				m.Conf("input", "meta.field", field)
			}

			// 搜索词汇
			bio := csv.NewReader(bytes.NewBufferString(m.Cmdx(ice.CLI_SYSTEM, "grep", "-rh", arg[0], m.Conf("input", "meta.store"))))
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
					for _, k := range []string{"id", "code", "text", "weight"} {
						// 输出词汇
						m.Push(k, line[kit.Int(field[k])])
					}
				}
			}
			m.Sort("weight", "int_r")
		}},
	},
}

func init() { cli.Index.Register(Index, nil) }
