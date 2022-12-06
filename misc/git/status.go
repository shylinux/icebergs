package git

import (
	"path"
	"runtime"
	"strings"
	"time"

	"shylinux.com/x/gogit"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
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
	}
	return "v0.0.1"
}
func _status_tags(m *ice.Message) {
	vs := ice.Maps{}
	m.Cmd(STATUS, func(value ice.Maps) {
		if value[mdb.TYPE] == "##" {
			if value[REPOS] == ice.RELEASE {
				value[REPOS] = ice.ICE
			}
			vs[value[REPOS]] = strings.Split(value[TAGS], "-")[0]
		}
	})
	web.GoToast(m, TAGS, func(toast func(string, int, int)) {
		count, total := 0, len(vs)
		defer toast(ice.SUCCESS, count, count)
		for k := range vs {
			if count++; k == ice.ICE {
				k = ice.RELEASE
			}
			change := false
			toast(k, count, total)
			m.Option(nfs.DIR_ROOT, _repos_path(k))
			mod := m.Cmdx(nfs.CAT, ice.GO_MOD, func(text string, line int) string {
				ls := kit.Split(strings.TrimPrefix(text, ice.REQUIRE))
				if len(ls) < 2 || !strings.Contains(ls[0], ice.PS) || !strings.Contains(ls[1], ice.PT) {
					return text
				}
				if v, ok := vs[kit.Slice(strings.Split(ls[0], ice.PS), -1)[0]]; ok && ls[1] != v {
					m.Logs(mdb.MODIFY, REPOS, ls[0], "from", ls[1], "to", v)
					text, change = strings.ReplaceAll(text, ls[1], v), true
				}
				return text
			})
			if mod == "" || !change {
				continue
			}
			m.Cmd(nfs.SAVE, ice.GO_MOD, mod)
			switch m.Option(cli.CMD_DIR, _repos_path(k)); k {
			case ice.RELEASE, ice.ICEBERGS, ice.TOOLKITS:
				m.Cmd(cli.SYSTEM, code.GO, cli.BUILD)
			case ice.CONTEXTS:
				defer m.Cmd(cli.SYSTEM, cli.MAKE)
			default:
				m.Cmd(cli.SYSTEM, cli.MAKE)
			}
		}
	})
}
func _status_each(m *ice.Message, title string, cmds ...string) {
	web.GoToast(m, title, func(toast func(string, int, int)) {
		list, count, total := []string{}, 0, len(m.Confm(REPOS, mdb.HASH))
		ReposList(m).Tables(func(value ice.Maps) {
			toast(value[REPOS], count, total)
			if msg := m.Cmd(cmds, kit.Dict(cli.CMD_DIR, value[nfs.PATH])); !cli.IsSuccess(msg) {
				web.Toast3s(m, msg.Append(cli.CMD_ERR), "error: "+value[REPOS]).Sleep3s()
				list = append(list, value[REPOS])
			}
			count++
		})
		if len(list) > 0 {
			web.Toast30s(m, strings.Join(list, ice.NL), ice.FAILURE)
		} else {
			toast(ice.SUCCESS, count, total)
		}
	})
}
func _status_stat(m *ice.Message, files, adds, dels int) (int, int, int) {
	for _, v := range kit.Split(_git_cmds(m, DIFF, "--shortstat") , ice.FS, ice.FS) {
		n := kit.Int(kit.Split(strings.TrimSpace(v))[0])
		switch {
		case strings.Contains(v, "file"):
			files += n
		case strings.Contains(v, "inser"):
			adds += n
		case strings.Contains(v, "delet"):
			dels += n
		}
	}
	return files, adds, dels
}
func _status_list(m *ice.Message) (files, adds, dels int, last time.Time) {
	ReposList(m).Tables(func(value ice.Maps) {
		if m.Option(cli.CMD_DIR, value[nfs.PATH]); runtime.GOOS != cli.DARWIN {
			files, adds, dels = _status_stat(m, files, adds, dels)
		}
		if repos, e := gogit.OpenRepository(_git_dir(value[nfs.PATH])); e == nil {
			if ci, e := repos.GetCommit(); e == nil && ci.Author.When.After(last) {
				last = ci.Author.When
			}
		}
		tags := kit.Format(mdb.Cache(m, m.PrefixKey(value[REPOS], TAGS), func() ice.Any { return _git_cmds(m, "describe", "--tags") }))
		for _, v := range strings.Split(strings.TrimSpace(_git_cmds(m, STATUS, "-sb")), ice.NL) {
			if v == "" {
				continue
			}
			vs := strings.SplitN(strings.TrimSpace(v), ice.SP, 2)
			switch kit.Ext(vs[1]) {
			case "swp", "swo", ice.BIN, ice.VAR:
				continue
			}
			switch m.Push(REPOS, value[REPOS]).Push(mdb.TYPE, vs[0]).Push(nfs.FILE, vs[1]); vs[0] {
			case "##":
				if m.Push(TAGS, strings.TrimSpace(tags)); strings.Contains(vs[1], "ahead") || !strings.Contains(vs[1], "...") {
					m.PushButton(PUSH)
				} else if tags == "" || strings.Contains(tags, "-") {
					m.PushButton(TAG)
				} else {
					m.PushButton("")
				}
			default:
				if m.Push(TAGS, ""); strings.Contains(vs[0], "??") {
					m.PushButton(ADD, nfs.TRASH)
				} else {
					m.PushButton(COMMIT)
				}
			}
		}
	})
	return
}

