package git

import (
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
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
func _status_tags(m *ice.Message) {
	vs := map[string]string{}
	m.Cmd(STATUS).Table(func(index int, value map[string]string, head []string) {
		if value[kit.MDB_TYPE] == "##" {
			if value[kit.MDB_NAME] == ice.RELEASE {
				value[kit.MDB_NAME] = ice.ICE
			}
			vs[value[kit.MDB_NAME]] = strings.Split(value[TAGS], "-")[0]
		}
	})

	m.GoToast(TAGS, func(toast func(string, int, int)) {
		count, total := 0, len(vs)
		toast(cli.BEGIN, count, total)

		for k := range vs {
			count++
			toast(k, count, total)

			if k == ice.ICE {
				k = ice.RELEASE
			}

			change := false
			m.Option(nfs.CAT_LOCAL, ice.TRUE)
			m.Option(nfs.DIR_ROOT, _repos_path(k))
			mod := m.Cmdx(nfs.CAT, ice.GO_MOD, func(text string, line int) string {
				ls := kit.Split(strings.TrimPrefix(text, ice.REQUIRE))
				if len(ls) < 2 || !strings.Contains(ls[0], ice.PS) || !strings.Contains(ls[1], ice.PT) {
					return text
				}
				if v, ok := vs[kit.Slice(strings.Split(ls[0], ice.PS), -1)[0]]; ok && ls[1] != v {
					m.Info("upgrade to %v %v from %v", ls[0], ls[1], v)
					text = strings.Replace(text, v, ls[1], 1)
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
			case ice.CONTEXTS:
				defer m.Cmd(cli.SYSTEM, cli.MAKE, ice.Option{cli.CMD_DIR, _repos_path(k)})
			case ice.ICEBERGS:
				m.Cmd(cli.SYSTEM, code.GO, cli.BUILD)
			default:
				m.Cmd(cli.SYSTEM, cli.MAKE)
			}
		}
		toast(ice.SUCCESS, count, total)
	})
}
func _status_each(m *ice.Message, title string, cmds ...string) {
	m.GoToast(title, func(toast func(string, int, int)) {
		count, total := 0, len(m.Confm(REPOS, kit.MDB_HASH))
		toast(cli.BEGIN, count, total)

		list := []string{}
		m.Cmd(REPOS, ice.OptionFields("name,path")).Table(func(index int, value map[string]string, head []string) {
			toast(value[kit.MDB_NAME], count, total)

			if msg := m.Cmd(cmds, ice.Option{cli.CMD_DIR, value[kit.MDB_PATH]}); !cli.IsSuccess(msg) {
				m.Toast3s(msg.Append(cli.CMD_ERR), "error: "+value[kit.MDB_NAME])
				list = append(list, value[kit.MDB_NAME])
				m.Sleep3s()
			}
			count++
		})

		if len(list) > 0 {
			m.Toast30s(strings.Join(list, ice.NL), ice.FAILURE)
		} else {
			toast(ice.SUCCESS, count, total)
		}
	})
}
func _status_stat(m *ice.Message, files, adds, dels int) (int, int, int) {
	for _, v := range kit.Split(m.Cmdx(cli.SYSTEM, GIT, DIFF, "--shortstat"), ice.FS, ice.FS) {
		n := kit.Int(kit.Split(strings.TrimSpace(v))[0])
		switch {
		case strings.Contains(v, "file"):
			files += n
		case strings.Contains(v, "insert"):
			adds += n
		case strings.Contains(v, "delete"):
			dels += n
		}
	}
	return files, adds, dels
}
func _status_list(m *ice.Message) (files, adds, dels int, last time.Time) {
	m.Cmd(REPOS, ice.OptionFields("name,path")).Table(func(index int, value map[string]string, head []string) {
		m.Option(cli.CMD_DIR, value[kit.MDB_PATH])
		diff := m.Cmdx(cli.SYSTEM, GIT, STATUS, "-sb")
		tags := m.Cmdx(cli.SYSTEM, GIT, "describe", "--tags")

		for _, v := range strings.Split(strings.TrimSpace(diff), ice.NL) {
			if v == "" {
				continue
			}
			vs := strings.SplitN(strings.TrimSpace(v), ice.SP, 2)
			switch kit.Ext(vs[1]) {
			case "swp", "swo", "bin", "var":
				continue
			}

			m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
			m.Push(kit.MDB_TYPE, vs[0])
			m.Push(kit.MDB_FILE, vs[1])

			list := []string{}
			switch vs[0] {
			case "##":
				m.Push(TAGS, strings.TrimSpace(tags))
				if tags == ice.ErrWarn {
					list = append(list, TAG)
				}

				if strings.Contains(vs[1], "ahead") {
					list = append(list, PUSH)
				} else if strings.Contains(tags, "-") {
					list = append(list, TAG)
				}
			default:
				m.Push(TAGS, "")
				if strings.Contains(vs[0], "??") {
					list = append(list, ADD)
				} else {
					list = append(list, COMMIT)
				}
			}
			m.PushButton(list)
		}

		files, adds, dels = _status_stat(m, files, adds, dels)
		now, _ := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, "log", `--pretty=%cd`, "--date=iso", "-n1")))
		if now.After(last) {
			last = now
		}
	})
	return
}

