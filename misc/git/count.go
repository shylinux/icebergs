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
	if m.Warn(len(arg) == 0 || arg[0] == "usr/", ice.ErrNotValid, nfs.DIR, "to many files, please select sub dir") {
		return
	}
	nfs.DirDeepAll(m, "", arg[0], func(value ice.Maps) {
		file := value[nfs.PATH]
		for _, p := range []string{"node_modules/", "bin/", "var/", "tags"} {
			if strings.Contains(file, p) {
				return
			}
		}
		switch kit.Ext(file) {
		case "sum", "log":
			return
		}
		cb(file)
	}, nfs.PATH)
}

const COUNT = "count"

func init() {
	Index.MergeCommands(ice.Commands{
		COUNT: {Name: "count path auto count order tags", Help: "代码行", Actions: ice.Actions{
			COUNT: {Help: "计数", Hand: func(m *ice.Message, arg ...string) {
				files := map[string]int{}
				lines := map[string]int{}
				_count_count(m, arg, func(file string) {
					files[mdb.TOTAL]++
					files[kit.Ext(file)]++
					m.Cmdy(nfs.CAT, file, func(text string) {
						if kit.Ext(file) == code.GO {
							switch {
							case strings.HasPrefix(text, "func "):
								lines["_func"]++
							case strings.HasPrefix(text, "type "):
								lines["_type"]++
							}
						}
						lines[mdb.TOTAL]++
						lines[kit.Ext(file)]++
					})
				})
				kit.Fetch(lines, func(k string, v int) { m.Push(mdb.TYPE, k).Push("files", files[k]).Push("lines", lines[k]) })
				m.StatusTimeCount().SortIntR("lines")
			}},
			"order": {Help: "排行", Hand: func(m *ice.Message, arg ...string) {
				files := map[string]int{}
				_count_count(m, arg, func(file string) {
					m.Cmdy(nfs.CAT, file, func(text string) { files[strings.TrimPrefix(file, arg[0])]++ })
				})
				kit.Fetch(files, func(k string, v int) { m.Push("files", k).Push("lines", v) })
				m.StatusTimeCount().SortIntR("lines")
			}},
			"tags": {Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				count := map[string]int{}
				m.Cmd(nfs.CAT, path.Join(arg[0], nfs.TAGS), func(line string) {
					if ls := strings.SplitN(line, ice.TB, 3); len(ls) < 3 {
						return
					} else if ls = strings.SplitN(ls[2], ";\"", 2); len(ls) < 2 {
						return
					} else {
						count[kit.Split(ls[1])[0]]++
					}
				})
				kit.Fetch(count, func(k string, v int) { m.Push(mdb.TYPE, k).Push(mdb.COUNT, v) })
				m.StatusTimeCount().SortIntR(mdb.COUNT)
			}},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg) }},
	})
}
