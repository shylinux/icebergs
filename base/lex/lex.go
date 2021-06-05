package lex

import (
	"sort"
	"strconv"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

type Seed struct {
	page int
	hash int
	word string
}
type Point struct {
	s int
	c byte
}
type State struct {
	star bool
	next int
	hash int
}
type Matrix struct {
	nlang int
	ncell int

	seed []*Seed
	page map[string]int
	hand map[int]string
	hash map[string]int
	word map[int]string

	trans map[byte][]byte
	state map[State]*State
	mat   []map[byte]*State

	*ice.Context

	nseed int
	npage int
	nhash int
	nline int
	nnode int
	nreal int
}

func NewMatrix(m *ice.Message, nlang, ncell int) *Matrix {
	mat := &Matrix{}
	mat.nlang = nlang
	mat.ncell = ncell

	mat.page = map[string]int{"nil": 0}
	mat.hand = map[int]string{0: "nil"}
	mat.hash = map[string]int{"nil": 0}
	mat.word = map[int]string{0: "nil"}

	mat.trans = map[byte][]byte{}
	for k, v := range map[byte]string{
		't': "\t", 'n': "\n", 'b': "\t ", 's': "\t \n",
		'd': "0123456789", 'x': "0123456789ABCDEFabcdef",
	} {
		mat.trans[k] = []byte(v)
	}

	mat.state = make(map[State]*State)
	mat.mat = make([]map[byte]*State, nlang)

	mat.nline = nlang
	return mat
}
func (mat *Matrix) char(c byte) []byte {
	if cs, ok := mat.trans[c]; ok {
		return cs
	}
	return []byte{c}
}
func (mat *Matrix) index(m *ice.Message, hash string, h string) int {
	which, names := mat.hash, mat.word
	if hash == NPAGE {
		which, names = mat.page, mat.hand
	}

	if x, e := strconv.Atoi(h); e == nil {
		if hash == NPAGE {
			m.Assert(x <= mat.npage)
		} else {
			mat.hash[h] = x
		}
		return x
	}

	if x, ok := which[h]; ok {
		return x
	}

	if hash == NPAGE {
		mat.npage++
		which[h] = mat.npage
	} else {
		mat.nhash++
		which[h] = mat.nhash
	}

	names[which[h]] = h
	m.Assert(hash != NPAGE || mat.npage < mat.nlang)
	return which[h]
}
func (mat *Matrix) train(m *ice.Message, page int, hash int, seed []byte) int {
	m.Debug("%s %s page: %v hash: %v seed: %v", "train", "lex", page, hash, string(seed))

	ss := []int{page}
	cn := make([]bool, mat.ncell)
	cc := make([]byte, 0, mat.ncell)
	sn := make([]bool, mat.nline)

	points := []*Point{}

	for p := 0; p < len(seed); p++ {

		switch seed[p] {
		case '[':
			set := true
			if p++; seed[p] == '^' {
				set, p = false, p+1
			}

			for ; seed[p] != ']'; p++ {
				if seed[p] == '\\' {
					p++
					for _, c := range mat.char(seed[p]) {
						cn[c] = true
					}
					continue
				}

				if seed[p+1] == '-' {
					begin, end := seed[p], seed[p+2]
					if begin > end {
						begin, end = end, begin
					}
					for c := begin; c <= end; c++ {
						cn[c] = true
					}
					p += 2
					continue
				}

				cn[seed[p]] = true
			}

			for c := 0; c < len(cn); c++ {
				if (set && cn[c]) || (!set && !cn[c]) {
					cc = append(cc, byte(c))
				}
				cn[c] = false
			}

		case '.':
			for c := 0; c < len(cn); c++ {
				cc = append(cc, byte(c))
			}

		case '\\':
			p++
			for _, c := range mat.char(seed[p]) {
				cc = append(cc, c)
			}
		default:
			cc = append(cc, seed[p])
		}

		m.Debug("page: \033[31m%d %v\033[0m", len(ss), ss)
		m.Debug("cell: \033[32m%d %v\033[0m", len(cc), cc)

		flag := '\000'
		if p+1 < len(seed) {
			switch flag = rune(seed[p+1]); flag {
			case '?', '+', '*':
				p++
			}
		}

		for _, s := range ss {
			for _, c := range cc {

				state := &State{}
				if mat.mat[s][c] != nil {
					*state = *mat.mat[s][c]
				} else {
					mat.nnode++
				}
				m.Debug("GET(%d,%d): %v", s, c, state)

				switch flag {
				case '+':
					state.star = true
				case '*':
					state.star = true
					sn[s] = true
				case '?':
					sn[s] = true
				}

				if state.next == 0 {
					mat.mat = append(mat.mat, make(map[byte]*State))
					sn = append(sn, false)
					state.next = mat.nline
					mat.nline++
				}
				sn[state.next] = true

				mat.mat[s][c] = state
				points = append(points, &Point{s, c})
				m.Debug("SET(%d,%d): %v(%d,%d)", s, c, state, mat.nnode, mat.nreal)
			}
		}

		cc, ss = cc[:0], ss[:0]
		for s, b := range sn {
			if sn[s] = false; b && s > 0 {
				ss = append(ss, s)
			}
		}
	}

	for _, s := range ss {
		if s < mat.nlang || s >= len(mat.mat) {
			continue
		}

		if len(mat.mat[s]) == 0 {
			last := mat.nline - 1
			mat.mat, mat.nline = mat.mat[:s], s
			m.Debug("DEL: %d-%d", last, mat.nline)
		}
	}

	for _, s := range ss {
		for _, p := range points {
			state := &State{}
			*state = *mat.mat[p.s][p.c]

			if state.next == s {
				m.Debug("GET(%d, %d): %v", p.s, p.c, state)
				if state.hash = hash; state.next >= len(mat.mat) {
					state.next = 0
				}
				mat.mat[p.s][p.c] = state
				m.Debug("SET(%d, %d): %v", p.s, p.c, state)
			}

			if x, ok := mat.state[*state]; !ok {
				mat.state[*state] = mat.mat[p.s][p.c]
				mat.nreal++
			} else {
				mat.mat[p.s][p.c] = x
			}
		}
	}

	m.Debug("%s %s npage: %v nhash: %v nseed: %v", "train", "lex", mat.npage, mat.nhash, len(mat.seed))
	return hash
}
func (mat *Matrix) parse(m *ice.Message, page int, line []byte) (hash int, rest []byte, word []byte) {
	m.Debug("%s %s page: %v line: %v", "parse", "lex", page, line)

	pos := 0
	for star, s := 0, page; s != 0 && pos < len(line); pos++ {

		c := line[pos]
		if c == '\\' && pos < len(line)-1 { //跳过转义
			pos++
			c = mat.char(line[pos])[0]
		}
		if c > 127 { //跳过中文
			word = append(word, c)
			continue
		}

		state := mat.mat[s][c]
		if state == nil {
			s, star, pos = star, 0, pos-1
			continue
		}
		m.Debug("GET (%d,%d): %v", s, c, state)

		word = append(word, c)

		if state.star {
			star = s
		} else if x, ok := mat.mat[star][c]; !ok || !x.star {
			star = 0
		}

		if s, hash = state.next, state.hash; s == 0 {
			s, star = star, 0
		}
	}

	if pos == len(line) {
		// hash, pos, word = -1, 0, word[:0]
	} else if hash == 0 {
		pos, word = 0, word[:0]
	}
	rest = line[pos:]

	m.Debug("%s %s hash: %v word: %v rest: %v", "parse", "lex", hash, word, rest)
	return
}
func (mat *Matrix) show(m *ice.Message, page string) {
	rows := map[int]bool{}
	cols := map[int]bool{}

	nrow := []int{mat.page[page]}
	for i := 0; i < len(nrow); i++ {
		line := nrow[i]
		rows[line] = true

		for i := 1; i < mat.ncell; i++ {
			if node := mat.mat[line][byte(i)]; node != nil {
				if cols[i] = true; node.next != 0 {
					nrow = append(nrow, node.next)
				}
			}
		}
	}

	nrow = nrow[:0]
	ncol := []int{}
	for k := range rows {
		nrow = append(nrow, k)
	}
	for k := range cols {
		ncol = append(ncol, k)
	}
	sort.Ints(nrow)
	sort.Ints(ncol)

	for _, i := range nrow {
		m.Push("0", kit.Select(kit.Format(i), mat.hand[i]))
		for _, j := range ncol {
			node := mat.mat[i][byte(j)]
			if node != nil {
				m.Push(kit.Format("%c", j), kit.Format("%v", node.next))
			} else {
				m.Push(kit.Format("%c", j), "")
			}
		}
	}
}

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

const LEX = "lex"

var Index = &ice.Context{Name: LEX, Help: "词法模块",
	Configs: map[string]*ice.Config{
		MATRIX: {Name: MATRIX, Help: "魔方矩阵", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Load()
			m.Richs(m.Prefix(MATRIX), "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				value = kit.GetMeta(value)

				mat := NewMatrix(m, kit.Int(kit.Select("32", value[NLANG])), kit.Int(kit.Select("256", value[NCELL])))
				m.Grows(m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
					page := mat.index(m, NPAGE, kit.Format(value[NPAGE]))
					hash := mat.index(m, NHASH, kit.Format(value[NHASH]))
					if mat.mat[page] == nil {
						mat.mat[page] = map[byte]*State{}
					}
					mat.seed = append(mat.seed, &Seed{page, hash, kit.Format(value[kit.MDB_TEXT])})
					mat.train(m, page, hash, []byte(kit.Format(value[kit.MDB_TEXT])))
				})
				value[MATRIX] = mat
			})
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Save()
		}},
		MATRIX: {Name: "matrix hash npage text auto", Help: "魔方矩阵", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create nlang=32 ncell=256", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				mat := NewMatrix(m, kit.Int(kit.Select("32", m.Option(NLANG))), kit.Int(kit.Select("256", m.Option(NCELL))))
				m.Rich(m.Prefix(MATRIX), "", kit.Data(kit.MDB_TIME, m.Time(), MATRIX, mat, NLANG, mat.nlang, NCELL, mat.ncell))
			}},
			mdb.INSERT: {Name: "insert npage=num nhash=num text=123", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(m.Prefix(MATRIX), "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
					value = kit.GetMeta(value)

					mat, _ := value[MATRIX].(*Matrix)
					page := mat.index(m, NPAGE, m.Option(NPAGE))
					hash := mat.index(m, NHASH, m.Option(NHASH))
					if mat.mat[page] == nil {
						mat.mat[page] = map[byte]*State{}
					}

					mat.seed = append(mat.seed, &Seed{page, hash, m.Option(kit.MDB_TEXT)})
					m.Grow(m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, key), kit.Dict(
						kit.MDB_TIME, m.Time(), NPAGE, m.Option(NPAGE), NHASH, m.Option(NHASH), kit.MDB_TEXT, m.Option(kit.MDB_TEXT),
					))

					mat.train(m, page, hash, []byte(m.Option(kit.MDB_TEXT)))

					value[NSEED] = len(mat.seed)
					value[NPAGE] = len(mat.page) - 1
					value[NHASH] = len(mat.hash) - 1
				})
			}},
			mdb.REMOVE: {Name: "create", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.Prefix(MATRIX), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if m.Action(mdb.CREATE); len(arg) == 0 { // 矩阵列表
				m.Fields(len(arg) == 0, "time,hash,npage,nhash")
				m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), "", mdb.HASH)
				m.PushAction(mdb.INSERT, mdb.REMOVE)
				return
			}

			if m.Action(mdb.INSERT); len(arg) == 1 { // 词法列表
				m.Fields(len(arg) == 1, "time,npage,nhash,text")
				m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST)
				return
			}

			m.Richs(m.Prefix(MATRIX), "", arg[0], func(key string, value map[string]interface{}) {
				value = kit.GetMeta(value)
				mat, _ := value[MATRIX].(*Matrix)
				m.Debug("what %#v", mat)

				if len(arg) == 2 { // 词法矩阵
					mat.show(m, arg[1])
					return
				}

				hash, rest, word := mat.parse(m, mat.index(m, NPAGE, arg[1]), []byte(arg[2]))
				m.Push("time", m.Time())
				m.Push("hash", mat.word[hash])
				m.Push("word", string(word))
				m.Push("rest", string(rest))
			})
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
