package git

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _count_count(m *ice.Message, arg []string, cb func(string)) {
	if len(arg) == 0 || arg[0] == "usr/" {
		m.Echo("to many file, please choice sub dir")
		return
	}
	m.Option(nfs.DIR_DEEP, ice.TRUE)
	m.Option(nfs.DIR_TYPE, nfs.TYPE_CAT)
	m.Cmdy(nfs.DIR, arg, func(file string) {
		if strings.Contains(file, "node_modules/") {
			return
		}
		if strings.Contains(file, "bin/") {
			return
		}
		if strings.Contains(file, "var/") {
			return
		}
		if strings.Contains(file, "tags") {
			return
		}
		switch kit.Ext(file) {
		case "sum", "log":
			return
		}
		cb(file)
	})
}

const COUNT = "count"

func init() {
	Index.MergeCommands(ice.Commands{
		COUNT: {Name: "count path auto count order tags", Help: "代码行", Actions: ice.Actions{
			"order": {Name: "order", Help: "排行", Hand: func(m *ice.Message, arg ...string) {
				files := map[string]int{}
				_count_count(m, arg, func(file string) {
					m.Cmdy(nfs.CAT, file, func(text string, line int) {
						files[strings.TrimPrefix(file, arg[0])]++
					})
				})
				for k, n := range files {
					m.Push("files", k)
					m.Push("lines", n)
				}
				m.StatusTimeCount().SortIntR("lines")
			}},
			"tags": {Name: "tags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				count := map[string]int{}
				m.Cmd(nfs.CAT, path.Join(arg[0], "tags"), func(line string) {
					ls := strings.SplitN(line, ice.TB, 3)
					if len(ls) < 3 {
						return
					}
					ls = strings.SplitN(ls[2], ";\"", 2)
					if len(ls) < 2 {
						return
					}
					ls = kit.Split(ls[1])
					count[ls[0]]++
				})
				for k, v := range count {
					m.Push("type", k)
					m.Push("count", v)
				}
				m.SortIntR("count")
			}},
			COUNT: {Name: "count", Help: "计数", Hand: func(m *ice.Message, arg ...string) {
				files := map[string]int{}
				lines := map[string]int{}
				_count_count(m, arg, func(file string) {
					files[mdb.TOTAL]++
					files[kit.Ext(file)]++
					m.Cmdy(nfs.CAT, file, func(text string, line int) {
						if kit.Ext(file) == code.GO {
							switch {
							case strings.HasPrefix(text, "func"):
								lines["_func"]++
							case strings.HasPrefix(text, "type"):
								lines["_type"]++
							}
						}

						lines[mdb.TOTAL]++
						lines[kit.Ext(file)]++
					})
				})
				for k := range lines {
					m.Push(mdb.TYPE, k)
					m.Push("files", files[k])
					m.Push("lines", lines[k])
				}
				m.StatusTime().SortIntR("lines")
			}},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg) }},
	})
}
