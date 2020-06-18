package alpha

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/wiki"
	"github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/task"

	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

func _alpha_find(m *ice.Message, method, word string) {
	// 搜索方法
	switch word = strings.TrimSpace(word); method {
	case LINE:
	case WORD:
		word = "," + word + "$"
	}

	// 搜索词汇
	msg := m.Cmd(cli.SYSTEM, "grep", "-rh", word, m.Conf(ALPHA, "meta.store"))
	msg.CSV(msg.Result(), kit.Simple(m.Confv(ALPHA, "meta.field"))...).Table(func(index int, line map[string]string, head []string) {
		if method == WORD && index == 0 {
			// 添加收藏
			m.Cmd(web.FAVOR, m.Conf(ALPHA, "meta.favor"), ALPHA, line["word"], line["translation"],
				"id", line["id"], "definition", line["definition"])
		}
		for _, k := range []string{"id", "word", "translation", "definition"} {
			// 输出词汇
			m.Push(k, line[k])
		}
	})
}
func _alpha_find2(m *ice.Message, method, word string) {
	p := path.Join(m.Conf(ALPHA, "meta.store"), ALPHA)
	if ls, e := ioutil.ReadDir(p); m.Assert(e) {
		args := []interface{}{}
		for _, v := range ls {
			args = append(args, v)
		}

		var mu sync.Mutex
		task.Sync(args, func(task *task.Task) error {
			info := task.Arg.(os.FileInfo)
			file := path.Join(p, info.Name())
			kit.CSV(file, 100000, func(index int, value map[string]string, head []string) {
				if value["word"] != word {
					return
				}

				mu.Lock()
				defer mu.Unlock()
				m.Push("word", value["word"])
				m.Push("translation", value["translation"])
				m.Push("definition", value["definition"])
			})
			return nil
		})
	}
}
func _alpha_load(m *ice.Message, file, name string) {
	// 清空数据
	meta := m.Confm(ALPHA, "meta")
	m.Assert(os.RemoveAll(path.Join(kit.Format(meta[kit.MDB_STORE]), name)))
	m.Conf(ALPHA, name, "")

	// 缓存配置
	m.Conf(ALPHA, kit.Keys(name, kit.MDB_META), kit.Dict(
		kit.MDB_STORE, meta[kit.MDB_STORE],
		kit.MDB_FSIZE, meta[kit.MDB_FSIZE],
		kit.MDB_LIMIT, meta[kit.MDB_LIMIT],
		kit.MDB_LEAST, meta[kit.MDB_LEAST],
	))

	m.Cmd(mdb.IMPORT, ALPHA, name, kit.MDB_LIST,
		m.Cmd(web.CACHE, "catch", "csv", file+".csv").Append(kit.MDB_DATA))

	// 保存词库
	m.Conf(ALPHA, kit.Keys(name, "meta.limit"), 0)
	m.Conf(ALPHA, kit.Keys(name, "meta.least"), 0)
	m.Echo("%s: %d", name, m.Grow(ALPHA, name, kit.Dict("word", " ")))
}

const ALPHA = "alpha"
const (
	WORD = "word"
	LINE = "line"
)

var Index = &ice.Context{Name: "alpha", Help: "英汉词典",
	Configs: map[string]*ice.Config{
		ALPHA: {Name: "alpha", Help: "英汉词典", Value: kit.Data(
			kit.MDB_STORE, "usr/export", kit.MDB_FSIZE, "2000000",
			kit.MDB_LIMIT, "50000", kit.MDB_LEAST, "1000",
			"repos", "word-dict", "local", "person",
			"field", []interface{}{"audio", "bnc", "collins", "definition", "detail", "exchange", "frq", "id", "oxford", "phonetic", "pos", "tag", "time", "translation", "word"},
			web.FAVOR, "alpha.word",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(ALPHA) }},

		"find": {Name: "find word=hi method auto", Help: "查找词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_alpha_find(m, kit.Select("word", arg, 1), arg[0])
		}},
		"find2": {Name: "find word=hi method auto", Help: "查找词汇", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_alpha_find2(m, kit.Select("word", arg, 1), arg[0])
		}},
		"load": {Name: "load file [name]", Help: "加载词库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if meta := m.Confm(ALPHA, "meta"); len(arg) == 0 {
				arg = append(arg, path.Join("usr", kit.Format(meta["repos"]), "ecdict"))
			}
			_alpha_load(m, arg[0], kit.Select(path.Base(arg[0]), arg, 1))
		}},
	},
}

func init() { wiki.Index.Register(Index, nil) }
