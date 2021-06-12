package lex

import (
	"sort"
	"strconv"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
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

	page map[string]int
	hand map[int]string
	hash map[string]int
	word map[int]string

	trans map[byte]string
	state map[State]*State
	mat   []map[byte]*State
}

func NewMatrix(m *ice.Message, nlang, ncell int) *Matrix {
	mat := &Matrix{nlang: nlang, ncell: ncell}
	mat.page = map[string]int{}
	mat.hand = map[int]string{}
	mat.hash = map[string]int{}
	mat.word = map[int]string{}

	mat.trans = map[byte]string{
		't': "\t", 'n': "\n", 'b': "\t ", 's': "\t \n",
		'd': "0123456789", 'x': "0123456789ABCDEFabcdef",
	}

	mat.state = make(map[State]*State)
	mat.mat = make([]map[byte]*State, nlang)
	return mat
}
func (mat *Matrix) char(c byte) []byte {
	if cs, ok := mat.trans[c]; ok {
		return []byte(cs)
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
func (mat *Matrix) Train(m *ice.Message, npage, nhash string, seed string) int {
	m.Debug("%s %s page: %v hash: %v seed: %v", TRAIN, LEX, npage, nhash, seed)

	page := mat.index(m, NPAGE, npage)
	hash := mat.index(m, NHASH, nhash)
	if mat.mat[page] == nil {
		mat.mat[page] = map[byte]*State{}
	}

	ss := []int{page}
	cn := make([]bool, mat.ncell)
	cc := make([]byte, 0, mat.ncell)
	sn := make([]bool, len(mat.mat))

	points := []*Point{}

	for i := 0; i < len(seed); i++ {
		switch seed[i] {
		case '[':
			set := true
			if i++; seed[i] == '^' {
				set, i = false, i+1
			}

			for ; seed[i] != ']'; i++ {
				if seed[i] == '\\' {
					i++
					for _, c := range mat.char(seed[i]) {
						cn[c] = true
					}
					continue
				}

				if seed[i+1] == '-' {
					begin, end := seed[i], seed[i+2]
					if begin > end {
						begin, end = end, begin
					}
					for c := begin; c <= end; c++ {
						cn[c] = true
					}
					i += 2
					continue
				}

				cn[seed[i]] = true
			}

			for c := 0; c < len(cn); c++ {
				if (set && cn[c]) || (!set && !cn[c]) {
					cc = append(cc, byte(c))
				}
				cn[c] = false
			}

		case '.':
			for c := 1; c < len(cn); c++ {
				cc = append(cc, byte(c))
			}

		case '\\':
			i++
			for _, c := range mat.char(seed[i]) {
				cc = append(cc, c)
			}
		default:
			cc = append(cc, seed[i])
		}

		m.Debug("page: \033[31m%d %v\033[0m", len(ss), ss)
		m.Debug("cell: \033[32m%d %v\033[0m", len(cc), cc)

		flag := '\000'
		if i+1 < len(seed) {
			switch flag = rune(seed[i+1]); flag {
			case '?', '+', '*':
				i++
			}
		}

		for _, s := range ss {
			for _, c := range cc {

				state := &State{}
				if mat.mat[s][c] != nil {
					*state = *mat.mat[s][c]
				}
				m.Debug("GET(%d,%d): %#v", s, c, state)

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
					state.next = len(mat.mat)
					mat.mat = append(mat.mat, make(map[byte]*State))
					sn = append(sn, true)
				} else {
					sn[state.next] = true
				}

				mat.mat[s][c] = state
				points = append(points, &Point{s, c})
				m.Debug("SET(%d,%d): %#v", s, c, state)
			}
		}

		cc, ss = cc[:0], ss[:0]
		for s, b := range sn {
			if sn[s] = false; b && s > 0 {
				ss = append(ss, s)
			}
		}
	}

	sort.Ints(ss)
	sort.Reverse(sort.IntSlice(ss))
	for _, s := range ss {
		if s < mat.nlang || s >= len(mat.mat) {
			continue
		}
		if len(mat.mat[s]) > 0 {
			break
		}
		mat.mat = mat.mat[:s]
		m.Debug("DEL: %d", len(mat.mat))
	}

	for _, s := range ss {
		for _, p := range points {
			state := &State{}
			*state = *mat.mat[p.s][p.c]

			if state.next == s {
				m.Debug("GET(%d, %d): %#v", p.s, p.c, state)
				if state.hash = hash; state.next >= len(mat.mat) {
					state.next = 0
				}
				mat.mat[p.s][p.c] = state
				m.Debug("SET(%d, %d): %#v", p.s, p.c, state)
			}

			if x, ok := mat.state[*state]; !ok {
				mat.state[*state] = mat.mat[p.s][p.c]
			} else {
				mat.mat[p.s][p.c] = x
			}
		}
	}

	m.Debug("%s %s npage: %v nhash: %v", "train", "lex", len(mat.page), len(mat.hash))
	return hash
}
func (mat *Matrix) Parse(m *ice.Message, npage string, line []byte) (hash int, word []byte, rest []byte) {
	// m.Debug("%s %s page: %v line: %v", "parse", "lex", npage, line)
	page := mat.index(m, NPAGE, npage)

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
		// m.Debug("GET (%d,%d): %v", s, c, state)

		if word = append(word, c); state.star {
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

	// m.Debug("%s %s hash: %v word: %v rest: %v", "parse", "lex", hash, word, rest)
	return
}
func (mat *Matrix) show(m *ice.Message) {
	show := map[int]bool{}
	for j := 1; j < mat.ncell; j++ {
		for i := 1; i < len(mat.mat); i++ {
			if node := mat.mat[i][byte(j)]; node != nil {
				show[j] = true
			}
		}
	}

	for i := 1; i < len(mat.mat); i++ {
		if len(mat.mat[i]) == 0 {
			continue
		}

		m.Push("00", kit.Select(kit.Format("%02d", i), mat.hand[i]))
		for j := 1; j < mat.ncell; j++ {
			if !show[j] {
				continue
			}
			key := kit.Format("%c", j)
			if node := mat.mat[i][byte(j)]; node != nil {
				if node.hash == 0 {
					m.Push(key, kit.Select(kit.Format("%02d", node.next), cli.ColorRed(m, mat.hand[node.next])))
				} else {
					m.Push(key, kit.Select(kit.Format("w%02d", node.hash), cli.ColorGreen(m, mat.word[node.hash])))
				}
			} else {
				m.Push(key, "")
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

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			MATRIX: {Name: MATRIX, Help: "魔方矩阵", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			MATRIX: {Name: "matrix hash npage text auto", Help: "魔方矩阵", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create nlang=32 ncell=256", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					mat := NewMatrix(m, kit.Int(kit.Select("32", m.Option(NLANG))), kit.Int(kit.Select("256", m.Option(NCELL))))
					h := m.Rich(m.Prefix(MATRIX), "", kit.Data(kit.MDB_TIME, m.Time(), MATRIX, mat, NLANG, mat.nlang, NCELL, mat.ncell))
					switch cb := m.Optionv("matrix.cb").(type) {
					case func(string, *Matrix):
						cb(h, mat)
					}
					m.Echo(h)
				}},
				mdb.INSERT: {Name: "insert hash npage=num nhash=num text=123", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(m.Prefix(MATRIX), "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)

						mat, _ := value[MATRIX].(*Matrix)
						m.Echo("%d", mat.Train(m, m.Option(NPAGE), m.Option(NHASH), m.Option(kit.MDB_TEXT)))
						m.Grow(m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, key), kit.Dict(
							kit.MDB_TIME, m.Time(), NPAGE, m.Option(NPAGE), NHASH, m.Option(NHASH), kit.MDB_TEXT, m.Option(kit.MDB_TEXT),
						))

						value[NPAGE] = len(mat.page)
						value[NHASH] = len(mat.hash)
					})
				}},
				mdb.REMOVE: {Name: "create", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(MATRIX), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				"show": {Name: "show", Help: "矩阵", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(m.Prefix(MATRIX), "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)
						mat, _ := value[MATRIX].(*Matrix)
						mat.show(m)
					})
					m.ProcessInner()
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				if m.Action(mdb.CREATE); len(arg) == 0 { // 矩阵列表
					m.Fields(len(arg) == 0, "time,hash,npage,nhash")
					m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), "", mdb.HASH)
					m.PushAction("show", mdb.INSERT, mdb.REMOVE)
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

					if len(arg) == 2 { // 词法矩阵
						mat.show(m)
						return
					}

					hash, word, rest := mat.Parse(m, arg[1], []byte(arg[2]))
					m.Push(kit.MDB_TIME, m.Time())
					m.Push(kit.MDB_HASH, mat.word[hash])
					m.Push("word", string(word))
					m.Push("rest", string(rest))
				})
			}},
		},
	})
}
