package wiki

import (
	"net/http"
	"strings"

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
				if len(arg) > 0 {
					m.OptionFields("path,size,time")
					mdb.HashSelect(m)
				}
				msg := m.Spawn(kit.Dict(nfs.DIR_DEEP, ice.TRUE))
				_wiki_list(msg, nfs.SRC)
				_wiki_list(msg, nfs.USR_ICEBERGS)
				msg.Table(func(value ice.Maps) {
					if !kit.HasPrefix(value[nfs.PATH], nfs.SRC_TEMPLATE, nfs.SRC_DOCUMENT) {
						m.Push("", value, kit.Split("path,size,time"))
					}
				})
				web.PushPodCmd(m.Spawn(), "").Table(func(value ice.Maps) {
					if !kit.HasPrefix(value[nfs.PATH], nfs.SRC_TEMPLATE, nfs.SRC_DOCUMENT) {
						value[nfs.PATH] = value[web.SPACE] + nfs.DF + value[nfs.PATH]
						m.Push("", value, kit.Split("path,size,time"))
					}
				})
			}},
			code.COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(kit.IsIn(kit.Split(m.Option(mdb.TEXT))[0], IMAGE, VIDEO, AUDIO), func() { m.Cmdy(FEEL).CutTo(nfs.PATH, mdb.NAME) })
			}},
		}, aaa.RoleAction(), WikiAction("", nfs.SHY), mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 {
				mdb.HashCreate(m.Spawn(), nfs.PATH, arg[0])
			}
			if len(arg) > 0 && strings.Contains(arg[0], nfs.DF) {
				ls := kit.Split(arg[0], nfs.DF)
				arg[0] = ls[1]
				defer web.ToastProcess(m)()
				defer m.StatusTime(web.SPACE, m.Option(web.SPACE, ls[0]))
			}
			if len(arg) == 0 {
				m.Option(nfs.DIR_DEEP, ice.TRUE)
				arg = append(arg, nfs.SRC)
			} else if web.PodCmd(m, web.SPACE, arg...) {
				return
			}
			kit.If(!_wiki_list(m, arg...), func() { _word_show(m, arg[0]) })
		}},
	})
}
func WordAlias(m *ice.Message, cmd string, cmds ...string) {
	mdb.Conf(m, WORD, kit.Keym(mdb.ALIAS, cmd), cmds)
}
