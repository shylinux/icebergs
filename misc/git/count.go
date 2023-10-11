package git

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _count_count(m *ice.Message, arg []string, cb func(string)) {
	if m.Warn(len(arg) == 0 || arg[0] == nfs.USR, ice.ErrNotValid, nfs.DIR, "to many files, please select sub dir") {
		return
	}
	nfs.DirDeepAll(m, "", arg[0], func(value ice.Maps) {
		if file := value[nfs.PATH]; kit.Contains(file, nfs.BIN, nfs.VAR, "node_modules/") {
			return
		} else if kit.IsIn(kit.Ext(file), "tags", "sum", "log") {
			return
		} else {
			cb(file)
		}
	}, nfs.PATH)
}

const COUNT = "count"

func init() {
	const (
		FILES = "files"
		LINES = "lines"
	)
	Index.MergeCommands(ice.Commands{
		COUNT: {Name: "count path@key auto order count package tags", Help: "代码行", Actions: ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, path.Dir(kit.Select(nfs.PWD, arg[1]))).CutTo(nfs.PATH, arg[0])
			}},
			cli.ORDER: {Help: "排行", Hand: func(m *ice.Message, arg ...string) {
				files := map[string]int{}
				_count_count(m, arg, func(file string) {
					m.Cmdy(nfs.CAT, file, func(text string) { files[strings.TrimPrefix(file, arg[0])]++ })
				})
				kit.For(files, func(k string, v int) { m.Push(FILES, k).Push(LINES, v) })
				m.SortIntR(LINES)
			}},
			COUNT: {Help: "计数", Hand: func(m *ice.Message, arg ...string) {
				files, lines := map[string]int{}, map[string]int{}
				_count_count(m, arg, func(file string) {
					files[mdb.TOTAL]++
					files[kit.Ext(file)]++
					m.Cmdy(nfs.CAT, file, func(text string) {
						if kit.Ext(file) == code.GO {
							switch {
							case strings.HasPrefix(text, "type "):
								lines["_type"]++
							case strings.HasPrefix(text, "func "):
								lines["_func"]++
							}
						}
						lines[kit.Ext(file)]++
						lines[mdb.TOTAL]++
					})
				})
				kit.For(lines, func(k string, v int) { m.Push(mdb.TYPE, k).Push(FILES, files[k]).Push(LINES, lines[k]) })
				m.SortIntR(LINES)
			}},
			code.PACKAGE: {Help: "依赖", Hand: func(m *ice.Message, arg ...string) {
				list := map[string]map[string]int{}
				ls := map[string]int{}
				pkg, block := "", false
				add := func(mod string) {
					if _, ok := list[pkg]; !ok {
						list[pkg] = map[string]int{}
					}
					kit.If(mod, func() { list[pkg][mod]++; ls[mod]++ })
				}
				_count_count(m, arg, func(file string) {
					m.Cmdy(nfs.CAT, file, func(text string) {
						if kit.Ext(file) == code.GO {
							switch {
							case strings.HasPrefix(text, "package "):
								pkg = kit.Split(text)[1]
							case strings.HasPrefix(text, "import ("):
								block = true
							case strings.HasPrefix(text, "import "):
								add(kit.Select("", kit.Split(text), -1))
							case strings.HasPrefix(text, ")"):
								block = false
							default:
								kit.If(block, func() { add(kit.Select("", kit.Split(text), -1)) })
							}
						}
					})
				})
				m.Appendv(ice.MSG_APPEND, []string{code.PACKAGE, mdb.COUNT})
				kit.For(ls, func(key string, value int) {
					if !strings.Contains(key, "shylinux.com") {
						return
					}
					count := 0
					m.Push(code.PACKAGE, key)
					kit.For(kit.SortedKey(list), func(k string) {
						if n := list[k][key]; n == 0 {
							m.Push(k, "")
						} else {
							m.Push(k, n)
							count++
						}
					})
					m.Push(mdb.COUNT, count)
				})
				m.SortIntR(mdb.COUNT)
			}},
			nfs.TAGS: {Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				count := map[string]int{}
				m.Cmd(nfs.CAT, path.Join(arg[0], nfs.TAGS), func(line string) {
					if ls := strings.SplitN(line, lex.TB, 3); len(ls) < 3 {
						return
					} else if ls = strings.SplitN(ls[2], ";\"", 2); len(ls) < 2 {
						return
					} else {
						count[kit.Split(ls[1])[0]]++
					}
				})
				kit.For(count, func(k string, v int) { m.Push(mdb.TYPE, k).Push(mdb.COUNT, v) })
				m.SortIntR(mdb.COUNT)
			}},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg) }},
	})
}
