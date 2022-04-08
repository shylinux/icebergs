package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _word_show(m *ice.Message, name string, arg ...string) {
	m.SetResult()
	m.Option(TITLE, map[string]int{})
	m.Option(MENU, kit.Dict(mdb.LIST, kit.List()))
	m.Option(ice.MSG_ALIAS, m.Configv(mdb.ALIAS))
	m.Cmdy(ssh.SOURCE, name, kit.Dict(nfs.DIR_ROOT, _wiki_path(m, WORD)))
}

const WORD = "word"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		WORD: {Name: WORD, Help: "语言文字", Value: kit.Data(
			nfs.PATH, "", lex.REGEXP, ".*\\.shy", mdb.ALIAS, kit.Dict(
				NAVMENU, kit.List(TITLE, NAVMENU),
				PREMENU, kit.List(TITLE, PREMENU),
				CHAPTER, kit.List(TITLE, CHAPTER),
				SECTION, kit.List(TITLE, SECTION),
				ENDMENU, kit.List(TITLE, ENDMENU),
				LABEL, kit.List(CHART, LABEL),
				CHAIN, kit.List(CHART, CHAIN),
			),
			mdb.SHORT, "type,name,text",
			mdb.FIELD, "time,hash,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		WORD: {Name: "word path=src/main.shy@key auto play", Help: "语言文字", Meta: kit.Dict(ice.DisplayLocal("")), Action: ice.MergeAction(map[string]*ice.Action{
			mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				m.PushSearch(mdb.TYPE, "shy", mdb.NAME, "src/main.shy", mdb.TEXT, m.MergeURL2("/chat/cmd/web.wiki.word"))
				m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.HASH).Table(func(index int, value map[string]string, head []string) {
					if arg[1] == "" {
						if value[mdb.TYPE] == SPARK {
							value[mdb.TEXT] = ice.Render(m, ice.RENDER_SCRIPT, value[mdb.TEXT])
						}
						m.PushSearch(value)
					}
				})
			}},
			mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, arg)
			}},
			"recent": {Name: "recent", Help: "最近", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(mdb.FIELD))
				m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.HASH).Table(func(index int, value map[string]string, head []string) {
					if value[mdb.TYPE] == "spark" {
						value[mdb.TEXT] = ice.Render(m, ice.RENDER_SCRIPT, value[mdb.TEXT])
					}
					m.Push("", value, head)
				})
				m.PushAction(mdb.REMOVE)
			}},

			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, "src/", kit.Dict(nfs.DIR_DEEP, ice.TRUE, nfs.DIR_REG, ".*\\.shy"), nfs.DIR_CLI_FIELDS)
				m.Cmdy(nfs.DIR, "src/help/", kit.Dict(nfs.DIR_DEEP, ice.TRUE, nfs.DIR_REG, ".*\\.shy"), nfs.DIR_CLI_FIELDS)
			}},
			web.STORY: {Name: "story", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(arg[0], ctx.ACTION, ice.RUN, arg[2:])
			}},
			ice.PLAY: {Name: "play", Help: "演示"},
		}, ctx.CmdAction(), mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(nfs.DIR_REG, m.Config(lex.REGEXP))
			if m.Option(nfs.DIR_DEEP, ice.TRUE); !_wiki_list(m, cmd, arg...) {
				_word_show(m, arg[0])
			}
		}},
	}})
}
