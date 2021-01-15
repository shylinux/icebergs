package alpha

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

func _alpha_find(m *ice.Message, method, word string) {
	if word == "" {
		return
	}

	// 搜索方法
	switch word = strings.TrimSpace(word); method {
	case "line":
	case "word":
		word = "," + word + "$"
	}

	// 搜索词汇
	msg := m.Cmd(cli.SYSTEM, "grep", "-rh", word, m.Conf(ALPHA, "meta.store"))
	msg.CSV(msg.Result(), kit.Simple(m.Confv(ALPHA, "meta.field"))...).Table(func(index int, value map[string]string, head []string) {
		if value["word"] == "" {
			return
		}
		m.PushSearch("cmd", ALPHA, "type", method, "name", value["word"], "text", value["translation"], value)
	})
	return
}
func _alpha_load(m *ice.Message, file, name string) {
	// 清空数据
	meta := m.Confm(ALPHA, kit.MDB_META)
	m.Assert(os.RemoveAll(path.Join(kit.Format(meta[kit.MDB_STORE]), name)))
	m.Conf(ALPHA, name, "")

	// 缓存配置
	m.Conf(ALPHA, kit.Keys(name, kit.MDB_META), kit.Dict(
		kit.MDB_STORE, meta[kit.MDB_STORE],
		kit.MDB_FSIZE, meta[kit.MDB_FSIZE],
		kit.MDB_LIMIT, meta[kit.MDB_LIMIT],
		kit.MDB_LEAST, meta[kit.MDB_LEAST],
		kit.MDB_FIELD, meta[kit.MDB_FIELD],
	))

	m.Cmd(mdb.IMPORT, ALPHA, name, kit.MDB_LIST, file)

	// 保存词库
	m.Conf(ALPHA, kit.Keys(name, "meta.limit"), 0)
	m.Conf(ALPHA, kit.Keys(name, "meta.least"), 0)
	m.Echo("%s: %d", name, m.Grow(ALPHA, name, kit.Dict("word", " ")))
}

const ALPHA = "alpha"

var Index = &ice.Context{Name: ALPHA, Help: "英汉词典",
	Configs: map[string]*ice.Config{
		ALPHA: {Name: ALPHA, Help: "英汉词典", Value: kit.Data(
			kit.MDB_LIMIT, "50000", kit.MDB_LEAST, "1000",
			kit.MDB_STORE, "usr/export/alpha", kit.MDB_FSIZE, "2000000",
			kit.SSH_REPOS, "word-dict", kit.MDB_FIELD, []interface{}{"audio", "bnc", "collins", "definition", "detail", "exchange", "frq", "id", "oxford", "phonetic", "pos", "tag", "time", "translation", "word"},
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(mdb.SEARCH, mdb.CREATE, ALPHA, m.Prefix(ALPHA))
			m.Load()
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		ALPHA: {Name: "alpha method=word,line word auto", Help: "英汉", Action: map[string]*ice.Action{
			mdb.IMPORT: {Name: "import file=usr/word-dict/ecdict name", Help: "加载词库", Hand: func(m *ice.Message, arg ...string) {
				_alpha_load(m, m.Option(kit.MDB_FILE), kit.Select(path.Base(m.Option(kit.MDB_FILE)), m.Option(kit.MDB_NAME)))
			}},
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == ALPHA {
					_alpha_find(m, kit.Select("word", arg, 2), arg[1])
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(mdb.FIELDS, "id,word,translation,definition")
			_alpha_find(m, arg[0], arg[1])
		}},
	},
}

func init() { wiki.Index.Register(Index, nil) }
