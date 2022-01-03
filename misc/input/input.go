package input

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _input_find(m *ice.Message, method, word, limit string) {
	switch method {
	case LINE:
	case WORD:
		word = "^" + word + ","
	}

	// 搜索词汇
	res := m.Cmdx(cli.SYSTEM, "grep", "-rn", word, m.Config(mdb.STORE))
	bio := csv.NewReader(bytes.NewBufferString(strings.Replace(res, ":", ",", -1)))

	for i := 0; i < kit.Int(limit); i++ {
		if line, e := bio.Read(); e != nil {
			break
		} else if len(line) < 3 {

		} else { // 输出词汇
			m.Push(mdb.ID, line[3])
			m.Push(CODE, line[2])
			m.Push(TEXT, line[4])
			m.Push(WEIGHT, line[6])
		}

	}
	m.SortIntR(WEIGHT)
}
func _input_load(m *ice.Message, file string, libs ...string) {
	if f, e := os.Open(file); m.Assert(e) {
		defer f.Close()

		// 清空数据
		lib := kit.Select(path.Base(file), libs, 0)
		m.Assert(os.RemoveAll(path.Join(m.Config(mdb.STORE), lib)))
		m.Cmd(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, mdb.ZONE, lib)
		prefix := kit.Keys(mdb.HASH, m.Rich(m.PrefixKey(), "", kit.Data(
			mdb.STORE, path.Join(m.Config(mdb.STORE), lib),
			m.ConfigSimple(mdb.FSIZE, mdb.LIMIT, mdb.LEAST),
			mdb.ZONE, lib, mdb.COUNT, 0,
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
			m.Grow(m.PrefixKey(), prefix, kit.Dict(TEXT, line[0], CODE, line[1], WEIGHT, kit.Select("999999", line, 2)))
		}

		// 保存词库
		m.Conf(m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LIMIT)), 0)
		m.Conf(m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LEAST)), 0)
		n := m.Grow(m.PrefixKey(), prefix, kit.Dict(TEXT, "成功", CODE, "z", WEIGHT, "0"))
		m.Log_IMPORT(m.PrefixKey(), lib, mdb.COUNT, n)
		m.Echo("%s: %d", lib, n)
	}
}
func _input_push(m *ice.Message, lib, text, code, weight string) {
	if m.Richs(m.PrefixKey(), "", lib, nil) == nil {
		m.Rich(m.PrefixKey(), "", kit.Data(
			mdb.STORE, path.Join(m.Config(mdb.STORE), lib),
			mdb.FSIZE, m.Config(mdb.FSIZE),
			mdb.LIMIT, m.Config(mdb.LIMIT),
			mdb.LEAST, m.Config(mdb.LEAST),
			mdb.ZONE, lib,
		))
	}

	m.Richs(m.PrefixKey(), "", lib, func(key string, value map[string]interface{}) {
		prefix := kit.Keys(mdb.HASH, key)
		m.Conf(m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LIMIT)), 0)
		m.Conf(m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LEAST)), 0)
		n := m.Grow(m.PrefixKey(), prefix, kit.Dict(TEXT, text, CODE, code, WEIGHT, weight))
		m.Log_IMPORT(CODE, code, TEXT, text)
		m.Echo("%s: %d", lib, n)
	})
}
func _input_save(m *ice.Message, file string, lib ...string) {
	if f, p, e := kit.Create(file); m.Assert(e) {
		defer f.Close()
		n := 0
		m.Option(ice.CACHE_LIMIT, -2)
		for _, lib := range lib {
			m.Richs(m.PrefixKey(), "", lib, func(key string, value map[string]interface{}) {
				m.Grows(m.PrefixKey(), kit.Keys(mdb.HASH, key), "", "", func(index int, value map[string]interface{}) {
					if value[CODE] != "z" {
						fmt.Fprintf(f, "%s %s %s\n", value[TEXT], value[CODE], value[WEIGHT])
						n++
					}
				})
			})
		}
		m.Log_EXPORT(FILE, p, mdb.COUNT, n)
		m.Echo("%s: %d", p, n)
	}
}

func _input_list(m *ice.Message, lib string) {
	if lib == "" {
		m.Richs(m.PrefixKey(), "", mdb.FOREACH, func(key string, value map[string]interface{}) {
			m.Push("", kit.GetMeta(value), kit.Split("time,zone,count,store"))
		})
		return
	}

	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_TYPE, nfs.CAT)
	m.Richs(m.PrefixKey(), "", lib, func(key string, value map[string]interface{}) {
		m.Cmdy(nfs.DIR, kit.Value(value, kit.Keym(mdb.STORE)), "time size line path")
	})
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
const INPUT = "input"

var Index = &ice.Context{Name: INPUT, Help: "输入法"}

func init() { code.Index.Register(Index, nil) }
