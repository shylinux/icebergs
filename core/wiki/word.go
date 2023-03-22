package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/misc/git"
	kit "shylinux.com/x/toolkits"
)

func _word_show(m *ice.Message, name string, arg ...string) {
	m.Options(ice.MSG_ALIAS, mdb.Configv(m, mdb.ALIAS), TITLE, map[string]int{}, MENU, kit.Dict(mdb.LIST, kit.List()))
	m.Cmdy(ssh.SOURCE, name, kit.Dict(nfs.DIR_ROOT, _wiki_path(m)))
}

const WORD = "word"

func init() {
	Index.MergeCommands(ice.Commands{
		WORD: {Name: "word path=src/main.shy@key list play", Help: "笔记文档", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				WordAlias(m, NAVMENU, TITLE, NAVMENU)
				WordAlias(m, PREMENU, TITLE, PREMENU)
				WordAlias(m, CHAPTER, TITLE, CHAPTER)
				WordAlias(m, SECTION, TITLE, SECTION)
				WordAlias(m, ENDMENU, TITLE, ENDMENU)
				WordAlias(m, SHELL, SPARK, SHELL)
				WordAlias(m, LABEL, CHART, LABEL)
				WordAlias(m, CHAIN, CHART, CHAIN)
				WordAlias(m, SEQUENCE, CHART, SEQUENCE)
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(git.REPOS, ice.OptionFields(nfs.PATH)).Tables(func(value ice.Maps) {
					if m.Option(nfs.DIR_DEEP, ice.TRUE); kit.Path(value[nfs.PATH]) == kit.Path("") {
						_wiki_list(m, "src/")
					} else {
						_wiki_list(m, value[nfs.PATH])
					}
				})
				m.Cut("path,size,time")
			}}, "play": {Name: "play", Help: "演示"},
			ice.STORY: {Hand: func(m *ice.Message, arg ...string) {
				if kit.Ext(m.Option(nfs.PATH)) == nfs.SHY {
					m.Option(nfs.PATH, "")
				}
				m.Cmdy(arg[0], ice.RUN, arg[2:])
			}},
		}, WikiAction("", nfs.SHY), ctx.CmdAction(), aaa.RoleAction("story.field")), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_DEEP, ice.TRUE); len(arg) == 0 {
				arg = append(arg, "src/")
			}
			if !_wiki_list(m, arg...) {
				_word_show(m, arg[0])
			}
		}},
	})
}
func WordAction(template string, arg ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(append([]ice.Any{nfs.TEMPLATE, template}, arg...)...)}
}
func WordAlias(m *ice.Message, cmd string, cmds ...string) {
	mdb.Conf(m, WORD, kit.Keym(mdb.ALIAS, cmd), cmds)
}
