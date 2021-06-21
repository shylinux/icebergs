package yac

import (
	"fmt"
	"strconv"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/lex"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

type Point struct {
	s int
	c int
}
type State struct {
	star int
	next int
	hash int
}

type Matrix struct {
	nlang int
	ncell int

	page map[string]int
	hand map[int]string
	hash map[string]int
	word map[int]string

	mat [][]*State

	lex_key string
	lex     *lex.Matrix
}

func NewMatrix(m *ice.Message, nlang, ncell int) *Matrix {
	mat := &Matrix{nlang: nlang, ncell: ncell}
	mat.page = map[string]int{}
	mat.hand = map[int]string{}
	mat.hash = map[string]int{}
	mat.word = map[int]string{}

	m.Option("matrix.cb", func(key string, lex *lex.Matrix) { mat.lex, mat.lex_key = lex, key })
	key := m.Cmdx("lex.matrix", mdb.CREATE, 32)
	m.Cmd("lex.matrix", mdb.INSERT, key, "space", "space", "[\t \n]+")

	mat.mat = make([][]*State, nlang)
	return mat
}

func (mat *Matrix) name(which map[int]string, index int) string {
	if name, ok := which[index]; ok {
		return name
	}
	return fmt.Sprintf("m%d", index)
}
func (mat *Matrix) index(m *ice.Message, hash string, h string) int {
	which, names := mat.hash, mat.word
	if hash == NPAGE {
		which, names = mat.page, mat.hand
	}

	if x, e := strconv.Atoi(h); e == nil {
		if hash == NPAGE {
			m.Assert(x <= len(mat.page))
		} else {
			mat.hash[h] = x
		}
		return x
	}

	if x, ok := which[h]; ok {
		return x
	}

	m.Assert(hash != NPAGE || len(which)+1 < mat.nlang)
	which[h], names[len(which)+1] = len(which)+1, h
	return which[h]
}
func (mat *Matrix) isVoid(page int) bool {
	for j := 1; j < len(mat.mat[page]); j++ {
		if mat.mat[page][j] != nil {
			return false
		}
	}
	return true
}
func (mat *Matrix) train(m *ice.Message, page, hash int, word []string, level int) (int, []*Point, []*Point) {
	m.Debug("%s %s\\%d page: %v hash: %v word: %v", TRAIN, strings.Repeat("#", level), level, page, hash, word)

	ss := []int{page}
	sn := make([]bool, len(mat.mat))
	points, ends := []*Point{}, []*Point{}

	for i, mul := 0, false; i < len(word); i++ {
		if !mul {
			if hash <= 0 && word[i] == "}" {
				m.Debug("%s %s/%d word: %d point: %d end: %d", TRAIN, strings.Repeat("#", level), level, len(word), len(points), len(ends))
				return i + 2, points, ends
			}
			ends = ends[:0]
		}

		for _, s := range ss {
			switch word[i] {
			case "opt{", "rep{":
				sn[s] = true
				num, point, end := mat.train(m, s, 0, word[i+1:], level+1)
				points = append(points, point...)

				for _, p := range end {
					state := mat.mat[p.s][p.c]
					if len(sn) <= state.next {
						sn = append(sn, make([]bool, state.next-len(sn)+1)...)
					}
					sn[state.next] = true

					if points = append(points, p); word[i] == "rep{" {
						state.star = s
						m.Debug("REP(%d, %d): %v", p.s, p.c, state)
					}
				}
				i += num - 1
			case "mul{":
				mul = true
				goto next
			case "}":
				if mul {
					mul = false
					goto next
				}
				fallthrough
			default:
				c, ok := mat.page[word[i]]
				if !ok {
					if c, _, _ = mat.lex.Parse(m, mat.name(mat.hand, s), []byte(word[i])); c == 0 {
						// c = mat.lex.Train(m, mat.name(s), fmt.Sprintf("%d", len(mat.mat[s])+1), []byte(word[i]))
						c = kit.Int(m.Cmdx("lex.matrix", mdb.INSERT, mat.lex_key, mat.name(mat.hand, s), len(mat.mat[s]), word[i]))
						mat.mat[s] = append(mat.mat[s], nil)
					}
				}

				state := mat.mat[s][c]
				if state == nil {
					state = &State{}
				}
				m.Debug("GET(%d,%d): %#v", s, c, state)

				if state.next == 0 {
					state.next = len(mat.mat)
					mat.mat = append(mat.mat, make([]*State, mat.ncell))
					sn = append(sn, true)
				} else {
					sn[state.next] = true
				}

				mat.mat[s][c] = state
				ends = append(ends, &Point{s, c})
				points = append(points, &Point{s, c})
				m.Debug("SET(%d,%d): %#v", s, c, state)
			}
		}
	next:
		if !mul {
			ss = ss[:0]
			for s, b := range sn {
				if sn[s] = false; b {
					ss = append(ss, s)
				}
			}
		}
	}

	trans := map[int]int{page: page}
	for i := mat.nlang; i < len(mat.mat); i++ { // 去空
		if !mat.isVoid(i) {
			trans[i] = i
			continue
		}

		for j := i + 1; j < len(mat.mat); j++ {
			if !mat.isVoid(j) {
				mat.mat[i] = mat.mat[j]
				mat.mat[j] = nil
				trans[j] = i
				break
			}
		}
		if mat.isVoid(i) {
			mat.mat = mat.mat[:i]
			break
		}
	}
	m.Debug("DEL: %v", trans)

	for _, p := range points { // 去尾
		p.s = trans[p.s]
		state := mat.mat[p.s][p.c]
		m.Debug("GET(%d, %d): %#v", p.s, p.c, state)
		if state.next = trans[state.next]; state.next == 0 {
			state.hash = hash
		}
		m.Debug("SET(%d, %d): %#v", p.s, p.c, state)
	}

	m.Debug("%s %s/%d word: %d point: %d end: %d", TRAIN, strings.Repeat("#", level), level, len(word), len(points), len(ends))
	return len(word), points, ends
}

func (mat *Matrix) Parse(m *ice.Message, rewrite Rewrite, page int, line []byte, level int) (hash int, word []string, rest []byte) {
	m.Debug("%s %s\\%d %s(%d): %s", PARSE, strings.Repeat("#", level), level, mat.name(mat.hand, page), page, string(line))

	rest = line
	h, w, r := 0, []byte{}, []byte{}
	for p, i := 0, page; i > 0 && len(rest) > 0; {
		// 解析空白
		h, w, r = mat.lex.Parse(m, "space", rest)
		// 解析单词
		h, w, r = mat.lex.Parse(m, mat.name(mat.hand, i), r)
		// 解析状态
		var s *State
		if h < len(mat.mat[i]) {
			s = mat.mat[i][h]
		}

		if s != nil { // 全局语法检查
			if hh, ww, _ := mat.lex.Parse(m, "key", rest); hh == 0 || len(ww) <= len(w) {
				word, rest = append(word, string(w)), r
			} else {
				s = nil
			}
		}

		if s == nil { // 嵌套语法递归解析
			for j := 1; j < len(mat.mat[i]); j++ {
				if n := mat.mat[i][j]; j < mat.nlang && n != nil {
					if _, w, r := mat.Parse(m, rewrite, j, rest, level+1); len(r) != len(rest) {
						s, word, rest = n, append(word, w...), r
						break
					}
				}
			}
		} else {
			m.Debug("%s %s|%d GET \033[33m%s\033[0m %#v", PARSE, strings.Repeat("#", level), level, w, s)
		}

		//语法切换
		if s == nil {
			i, p = p, 0
		} else if i, p, hash = s.next, s.star, s.hash; i == 0 {
			i, p = p, 0
		}
	}

	if hash == 0 {
		word, rest = word[:0], line
	} else {
		hash, word, rest = rewrite(m, mat.word[hash], hash, word, rest)
	}

	m.Debug("%s %s/%d %s(%d): %v %v", PARSE, strings.Repeat("#", level), level, mat.hand[hash], hash, word, rest)
	return hash, word, rest
}
func (mat *Matrix) show(m *ice.Message) {
	showCol := map[int]bool{} // 有效列
	for i := 1; i < len(mat.mat); i++ {
		for j := 1; j < len(mat.mat[i]); j++ {
			if node := mat.mat[i][j]; node != nil {
				showCol[j] = true
			}
		}
	}

	for i := 1; i < len(mat.mat); i++ {
		if mat.isVoid(i) { // 无效行
			continue
		}

		m.Push("00", kit.Select(kit.Format("%02d", i), mat.hand[i]))
		for j := 1; j < len(mat.mat[i]); j++ {
			if !showCol[j] { // 无效列
				continue
			}
			key, value := kit.Format("%v", mat.name(mat.hand, j)), []string{}
			if node := mat.mat[i][j]; node != nil {
				if node.star > 0 {
					value = append(value, cli.ColorYellow(m, mat.name(mat.hand, node.star)))
				}
				if node.next > 0 {
					value = append(value, cli.ColorGreen(m, node.next))
				}
				if node.hash > 0 {
					value = append(value, cli.ColorRed(m, mat.name(mat.hand, node.hash)))
				}
			}
			m.Push(key, strings.Join(value, ","))
		}
	}

	m.Status(NLANG, mat.nlang, NCELL, mat.ncell, NPAGE, len(mat.page), NHASH, len(mat.hash))
}

type Rewrite func(m *ice.Message, nhash string, hash int, word []string, rest []byte) (int, []string, []byte)

const (
	NLANG = "nlang"
	NCELL = "ncell"
)
const (
	NPAGE = "npage"
	NHASH = "nhash"
)
const (
	TRAIN = "train"
	PARSE = "parse"
)
const MATRIX = "matrix"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			MATRIX: {Name: MATRIX, Help: "魔方矩阵", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			MATRIX: {Name: "matrix name npage text auto", Help: "魔方矩阵", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name=shy nlang=32 ncell=32", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					mat := NewMatrix(m, kit.Int(kit.Select("32", m.Option(NLANG))), kit.Int(kit.Select("32", m.Option(NCELL))))
					h := m.Rich(m.Prefix(MATRIX), "", kit.Data(
						kit.MDB_TIME, m.Time(), kit.MDB_NAME, m.Option(kit.MDB_NAME),
						MATRIX, mat, NLANG, mat.nlang, NCELL, mat.ncell,
					))
					switch cb := m.Optionv(kit.Keycb(MATRIX)).(type) {
					case func(string, *Matrix):
						cb(h, mat)
					}
					m.Echo(h)
				}},
				mdb.INSERT: {Name: "insert name=shy npage=num nhash=num text=123", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(m.Prefix(MATRIX), "", m.Option(kit.MDB_NAME), func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)

						mat, _ := value[MATRIX].(*Matrix)

						page := mat.index(m, NPAGE, m.Option(NPAGE))
						hash := mat.index(m, NHASH, m.Option(NHASH))
						if len(mat.mat[page]) == 0 {
							mat.mat[page] = make([]*State, mat.ncell)
						}

						text := kit.Split(m.Option(kit.MDB_TEXT), " ", " ", " ")
						mat.train(m, page, hash, text, 1)
						m.Grow(m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, key), kit.Dict(
							kit.MDB_TIME, m.Time(), NPAGE, m.Option(NPAGE), NHASH, m.Option(NHASH), kit.MDB_TEXT, text,
						))

						value[NPAGE] = len(mat.page)
						value[NHASH] = len(mat.hash)
					})
				}},
				mdb.REMOVE: {Name: "create", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(MATRIX), "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
				PARSE: {Name: "parse name npage text=123", Help: "解析", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(m.Prefix(MATRIX), "", m.Option(kit.MDB_NAME), func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)
						mat, _ := value[MATRIX].(*Matrix)

						for text := []byte(m.Option(kit.MDB_TEXT)); len(text) > 0; {
							hash, _, rest := mat.Parse(m, func(m *ice.Message, nhash string, hash int, word []string, rest []byte) (int, []string, []byte) {
								switch cb := m.Optionv(kit.Keycb(MATRIX)).(type) {
								case func(string, int, []string, []byte) (int, []string, []byte):
									return cb(nhash, hash, word, rest)
								}
								return hash, word, rest
							}, mat.index(m, NPAGE, m.Option(NPAGE)), text, 1)
							if text = rest; hash == 0 {
								break
							}
						}
					})
					m.ProcessInner()
				}},
				"show": {Name: "show", Help: "矩阵", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(m.Prefix(MATRIX), "", kit.Select(m.Option(kit.MDB_NAME), arg, 0), func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)
						value[MATRIX].(*Matrix).show(m)
					})
					m.ProcessInner()
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				if m.Action(mdb.CREATE); len(arg) == 0 { // 矩阵列表
					m.Fields(len(arg) == 0, "time,name,npage,nhash")
					m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), "", mdb.HASH)
					m.PushAction(mdb.INSERT, "show", mdb.REMOVE)
					return
				}

				if m.Action(mdb.INSERT, "show"); len(arg) == 1 { // 词法列表
					m.Fields(len(arg) == 1, "time,npage,nhash,text")
					m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, kit.Hashs(arg[0])), mdb.LIST)
					m.PushAction(PARSE)
					return
				}

				m.Richs(m.Prefix(MATRIX), "", arg[0], func(key string, value map[string]interface{}) {
					value = kit.GetMeta(value)
					mat, _ := value[MATRIX].(*Matrix)

					if len(arg) == 2 { // 词法矩阵
						mat.show(m)
						return
					}

					hash, word, rest := mat.Parse(m, func(m *ice.Message, nhash string, hash int, word []string, rest []byte) (int, []string, []byte) {
						m.Debug("\033[32mrun --- %v %v %v\033[0m", nhash, word, rest)
						return hash, word, rest
					}, mat.index(m, NPAGE, arg[1]), []byte(arg[2]), 1)

					m.Push(kit.MDB_TIME, m.Time())
					m.Push(kit.MDB_HASH, mat.word[hash])
					m.Push("word", kit.Format(word))
					m.Push("rest", string(rest))
				})
			}},
		},
	})
}
