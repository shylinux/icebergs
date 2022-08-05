package input

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	TEXT   = "text"
	CODE   = "code"
	WEIGHT = "weight"
)
const (
	WORD = "word"
	LINE = "line"
)

type input struct {
	insert string `name:"insert zone=person text code weight" help:"添加"`
	load   string `name:"load file=usr/wubi-dict/wubi86 zone=wubi86" help:"加载"`
	save   string `name:"save file=usr/wubi-dict/person zone=person" help:"保存"`
	list   string `name:"list method code auto" help:"输入法"`
}

func (i input) Load(m *ice.Message, arg ...string) {
	if f, e := os.Open(m.Option(nfs.FILE)); m.Assert(e) {
		defer f.Close()

		// 清空数据
		lib := kit.Select(path.Base(m.Option(nfs.FILE)), m.Option(mdb.ZONE))
		m.Assert(os.RemoveAll(path.Join(m.Config(mdb.STORE), lib)))
		m.Cmd(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, mdb.ZONE, lib)
		prefix := kit.Keys(mdb.HASH, mdb.Rich(m.Message, m.PrefixKey(), "", kit.Data(
			mdb.STORE, path.Join(m.Config(mdb.STORE), lib),
			m.ConfigSimple(mdb.FSIZE, mdb.LIMIT, mdb.LEAST),
			mdb.ZONE, lib, mdb.COUNT, 0,
		)))

		// 加载词库
		for bio := bufio.NewScanner(f); bio.Scan(); {
			if strings.HasPrefix(bio.Text(), "# ") {
				continue
			}
			line := kit.Split(bio.Text())
			if len(line) < 2 || (len(line) > 2 && line[2] == "0") {
				continue
			}
			mdb.Grow(m.Message, m.PrefixKey(), prefix, kit.Dict(TEXT, line[0], CODE, line[1], WEIGHT, kit.Select("999999", line, 2)))
		}

		// 保存词库
		m.Conf(m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LIMIT)), 0)
		m.Conf(m.PrefixKey(), kit.Keys(prefix, kit.Keym(mdb.LEAST)), 0)
		n := mdb.Grow(m.Message, m.PrefixKey(), prefix, kit.Dict(TEXT, "成功", CODE, "z", WEIGHT, "0"))
		m.Logs(mdb.IMPORT, m.PrefixKey(), lib, mdb.COUNT, n)
		m.Echo("%s: %d", lib, n)
	}
}
func (i input) Save(m *ice.Message, arg ...string) {
	if f, p, e := kit.Create(m.Option(nfs.FILE)); m.Assert(e) {
		defer f.Close()
		n := 0
		m.Option(mdb.CACHE_LIMIT, -2)
		for _, lib := range kit.Split(m.Option(mdb.ZONE)) {
			mdb.Richs(m.Message, m.PrefixKey(), "", lib, func(key string, value ice.Map) {
				mdb.Grows(m.Message, m.PrefixKey(), kit.Keys(mdb.HASH, key), "", "", func(index int, value ice.Map) {
					if value[CODE] != "z" {
						fmt.Fprintf(f, "%s %s %s\n", value[TEXT], value[CODE], value[WEIGHT])
						n++
					}
				})
			})
		}
		m.Logs(mdb.EXPORT, nfs.FILE, p, mdb.COUNT, n)
		m.Echo("%s: %d", p, n)
	}
}
func (i input) List(m *ice.Message, arg ...string) {
	if len(arg) < 2 || arg[1] == "" {
		return
	}
	switch arg[0] {
	case LINE:
	case WORD:
		arg[1] = "^" + arg[1] + ice.FS
	}

	// 搜索词汇
	res := m.Cmdx(cli.SYSTEM, "grep", "-rn", arg[1], m.Config(mdb.STORE))
	bio := csv.NewReader(bytes.NewBufferString(strings.Replace(res, ":", ",", -1)))

	for i := 0; i < kit.Int(10); i++ {
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
	m.StatusTimeCount()
}
