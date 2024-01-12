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
		STATUS: {Name: "status repos:text auto", Help: "代码库", Icon: "git.png", Role: aaa.VOID, Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict("actions", "操作", "message", "信息")),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case aaa.EMAIL:
					m.Push(arg[0], _configs_get(m, USER_EMAIL), ice.Info.Make.Email)
				case aaa.USERNAME:
					m.Push(arg[0], kit.Select(m.Option(ice.MSG_USERNAME), _configs_get(m, USER_NAME)), ice.Info.Make.Username)
				default:
					m.Cmdy(REPOS, mdb.INPUTS, arg)
				}
			}},
			OAUTH: {Name: "oauth remote", Help: "授权", Icon: "bi bi-person-check", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(kit.MergeURL2(kit.Select(ice.Info.Make.Domain, m.Cmdx(REPOS, REMOTE_URL), m.Option(REMOTE)),
					web.ChatCmdPath(m, web.TOKEN, "gen"), mdb.TYPE, web.CODE_GIT_STATUS, tcp.HOST, web.UserHost(m)))
			}},
			CONFIGS: {Name: "configs email* username*", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, kit.HomePath(_GITCONFIG), kit.Format(nfs.Template(m, "gitconfig"), m.Option(aaa.USERNAME), m.Option(aaa.EMAIL)))
				mdb.Config(m, aaa.USERNAME, m.Option(aaa.USERNAME))
				mdb.Config(m, aaa.EMAIL, m.Option(aaa.EMAIL))
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if !nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME), _GIT)) {
					m.Push(mdb.TEXT, "")
					m.PushButton(kit.Dict(m.CommandKey(), "源码"))
					return
				}
				text := []string{}
				for _, line := range kit.Split(m.Cmdx(cli.SYSTEM, GIT, DIFF, "--shortstat", kit.Dict(cli.CMD_DIR, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))), mdb.FS, mdb.FS) {
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
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, nil, arg...) }},
		}, Prefix(REPOS)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmdy(REPOS, arg)
			} else if config, err := config.LoadConfig(config.GlobalScope); err == nil && config.User.Email == "" && mdb.Config(m, aaa.EMAIL) == "" {
				m.EchoInfoButton("please config email and name. ", CONFIGS)
			} else if !nfs.Exists(m, _GIT) {
				m.EchoInfoButton("please init repos. ", INIT)
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