const (
	PULL   = "pull"
	PUSH   = "push"
	DIFF   = "diff"
	TAGS   = "tags"
	STASH  = "stash"
	COMMIT = "commit"

	ADD = "add"
	OPT = "opt"
	PRO = "pro"
	TAG = "tag"
	PIE = "pie"

	COMMENT = "comment"
	VERSION = "version"
)
const STATUS = "status"

func init() {
	Index.MergeCommands(ice.Commands{
		STATUS: {Name: "status repos auto", Help: "状态机", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case COMMENT:
					m.Push(mdb.TEXT, m.Option(nfs.FILE))
					for _, v := range kit.Split(m.Option(nfs.FILE), " /") {
						m.Push(mdb.TEXT, v)
					}
				case VERSION, TAGS:
					if m.Option(TAGS) == "" {
						m.Push(VERSION, "v0.0.1")
					} else {
						m.Push(VERSION, _status_tag(m, m.Option(TAGS)))
					}
				case aaa.EMAIL:
					m.Push(arg[0], _configs_get(m, "user.email"))
				case aaa.USERNAME:
					m.Push(arg[0], kit.Select(m.Option(ice.MSG_USERNAME), _configs_get(m, "user.name")))
				}
			}},
			CONFIGS: {Name: "configs email username", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				_configs_set(m, "user.name", m.Option(aaa.USERNAME))
				_configs_set(m, "user.email", m.Option(aaa.EMAIL))
			}},
			INIT: {Name: "init origin*='https://shylinux.com/x/volcanos' name path", Help: "克隆", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, mdb.CREATE)
			}},
			PULL: {Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_status_each(m, PULL, cli.SYSTEM, GIT, PULL)
			}},
			PUSH: {Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(REPOS) == "" {
					_status_each(m, PUSH, cli.SYSTEM, GIT, PUSH)
					return
				}
				m.Option(cli.CMD_DIR, _repos_path(m.Option(REPOS)))
				if strings.TrimSpace(_git_cmds(m, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")) == "" {
					_git_cmd(m, PUSH, "--set-upstream", ORIGIN, MASTER)
				} else {
					_git_cmd(m, PUSH)
				}
				_git_cmd(m, PUSH, "--tags")
			}},
			ADD: {Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_repos_cmd(m, m.Option(REPOS), ADD, m.Option(nfs.FILE))
			}}, OPT: {Help: "优化"}, PRO: {Help: "升级"},
			COMMIT: {Name: "commit action=opt,add,pro comment=some", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				_repos_cmd(m, m.Option(REPOS), COMMIT, "-am", m.Option("action")+ice.SP+m.Option(COMMENT))
				mdb.Cache(m, m.PrefixKey(m.Option(REPOS), TAGS), nil)
				m.ProcessBack()
			}},
			TAG: {Name: "tag version", Help: "标签", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(VERSION) == "" {
					m.Option(VERSION, _status_tag(m, m.Option(TAGS)))
				}
				_repos_cmd(m, m.Option(REPOS), TAG, m.Option(VERSION))
				_repos_cmd(m, m.Option(REPOS), PUSH, "--tags")
				mdb.Cache(m, m.PrefixKey(m.Option(REPOS), TAGS), nil)
				ctx.ProcessRefresh(m)
			}},
			TAGS: {Help: "标签", Hand: func(m *ice.Message, arg ...string) { _status_tags(m) }},
			PIE:  {Help: "饼图", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TOTAL, PIE) }},
			STASH: {Help: "缓存", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 && m.Option(REPOS) == "" {
					_status_each(m, STASH, cli.SYSTEM, GIT, STASH)
				} else {
					_repos_cmd(m, kit.Select(m.Option(REPOS), arg, 0), STASH)
				}
			}},
			BRANCH: {Help: "分支", Hand: func(m *ice.Message, arg ...string) {
				for _, line := range kit.Split(_repos_cmd(m.Spawn(), arg[0], BRANCH).Result(), ice.NL, ice.NL) {
					if strings.HasPrefix(line, "*") {
						m.Push(BRANCH, strings.TrimPrefix(line, "* ")).PushButton("")
					} else {
						m.Push(BRANCH, strings.TrimSpace(line)).PushButton("branch_switch")
					}
				}
			}},
			"branch_switch": {Help: "切换", Hand: func(m *ice.Message, arg ...string) {
				_repos_cmd(m, m.Option(REPOS), "checkout", m.Option(BRANCH))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, path.Join(_repos_path(m.Option(REPOS)), m.Option(nfs.FILE)))
			}},
			code.COMPILE: {Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				defer web.ToastProcess(m)()
				m.Cmdy(code.VIMER, code.COMPILE)
			}},
			code.PUBLISH: {Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.PUBLISH, ice.CONTEXTS, ice.MISC, ice.CORE)
			}},
			code.BINPACK: {Help: "发布模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_RELEASE, ice.GO_SUM))
				m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_RELEASE, ice.GO_MOD))
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(code.VIMER, code.BINPACK)
			}},
			code.DEVPACK: {Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.VIMER, code.DEVPACK)
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
					} else if strings.Contains(line, "dele") {
						text = append(text, list[0]+" ---")
					}
				}
				m.Push(mdb.TEXT, strings.Join(text, ", "))
			}},
		}, gdb.EventAction(web.DREAM_TABLES), ctx.CmdAction(), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if _configs_get(m, "user.email") == "" {
				m.Echo("please config user.email").Action(CONFIGS)
			} else if len(arg) == 0 {
				defer web.ToastProcess(m)()
				files, adds, dels, last := _status_list(m)
				m.StatusTimeCount("files", files, "adds", adds, "dels", dels, "last", last.Format(ice.MOD_TIME))
				m.Action(PULL, PUSH, TAGS, PIE, code.COMPILE, code.PUBLISH)
			} else {
				_repos_cmd(m, arg[0], DIFF)
				files, adds, dels := _status_stat(m, 0, 0, 0)
				m.StatusTime("files", files, "adds", adds, "dels", dels)
				m.Action(COMMIT, TAGS, STASH, BRANCH)
			}
		}},
	})
}
