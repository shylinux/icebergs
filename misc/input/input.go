package input

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/task"

	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

func _input_list(m *ice.Message, lib string) {
	if lib == "" {
		m.Richs(INPUT, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			m.Push(key, value[kit.MDB_META], []string{kit.MDB_ZONE, kit.MDB_COUNT, kit.MDB_STORE})
		})
		return
	}

	m.Richs(INPUT, "", lib, func(key string, value map[string]interface{}) {
		m.Grows(INPUT, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			m.Push(key, value, []string{kit.MDB_ID, CODE, TEXT, WEIGHT})
		})
	})
}
func _input_push(m *ice.Message, lib, text, code, weight string) {
	m.Richs(INPUT, "", lib, func(key string, value map[string]interface{}) {
		prefix := kit.Keys(kit.MDB_HASH, key)
		m.Conf(INPUT, kit.Keys(prefix, "meta.limit"), 0)
		m.Conf(INPUT, kit.Keys(prefix, "meta.least"), 0)
		n := m.Grow(INPUT, prefix, kit.Dict(TEXT, text, CODE, code, WEIGHT, weight))
		m.Log_IMPORT(CODE, code, TEXT, text)
		m.Echo("%s: %d", lib, n)
	})
}
func _input_find(m *ice.Message, method, word, limit string) {
	// 搜索方法
	switch method {
	case LINE:
	case WORD:
		word = "^" + word + ","
	}

	// 搜索词汇
	res := m.Cmdx(cli.SYSTEM, "grep", "-rn", word, m.Conf(INPUT, "meta.store"))
	bio := csv.NewReader(bytes.NewBufferString(strings.Replace(res, ":", ",", -1)))

	for i := 0; i < kit.Int(limit); i++ {
		if line, e := bio.Read(); e != nil {
			break
		} else if len(line) < 3 {

		} else {
			if method == WORD && i == 0 {
				// 添加收藏
				// web.FavorInsert(m.Spawn(), "input.word", "input", line[2], line[4], "id", line[3], WEIGHT, line[6])
			}

			// 输出词汇
			m.Push(FILE, path.Base(line[0]))
			m.Push(kit.MDB_ID, line[3])
			m.Push(CODE, line[2])
			m.Push(TEXT, line[4])
			m.Push(WEIGHT, line[6])
		}
	}
	m.Sort(WEIGHT, "int_r")
}
func _input_find2(m *ice.Message, method, word, limit string) {
	list := []interface{}{}
	files := map[string]bool{}
	m.Richs(INPUT, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		kit.Fetch(kit.Value(value, "meta.record"), func(index int, value map[string]interface{}) {
			file := value["file"].(string)
			if _, ok := files[file]; ok {
				list = append(list, file)
			} else {
				files[file] = true
			}
		})
	})
	defer m.Cost("some")

	var mu sync.Mutex
	task.Sync(list, func(task *task.Task, lock *task.Lock) error {
		kit.CSV(kit.Format(task.Arg), 100000, func(index int, value map[string]string, head []string) {
			if value["code"] != word {
				return
			}
			mu.Lock()
			defer mu.Unlock()

			m.Push(FILE, task.Arg)
			m.Push(kit.MDB_ID, value[kit.MDB_ID])
			m.Push(CODE, value["code"])
			m.Push(TEXT, value["text"])
			m.Push(WEIGHT, value["weight"])
			m.Push(kit.MDB_TIME, value["time"])
		})
		return nil
	})
}
func _input_save(m *ice.Message, file string, lib ...string) {
	if f, p, e := kit.Create(file); m.Assert(e) {
		defer f.Close()
		n := 0
		m.Option("cache.limit", -2)
		for _, lib := range lib {
			m.Richs(INPUT, "", lib, func(key string, value map[string]interface{}) {
				m.Grows(INPUT, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
					if value[CODE] != "z" {
						fmt.Fprintf(f, "%s %s %s\n", value[TEXT], value[CODE], value[WEIGHT])
						n++
					}
				})
			})
		}
		m.Log_EXPORT(FILE, p, kit.MDB_COUNT, n)
		m.Echo("%s: %d", p, n)
	}
}
func _input_load(m *ice.Message, file string, libs ...string) {
	if f, e := os.Open(file); m.Assert(e) {
		defer f.Close()

		// 清空数据
		lib := kit.Select(path.Base(file), libs, 0)
		m.Assert(os.RemoveAll(path.Join(m.Conf(INPUT, "meta.store"), lib)))
		prefix := kit.Keys(kit.MDB_HASH, m.Rich(INPUT, "", kit.Data(
			kit.MDB_STORE, path.Join(m.Conf(INPUT, "meta.store"), lib),
			kit.MDB_FSIZE, m.Conf(INPUT, "meta.fsize"),
			kit.MDB_LIMIT, m.Conf(INPUT, "meta.limit"),
			kit.MDB_LEAST, m.Conf(INPUT, "meta.least"),
			"zone", lib,
		)))

		// 缓存配置

		// 加载词库
		for bio := bufio.NewScanner(f); bio.Scan(); {
			if strings.HasPrefix(bio.Text(), "#") {
				continue
			}
			line := kit.Split(bio.Text())
			if len(line) < 3 || line[2] == "0" {
				continue
			}
			m.Grow(INPUT, prefix, kit.Dict(TEXT, line[0], CODE, line[1], WEIGHT, line[2]))
		}

		// 保存词库
		m.Conf(INPUT, kit.Keys(prefix, "meta.limit"), 0)
		m.Conf(INPUT, kit.Keys(prefix, "meta.least"), 0)
		n := m.Grow(INPUT, prefix, kit.Dict(TEXT, "成功", CODE, "z", WEIGHT, "0"))
		m.Log_IMPORT(INPUT, lib, kit.MDB_COUNT, n)
		m.Echo("%s: %d", lib, n)
	}
}

const INPUT = "input"
const (
	WORD = "word"
	LINE = "line"
)
const (
	FILE   = "file"
	CODE   = "code"
	TEXT   = "text"
	WEIGHT = "weight"
)

var Index = &ice.Context{Name: "input", Help: "输入法",
	Configs: map[string]*ice.Config{
		INPUT: {Name: "input", Help: "输入法", Value: kit.Data(
			"repos", "wubi-dict", "local", "person",
			kit.MDB_STORE, "usr/export/input", kit.MDB_FSIZE, "200000",
			kit.MDB_LIMIT, "5000", kit.MDB_LEAST, "1000",
			kit.MDB_SHORT, "zone",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(INPUT) }},

		"list": {Name: "list [lib]", Help: "查看词库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_input_list(m, kit.Select("", arg, 0))
		}},
		"push": {Name: "push lib text code [weight]", Help: "添加字码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_input_push(m, arg[0], arg[1], arg[2], kit.Select("90919495", arg, 3))
		}},
		"find": {Name: "find key [word|line [limit]]", Help: "查找字码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				web.FavorList(m, "input.word", "")
				return
			}
			_input_find(m, kit.Select(WORD, arg, 1), arg[0], kit.Select("100", arg, 2))
		}},
		"find2": {Name: "find2 key [word|line [limit]]", Help: "查找字码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_input_find2(m, kit.Select(WORD, arg, 1), arg[0], kit.Select("100", arg, 2))
		}},

		"save": {Name: "save file lib...", Help: "导出词库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_input_save(m, arg[0], arg[1:]...)
		}},
		"load": {Name: "load file lib", Help: "导入词库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_input_load(m, kit.Select("usr/wubi-dict/wubi86", arg, 0))
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
