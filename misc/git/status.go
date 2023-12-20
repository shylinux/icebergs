package git

import (
	"path"
	"strings"

	"shylinux.com/x/go-git/v5/config"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const (
	INIT      = "init"
	DIFF      = "diff"
	INSTEADOF = "insteadof"
	OAUTH     = "oauth"
)

const STATUS = "status"

func init() {
	Index.MergeCommands(ice.Commands{
		STATUS: {Name: "status repos:text auto", Help: "代码库", Icon: "git.png", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict("actions", "操作", "message", "信息")),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case INIT:
					m.Cmd("web.spide", ice.OptionFields(web.CLIENT_ORIGIN), func(value ice.Maps) { m.Push(arg[0], value[web.CLIENT_ORIGIN]+"/x/"+path.Base(kit.Path(""))) })
					m.Push(arg[0], web.UserHost(m)+"/x/")
				case INSTEADOF:
					m.Cmd("web.spide", ice.OptionFields(web.CLIENT_ORIGIN), func(value ice.Maps) { m.Push(arg[0], value[web.CLIENT_ORIGIN]+"/x/") })
					m.Push(arg[0], web.UserHost(m)+"/x/")
				default:
					switch arg[0] {
					case aaa.USERNAME:
						m.Push(arg[0], kit.Select(m.Option(ice.MSG_USERNAME), _configs_get(m, USER_NAME)), ice.Info.Make.Username)
					case aaa.EMAIL:
						m.Push(arg[0], _configs_get(m, USER_EMAIL), ice.Info.Make.Email)
					default:
						m.Cmdy(REPOS, mdb.INPUTS, arg)
					}
				}
			}},
			INIT: {Name: "init origin*", Help: "初始化", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(REPOS, INIT) }},
			CONFIGS: {Name: "configs email* username* token", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, kit.HomePath(".gitconfig"), kit.Format(nfs.Template(m, "gitconfig"), m.Option(aaa.USERNAME), m.Option(aaa.EMAIL)))
				kit.If(m.Option(web.TOKEN), func() { m.Cmd(web.TOKEN, "set") })
				mdb.Config(m, aaa.USERNAME, m.Option(aaa.USERNAME))
				mdb.Config(m, aaa.EMAIL, m.Option(aaa.EMAIL))
			}},
			INSTEADOF: {Name: "insteadof remote", Help: "代理", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(REPOS, INSTEADOF, arg) }},
			OAUTH: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(kit.MergeURL2(kit.Select(ice.Info.Make.Domain, m.Cmdx(REPOS, "remoteURL")), web.ChatCmdPath(m, web.TOKEN, "gen"),
					mdb.TYPE, "web.code.git.status", tcp.HOST, m.Option(ice.MSG_USERWEB)))
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if !kit.IsIn(m.Option(mdb.TYPE), web.WORKER, web.SERVER) {
					return
				} else if !nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME), ".git")) {
					m.Push(mdb.TEXT, "")
					return
				}
				text := []string{}
				for _, line := range kit.Split(m.Cmdx(web.SPACE, m.Option(mdb.NAME), cli.SYSTEM, GIT, DIFF, "--shortstat"), mdb.FS, mdb.FS) {
					if list := kit.Split(line); strings.Contains(line, nfs.FILE) {
						text = append(text, list[0]+" file")
					} else if strings.Contains(line, "ins") {
						text = append(text, list[0]+" +++")
					} else if strings.Contains(line, "del") {
						text = append(text, list[0]+" ---")
					}
				}
				m.Push(mdb.TEXT, strings.Join(text, ", "))
				m.PushButton(kit.Dict(m.CommandKey(), "源码"))
			}},
		}, aaa.RoleAction(), web.DreamAction(), Prefix(REPOS)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmdy(REPOS, arg)
			} else if config, err := config.LoadConfig(config.GlobalScope); err == nil && config.User.Email == "" && mdb.Config(m, aaa.EMAIL) == "" {
				m.Action(CONFIGS).Echo("please config email and name. ").EchoButton(CONFIGS)
			} else if !nfs.Exists(m, ".git") {
				m.Action(INIT).Echo("please init repos. ").EchoButton(INIT)
			} else if len(arg) == 0 {
				kit.If(config != nil, func() { m.Option(aaa.EMAIL, kit.Select(mdb.Config(m, aaa.EMAIL), config.User.Email)) })
				m.Cmdy(REPOS, STATUS).Action(PULL, PUSH, INSTEADOF, OAUTH, CONFIGS)
				kit.If(!m.IsCliUA(), func() { m.Cmdy(code.PUBLISH, ice.CONTEXTS, ice.DEV) })
			} else {
				_repos_cmd(m, arg[0], DIFF)
			}
		}},
	})
}
