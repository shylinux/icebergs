package wiki

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _word_show(m *ice.Message, name string, arg ...string) {
	kit.If(kit.HasPrefix(name, nfs.PS, web.HTTP), func() { m.Option(nfs.CAT_CONTENT, m.Cmdx(web.SPIDE, ice.OPS, web.SPIDE_RAW, http.MethodGet, name)) })
	m.Options(ice.SSH_TARGET, m.Target(), ice.SSH_ALIAS, mdb.Configv(m, mdb.ALIAS), TITLE, map[string]int{})
	m.Cmdy(ssh.SOURCE, name, kit.Dict(nfs.DIR_ROOT, _wiki_path(m)))
}

const WORD = "word"

func init() {
	Index.MergeCommands(ice.Commands{
		WORD: {Name: "word path=src/main.shy@key auto play", Icon: "Books.png", Help: "上下文", Actions: ice.MergeActions(ice.Actions{
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
				m.Option(nfs.DIR_DEEP, ice.TRUE)
				_wiki_list(m, nfs.SRC)
				_wiki_list(m, nfs.USR_ICEBERGS)
				m.Cut("path,size,time")
			}},
			code.COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(kit.IsIn(kit.Split(m.Option(mdb.TEXT))[0], IMAGE, VIDEO, AUDIO), func() { m.Cmdy(FEEL).CutTo(nfs.PATH, mdb.NAME) })
			}},
		}, aaa.RoleAction(), WikiAction("", nfs.SHY)), Hand: func(m *ice.Message, arg ...string) {
			m.Option(nfs.DIR_DEEP, ice.TRUE)
			kit.If(len(arg) == 0, func() { arg = append(arg, nfs.SRC) })
			kit.If(!_wiki_list(m, arg...), func() { _word_show(m, arg[0]) })
		}},
	})
}
func WordAlias(m *ice.Message, cmd string, cmds ...string) {
	mdb.Conf(m, WORD, kit.Keym(mdb.ALIAS, cmd), cmds)
}
