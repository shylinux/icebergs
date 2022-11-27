package git

import (
	"path"
	"strings"
	"time"

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
func _status_tags(m *ice.Message, repos string) {
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
		toast(cli.BEGIN, count, total)
		defer web.PushNoticeRefresh(m)

		for k := range vs {
			if k != repos && repos != "" {
				continue
			}
			if k == ice.ICE {
				k = ice.RELEASE
			}
			count++
			toast(k, count, total)

			change := false
			m.Option(nfs.DIR_ROOT, _repos_path(k))
			mod := m.Cmdx(nfs.CAT, ice.GO_MOD, func(text string, line int) string {
				ls := kit.Split(strings.TrimPrefix(text, ice.REQUIRE))
				if len(ls) < 2 || !strings.Contains(ls[0], ice.PS) || !strings.Contains(ls[1], ice.PT) {
					return text
				}
				if v, ok := vs[kit.Slice(strings.Split(ls[0], ice.PS), -1)[0]]; ok && ls[1] != v {
					m.Logs(mdb.MODIFY, REPOS, ls[0], "from", ls[1], "to", v)
					text = strings.ReplaceAll(text, ls[1], v)
					change = true
				}
				return text
			})
			if !change || mod == "" {
				continue
			}

			m.Cmd(nfs.SAVE, ice.GO_SUM, "")
			m.Cmd(nfs.SAVE, ice.GO_MOD, mod)
			switch m.Option(cli.CMD_DIR, _repos_path(k)); k {
			case ice.RELEASE, ice.ICEBERGS, ice.TOOLKITS:
				m.Cmd(cli.SYSTEM, code.GO, cli.BUILD)
			default:
				m.Cmd(cli.SYSTEM, cli.MAKE)
			}
		}
		toast(ice.SUCCESS, count, count)
	})
}
func _status_each(m *ice.Message, title string, cmds ...string) {
	web.GoToast(m, title, func(toast func(string, int, int)) {
		count, total := 0, len(m.Confm(REPOS, mdb.HASH))
		toast(cli.BEGIN, count, total)

		list := []string{}
		m.Cmd(REPOS, ice.OptionFields("name,path"), func(value ice.Maps) {
			toast(value[REPOS], count, total)

			if msg := m.Cmd(cmds, ice.Option{cli.CMD_DIR, value[nfs.PATH]}); !cli.IsSuccess(msg) {
				web.Toast3s(m, msg.Append(cli.CMD_ERR), "error: "+value[REPOS])
				list = append(list, value[REPOS])
				m.Sleep3s()
			}
			count++
		})

		if len(list) > 0 {
			web.Toast30s(m, strings.Join(list, ice.NL), ice.FAILURE)
		} else {
			toast(ice.SUCCESS, count, total)
			web.PushNoticeRefresh(m)
		}
	})
	m.ProcessHold()
}
func _status_stat(m *ice.Message, files, adds, dels int) (int, int, int) {
	res := _git_cmds(m, DIFF, "--shortstat")
	for _, v := range kit.Split(res, ice.FS, ice.FS) {
		n := kit.Int(kit.Split(strings.TrimSpace(v))[0])
		switch {
		case strings.Contains(v, "file"):
			files += n
		case strings.Contains(v, "insert"):
			adds += n
		case strings.Contains(v, "delet"):
			dels += n
		}
	}
	return files, adds, dels
}
func _status_list(m *ice.Message) (files, adds, dels int, last time.Time) {
	defer m.Sort("repos,type")
	m.Cmd(REPOS, ice.OptionFields("name,path")).Tables(func(value ice.Maps) {
		msg := m.Spawn(kit.Dict(cli.CMD_DIR, value[nfs.PATH]))
		diff := _git_cmds(msg, STATUS, "-sb")
		tags := _git_cmds(msg, "describe", "--tags")
		_files, _adds, _dels := _status_stat(msg, 0, 0, 0)
		now, _ := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(_git_cmds(msg, "log", `--pretty=%cd`, "--date=iso", "-n1")))

		if files, adds, dels = files+_files, adds+_adds, dels+_dels; now.After(last) {
			last = now
		}

		for _, v := range strings.Split(strings.TrimSpace(diff), ice.NL) {
			if v == "" {
				continue
			}
			vs := strings.SplitN(strings.TrimSpace(v), ice.SP, 2)
			switch kit.Ext(vs[1]) {
			case "swp", "swo", ice.BIN, ice.VAR:
				continue
			}

			m.Push(REPOS, value[REPOS])
			m.Push(mdb.TYPE, vs[0])
			m.Push(nfs.FILE, vs[1])

			list := []string{}
			switch vs[0] {
			case "##":
				m.Push(TAGS, strings.TrimSpace(tags))
				if tags == ice.ErrWarn || tags == "" {
					list = append(list, TAG)
				}

				if strings.Contains(vs[1], "ahead") || !strings.Contains(vs[1], "...") {
					list = append(list, PUSH)
				} else if strings.Contains(tags, "-") {
					list = append(list, TAG)
				}
			default:
				m.Push(TAGS, "")
				if strings.Contains(vs[0], "??") {
					list = append(list, ADD, nfs.TRASH)
				} else {
					list = append(list, COMMIT)
				}
			}
			m.PushButton(list)
		}
	})
	return
}

