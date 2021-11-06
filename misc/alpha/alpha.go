package alpha

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

func _alpha_find(m *ice.Message, method, word string) {
	if word == "" {
		return
	}

	// 搜索方法
	switch word = strings.TrimSpace(word); method {
	case LINE:
	case WORD:
		word = "^" + word + ","
	}

	// 搜索词汇
	msg := m.Cmd(cli.SYSTEM, "grep", "-rh", word, m.Config(kit.MDB_STORE))
	msg.CSV(msg.Result(), kit.Split(m.Config(kit.MDB_FIELD))...).Table(func(index int, value map[string]string, head []string) {
		m.PushSearch(ice.CMD, ALPHA, kit.MDB_TYPE, method, kit.MDB_NAME, value[WORD], kit.MDB_TEXT, value["translation"], value)
	})
}
func _alpha_load(m *ice.Message, file, name string) {
	// 清空数据
	meta := m.Confm(ALPHA, kit.MDB_META)
	m.Assert(os.RemoveAll(path.Join(kit.Format(meta[kit.MDB_STORE]), name)))
	m.Conf(ALPHA, name, "")

	// 缓存配置
	m.Conf(ALPHA, kit.Keys(name, kit.MDB_META), kit.Dict(meta))
	m.Cmd(mdb.IMPORT, ALPHA, name, kit.MDB_LIST, file)

	// 保存词库
	m.Conf(ALPHA, kit.Keys(name, kit.Keym(kit.MDB_LIMIT)), 0)
	m.Conf(ALPHA, kit.Keys(name, kit.Keym(kit.MDB_LEAST)), 0)
	m.Echo("%s: %d", name, m.Grow(ALPHA, name, kit.Dict(WORD, ice.SP)))
}

const (
	WORD = "word"
	LINE = "line"
)

const ALPHA = "alpha"

var Index = &ice.Context{Name: ALPHA, Help: "英汉词典", Configs: map[string]*ice.Config{
	ALPHA: {Name: ALPHA, Help: "英汉词典", Value: kit.Data(
		kit.SSH_REPOS, "word-dict", kit.MDB_FIELD, "word,translation,definition",
		kit.MDB_STORE, path.Join(ice.USR_LOCAL_EXPORT, ALPHA), kit.MDB_FSIZE, "300000",
		kit.MDB_LIMIT, "50000", kit.MDB_LEAST, "1000",
	)},
}, Commands: map[string]*ice.Command{
	ALPHA: {Name: "alpha method=word,line word auto", Help: "英汉", Action: map[string]*ice.Action{
		mdb.IMPORT: {Name: "import file=usr/word-dict/ecdict name=ecdict", Help: "加载词库", Hand: func(m *ice.Message, arg ...string) {
			_alpha_load(m, m.Option(kit.MDB_FILE), kit.Select(path.Base(m.Option(kit.MDB_FILE)), m.Option(kit.MDB_NAME)))
		}},
		mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
			if arg[0] == ALPHA {
				_alpha_find(m, kit.Select(WORD, arg, 2), arg[1])
			}
		}},
	}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if len(arg) < 2 {
			return
		}
		defer m.StatusTimeCountTotal(m.Config(kit.MDB_COUNT))
		if arg[0] == WORD {
			if msg := m.Cmd(CACHE, arg[1]); msg.Length() > 0 {
				m.Copy(msg)
				return
			}
		}
		m.OptionFields(m.Config(kit.MDB_FIELD))
		if _alpha_find(m, arg[0], arg[1]); arg[0] == WORD && m.Length() > 0 {
			m.Cmd(CACHE, mdb.CREATE, m.AppendSimple())
		}
	}},
}}

func init() { wiki.Index.Register(Index, nil) }
