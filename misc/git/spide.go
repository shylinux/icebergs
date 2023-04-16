package git

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const SPIDE = "spide"

func init() {
	Index.MergeCommands(ice.Commands{
		SPIDE: {Name: "spide repos auto", Help: "构架图", Actions: ice.MergeActions(ice.Actions{
			"depend": {Name: "depend path*=icebergs/base pkg=shy,all", Help: "依赖", Hand: func(m *ice.Message, arg ...string) {
				list, keys := map[string]map[string]bool{}, map[string]bool{}
				kit.SplitKV(ice.TB, ice.NL, m.Cmdx(cli.SYSTEM, "gotags", "-R", path.Join(ice.USR, m.Option(nfs.PATH))+ice.PS), func(text string, ls []string) {
					if strings.HasPrefix(text, "!_") {
						return
					} else if kit.Select("", ls, 3) != "i" {
						return
					} else if !strings.Contains(ls[0], m.Option(nfs.PATH)) && m.Option("pkg") == "shy" {
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
					m.Push("pkg", k).Push(mdb.COUNT, len(v))
					for _, i := range item {
						m.Push(i, kit.Select("", ice.OK, v[i]))
					}
				}
				m.StatusTimeCount().SortIntR(mdb.COUNT)
			}}, code.INNER: {Name: web.CODE_INNER},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(kit.Slice(arg, 0, 1)) == 0 {
				m.Cmdy(REPOS)
			} else if len(arg) == 1 {
				color := []string{cli.YELLOW, cli.BLUE, cli.CYAN, cli.RED}
				ctx.DisplayStory(m, "", mdb.FIELD, nfs.PATH, aaa.ROOT, arg[0])
				nfs.DirDeepAll(m, _repos_path(m, arg[0]), "", func(value ice.Maps) {
					m.Push(cli.COLOR, color[strings.Count(value[nfs.PATH], ice.PS)%len(color)])
					m.Push("", value, []string{nfs.PATH})
				}, nfs.PATH)
				m.Option(nfs.DIR_ROOT, _repos_path(m, arg[0]))
				m.StatusTimeCount()
			} else if len(arg) == 2 {

			}
		}},
	})
}
