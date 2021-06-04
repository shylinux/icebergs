package git

import (
	"strings"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _status_each(m *ice.Message, title string, cmds ...string) {
	m.GoToast(title, func(toast func(string, int, int)) {
		count, total := 0, len(m.Confm(REPOS, kit.MDB_HASH))
		toast("begin", count, total)

		list := []string{}
		m.Cmd(REPOS, ice.OptionFields("name,path")).Table(func(index int, value map[string]string, head []string) {
			toast(value[kit.MDB_NAME], count, total)

			msg := m.Cmd(cmds, ice.Option{cli.CMD_DIR, value[kit.MDB_PATH]})
			if msg.Append(cli.CMD_CODE) != "0" {
				list = append(list, value[kit.MDB_NAME])
				m.Toast(msg.Append(cli.CMD_ERR), "error: "+value[kit.MDB_NAME], "3s")
				m.Sleep("3s")
			}
			count++
		})

		if len(list) > 0 {
			m.Toast(strings.Join(list, "\n"), "failure", "30s")
		} else {
			toast("success", count, total)
		}

	})
}
func _status_stat(m *ice.Message, files, adds, dels int) (int, int, int) {
	ls := kit.Split(m.Cmdx(cli.SYSTEM, GIT, DIFF, "--shortstat"), ",", ",")
	for _, v := range ls {
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
	m.Cmd(REPOS, ice.OptionFields("name,path")).Table(func(index int, value map[string]string, head []string) {
		m.Option(cli.CMD_DIR, value[kit.MDB_PATH])
		diff := m.Cmdx(cli.SYSTEM, GIT, STATUS, "-sb")

		for _, v := range strings.Split(strings.TrimSpace(diff), "\n") {
			vs := strings.SplitN(strings.TrimSpace(v), " ", 2)
			switch kit.Ext(vs[1]) {
			case "swp", "swo", "bin":
				continue
			}

			m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
			m.Push(kit.MDB_TYPE, vs[0])
			m.Push(kit.MDB_FILE, vs[1])

			list := []string{}
			switch vs[0] {
			case "##":
				if strings.Contains(vs[1], "ahead") {
					list = append(list, PUSH)
				}
			default:
				if strings.Contains(vs[0], "??") {
					list = append(list, ADD)
				} else {
					list = append(list, COMMIT)
				}
			}
			m.PushButton(strings.Join(list, ","))
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

	ADD = "add"
	OPT = "opt"
	PRO = "pro"

	DIFF    = "diff"
	COMMIT  = "commit"
	COMMENT = "comment"
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
				m.Toast("building", MAKE, 100000)
				defer m.Toast("success", MAKE, 1000)

				m.Cmdy(cli.SYSTEM, MAKE)
			}},
			PUSH: {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_NAME) == "" {
					_status_each(m, PUSH, cli.SYSTEM, GIT, PUSH)
					m.ProcessHold()
					return
				}

				m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))
				m.Cmdy(cli.SYSTEM, GIT, PUSH)
			}},

			ADD: {Name: "add", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))
				m.Cmdy(cli.SYSTEM, GIT, ADD, m.Option(kit.MDB_FILE))
			}}, OPT: {Name: "opt", Help: "优化"}, PRO: {Name: "pro", Help: "自举"},

			COMMIT: {Name: "commit action=opt,add,pro comment=some@key", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == kit.MDB_ACTION {
					m.Option(kit.MDB_TEXT, arg[1]+" "+arg[3])
				} else {
					m.Option(kit.MDB_TEXT, kit.Select("opt some", strings.Join(arg, " ")))
				}

				m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))
				m.Cmdy(cli.SYSTEM, GIT, COMMIT, "-am", m.Option(kit.MDB_TEXT))
				m.ProcessBack()
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_NAME:
					m.Cmdy(REPOS, ice.OptionFields("name,time"))

				case COMMENT:
					ls := []string{}
					ls = append(ls, kit.Split(m.Option(kit.MDB_FILE), " /")...)

					m.Push(kit.MDB_TEXT, m.Option(kit.MDB_FILE))
					for _, v := range ls {
						m.Push(kit.MDB_TEXT, v)
					}
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Action(PULL, MAKE, PUSH)

				files, adds, dels, last := _status_list(m)
				m.Status("files", files, "adds", adds, "dels", dels, "last", last.Format(ice.MOD_TIME))
				m.Toast(kit.Format("files: %d, adds: %d, dels: %d", files, adds, dels), ice.CONTEXTS, "3s")
				return
			}

			m.Option(cli.CMD_DIR, _repos_path(arg[0]))
			m.Echo(m.Cmdx(cli.SYSTEM, GIT, DIFF))
			m.Action(COMMIT)

			files, adds, dels := _status_stat(m, 0, 0, 0)
			m.Status("files", files, "adds", adds, "dels", dels)
			m.Toast(kit.Format("files: %d, adds: %d, dels: %d", files, adds, dels), arg[0], "3s")
		}},
	}})
}
