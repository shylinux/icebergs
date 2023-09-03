package git

import (
	"net/http"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat/macos"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const SEARCH = "search"
	Index.MergeCommands(ice.Commands{
		SEARCH: {Name: "search repos keyword auto", Help: "代码源", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				macos.AppInstall(m, "usr/icons/gitea.png", m.PrefixKey(), ctx.ARGS, kit.Format([]string{"repos"}))
			}},
			cli.START: {Name: "start name", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.DREAM, cli.START)
			}},
			CLONE: {Name: "clone name", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, CLONE, m.Option(REPOS))
			}},
			cli.OPEN: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.Option("html_url"))
			}},
			"origin": {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.Cmdv("web.spide", "repos", "client.origin") + "/explore/repos")
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy("web.spide").RenameAppend("client.name", "repos", "client.url", "origin").Cut("time,repos,origin")
				return
			}
			res := kit.UnMarshal(m.Cmdx("web.spide", REPOS, web.SPIDE_RAW, http.MethodGet, "/api/v1/repos/search",
				"q", kit.Select("", arg, 1), "sort", "updated", "order", "desc", "page", "1", "limit", "100",
			))
			kit.For(kit.Value(res, "data"), func(value ice.Map) {
				value["size"] = kit.FmtSize(kit.Int(value["size"]) * 1000)
				if t, e := time.Parse(time.RFC3339, kit.Format(value["updated_at"])); e == nil {
					value["updated_at"] = t.Format("01-02 15:04")
				}
				m.Push("", value, []string{
					"avatar_url",
					"name",

					"language",
					"forks_count",
					"stars_count",
					"watchers_count",
					"size", "updated_at",

					"description",
					"clone_url",
					"html_url",
					"website",
				})
				m.PushButton(cli.START, CLONE, cli.OPEN)
			})
			m.RenameAppend("clone_url", "repos").StatusTimeCount().Display("")
			m.Action("origin")
			// m.Echo("%v", kit.Formats(res))
		}},
	})
}
