package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _word_show(m *ice.Message, name string, arg ...string) {
	m.SetResult()
	m.Option(TITLE, map[string]int{})
	m.Option(MENU, kit.Dict(mdb.LIST, kit.List()))
	m.Option(ice.MSG_ALIAS, m.Configv(mdb.ALIAS))
	m.Cmdy("ssh.source", name, kit.Dict(nfs.DIR_ROOT, _wiki_path(m, WORD)))
}

const WORD = "word"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		WORD: {Name: WORD, Help: "笔记文档", Value: kit.Data(
			nfs.PATH, "", lex.REGEXP, ".*\\.shy", mdb.ALIAS, kit.Dict(
				NAVMENU, kit.List(TITLE, NAVMENU),
				PREMENU, kit.List(TITLE, PREMENU),
				CHAPTER, kit.List(TITLE, CHAPTER),
				SECTION, kit.List(TITLE, SECTION),
				ENDMENU, kit.List(TITLE, ENDMENU),
				LABEL, kit.List(CHART, LABEL),
				CHAIN, kit.List(CHART, CHAIN),
				SEQUENCE, kit.List(CHART, SEQUENCE),
			),
			mdb.SHORT, "type,name,text",
			mdb.FIELD, "time,hash,type,name,text",
		)},
	}, Commands: ice.Commands{
		WORD: {Name: "word path=src/main.shy@key list play", Help: "笔记文档", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.PrefixKey())
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, ice.SRC_MAIN_SHY)
			}},
			mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if (arg[0] != mdb.FOREACH && arg[0] != m.CommandKey()) || arg[1] == "" {
					return
				}
				if arg[1] == "" {
					m.PushSearch(mdb.TYPE, nfs.SHY, mdb.NAME, ice.SRC_MAIN_SHY, mdb.TEXT, web.MergePodCmd(m, "", ""))
				}

				m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, func(value ice.Maps) {
					if arg[1] == "" {
						if value[mdb.TYPE] == SPARK {
							value[mdb.TEXT] = ice.Render(m, ice.RENDER_SCRIPT, value[mdb.TEXT])
						}
						m.PushSearch(value)
					}
				})
				m.Cmd("", mdb.INPUTS).Tables(func(value ice.Maps) {
					if strings.Contains(value[nfs.PATH], arg[1]) {
						m.PushSearch(mdb.TYPE, "shy", mdb.NAME, value[nfs.PATH], value)
					}
				})
			}},
			mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, arg)
			}},
			"recent": {Name: "recent", Help: "最近", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(mdb.FIELD))
				m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.HASH).Table(func(index int, value ice.Maps, head []string) {
					if value[mdb.TYPE] == "spark" {
						value[mdb.TEXT] = ice.Render(m, ice.RENDER_SCRIPT, value[mdb.TEXT])
					}
					m.Push("", value, head)
				})
				m.PushAction(mdb.REMOVE)
			}},

			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				for _, p := range []string{"src/", "src/help/", "usr/icebergs/", "usr/linux-story/", "usr/nginx-story/", "usr/golang-story/", "usr/redis-story/", "usr/mysql-story/"} {
					m.Cmdy(nfs.DIR, p, kit.Dict(nfs.DIR_DEEP, ice.TRUE, nfs.DIR_REG, ".*\\.shy"), nfs.DIR_CLI_FIELDS)
				}
			}},
			ice.STORY: {Name: "story", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(arg[0], ice.RUN, arg[2:])
			}},
			ice.PLAY: {Name: "play", Help: "演示"},
		}, ctx.CmdAction(), mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "src/")
			}
			m.Option(nfs.DIR_REG, m.Config(lex.REGEXP))
			if m.Option(nfs.DIR_DEEP, ice.TRUE); !_wiki_list(m, m.CommandKey(), arg...) {
				if !nfs.ExistsFile(m, arg[0]) && nfs.ExistsFile(m, path.Join(ice.SRC, arg[0])) {
					arg[0] = path.Join(ice.SRC, arg[0])
				}
				ctx.DisplayLocal(m, "")
				_word_show(m, arg[0])
				m.StatusTime()
			}
		}},
	}})
}
