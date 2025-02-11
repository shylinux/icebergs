package git

import (
	"net/http"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const (
		EXPLORE_REPOS = "/explore/repos"
		REPOS_SEARCH  = "/api/v1/repos/search"
	)
	const (
		DESCRIPTION = "description"
		UPDATED_AT  = "updated_at"
		CLONE_URL   = "clone_url"
		HTML_URL    = "html_url"
		WEBSITE     = "website"
	)
	const SEARCH = "search"
	Index.MergeCommands(ice.Commands{
		SEARCH: {Name: "search keyword auto", Help: "源码库", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			cli.START: {Name: "start name*", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.DREAM, mdb.CREATE); m.Cmdy(web.DREAM, cli.START) }},
			CLONE:     {Name: "clone name*", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(REPOS, CLONE, m.Option(REPOS)) }},
			HTML_URL:  {Help: "源码", Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(m.Option(HTML_URL)) }},
			WEBSITE:   {Help: "官网", Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(m.Option(WEBSITE)) }},
			ORIGIN: {Help: "平台", Icon: "bi bi-box-arrow-up-right", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(web.SpideOrigin(m, REPOS) + EXPLORE_REPOS)
			}},
			web.DREAM_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(arg[0] == REPOS, func() {
					m.Cmds("", nfs.REPOS).Table(func(value ice.Maps) {
						m.Push(nfs.REPOS, value[HTML_URL]).Push(nfs.VERSION, "").Push(mdb.TIME, value[UPDATED_AT])
					})
				})
			}},
		}, web.DreamAction(), ctx.ConfAction(ctx.TOOLS, web.DREAM)), Hand: func(m *ice.Message, arg ...string) {
			kit.For(kit.Value(kit.UnMarshal(m.Cmdx(web.SPIDE, REPOS, web.SPIDE_RAW, http.MethodGet, REPOS_SEARCH,
				"q", kit.Select("", arg, 0), mdb.SORT, "updated", mdb.ORDER, "desc", mdb.PAGE, "1", mdb.LIMIT, "30")), mdb.DATA), func(value ice.Map) {
				value[nfs.SIZE] = kit.FmtSize(kit.Int(value[nfs.SIZE]) * 1000)
				if t, e := time.Parse(time.RFC3339, kit.Format(value[UPDATED_AT])); e == nil {
					value[UPDATED_AT] = t.Format("01-02 15:04")
				}
				m.Push("", value, []string{aaa.AVATAR_URL, mdb.NAME, DESCRIPTION,
					aaa.LANGUAGE, "forks_count", "stars_count", "watchers_count",
					nfs.SIZE, UPDATED_AT, CLONE_URL, HTML_URL, WEBSITE,
				})
				button := []ice.Any{}
				kit.If(!kit.IsIn(kit.Format(value[mdb.NAME]), ice.ICEBERGS, ice.VOLCANOS), func() { button = append(button, cli.START) })
				button = append(button, CLONE)
				kit.For([]string{HTML_URL, WEBSITE}, func(key string) { kit.If(kit.Format(value[key]), func() { button = append(button, key) }) })
				m.PushButton(button...)
				// m.Echo(kit.Formats(value))
			})
			m.RenameAppend(CLONE_URL, REPOS).Action(ORIGIN).Display("")
			ctx.Toolkit(m, "")
		}},
	})
}
