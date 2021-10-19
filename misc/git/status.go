package git

import (
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _status_each(m *ice.Message, title string, cmds ...string) {
	m.GoToast(title, func(toast func(string, int, int)) {
		count, total := 0, len(m.Confm(REPOS, kit.MDB_HASH))
		toast(cli.BEGIN, count, total)

		list := []string{}
		m.Cmd(REPOS, ice.OptionFields("name,path")).Table(func(index int, value map[string]string, head []string) {
			toast(value[kit.MDB_NAME], count, total)

			if msg := m.Cmd(cmds, ice.Option{cli.CMD_DIR, value[kit.MDB_PATH]}); msg.Append(cli.CMD_CODE) != "0" {
				m.Toast(msg.Append(cli.CMD_ERR), "error: "+value[kit.MDB_NAME], "3s")
				list = append(list, value[kit.MDB_NAME])
				m.Sleep("3s")
			}
			count++
		})

		if len(list) > 0 {
			m.Toast(strings.Join(list, ice.NL), ice.FAILURE, "30s")
		} else {
			toast(ice.SUCCESS, count, total)
		}
	})
}
func _status_stat(m *ice.Message, files, adds, dels int) (int, int, int) {
	for _, v := range kit.Split(m.Cmdx(cli.SYSTEM, GIT, DIFF, "--shortstat"), ",", ",") {
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
			vs := strings.SplitN(strings.TrimSpace(v), ice.SP, 2)
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
				m.Toast(ice.PROCESS, MAKE, "30s")
				defer m.Toast(ice.SUCCESS, MAKE, "3s")
				m.Cmdy(cli.SYSTEM, MAKE)
			}},
			TAGS: {Name: "tags", Help: "标签", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessHold()
				vs := map[string]string{}
				m.Cmd(STATUS).Table(func(index int, value map[string]string, head []string) {
					if value["type"] == "##" {
						if value["name"] == "release" {
							value["name"] = "ice"
						}
						vs[value["name"]] = strings.Split(value["tags"], "-")[0]
					}
				})

				m.GoToast(TAGS, func(toast func(string, int, int)) {
					count, total := 0, len(vs)
					toast(cli.BEGIN, count, total)

					for k := range vs {
						count++
						toast(k, count, total)

						change := false
						m.Option(nfs.DIR_ROOT, _repos_path(k))
						mod := m.Cmdx(nfs.CAT, ice.GO_MOD, func(text string, line int) string {
							ls := kit.Split(text)
							if len(ls) < 2 || ls[1] == "=>" {
								return text
							}
							if v, ok := vs[kit.Slice(strings.Split(ls[0], "/"), -1)[0]]; ok {
								text = ice.TB + ls[0] + ice.SP + v
								change = true
							}
							return text
						})

						if change {
							m.Cmd(nfs.SAVE, ice.GO_SUM, "")
							m.Cmd(nfs.SAVE, ice.GO_MOD, mod)
							m.Option(cli.CMD_DIR, _repos_path(k))
							if k == ice.CONTEXTS {
								continue
							}
							if k == ice.ICEBERGS {
								m.Cmd(cli.SYSTEM, "go", "build")
								continue
							}
							m.Cmd(cli.SYSTEM, cli.MAKE)
						}
					}
					toast(ice.SUCCESS, count, total)
				})
			}},
			PUSH: {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_NAME) == "" {
					_status_each(m, PUSH, cli.SYSTEM, GIT, PUSH)
					m.ProcessHold()
					return
				}

				m.Cmdy(cli.SYSTEM, GIT, PUSH, ice.Option{cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME))})
				m.Cmdy(cli.SYSTEM, GIT, PUSH, "--tags")
			}},

			TAG: {Name: "tag version@key", Help: "标签", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))
				m.Cmdy(cli.SYSTEM, GIT, TAG, m.Option(VERSION))
				m.Cmdy(cli.SYSTEM, GIT, PUSH, "--tags")
			}},
			ADD: {Name: "add", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, GIT, ADD, m.Option(kit.MDB_FILE), ice.Option{cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME))})
			}}, OPT: {Name: "opt", Help: "优化"}, PRO: {Name: "pro", Help: "升级"},
			COMMIT: {Name: "commit action=opt,add,pro comment=some@key", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == ctx.ACTION {
					m.Option(kit.MDB_TEXT, arg[1]+ice.SP+arg[3])
				} else {
					m.Option(kit.MDB_TEXT, kit.Select("opt some", strings.Join(arg, ice.SP)))
				}

				m.Cmdy(cli.SYSTEM, GIT, COMMIT, "-am", m.Option(kit.MDB_TEXT), ice.Option{cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME))})
				m.ProcessBack()
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_NAME:
					m.Cmdy(REPOS, ice.OptionFields("name,time"))

				case TAGS, VERSION:
					if m.Option(TAGS) == ice.ErrWarn {
						m.Push(VERSION, kit.Format("v0.0.%d", 1))
						return
					}

					ls := kit.Split(strings.TrimPrefix(kit.Split(m.Option(TAGS), "-")[0], "v"), ".")
					if v := kit.Int(ls[2]); v < 9 {
						m.Push(VERSION, kit.Format("v%v.%v.%v", ls[0], ls[1], v+1))
					} else if v := kit.Int(ls[1]); v < 9 {
						m.Push(VERSION, kit.Format("v%v.%v.0", ls[0], v+1))
					} else if v := kit.Int(ls[0]); v < 9 {
						m.Push(VERSION, kit.Format("v%v.0.0", v+1))
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