const (
	PULL = "pull"
	MAKE = "make"
	PUSH = "push"

	ADD = "add"
	OPT = "opt"
	PRO = "pro"
	TAG = "tag"
	PIE = "pie"

	TAGS    = "tags"
	DIFF    = "diff"
	COMMIT  = "commit"
	COMMENT = "comment"
	VERSION = "version"
	STASH   = "stash"
)
const STATUS = "status"

func init() {
	Index.MergeCommands(ice.Commands{
		STATUS: {Name: "status repos auto", Help: "状态机", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.NAME, REPOS:
					m.Cmdy(REPOS).Cut(REPOS)

				case TAGS, VERSION:
					if m.Option(TAGS) == ice.ErrWarn {
						m.Push(VERSION, "v0.0.1")
					} else {
						m.Push(VERSION, _status_tag(m, m.Option(TAGS)))
					}

				case COMMENT:
					m.Push(mdb.TEXT, m.Option(nfs.FILE))
					for _, v := range kit.Split(m.Option(nfs.FILE), " /") {
						m.Push(mdb.TEXT, v)
					}
				case aaa.USERNAME:
					m.Push(arg[0], kit.Select(m.Option(ice.MSG_USERNAME), _configs_get(m, "user.name")))
				case "email":
					m.Push(arg[0], _configs_get(m, "user.email"))
				}
			}},
			CONFIGS: {Name: "configs email username", Help: "配置", Hand: func(m *ice.Message, arg ...string) {
				_configs_set(m, "user.name", m.Option(aaa.USERNAME))
				_configs_set(m, "user.email", m.Option(aaa.EMAIL))
			}},
			CLONE: {Name: "clone repos='https://shylinux.com/x/volcanos' path=", Help: "克隆", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, mdb.CREATE)
			}},
			PULL: {Name: "pull", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_status_each(m, PULL, cli.SYSTEM, GIT, PULL)
			}},
			code.COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				web.ToastProcess(m)
				defer web.ToastSuccess(m)
				m.Cmdy(code.VIMER, code.COMPILE)
			}},
			PUSH: {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(REPOS) == "" {
					_status_each(m, PUSH, cli.SYSTEM, GIT, PUSH)
					return
				}

				m.Option(cli.CMD_DIR, _repos_path(m.Option(REPOS)))
				if strings.TrimSpace(_git_cmds(m, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")) == "" {
					_git_cmd(m, PUSH, "--set-upstream", "origin", "master")
				} else {
					_git_cmd(m, PUSH)
				}
				_git_cmd(m, PUSH, "--tags")
			}},
			TAGS: {Name: "tags", Help: "标签", Hand: func(m *ice.Message, arg ...string) {
				_status_tags(m, kit.Select("", arg, 0))
				m.ProcessHold()
			}},
			PIE: {Name: "pie", Help: "饼图", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TOTAL, PIE)
			}},
			STASH: {Name: "stash", Help: "缓存", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 && m.Option(REPOS) == "" {
					_status_each(m, STASH, cli.SYSTEM, GIT, STASH)
				} else {
					_git_cmd(m, STASH)
				}
			}},

			ADD: {Name: "add", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_repos_cmd(m, m.Option(REPOS), ADD, m.Option(nfs.FILE)).SetAppend()
			}}, OPT: {Name: "opt", Help: "优化"}, PRO: {Name: "pro", Help: "升级"},
			COMMIT: {Name: "commit action=opt,add,pro comment=some@key", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == ctx.ACTION {
					m.Option(mdb.TEXT, arg[1]+ice.SP+arg[3])
				} else {
					m.Option(mdb.TEXT, kit.Select("opt some", strings.Join(arg, ice.SP)))
				}
				_repos_cmd(m, m.Option(REPOS), COMMIT, "-am", m.Option(mdb.TEXT))
				m.ProcessBack()
			}},
			TAG: {Name: "tag version@key", Help: "标签", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(VERSION) == "" {
					m.Option(VERSION, _status_tag(m, m.Option(TAGS)))
				}
				_repos_cmd(m, m.Option(REPOS), TAG, m.Option(VERSION))
				_repos_cmd(m, m.Option(REPOS), PUSH, "--tags")
				m.ProcessRefresh()
			}},
			BRANCH: {Name: "branch", Help: "分支", Hand: func(m *ice.Message, arg ...string) {
				for _, line := range kit.Split(_repos_cmd(m.Spawn(), arg[0], BRANCH).Result(), ice.NL, ice.NL) {
					if strings.HasPrefix(line, "*") {
						m.Push(BRANCH, strings.TrimPrefix(line, "* "))
						m.PushButton("")
					} else {
						m.Push(BRANCH, strings.TrimSpace(line))
						m.PushButton("branch_switch")
					}
				}
			}},
			"branch_switch": {Name: "branch_switch", Help: "切换", Hand: func(m *ice.Message, arg ...string) {
				_repos_cmd(m.Spawn(), m.Option(REPOS), "checkout", m.Option(BRANCH))
			}},
			code.PUBLISH: {Name: "publish", Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.PUBLISH, ice.CONTEXTS, ice.MISC, ice.CORE)
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, path.Join(_repos_path(m.Option(REPOS)), m.Option(nfs.FILE)))
			}},
			code.BINPACK: {Name: "binpack", Help: "发布模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_RELEASE, ice.GO_SUM))
				m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_RELEASE, ice.GO_MOD))
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(code.VIMER, code.BINPACK)
			}},
			code.DEVPACK: {Name: "devpack", Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.VIMER, code.DEVPACK)
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TYPE) != web.WORKER {
					return
				}
				text := []string{}
				for _, line := range kit.Split(m.Cmdx(web.SPACE, m.Option(mdb.NAME), cli.SYSTEM, "git", "diff", "--shortstat"), ice.FS, ice.FS) {
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
				m.Echo("please config user.email")
				m.Action(CONFIGS)
				return
			}
			if len(arg) == 0 {
				web.ToastProcess(m, "status")
				m.Action(PULL, code.COMPILE, PUSH, TAGS, PIE, code.PUBLISH)
				files, adds, dels, last := _status_list(m)
				m.Status("cost", m.FormatCost(), "repos", m.Length(), "files", files, "adds", adds, "dels", dels, "last", last.Format(ice.MOD_TIME))
				web.Toast3s(m, kit.Format("files: %d, adds: %d, dels: %d", files, adds, dels), ice.CONTEXTS)
				return
			}

			_repos_cmd(m, arg[0], DIFF)
			m.Action(COMMIT, TAGS, STASH, BRANCH)
			files, adds, dels := _status_stat(m, 0, 0, 0)
			m.Status("files", files, "adds", adds, "dels", dels)
			web.Toast3s(m, kit.Format("files: %d, adds: %d, dels: %d", files, adds, dels), arg[0])
		}},
	})
}
