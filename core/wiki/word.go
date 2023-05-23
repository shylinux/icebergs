package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	"shylinux.com/x/icebergs/misc/git"
	kit "shylinux.com/x/toolkits"
)

func _word_show(m *ice.Message, name string, arg ...string) {
	m.Options(ice.MSG_ALIAS, mdb.Configv(m, mdb.ALIAS), TITLE, map[string]int{})
	m.Cmdy(ssh.SOURCE, name, kit.Dict(nfs.DIR_ROOT, _wiki_path(m)))
}

const WORD = "word"

func init() {
	Index.MergeCommands(ice.Commands{
		WORD: {Name: "word path=src/main.shy@key auto play", Help: "笔记文档", Actions: ice.MergeActions(ice.Actions{
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
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				mdb.IsSearchForEach(m, arg, func() []string { return []string{web.LINK, m.CommandKey(), m.MergePodCmd("", "")} })
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(git.REPOS, ice.OptionFields(nfs.PATH)).Table(func(value ice.Maps) {
					if m.Option(nfs.DIR_DEEP, ice.TRUE); kit.Path(value[nfs.PATH]) == kit.Path("") {
						_wiki_list(m, nfs.SRC)
					} else {
						_wiki_list(m, value[nfs.PATH])
					}
				})
				m.Cut("path,size,time")
			}}, "play": {Help: "演示"},
			ice.STORY: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(arg[0], ice.RUN, arg[2:]) }},
			code.COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				ls := kit.Split(m.Option(mdb.TEXT))
				kit.If(kit.IsIn(ls[0], IMAGE, VIDEO, AUDIO), func() { m.Cmdy(FEEL).CutTo(nfs.PATH, mdb.NAME) })
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() { m.PushButton(kit.Dict(m.CommandKey(), "文档")) })
			}},
		}, aaa.RoleAction("story.field"), ctx.CmdAction(), WikiAction("", nfs.SHY)), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_DEEP, ice.TRUE); len(arg) == 0 {
				arg = append(arg, nfs.SRC)
			}
			kit.If(!_wiki_list(m, arg...), func() { _word_show(m, arg[0]) })
		}},
	})
}
func WordAlias(m *ice.Message, cmd string, cmds ...string) {
	mdb.Conf(m, WORD, kit.Keym(mdb.ALIAS, cmd), cmds)
}
