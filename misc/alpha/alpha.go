package alpha

import (
	"bytes"
	"encoding/csv"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/wiki"
	"github.com/shylinux/toolkits"
	"math/rand"
	"os"
	"path"
)

var Index = &ice.Context{Name: "alpha", Help: "英汉词典",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"alpha": {Name: "alpha", Help: "英汉词典", Value: kit.Data(
			"store", "var/alpha/", "limit", "2000", "least", "1000",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", "alpha.json")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "alpha.json", "alpha")
		}},

		"alpha": {Name: "alpha [load|list]", Help: "英汉词典", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			count := kit.Int(m.Conf("alpha", "meta.count"))
			if len(arg) == 0 {
				arg = append(arg, "list", kit.Format(count-rand.Intn(count)))
			}

			switch arg[0] {
			case "load":
				// 加载词库
				m.Cmd(ice.MDB_IMPORT, "web.wiki.alpha.alpha", "", "list",
					m.Cmd(ice.WEB_CACHE, "catch", "csv", arg[1]).Append("data"))
			case "list":
				// 词汇列表
				m.Option("cache.offend", kit.Select("0", arg, 1))
				m.Option("cache.limit", kit.Select("10", arg, 2))
				m.Grows("alpha", nil, "", "", func(index int, value map[string]interface{}) {
					m.Push("", value, []string{"id", "word", "translation", "definition"})
				})
			}
		}},
		"random": {Name: "random [count]", Help: "随机词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			count := kit.Int(m.Conf("alpha", "meta.count"))
			for i := 0; i < kit.Int(kit.Select("10", arg, 0)); i++ {
				m.Cmdy("alpha", "list", count-rand.Intn(count), 1)
			}
		}},
		"search": {Name: "search [word [method]]", Help: "查找词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 收藏列表
				m.Cmdy(ice.WEB_FAVOR, "alpha.word")
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
				arg[0] = "," + arg[0] + "$"
			}

			// 字段列表
			field := m.Confm("alpha", "meta.field")
			if field == nil {
				field = map[string]interface{}{}
				head := []string{}
				if f, e := os.Open(path.Join(m.Conf("alpha", "meta.store"), "web.wiki.alpha.alpha.csv")); m.Assert(e) {
					defer f.Close()
					bio := csv.NewReader(f)
					head, e = bio.Read()
				}
				for i, k := range head {
					field[k] = i
				}
				m.Conf("alpha", "meta.field", field)
			}

			// 搜索词汇
			bio := csv.NewReader(bytes.NewBufferString(m.Cmdx(ice.CLI_SYSTEM, "grep", "-rh", arg[0], m.Conf("alpha", "meta.store"))))
			for i := 0; i < 100; i++ {
				if line, e := bio.Read(); e != nil {
					break
				} else {
					if method == "word" && i == 0 {
						// 添加收藏
						m.Cmd(ice.WEB_FAVOR, "alpha.word", "alpha",
							line[kit.Int(field["word"])], line[kit.Int(field["translation"])],
							"id", line[kit.Int(field["id"])], "definition", line[kit.Int(field["definition"])],
						)
					}
					for _, k := range []string{"id", "word", "translation", "definition"} {
						// 输出词汇
						m.Push(k, line[kit.Int(field[k])])
					}
				}
			}
		}},
	},
}

func init() { wiki.Index.Register(Index, nil) }
