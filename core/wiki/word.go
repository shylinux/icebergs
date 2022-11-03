package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _word_show(m *ice.Message, name string, arg ...string) {
	m.SetResult()
	defer m.StatusTime()
	m.Option(TITLE, map[string]int{})
	m.Option(MENU, kit.Dict(mdb.LIST, kit.List()))
	m.Option(ice.MSG_ALIAS, m.Configv(mdb.ALIAS))
	m.Cmdy("ssh.source", name, kit.Dict(nfs.DIR_ROOT, _wiki_path(m)))
}

const WORD = "word"

func init() {
	Index.MergeCommands(ice.Commands{
		WORD: {Name: "word path=src/main.shy@key list play", Help: "笔记文档", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.PrefixKey())
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, ice.SRC_MAIN_SHY)
				WordAlias(m, NAVMENU, TITLE, NAVMENU)
				WordAlias(m, PREMENU, TITLE, PREMENU)
				WordAlias(m, CHAPTER, TITLE, CHAPTER)
				WordAlias(m, SECTION, TITLE, SECTION)
				WordAlias(m, ENDMENU, TITLE, ENDMENU)
				WordAlias(m, LABEL, CHART, LABEL)
				WordAlias(m, CHAIN, CHART, CHAIN)
				WordAlias(m, SEQUENCE, CHART, SEQUENCE)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.DIR_DEEP, ice.TRUE)
				for _, p := range []string{"src/", "usr/icebergs/", "usr/learning/", "usr/linux-story/", "usr/nginx-story/", "usr/golang-story/", "usr/redis-story/", "usr/mysql-story/"} {
					_wiki_list(m, p)
				}
			}}, "play": {Name: "play", Help: "演示"},
			ice.STORY: {Name: "story", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(arg[0], ice.RUN, arg[2:]) }},
		}, WikiAction("", nfs.SHY), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_DEEP, ice.TRUE); len(arg) == 0 {
				arg = append(arg, "src/")
			}
			if !_wiki_list(m, arg...) {
				_word_show(m, arg[0])
			}
		}},
	})
}
func WordAlias(m *ice.Message, cmd string, cmds ...string) {
	m.Conf(WORD, kit.Keym(mdb.ALIAS, cmd), cmds)
}
func WordAction(template string, arg ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: &ice.Action{Hand: func(m *ice.Message, args ...string) {
		if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
			cs[m.CommandKey()] = &ice.Config{Value: kit.Data()}
			ice.Info.Load(m, m.CommandKey())
		}
		m.Config(nfs.TEMPLATE, template)
		for i := 0; i < len(arg)-1; i += 2 {
			m.Config(kit.Format(arg[i]), arg[i+1])
		}
	}}}
}
