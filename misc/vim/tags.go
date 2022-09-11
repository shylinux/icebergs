package vim

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const TAGS = "tags"

func init() {
	_tags_split := func(pre, col string) []string {
		ls := kit.Split(pre[:kit.Int(col)-1])
		ls[len(ls)-1] += kit.Split(pre[kit.Int(col)-1:])[0]
		return ls
	}
	_tags_field := func(m *ice.Message, arg ...string) {
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

		m.OptionFields(ctx.INDEX)
		list0 := map[string]bool{}
		list := map[string]bool{}
		push := func(index string) {
			if strings.HasPrefix(index, pre) {
				p := kit.Split(sp+strings.TrimPrefix(index, pre), ice.PT)[0]
				list0[p+kit.Select("", ice.PT, !strings.HasSuffix(index, p))] = true
			}
			list[strings.TrimPrefix(index, kit.Join(kit.Slice(kit.Split(pre, ice.PT), 0, -1), ice.PT)+ice.PT)] = true
		}
		m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND).Tables(func(value ice.Maps) {
			if ls := kit.Split(pre, ice.PT); len(ls) == 1 && strings.Contains(value[ctx.INDEX], pre) && !strings.HasSuffix(arg[0], ice.PT) {
				push(value[ctx.INDEX])

			} else if len(ls) > 1 && strings.HasPrefix(value[ctx.INDEX], kit.Join(ls[:len(ls)-1], ice.PT)) && strings.Contains(value[ctx.INDEX], ls[len(ls)-1]) && !strings.HasSuffix(arg[0], ice.PT) {
				push(value[ctx.INDEX])

			} else if strings.HasPrefix(value[ctx.INDEX], pre) {
				res := sp + strings.TrimPrefix(value[ctx.INDEX], pre)
				ls := kit.Split(res, ice.PT)
				if len(ls) == 0 {
					return
				}
				if len(ls) > 1 {
					ls[0] += ice.PT
				}
				if strings.HasPrefix(res, ice.PT) {
					list[ice.PT+ls[0]] = true
				} else {
					list[ls[0]] = true
				}
			}
		})
		for _, k := range kit.SortedKey(list0) {
			m.Echo("%s\n", k)
		}
		for _, k := range kit.SortedKey(list) {
			m.Echo("%s\n", k)
		}
	}
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
		"/tags": {Name: "/tags", Help: "跳转", Actions: ice.Actions{
			"server": {Name: "server", Help: "服务", Hand: func(m *ice.Message, arg ...string) {
				switch args := _tags_split(m.Option(PRE), m.Option(COL)); args[0] {
				case "field":
					m.Echo(`!curl "localhost:9020/code/bash/qrcode?text=%s"`, kit.Format("http://2022.shylinux.com:9020/chat/cmd/%s?topic=black", args[1]))
				case "qrcode":
					m.Echo(`!curl "localhost:9020/code/bash/qrcode?text=%s"`, args[1])
				}
			}},
			"source": {Name: "source", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				switch args := _tags_split(m.Option(PRE), m.Option(COL)); args[0] {
				case "field":
					m.Search(kit.Select(args[1], args, 2), func(key string, cmd *ice.Command) {
						ls := kit.Split(cmd.GetFileLine(), ":")
						m.Echo("vi +%s %s", ls[1], ls[0])
					})
				case "qrcode":
					m.Echo(`!curl "localhost:9020/code/bash/qrcode?text=%s"`, args[1])
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
		TAGS: {Name: "tags zone id auto", Help: "索引", Actions: ice.MergeActions(ice.Actions{
			"listTags": {Name: "listTags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				kit.Fetch(kit.UnMarshal(m.Option(mdb.TEXT)), func(index int, value ice.Map) {
					if value == nil {
						return
					}
					m.Cmd(TAGS, mdb.INSERT, mdb.ZONE, value[mdb.ZONE], kit.Simple(value))
				})
				m.ProcessRefresh300ms()
			}},
			mdb.INSERT: {Name: "insert zone=core type name=hi text=hello path file line", Help: "添加"},
			code.INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessCommand(m, code.INNER, m.OptionSplit("path,file,line"), arg...)
			}},
			INPUT: {Name: "input name text", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if kit.Ext(m.Option(BUF)) == nfs.SHY && arg[1] == "" {
					for _, k := range []string{
						"field",
						"shell",
						"refer",
						"section",
						"chapter",
						"title",
					} {
						if strings.HasPrefix(k, arg[0]) {
							m.Echo("%s \n", k)
						}
					}
					_tags_field(m, arg...)
					return
				}
				if arg[1] == "" {
					_tags_field(m, arg...)
					return
				}
				if kit.Ext(m.Option(BUF)) == nfs.SHY && strings.HasPrefix(arg[1], "field") {
					_tags_field(m, arg...)
					return
				}
				name := kit.Select("", kit.Slice(kit.Split(arg[1], "\t \n."), -1), 0)
				switch name {
				case "can":
					mdb.ZoneSelectCB(m, "", func(value ice.Maps) {
						m.Echo(value[mdb.NAME] + ice.NL)
					})
					return
				}
				mdb.ZoneSelectCB(m, name, func(value ice.Maps) {
					if !strings.Contains(value[mdb.NAME], arg[0]) && arg[0] != ice.PT {
						return
					}
					if m.Length() == 0 {
						m.Echo("func" + ice.NL)
					}
					m.Echo(value[mdb.NAME] + ice.NL)
					m.Echo("%s: %s"+ice.NL, value[mdb.NAME], strings.Split(value[mdb.TEXT], ice.NL)[0])
				})
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelectAll(m, arg...); len(arg) == 0 {
				m.Action("listTags", mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
			} else {
				m.Action(mdb.INSERT).PushAction(code.INNER).StatusTimeCount()
			}
		}},
	})
}
