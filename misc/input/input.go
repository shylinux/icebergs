package input

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strings"
)

func _input_list(m *ice.Message, lib string) {
	if lib == "" {
		m.Richs(INPUT, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			m.Push(kit.MDB_TIME, kit.Value(value, "meta.time"))
			m.Push(kit.MDB_ZONE, kit.Value(value, "meta.zone"))
			m.Push(kit.MDB_COUNT, kit.Value(value, "meta.count"))
			m.Push(kit.MDB_STORE, kit.Value(value, "meta.store"))
		})
		return
	}

	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_TYPE, nfs.CAT)
	m.Richs(INPUT, "", lib, func(key string, value map[string]interface{}) {
		m.Cmdy(nfs.DIR, kit.Value(value, "meta.store"), "time size line path")
	})
}
func _input_push(m *ice.Message, lib, text, code, weight string) {
	if m.Richs(INPUT, "", lib, nil) == nil {
		m.Rich(INPUT, "", kit.Data(
			kit.MDB_STORE, path.Join(m.Conf(INPUT, "meta.store"), lib),
			kit.MDB_FSIZE, m.Conf(INPUT, "meta.fsize"),
			kit.MDB_LIMIT, m.Conf(INPUT, "meta.limit"),
			kit.MDB_LEAST, m.Conf(INPUT, "meta.least"),
			kit.MDB_ZONE, lib,
		))
	}

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
			// 输出词汇
			// m.Push(FILE, path.Base(line[0]))
			m.Push(kit.MDB_ID, line[3])
			m.Push(CODE, line[2])
			m.Push(TEXT, line[4])
			m.Push(WEIGHT, line[6])
		}
	}
	m.Sort(WEIGHT, "int_r")
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
			kit.MDB_ZONE, lib,
		)))

		// 加载词库
		for bio := bufio.NewScanner(f); bio.Scan(); {
			if strings.HasPrefix(bio.Text(), "#") {
				continue
			}
			line := kit.Split(bio.Text())
			if len(line) < 2 || (len(line) > 2 && line[2] == "0") {
				continue
			}
			m.Grow(INPUT, prefix, kit.Dict(TEXT, line[0], CODE, line[1], WEIGHT, kit.Select("999999", line, 2)))
		}

		// 保存词库
		m.Conf(INPUT, kit.Keys(prefix, "meta.limit"), 0)
		m.Conf(INPUT, kit.Keys(prefix, "meta.least"), 0)
		n := m.Grow(INPUT, prefix, kit.Dict(TEXT, "成功", CODE, "z", WEIGHT, "0"))
		m.Log_IMPORT(INPUT, lib, kit.MDB_COUNT, n)
		m.Echo("%s: %d", lib, n)
	}
}

const (
	ZONE   = "zone"
	FILE   = "file"
	CODE   = "code"
	TEXT   = "text"
	WEIGHT = "weight"
)
const (
	WORD = "word"
	LINE = "line"
)
const (
	WUBI = "wubi"
)
const INPUT = "input"

var Index = &ice.Context{Name: INPUT, Help: "输入法",
	Configs: map[string]*ice.Config{
		INPUT: {Name: INPUT, Help: "输入法", Value: kit.Data(
			kit.MDB_STORE, "usr/local/export/input", kit.MDB_FSIZE, "200000",
			kit.MDB_LIMIT, "5000", kit.MDB_LEAST, "1000",
			kit.MDB_SHORT, "zone", "repos", "wubi-dict",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		WUBI: {Name: "wubi method=word,line code auto", Help: "五笔", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=person text code weight", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_input_push(m, kit.Select("person", m.Option(ZONE)), m.Option(TEXT), m.Option(CODE), m.Option(WEIGHT))
			}},
			mdb.EXPORT: {Name: "export file=usr/wubi-dict/person zone=person", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				// _input_save(m, kit.Select("usr/wubi-dict/person", m.Option("file")), m.Option("zone"))
			}},
			mdb.IMPORT: {Name: "import file=usr/wubi-dict/person zone=", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				_input_load(m, kit.Select("usr/wubi-dict/person", m.Option(FILE)), m.Option(ZONE))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_input_find(m, arg[0], arg[1], m.Option("cache.limit"))
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
