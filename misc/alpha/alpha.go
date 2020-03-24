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
			"store", "var/data/alpha", "fsize", "200000", "limit", "5000", "least", "1000",
			"repos", "word-dict", "local", "some",
			"field", []interface{}{"audio", "bnc", "collins", "definition", "detail", "exchange", "frq", "id", "oxford", "phonetic", "pos", "tag", "time", "translation", "word"},
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("alpha")
		}},
		"load": {Name: "load file [name]", Help: "加载词库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 下载词库
				if m.Cmd("web.code.git.repos", m.Conf("alpha", "meta.repos"), "usr/"+m.Conf("alpha", "meta.repos")); m.Confs("alpha", "ecdict") {
					m.Echo("ecdict: %v", m.Conf("alpha", "ecdict.meta.count"))
					return
				}
				arg = append(arg, path.Join("usr", m.Conf("alpha", "meta.repos"), "ecdict"))
			}

			// 清空数据
			lib := kit.Select(path.Base(arg[0]), arg, 1)
			m.Assert(os.RemoveAll(path.Join(m.Conf("alpha", "meta.store"), lib)))
			m.Conf("alpha", lib, "")

			// 缓存配置
			m.Conf("alpha", kit.Keys(lib, "meta.store"), path.Join(m.Conf("alpha", "meta.store"), lib))
			m.Conf("alpha", kit.Keys(lib, "meta.fsize"), m.Conf("alpha", "meta.fsize"))
			m.Conf("alpha", kit.Keys(lib, "meta.limit"), m.Conf("alpha", "meta.limit"))
			m.Conf("alpha", kit.Keys(lib, "meta.least"), m.Conf("alpha", "meta.least"))

			m.Cmd(ice.MDB_IMPORT, "alpha", lib, "list",
				m.Cmd(ice.WEB_CACHE, "catch", "csv", arg[0]+".csv").Append("data"))

			// 保存词库
			m.Conf("alpha", kit.Keys(lib, "meta.limit"), 0)
			m.Conf("alpha", kit.Keys(lib, "meta.least"), 0)
			m.Echo("%s: %d", lib, m.Grow("alpha", lib, kit.Dict("word", " ")))
		}},
		"push": {Name: "push lib word text", Help: "添加词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf("alpha", kit.Keys(arg[0], "meta.limit"), 0)
			m.Conf("alpha", kit.Keys(arg[0], "meta.least"), 0)
			m.Echo("%s: %d", arg[0], m.Grow("alpha", arg[0], kit.Dict("word", arg[1], "translation", arg[2])))
		}},
		"list": {Name: "list [lib [offend [limit]]]", Help: "查看词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				kit.Fetch(m.Confv("alpha"), func(key string, value map[string]interface{}) {
					if key != "meta" {
						m.Push(key, value["meta"], []string{"key", "count"})
					}
				})
				return
			}

			lib := kit.Select("ecdict", arg, 0)
			m.Option("cache.offend", kit.Select("0", arg, 1))
			m.Option("cache.limit", kit.Select("10", arg, 2))
			m.Grows("alpha", lib, "", "", func(index int, value map[string]interface{}) {
				m.Push("", value, []string{"id", "word", "translation"})
			})
		}},
		"save": {Name: "save lib [filename]", Help: "导出词库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cache.offend", 0)
			m.Option("cache.limit", -2)
			m.Cmdy(ice.WEB_STORY, "watch", m.Cmdx(ice.MDB_EXPORT, "alpha", arg[0], "list"), arg[1:])
		}},

		"random": {Name: "random [count]", Help: "随机词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			count := kit.Int(m.Conf("alpha", "ecdict.meta.count")) + 1
			for i := 0; i < kit.Int(kit.Select("10", arg, 0)); i++ {
				m.Cmdy("list", "ecdict", count-rand.Intn(count), 1)
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

			// 搜索词汇
			field := map[string]int{}
			for i, k := range kit.Simple(m.Confv("alpha", "meta.field")) {
				field[k] = i
			}
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
