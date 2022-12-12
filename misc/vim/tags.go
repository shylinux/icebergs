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
	kit "shylinux.com/x/toolkits"
)

func _tags_split(pre, col string) []string {
	ls := kit.Split(pre[:kit.Int(col)-1])
	ls[len(ls)-1] += kit.Split(pre[kit.Int(col)-1:])[0]
	return ls
}
func _tags_field(m *ice.Message, arg ...string) {
	if arg[0] == "" {
		return
	}
	pre, sp := "", ""
	if word := kit.Slice(kit.Split(arg[1]+arg[0]), -1)[0]; arg[0] == ice.SP {
		sp = ice.SP
	} else if strings.HasSuffix(word, ice.PT) {
		pre = strings.TrimSuffix(word, ice.PT)
	} else if p := kit.Split(word, ice.PT); true {
		sp, pre = p[len(p)-1], word
	}
	list0, list := map[string]bool{}, map[string]bool{}
	push := func(index string) {
		if strings.HasPrefix(index, pre) {
			p := kit.Split(sp+strings.TrimPrefix(index, pre), ice.PT)[0]
			list0[p+kit.Select("", ice.PT, !strings.HasSuffix(index, p))] = true
		}
		list[strings.TrimPrefix(index, kit.Join(kit.Slice(kit.Split(pre, ice.PT), 0, -1), ice.PT)+ice.PT)] = true
	}
	ctx.CmdList(m).Tables(func(value ice.Maps) {
		if ls := kit.Split(pre, ice.PT); len(ls) == 1 && strings.Contains(value[ctx.INDEX], pre) && !strings.HasSuffix(arg[0], ice.PT) {
			push(value[ctx.INDEX])
		} else if len(ls) > 1 && strings.HasPrefix(value[ctx.INDEX], kit.Join(ls[:len(ls)-1], ice.PT)) && strings.Contains(value[ctx.INDEX], ls[len(ls)-1]) && !strings.HasSuffix(arg[0], ice.PT) {
			push(value[ctx.INDEX])
		} else if strings.HasPrefix(value[ctx.INDEX], pre) {
			res := sp + strings.TrimPrefix(value[ctx.INDEX], pre)
			ls := kit.Split(res, ice.PT)
			if len(ls) == 0 {
				return
			} else if len(ls) > 1 {
				ls[0] += ice.PT
			}
			if strings.HasPrefix(res, ice.PT) {
				list[ice.PT+ls[0]] = true
			} else {
				list[ls[0]] = true
			}
		}
	})
	kit.Fetch(list0, func(k string) { m.Echo("%s\n", k) })
	kit.Fetch(list, func(k string) { m.Echo("%s\n", k) })
}
func _tags_input(m *ice.Message, arg ...string) {
	if kit.Ext(m.Option(BUF)) == nfs.SHY && arg[1] == "" {
		kit.Fetch([]string{"field", "shell", "refer", "section", "chapter", "title"}, func(k string) {
			kit.If(strings.HasPrefix(k, arg[0]), func() { m.Echo("%s \n", k) })
		})
		_tags_field(m, arg...)
		return
	} else if arg[1] == "" {
		_tags_field(m, arg...)
		return
	} else if kit.Ext(m.Option(BUF)) == nfs.SHY && strings.HasPrefix(arg[1], "field") {
		_tags_field(m, arg...)
		return
	}
	name := kit.Select("", kit.Slice(kit.Split(arg[1], "\t \n."), -1), 0)
	switch name {
	case "can":
		mdb.ZoneSelectCB(m, "", func(value ice.Maps) { m.Echo(value[mdb.NAME] + ice.NL) })
		return
	}
	mdb.ZoneSelectCB(m, name, func(value ice.Maps) {
		if !strings.Contains(value[mdb.NAME], arg[0]) && arg[0] != ice.PT {
			return
		} else if m.Length() == 0 {
			m.Echo("func" + ice.NL)
		}
		m.Echo(value[mdb.NAME] + ice.NL)
		m.Echo("%s: %s"+ice.NL, value[mdb.NAME], strings.Split(value[mdb.TEXT], ice.NL)[0])
	})
}

const TAGS = "tags"

func init() {
	const (
		MODULE  = "module"
		PATTERN = "pattern"

		ONIMPORT = "onimport"
		ONACTION = "onaction"
		ONEXPORT = "onexport"

		defs_pattern = "4\n%s\n/\\<%s: /\n"
		func_pattern = "4\n%s\n/\\<%s: \\(shy\\|func\\)/\n"
		libs_pattern = "4\nusr/volcanos/lib/%s.js\n/\\<%s: \\(shy\\|func\\)/\n"
	)
	Index.MergeCommands(ice.Commands{
		TAGS: {Name: "tags zone id auto insert", Help: "索引", Actions: ice.MergeActions(ice.Actions{
			"listTags": {Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				kit.Fetch(kit.UnMarshal(m.Option(mdb.TEXT)), func(value ice.Map) {
					kit.If(value != nil, func() { mdb.ZoneInsert(m, value[mdb.ZONE], kit.Simple(value)) })
				})
			}},
			code.INNER: {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessField(m, "", m.OptionSplit("path,file,line"), arg...) }},
			INPUT:      {Hand: func(m *ice.Message, arg ...string) { _tags_input(m, arg...) }},
		}, mdb.ZoneAction(mdb.FIELD, "time,id,type,name,text,path,file,line"), ctx.ACTION, code.INNER), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelectAll(m, arg...); len(arg) == 0 {
				m.Action("listTags", mdb.EXPORT, mdb.IMPORT)
			}
		}},
		web.P(TAGS): {Actions: ice.Actions{
			tcp.SERVER: {Hand: func(m *ice.Message, arg ...string) {
				switch args := _tags_split(m.Option(PRE), m.Option(COL)); args[0] {
				case cli.QRCODE:
					m.Echo(`!curl "http://localhost:9020/code/bash/qrcode?text=%s"`, args[1])
				case wiki.FIELD:
					m.Echo(`!curl "http://localhost:9020/code/bash/qrcode?text=%s"`, kit.Format("http://2022.shylinux.com:9020/chat/cmd/%s?topic=black", args[1]))
				}
			}},
			nfs.SOURCE: {Hand: func(m *ice.Message, arg ...string) {
				switch args := _tags_split(m.Option(PRE), m.Option(COL)); args[0] {
				case cli.QRCODE:
					m.Echo(`!curl "http://localhost:9020/code/bash/qrcode?text=%s"`, args[1])
				case wiki.FIELD:
					m.Search(kit.Select(args[1], args, 2), func(key string, cmd *ice.Command) {
						ls := kit.Split(cmd.GetFileLines(), ice.DF)
						m.Echo("vi +%s %s", ls[1], ls[0])
					})
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			switch m.Option(MODULE) {
			case ONIMPORT, ONACTION, ONEXPORT:
				m.Echo(func_pattern, m.Option(BUF), m.Option(PATTERN))
			case "msg", "res":
				m.Echo(libs_pattern, ice.MISC, m.Option(PATTERN))
			default:
				if mdb.ZoneSelectCB(m, m.Option(MODULE), func(value ice.Maps) {
					if value[mdb.NAME] == m.Option(PATTERN) {
						m.Echo(kit.Select(defs_pattern, func_pattern, value[mdb.TYPE] == "function"),
							path.Join(value[nfs.PATH], value[nfs.FILE]), m.Option(PATTERN))
					}
				}); m.Length() == 0 {
					m.Echo(defs_pattern, "usr/volcanos/proto.js", m.Option(PATTERN))
				}
			}
		}},
	})
}
