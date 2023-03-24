package vim

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	"shylinux.com/x/icebergs/core/wiki"
	"shylinux.com/x/icebergs/misc/bash"
	kit "shylinux.com/x/toolkits"
)

func _tags_input(m *ice.Message, arg ...string) {
	if kit.Ext(m.Option(BUF)) == nfs.SHY {
		if arg[1] == "" {
			kit.For([]string{"field", "shell", "refer", "section", "chapter", "title"}, func(k string) { kit.If(strings.HasPrefix(k, arg[0]), func() { m.EchoLine(k) }) })
		}
		m.EchoLine(kit.Join(bash.Complete(m, true, kit.Split(m.Option(PRE)+kit.Select("", ice.SP, !strings.HasSuffix(m.Option(PRE), ice.PT))+m.Option("cmds"))...), ice.SP+ice.NL))
		return
	}
	switch name := kit.Select("", kit.Slice(kit.Split(arg[1], "\t \n."), -1), 0); name {
	case "can", "sup", "sub":
		mdb.ZoneSelect(m).Table(func(value ice.Maps) {
			if strings.Contains(value[mdb.ZONE], arg[0]) || arg[0] == ice.PT {
				m.EchoLine(value[mdb.ZONE])
			}
		})
	default:
		mdb.ZoneSelectCB(m.Echo("func").Echo(ice.NL), name, func(value ice.Maps) {
			if strings.Contains(value[mdb.NAME], arg[0]) || arg[0] == ice.PT {
				m.EchoLine(value[mdb.NAME]+kit.Select("", "(", value[mdb.TYPE] == "function")).EchoLine("%s: %s", value[mdb.NAME], strings.Split(value[mdb.TEXT], ice.NL)[0])
			}
		})
	}
}

const TAGS = "tags"

func init() {
	Index.MergeCommands(ice.Commands{
		TAGS: {Name: "tags zone id auto insert", Help: "索引", Actions: ice.MergeActions(ice.Actions{
			nfs.LOAD: {Hand: func(m *ice.Message, arg ...string) {
				kit.For(kit.UnMarshal(m.Option(mdb.TEXT)), func(value ice.Map) { mdb.ZoneInsert(m, value[mdb.ZONE], kit.Simple(value)) })
			}},
			INPUT:      {Hand: func(m *ice.Message, arg ...string) { _tags_input(m, arg...) }},
			code.INNER: {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessField(m, "", m.OptionSplit("path,file,line"), arg...) }},
		}, mdb.ZoneAction(mdb.FIELD, "time,id,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelectAll(m, arg...); len(arg) == 0 {
				m.Action(nfs.LOAD, mdb.EXPORT, mdb.IMPORT)
			} else {
				m.PushAction(code.INNER)
			}
		}},
		web.PP(TAGS): {Actions: ice.Actions{
			tcp.SERVER: {Hand: func(m *ice.Message, arg ...string) {
				switch args := kit.Split(m.Option(PRE)); args[0] {
				case cli.QRCODE:
					Qrcode(m, args[1])
				case wiki.FIELD:
					m.Option(ice.MSG_USERWEB, m.Cmdx(web.SPACE, web.DOMAIN))
					Qrcode(m, web.MergePodCmd(m, "", kit.Select(args[1], args, 2)))
				default:
					m.Option(ice.MSG_USERWEB, m.Cmdx(web.SPACE, web.DOMAIN))
					Qrcode(m, web.MergePodCmd(m, "", args[0]))
				}
			}},
			nfs.SOURCE: {Hand: func(m *ice.Message, arg ...string) {
				switch args := kit.Split(m.Option(PRE)); args[0] {
				case cli.QRCODE:
					Qrcode(m, args[1])
				case wiki.FIELD:
					m.Search(kit.Select(args[1], args, 2), func(key string, cmd *ice.Command) {
						ls := kit.Split(cmd.FileLine(), ice.DF)
						m.Echo("vi +%s %s", ls[1], ls[0])
					})
				default:
					m.Search(args[0], func(key string, cmd *ice.Command) {
						ls := kit.Split(cmd.FileLine(), ice.DF)
						m.Echo("vi +%s %s", ls[1], ls[0])
					})
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			const (
				ONIMPORT = "onimport"
				ONACTION = "onaction"
				ONEXPORT = "onexport"
			)
			switch m.Option(mdb.ZONE) {
			case ONIMPORT, ONACTION, ONEXPORT:
				m.Echo(m.Option(BUF))
			case "msg", "res":
				m.Echo("usr/volcanos/lib/misc.js")
			default:
				if mdb.ZoneSelectAll(m, m.Option(mdb.ZONE)).Table(func(value ice.Maps) {
					kit.If(value[mdb.NAME] == m.Option(mdb.NAME), func() { m.Echo(path.Join(value[nfs.PATH], value[nfs.FILE])) })
				}); m.Result() == "" {
					m.Echo("usr/volcanos/proto.js")
				}
			}
		}},
	})
}

func Qrcode(m *ice.Message, arg ...string) {
	m.Echo(`!curl "http://localhost:9020/code/bash/qrcode?text=%s"`, arg[0])
}
