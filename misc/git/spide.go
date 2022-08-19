package git

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _spide_for(text string, cb func([]string)) {
	for _, line := range strings.Split(text, ice.NL) {
		if len(line) == 0 || strings.HasPrefix(line, "!_") {
			continue
		}
		cb(kit.Split(line, "\t ", "\t ", "\t "))
	}
}
func _spide_go(m *ice.Message, file string) {
	_spide_for(m.Cmdx(cli.SYSTEM, "gotags", file), func(ls []string) {
		switch ls[3] {
		case "i":
			return
		case "w", "e":
			return
			ls[0] = "-" + ls[0] + ice.DF + strings.TrimPrefix(ls[len(ls)-1], "type:")
		case "m":
			if strings.HasPrefix(ls[6], "ntype") {
				return
			} else if strings.HasPrefix(ls[5], "ctype") {
				ls[0] = strings.TrimPrefix(ls[5], "ctype:") + ice.DF + ls[0]
			} else {
				ls[0] = ls[3] + ice.DF + ls[0]
			}
		default:
			ls[0] = ls[3] + ice.DF + ls[0]
		}

		m.Push(mdb.NAME, ls[0])
		m.Push(nfs.FILE, ls[1])
		m.Push(nfs.LINE, strings.TrimSuffix(ls[2], ";\""))
		m.Push(mdb.TYPE, ls[3])
		m.Push(mdb.EXTRA, strings.Join(ls[4:], ice.SP))
	})
}
func _spide_c(m *ice.Message, file string) {
	_spide_for(m.Cmdx(cli.SYSTEM, "ctags", "-f", file), func(ls []string) {
		m.Push(mdb.NAME, ls[0])
		m.Push(nfs.FILE, ls[1])
		m.Push(nfs.LINE, "1")
	})
}

const SPIDE = "spide"

func init() {
	Index.MergeCommands(ice.Commands{
		SPIDE: {Name: "spide repos auto", Help: "构架图", Actions: ice.MergeActions(ice.Actions{
			"depend": {Name: "depend path=icebergs/base", Help: "依赖", Hand: func(m *ice.Message, arg ...string) {
				keys := map[string]bool{}
				list := map[string]map[string]bool{}
				dir := path.Join(ice.USR, m.Option(nfs.PATH)) + ice.PS
				_spide_for(m.Cmdx(cli.SYSTEM, "gotags", "-R", dir), func(ls []string) {
					if kit.Select("", ls, 3) != "i" {
						return
					}
					if !strings.Contains(ls[0], m.Option(nfs.PATH)) {
						return
					}

					item, ok := list[ls[0]]
					if !ok {
						item = map[string]bool{}
						list[ls[0]] = item
					}

					p := strings.TrimPrefix(path.Dir(ls[1]), path.Join(ice.USR, m.Option(nfs.PATH)))
					keys[p], item[p] = true, true
				})

				item := kit.SortedKey(keys)

				for k, v := range list {
					m.Push("pkg", k)
					m.Push(mdb.COUNT, len(v))
					for _, i := range item {
						m.Push(i, kit.Select("", ice.OK, v[i]))
					}
				}
				m.SortIntR(mdb.COUNT)
				m.ProcessInner()
			}}, code.INNER: {Name: "web.code.inner"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(kit.Slice(arg, 0, 1)) == 0 { // 仓库列表
				m.Cmdy(REPOS)
				return
			}

			if arg[0] = kit.Replace(arg[0], ice.SRC, ice.CONTEXTS); arg[0] == path.Base(kit.Pwd()) {
				m.Option(nfs.DIR_ROOT, path.Join(ice.SRC)+ice.PS)
			} else {
				m.Option(nfs.DIR_ROOT, path.Join(ice.USR, arg[0])+ice.PS)
			}
			ctx.DisplayStory(m, "spide.js", mdb.FIELD, nfs.PATH, "root", arg[0])

			if len(arg) == 1 || !strings.HasSuffix(arg[1], arg[2]) { // 目录列表
				m.Option(nfs.DIR_DEEP, ice.TRUE)
				color := []string{cli.YELLOW, cli.BLUE, cli.CYAN, cli.RED}
				nfs.Dir(m, nfs.PATH).Tables(func(value ice.Maps) {
					m.Push(cli.COLOR, color[strings.Count(value[nfs.PATH], ice.PS)%len(color)])
				})
				return
			}

			// 语法解析
			switch m.Option(cli.CMD_DIR, m.Option(nfs.DIR_ROOT)); kit.Ext(arg[1]) {
			case code.GO:
				_spide_go(m, arg[1])
			case code.JS:
			default:
				_spide_c(m, arg[1])
			}
			m.SortInt(nfs.LINE)
		}},
	})
}
