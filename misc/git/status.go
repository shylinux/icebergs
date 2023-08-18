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
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _status_tag(m *ice.Message, tags string) string {
	if tags == "" {
		return "v0.0.1"
	}
	ls := kit.Split(strings.TrimPrefix(kit.Split(tags, "-")[0], "v"), nfs.PT)
	if v := kit.Int(ls[2]); v < 9 {
		return kit.Format("v%v.%v.%v", ls[0], ls[1], v+1)
	} else if v := kit.Int(ls[1]); v < 9 {
		return kit.Format("v%v.%v.0", ls[0], v+1)
	} else if v := kit.Int(ls[0]); v < 9 {
		return kit.Format("v%v.0.0", v+1)
	} else {
		return "v0.0.1"
	}
}
func _status_stat(m *ice.Message, files, adds, dels int) (int, int, int) {
	kit.SplitKV(lex.SP, mdb.FS, _git_diff(m), func(text string, ls []string) {
		n := kit.Int(ls[0])
		switch {
		case strings.Contains(text, "file"):
			files += n
		case strings.Contains(text, "inser"):
			adds += n
		case strings.Contains(text, "delet"):
			dels += n
		}
	})
	return files, adds, dels
}
func _status_list(m *ice.Message) (files, adds, dels int, last string) {
	onlychange := m.Option(ice.MSG_MODE) == mdb.ZONE
	defer m.Option(cli.CMD_DIR, "")
	ReposList(m).Table(func(value ice.Maps) {
		m.Option(cli.CMD_DIR, value[nfs.PATH])
		files, adds, dels = _status_stat(m, files, adds, dels)
		_last := m.Cmdv(REPOS, path.Base(value[nfs.PATH]), mdb.TIME)
		kit.If(_last > last, func() { last = _last })
		tags := _git_tags(m)
		kit.SplitKV(lex.SP, lex.NL, _git_status(m), func(text string, ls []string) {
			switch kit.Ext(ls[1]) {
			case "swp", "swo", ice.BIN, ice.VAR:
				return
			}
			if onlychange && ls[0] == "##" {
				return
			}
			if m.Push(REPOS, value[REPOS]).Push(mdb.TYPE, ls[0]).Push(nfs.FILE, ls[1]); onlychange {
				m.Push(nfs.PATH, value[nfs.PATH]).Push(mdb.VIEW, kit.Format("%s %s", ls[0]+strings.Repeat(lex.SP, len(ls[0])-9), kit.Select("", nfs.USR+value[REPOS]+nfs.PS, value[REPOS] != ice.CONTEXTS)+ls[1]))
			}
			switch ls[0] {
			case "##":
				if m.Push(TAGS, tags); strings.Contains(ls[1], "ahead") || !strings.Contains(ls[1], "...") {
					m.PushButton(PUSH)
				} else if tags == "" || strings.Contains(tags, "-") {
					m.PushButton(TAG)
				} else {
					m.PushButton("")
				}
			default:
				if m.Push(TAGS, ""); strings.Contains(ls[0], "??") {
					m.PushButton(ADD, nfs.TRASH)
				} else {
					m.PushButton(COMMIT)
				}
			}
		})
	})
	return
}

const (
	OAUTH = "oauth"
	DIFF  = "diff"
	OPT   = "opt"
	FIX   = "fix"

	TAGS    = "tags"
	VERSION = "version"
	COMMENT = "comment"
)
const STATUS = "status"

func init() {
	Index.MergeCommands(ice.Commands{
		STATUS: {Name: "status repos:text auto", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case INIT:
					switch arg[0] {
					case ORIGIN:
						m.Cmd("web.spide", func(value ice.Maps) {
							m.Push(arg[0], kit.ParseURLMap(value[web.CLIENT_URL])[ORIGIN]+"/x/"+path.Base(kit.Path("")))
						})
					}
				case INSTEADOF:
					switch arg[0] {
					case nfs.FROM:
						m.Push(arg[0], kit.MergeURL2(ice.Info.Make.Remote, nfs.PS))
					case nfs.TO:
						m.Cmd(web.BROAD, func(value ice.Maps) { m.Push(arg[0], kit.Format("http://%s:%s/", value[tcp.HOST], value[tcp.PORT])) })
					case REMOTE:
						m.Cmd("web.spide", func(value ice.Maps) { m.Push(arg[0], kit.ParseURLMap(value[web.CLIENT_URL])[ORIGIN]+"/x/") })
					}
					return
				}
				switch arg[0] {
				case aaa.EMAIL:
					m.Push(arg[0], _configs_get(m, USER_EMAIL), ice.Info.Make.Email)
				case aaa.USERNAME:
					m.Push(arg[0], kit.Select(m.Option(ice.MSG_USERNAME), _configs_get(m, USER_NAME)), ice.Info.Make.Username)
				default:
					m.Cmdy(REPOS, mdb.INPUTS, arg)
				}
			}},
			INIT: {Name: "init origin", Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, INIT)
			}},
			CONFIGS: {Name: "configs email* username* token", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, kit.HomePath(".gitconfig"), nfs.Template(m, "gitconfig", m.Option(aaa.USERNAME), m.Option(aaa.EMAIL)))
				mdb.Config(m, aaa.USERNAME, m.Option(aaa.USERNAME))
				mdb.Config(m, aaa.EMAIL, m.Option(aaa.EMAIL))
				kit.If(m.Option(web.TOKEN), func() { m.Cmd(web.TOKEN, "set") })
			}},
			OAUTH: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(kit.MergeURL2(kit.Select(ice.Info.Make.Domain, m.Cmdx(REPOS, "remoteURL")), web.ChatCmdPath(m, web.TOKEN, "gen"), tcp.HOST, m.Option(ice.MSG_USERWEB)))
			}},
			INSTEADOF: {Name: "insteadof remote", Help: "代理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, INSTEADOF, arg)
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
					if list := kit.Split(line); strings.Contains(line, "file") {
						text = append(text, list[0]+" file")
					} else if strings.Contains(line, "ins") {
						text = append(text, list[0]+" +++")
					} else if strings.Contains(line, "del") {
						text = append(text, list[0]+" ---")
					}
				}
				m.Push(mdb.TEXT, strings.Join(text, ", "))
			}},
		}, aaa.RoleAction(), web.DreamAction(), Prefix(REPOS), mdb.ImportantHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmdy(REPOS, arg)
			} else if config, err := config.LoadConfig(config.GlobalScope); err == nil && config.User.Email == "" && mdb.Config(m, aaa.EMAIL) == "" {
				m.Action(CONFIGS).Echo("please config email and name. ").EchoButton(CONFIGS)
			} else if !nfs.Exists(m, ".git") {
				m.Action("init").Echo("please init repos. ").EchoButton("init")
			} else if len(arg) == 0 {
				kit.If(config != nil, func() { m.Option(aaa.EMAIL, kit.Select(mdb.Config(m, aaa.EMAIL), config.User.Email)) })
				m.Cmdy(REPOS, STATUS).Action(PULL, PUSH, INSTEADOF, "oauth", CONFIGS)
				m.Cmdy(code.PUBLISH, ice.CONTEXTS, "dev")
			} else {
				m.Cmdy(REPOS, arg[0], MASTER, INDEX, m.Cmdv(REPOS, arg[0], MASTER, INDEX, nfs.FILE))
			}
		}},
	})
}
