package yac

import (
	"fmt"
	"strconv"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/lex"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

type Seed struct {
	page int
	hash int
	word []string
}
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

	state map[State]*State
	mat   [][]*State

	lex     *lex.Matrix
	lex_key string
}

func NewMatrix(m *ice.Message, nlang, ncell int) *Matrix {
	mat := &Matrix{nlang: nlang, ncell: ncell}
	mat.page = map[string]int{}
	mat.hand = map[int]string{}
	mat.hash = map[string]int{}
	mat.word = map[int]string{}

	m.Option("matrix.cb", func(key string, lex *lex.Matrix) { mat.lex, mat.lex_key = lex, key })
	key := m.Cmdx("lex.matrix", mdb.CREATE, 32, 256)
	m.Cmd("lex.matrix", mdb.INSERT, key, "space", "space", "[\t \n]")

	mat.state = make(map[State]*State)
	mat.mat = make([][]*State, nlang)
	return mat
}

func (mat *Matrix) name(page int) string {
	if name, ok := mat.word[page]; ok {
		return name
	}
	return fmt.Sprintf("m%d", page)
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

	if hash == NPAGE {
		which[h] = len(mat.page) + 1
	} else {
		which[h] = len(mat.hash) + 1
	}

	names[which[h]] = h
	m.Assert(hash != NPAGE || len(mat.page) < mat.nlang)
	return which[h]
}
func (mat *Matrix) train(m *ice.Message, page, hash int, word []string, level int) (int, []*Point, []*Point) {
	m.Debug("%s %s\\%d page: %v hash: %v word: %v", TRAIN, strings.Repeat("#", level), level, page, hash, word)

	ss := []int{page}
	sn := make([]bool, len(mat.mat))
	points, ends := []*Point{}, []*Point{}

	for i, mul := 0, false; i < len(word); i++ {
		if !mul {
			if hash <= 0 && word[i] == "}" {
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
				i += num - 1

				for _, x := range end {
					state := &State{}
					*state = *mat.mat[x.s][x.c]
					for i := len(sn); i <= state.next; i++ {
						sn = append(sn, false)
					}
					sn[state.next] = true

					points = append(points, x)
					if word[i] == "rep{" {
						state.star = s
						mat.mat[x.s][x.c] = state
						m.Debug("REP(%d, %d): %v", x.s, x.c, state)
					}
				}
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
				x, ok := mat.page[word[i]]
				if !ok {
					if x, _, _ = mat.lex.Parse(m, mat.name(s), []byte(word[i])); x == 0 {
						// x = mat.lex.Train(m, mat.name(s), fmt.Sprintf("%d", len(mat.mat[s])+1), []byte(word[i]))
						x = kit.Int(m.Cmdx("lex.matrix", mdb.INSERT, mat.lex_key, mat.name(s), len(mat.mat[s]), word[i]))
						mat.mat[s] = append(mat.mat[s], nil)
					}
				}

				c := x
				state := &State{}
				if mat.mat[s][c] != nil {
					*state = *mat.mat[s][c]
				}
				m.Debug("GET(%d,%d): %v", s, c, state)

				if state.next == 0 {
					state.next = len(mat.mat)
					mat.mat = append(mat.mat, make([]*State, mat.ncell))
					sn = append(sn, false)
				}
				sn[state.next] = true

				mat.mat[s][c] = state
				m.Debug("SET(%d,%d): %v", s, c, state)
				ends = append(ends, &Point{s, c})
				points = append(points, &Point{s, c})
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

	for _, s := range ss {
		if s < mat.nlang || s >= len(mat.mat) {
			continue
		}
		void := true
		for _, x := range mat.mat[s] {
			if x != nil {
				void = false
				break
			}
		}
		if void {
			mat.mat = mat.mat[:s]
			m.Debug("DEL: %d", len(mat.mat))
		}
	}

	for _, s := range ss {
		for _, p := range points {
			state := &State{}
			*state = *mat.mat[p.s][p.c]

			if state.next == s {
				m.Debug("GET(%d, %d): %v", p.s, p.c, state)
				if state.next >= len(mat.mat) {
					state.next = 0
				}
				if hash > 0 {
					state.hash = hash
				}
				mat.mat[p.s][p.c] = state
				m.Debug("SET(%d, %d): %v", p.s, p.c, state)
			}
			if x, ok := mat.state[*state]; !ok {
				mat.state[*state] = mat.mat[p.s][p.c]
			} else {
				mat.mat[p.s][p.c] = x
			}
		}
	}

	m.Debug("%s %s/%d word: %d point: %d end: %d", TRAIN, strings.Repeat("#", level), level, len(word), len(points), len(ends))
	return len(word), points, ends
}

func (mat *Matrix) Parse(m *ice.Message, rewrite Rewrite, page int, line []byte, level int) (hash int, word []string, rest []byte) {
	// m.Debug("%s %s\\%d %s(%d): %s", PARSE, strings.Repeat("#", level), level, mat.name(page), page, string(line))

	rest = line
	h, w, r := 0, []byte{}, []byte{}
	for p, i := 0, page; i > 0 && len(rest) > 0; {
		// 解析空白
		_, _, r = mat.lex.Parse(m, "space", rest)
		// 解析单词
		h, w, r = mat.lex.Parse(m, mat.name(i), r)
		// 解析状态
		s := mat.mat[i][h]

		if s != nil { // 全局语法检查
			if hh, ww, _ := mat.lex.Parse(m, "key", rest); hh == 0 || len(ww) <= len(w) {
				word, rest = append(word, string(w)), r
			} else {
				s = nil
			}
		}

		if s == nil { // 嵌套语法递归解析
			for j := 0; j < mat.ncell; j++ {
				if n := mat.mat[i][j]; j < mat.nlang && n != nil {
					if _, w, r := mat.Parse(m, rewrite, j, rest, level+1); len(r) != len(rest) {
						s, word, rest = n, append(word, w...), r
						break
					}
				}
			}
		} else {
			// m.Debug("%s %s|%d GET \033[33m%s\033[0m", PARSE, strings.Repeat("#", level), level, w)
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
		hash, word, rest = rewrite(m, mat.hand[hash], hash, word, rest)
	}

	// m.Debug("%s %s/%d %s(%d): %v %v", PARSE, strings.Repeat("#", level), level, mat.hand[hash], hash, word, rest)
	return hash, word, rest
}
func (mat *Matrix) show(m *ice.Message) {
	max := mat.ncell
	for i := 1; i < len(mat.mat); i++ {
		if len(mat.mat[i]) > max {
			max = len(mat.mat[i])
		}
	}
	for i := 1; i < len(mat.mat); i++ {
		if len(mat.mat[i]) == 0 {
			continue
		}

		m.Push("00", kit.Select(kit.Format("%02d", i), mat.hand[i]))
		for j := 1; j < max; j++ {
			if j > len(mat.page) && j < mat.ncell {
				continue
			}
			key := kit.Select(kit.Format("w%02d", j), mat.hand[j])
			if j < len(mat.mat[i]) {
				if node := mat.mat[i][j]; node != nil {
					if node.next == 0 {
						m.Push(key, mat.word[node.hash])
					} else {
						m.Push(key, kit.Select(kit.Format("%02d", node.next), mat.hand[node.next]))
					}
					continue
				}
			}
			m.Push(key, "")
		}
	}
}

type Rewrite func(m *ice.Message, nhash string, hash int, word []string, rest []byte) (int, []string, []byte)

const (
	NLANG = "nlang"
	NCELL = "ncell"
	NSEED = "nseed"
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
					h := m.Rich(m.Prefix(MATRIX), "", kit.Data(kit.MDB_TIME, m.Time(), kit.MDB_NAME, m.Option(kit.MDB_NAME), MATRIX, mat, NLANG, mat.nlang, NCELL, mat.ncell))
					switch cb := m.Optionv("matrix.cb").(type) {
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

						mat.train(m, page, hash, kit.Split(m.Option(kit.MDB_TEXT)), 1)
						m.Grow(m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, key), kit.Dict(
							kit.MDB_TIME, m.Time(), NPAGE, m.Option(NPAGE), NHASH, m.Option(NHASH), kit.MDB_TEXT, m.Option(kit.MDB_TEXT),
						))

						value[NPAGE] = len(mat.page)
						value[NHASH] = len(mat.hash)
					})
				}},
				mdb.REMOVE: {Name: "create", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(MATRIX), "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
				"show": {Name: "show", Help: "矩阵", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(m.Prefix(MATRIX), "", m.Option(kit.MDB_NAME), func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)
						mat, _ := value[MATRIX].(*Matrix)
						mat.show(m)
					})
					m.ProcessInner()
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				if m.Action(mdb.CREATE); len(arg) == 0 { // 矩阵列表
					m.Fields(len(arg) == 0, "time,name,npage,nhash")
					m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), "", mdb.HASH)
					m.PushAction("show", mdb.INSERT, mdb.REMOVE)
					return
				}

				if m.Action(mdb.INSERT); len(arg) == 1 { // 词法列表
					m.Fields(len(arg) == 1, "time,npage,nhash,text")
					m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, kit.Hashs(arg[0])), mdb.LIST)
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
