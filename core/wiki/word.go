package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _word_show(m *ice.Message, name string, arg ...string) {
	m.Set(ice.MSG_RESULT)
	m.Option(TITLE, map[string]int{})
	m.Option(kit.MDB_MENU, kit.Dict(kit.MDB_LIST, []interface{}{}))

	m.Option(ice.MSG_ALIAS, m.Confv(WORD, kit.Keym(kit.MDB_ALIAS)))
	m.Option(nfs.DIR_ROOT, _wiki_path(m, WORD))
	m.Cmdy(ssh.SOURCE, name)
}

const WORD = "word"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		WORD: {Name: WORD, Help: "语言文字", Value: kit.Data(
			kit.MDB_PATH, "", kit.MDB_REGEXP, ".*\\.shy", kit.MDB_ALIAS, kit.Dict(
				NAVMENU, []interface{}{TITLE, NAVMENU},
				PREMENU, []interface{}{TITLE, PREMENU},
				CHAPTER, []interface{}{TITLE, CHAPTER},
				SECTION, []interface{}{TITLE, SECTION},
				ENDMENU, []interface{}{TITLE, ENDMENU},
				LABEL, []interface{}{CHART, LABEL},
				CHAIN, []interface{}{CHART, CHAIN},
			),
		)},
	}, Commands: map[string]*ice.Command{
		WORD: {Name: "word path=src/main.shy auto play", Help: "语言文字", Meta: kit.Dict(
			ice.Display("/plugin/local/wiki/word.js"),
		), Action: ice.MergeAction(map[string]*ice.Action{
			web.STORY: {Name: "story", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(arg[0], ctx.ACTION, ice.RUN, arg[2:])
			}},
			"play": {Name: "play", Help: "演示"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(nfs.DIR_REG, m.Config(kit.MDB_REGEXP))
			if m.Option(nfs.DIR_DEEP, ice.TRUE); !_wiki_list(m, cmd, arg...) {
				_word_show(m, arg[0])
			}
		}},
	}})
}