const (
	PULL = "pull"
	MAKE = "make"
	PUSH = "push"

	TAG = "tag"
	ADD = "add"
	OPT = "opt"
	PRO = "pro"

	TAGS    = "tags"
	DIFF    = "diff"
	COMMIT  = "commit"
	COMMENT = "comment"
	VERSION = "version"
)
const STATUS = "status"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		STATUS: {Name: "status name auto", Help: "状态机", Action: map[string]*ice.Action{
			PULL: {Name: "pull", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_status_each(m, PULL, cli.SYSTEM, GIT, PULL)
				m.ProcessHold()
			}},
			MAKE: {Name: "make", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				web.PushStream(m)
				m.Cmdy(cli.SYSTEM, MAKE)
				m.Toast(ice.SUCCESS)
				m.ProcessHold()
			}},
			TAGS: {Name: "tags", Help: "标签", Hand: func(m *ice.Message, arg ...string) {
				_status_tags(m)
				m.ProcessHold()
			}},
			PUSH: {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_NAME) == "" {
					_status_each(m, PUSH, cli.SYSTEM, GIT, PUSH)
					m.ProcessHold()
					return
				}

				_repos_cmd(m, m.Option(kit.MDB_NAME), PUSH)
				_repos_cmd(m, m.Option(kit.MDB_NAME), PUSH, "--tags")
			}},

			TAG: {Name: "tag version@key", Help: "标签", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(VERSION) == "" {
					m.Option(VERSION, _status_tag(m, m.Option(TAGS)))
				}
				_repos_cmd(m, m.Option(kit.MDB_NAME), TAG, m.Option(VERSION))
				_repos_cmd(m, m.Option(kit.MDB_NAME), PUSH, "--tags")
			}},
			ADD: {Name: "add", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_repos_cmd(m, m.Option(kit.MDB_NAME), ADD, m.Option(kit.MDB_FILE))
			}}, OPT: {Name: "opt", Help: "优化"}, PRO: {Name: "pro", Help: "升级"},
			COMMIT: {Name: "commit action=opt,add,pro comment=some@key", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == ctx.ACTION {
					m.Option(kit.MDB_TEXT, arg[1]+ice.SP+arg[3])
				} else {
					m.Option(kit.MDB_TEXT, kit.Select("opt some", strings.Join(arg, ice.SP)))
				}

				_repos_cmd(m, m.Option(kit.MDB_NAME), COMMIT, "-am", m.Option(kit.MDB_TEXT))
				m.ProcessBack()
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_NAME:
					m.Cmdy(REPOS, ice.OptionFields("name,time"))

				case TAGS, VERSION:
					if m.Option(TAGS) == ice.ErrWarn {
						m.Push(VERSION, "v0.0.1")
					} else {
						m.Push(VERSION, _status_tag(m, m.Option(TAGS)))
					}

				case COMMENT:
					m.Push(kit.MDB_TEXT, m.Option(kit.MDB_FILE))
					for _, v := range kit.Split(m.Option(kit.MDB_FILE), " /") {
						m.Push(kit.MDB_TEXT, v)
					}
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Action(PULL, MAKE, PUSH, TAGS)

				files, adds, dels, last := _status_list(m)
				m.Status("files", files, "adds", adds, "dels", dels, "last", last.Format(ice.MOD_TIME))
				m.Toast3s(kit.Format("files: %d, adds: %d, dels: %d", files, adds, dels), ice.CONTEXTS)
				return
			}

			m.Option(cli.CMD_DIR, _repos_path(arg[0]))
			m.Echo(m.Cmdx(cli.SYSTEM, GIT, DIFF))
			m.Action(COMMIT)

			files, adds, dels := _status_stat(m, 0, 0, 0)
			m.Status("files", files, "adds", adds, "dels", dels)
			m.Toast3s(kit.Format("files: %d, adds: %d, dels: %d", files, adds, dels), arg[0])
		}},
	}})
}
