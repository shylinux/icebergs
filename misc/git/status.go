package git

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _status_tag(m *ice.Message, tags string) string {
	if tags == "" {
		return "v0.0.1"
	}
	ls := kit.Split(strings.TrimPrefix(kit.Split(tags, "-")[0], "v"), ice.PT)
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
	kit.SplitKV(ice.SP, ice.FS, _git_diff(m), func(text string, ls []string) {
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
		kit.SplitKV(ice.SP, ice.NL, _git_status(m), func(text string, ls []string) {
			switch kit.Ext(ls[1]) {
			case "swp", "swo", ice.BIN, ice.VAR:
				return
			}
			if onlychange && ls[0] == "##" {
				return
			}
			if m.Push(REPOS, value[REPOS]).Push(mdb.TYPE, ls[0]).Push(nfs.FILE, ls[1]); onlychange {
				m.Push(nfs.PATH, value[nfs.PATH]).Push(mdb.VIEW, kit.Format("%s %s", ls[0]+strings.Repeat(ice.SP, len(ls[0])-9), kit.Select("", nfs.USR+value[REPOS]+nfs.PS, value[REPOS] != ice.CONTEXTS)+ls[1]))
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
	INSTEADOF = "insteadof"
	OAUTH     = "oauth"
	DIFF      = "diff"
	OPT       = "opt"
	FIX       = "fix"

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
				case INSTEADOF:
					switch arg[0] {
					case nfs.FROM:
						m.Push(arg[0], kit.MergeURL2(ice.Info.Make.Remote, ice.PS))
					case nfs.TO:
						m.Cmd(web.BROAD, func(value ice.Maps) { m.Push(arg[0], kit.Format("http://%s:%s/", value[tcp.HOST], value[tcp.PORT])) })
					}
					return
				}
				switch arg[0] {
				case COMMENT:
					ls := kit.Split(m.Option(nfs.FILE), " /")
					m.Push(arg[0], kit.Join(kit.Slice(ls, -1), ice.PS))
					m.Push(arg[0], kit.Join(kit.Slice(ls, -2), ice.PS))
					m.Push(arg[0], m.Option(nfs.FILE))
				case VERSION:
					m.Push(VERSION, _status_tag(m, m.Option(TAGS)))
				case aaa.EMAIL:
					m.Push(arg[0], _configs_get(m, USER_EMAIL), ice.Info.Make.Email)
				case aaa.USERNAME:
					m.Push(arg[0], kit.Select(m.Option(ice.MSG_USERNAME), _configs_get(m, USER_NAME)), ice.Info.Make.Username)
				}
			}},
			CONFIGS: {Name: "configs email username", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				_configs_set(m, USER_NAME, m.Option(aaa.USERNAME))
				_configs_set(m, USER_EMAIL, m.Option(aaa.EMAIL))
			}},
			INSTEADOF: {Name: "insteadof from* to", Help: "代理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(CONFIGS, func(value ice.Maps) {
					kit.If(value[mdb.VALUE] == m.Option(nfs.FROM), func() { _configs_set(m, "--unset", value[mdb.NAME]) })
				})
				kit.If(m.Option(nfs.TO), func() { _git_cmd(m, CONFIG, "--global", "url."+m.Option(nfs.TO)+".insteadof", m.Option(nfs.FROM)) })
			}},
			OAUTH: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(kit.MergeURL2(kit.Select(ice.Info.Make.Domain, _git_remote(m)), "/chat/cmd/web.code.git.token/gen/", tcp.HOST, m.Option(ice.MSG_USERWEB)))
			}},
			TAG: {Name: "tag version", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, m.ActionKey(), arg)
			}},
			COMMIT: {Name: "commit actions=add,opt,fix comment*=some", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, m.ActionKey(), arg)
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TYPE) != web.WORKER {
					return
				}
				text := []string{}
				for _, line := range kit.Split(m.Cmdx(web.SPACE, m.Option(mdb.NAME), cli.SYSTEM, GIT, DIFF, "--shortstat"), ice.FS, ice.FS) {
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
		}, gdb.EventAction(web.DREAM_TABLES), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmdy(REPOS, arg)
			} else if len(arg) == 0 {
				m.Cmdy(REPOS, STATUS).Action(PULL, PUSH, "oauth", "insteadof")
				return
				files, adds, dels, last := _status_list(m)
				m.StatusTimeCount("files", files, "adds", adds, "dels", dels, "last", last, nfs.ORIGIN, _git_remote(m))
				m.Action(PULL, PUSH, "insteadof", "oauth").Sort("repos,type,file")
			} else {
				m.Cmdy(REPOS, arg[0], MASTER, INDEX, "README.md")
				return
				_repos_cmd(m, arg[0], DIFF)
				files, adds, dels := _status_stat(m, 0, 0, 0)
				m.StatusTime("files", files, "adds", adds, "dels", dels)
				m.Action(COMMIT, STASH)
			}
		}},
	})
}
