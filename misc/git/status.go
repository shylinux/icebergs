package git

import (
	"path"
	"strings"

	"shylinux.com/x/go-git/v5/config"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const (
	INIT      = "init"
	DIFF      = "diff"
	INSTEADOF = "insteadof"
	OAUTH     = "oauth"
	STASH     = "stash"
	CHECKOUT  = "checkout"
)

const STATUS = "status"

func init() {
	Index.MergeCommands(ice.Commands{
		STATUS: {Name: "status repos:text auto", Help: "源代码", Icon: "git.png", Role: aaa.VOID, Meta: kit.Dict(
			ctx.ICONS, kit.Dict("message", "bi bi-info-square"),
			ctx.TRANS, kit.Dict(ctx.INPUT, kit.Dict("actions", "操作", "message", "信息", "remote", "远程库")),
		), Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: web.DreamWhiteHandle},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case aaa.EMAIL:
					m.Push(arg[0], ice.Info.NodeName+"@"+kit.Keys(kit.Slice(kit.Split(web.UserWeb(m).Hostname(), nfs.PT), -2)))
					m.Push(arg[0], _configs_get(m, USER_EMAIL), ice.Info.Make.Email)
				case aaa.USERNAME:
					m.Push(arg[0], ice.Info.NodeName)
					m.Push(arg[0], kit.Select(m.Option(ice.MSG_USERNAME), _configs_get(m, USER_NAME)), ice.Info.Make.Username)
				default:
					m.Cmdy(REPOS, mdb.INPUTS, arg)
				}
			}},
			ctx.CONFIG: {Name: "config email* username*", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, kit.HomePath(_GITCONFIG), kit.Format(nfs.Template(m, "gitconfig"), m.Option(aaa.USERNAME), m.Option(aaa.EMAIL)))
				mdb.Config(m, aaa.USERNAME, m.Option(aaa.USERNAME))
				mdb.Config(m, aaa.EMAIL, m.Option(aaa.EMAIL))
				m.ProcessRefresh()
			}},
			STASH: {Help: "清空", Icon: "bi bi-trash", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.SYSTEM, GIT, STASH)
				m.Cmd(cli.SYSTEM, GIT, CHECKOUT, ".")
				m.Go(func() { m.Sleep30ms(ice.QUIT, 1) })
				m.ProcessHold()
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if !m.IsDebug() || !aaa.IsTechOrRoot(m) || !nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME), _GIT)) {
					m.Push(mdb.TEXT, "")
					return
				}
				m.Push(mdb.TEXT, web.DreamStat(m, m.Option(mdb.NAME)))
				m.PushButton(kit.Dict(m.CommandKey(), "源码"))
			}},
			mdb.DEV_REQUEST: {Name: "dev.request origin*", Help: "授权"},
			web.DEV_CREATE_TOKEN: {Hand: func(m *ice.Message, arg ...string) {
				const FILE = ".git-credentials"
				host, list := ice.Map{kit.ParseURL(m.Option(web.ORIGIN)).Host: true}, []string{strings.Replace(m.Option(web.ORIGIN), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), m.Option(web.TOKEN)), 1)}
				m.Cmd(nfs.CAT, kit.HomePath(FILE), func(line string) {
					line = strings.ReplaceAll(line, "%3a", ":")
					kit.IfNoKey(host, kit.ParseURL(line).Host, func(p string) { list = append(list, line) })
				}).Cmd(nfs.SAVE, kit.HomePath(FILE), strings.Join(list, lex.NL)+lex.NL)
				m.Cmd(CONFIGS, mdb.CREATE, "credential.helper", "store")
				m.ProcessClose()
			}},
		}, web.DreamTablesAction(), web.DevTokenAction(web.ORIGIN, web.ORIGIN), Prefix(REPOS)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmdy(REPOS, arg)
			} else if config, err := config.LoadConfig(config.GlobalScope); err == nil && config.User.Email == "" && mdb.Config(m, aaa.EMAIL) == "" {
				m.EchoInfoButton(nfs.Template(m, "email.html"), ctx.CONFIG)
			} else if !nfs.Exists(m, _GIT) {
				m.EchoInfoButton(nfs.Template(m, "init.html"), INIT)
			} else if len(arg) == 0 {
				kit.If(config != nil, func() { m.Option(aaa.EMAIL, kit.Select(mdb.Config(m, aaa.EMAIL), config.User.Email)) })
				m.Cmdy(REPOS, STATUS)
				if m.IsMobileUA() {
					m.Action(PULL, PUSH)
				} else {
					m.Action(PULL, PUSH, INSTEADOF, mdb.DEV_REQUEST, ctx.CONFIG, STASH)
				}
				kit.If(!m.IsCliUA(), func() { m.Cmdy(code.PUBLISH, ice.CONTEXTS, ice.DEV) })
				ctx.Toolkit(m)
			} else {
				_repos_cmd(m, arg[0], DIFF)
			}
		}},
	})
}
