package git

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	kit "github.com/shylinux/toolkits"
)

func _status_each(m *ice.Message, title string, cmds ...string) {
	m.GoToast(title, func(toast func(string, int, int)) {
		count, total := 0, len(m.Confm(REPOS, kit.MDB_HASH))
		toast("begin", count, total)

		m.Cmd(REPOS, ice.OptionFields("name,path")).Table(func(index int, value map[string]string, head []string) {
			toast(value[kit.MDB_NAME], count, total)

			m.Cmd(cmds, ice.Option{cli.CMD_DIR, value[kit.MDB_PATH]})
			count++
		})

		toast("success", count, total)
	})
}
func _status_list(m *ice.Message) {
	m.Cmd(REPOS, ice.OptionFields("name,path")).Table(func(index int, value map[string]string, head []string) {
		m.Option(cli.CMD_DIR, value[kit.MDB_PATH])
		diff := m.Cmdx(cli.SYSTEM, GIT, STATUS, "-sb")

		for _, v := range strings.Split(strings.TrimSpace(diff), "\n") {
			vs := strings.SplitN(strings.TrimSpace(v), " ", 2)
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
	})
}

const (
	PULL = "pull"
	MAKE = "make"
	PUSH = "push"

	ADD    = "add"
	DIFF   = "diff"
	COMMIT = "commit"
)
const STATUS = "status"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		STATUS: {Name: "status name auto", Help: "代码状态", Action: map[string]*ice.Action{
			PULL: {Name: "pull", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_status_each(m, PULL, cli.SYSTEM, GIT, PULL)
				m.ProcessHold()
			}},
			MAKE: {Name: "make", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				m.Toast("doing", "make", 100000)
				defer m.Toast("success", "make", 1000)

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
			}},
			COMMIT: {Name: "commit action=opt,add comment=some", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == kit.MDB_ACTION {
					m.Option(kit.MDB_TEXT, arg[1]+" "+arg[3])
				} else {
					m.Option(kit.MDB_TEXT, kit.Select("opt some", strings.Join(arg, " ")))
				}

				m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))
				m.Cmdy(cli.SYSTEM, GIT, COMMIT, "-am", m.Option(kit.MDB_TEXT))
				m.ProcessBack()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				_status_list(m)
				m.Action(PULL, MAKE, PUSH)
				return
			}

			m.Echo(m.Cmdx(cli.SYSTEM, GIT, DIFF, ice.Option{cli.CMD_DIR, _repos_path(arg[0])}))
			m.Action(COMMIT)
		}},
	}})
}
