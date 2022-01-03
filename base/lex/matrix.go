package lex

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

type Stream struct {
	r io.Reader
	b []byte
	P int
}

func NewStream(r io.Reader) *Stream {
	return &Stream{r: r}
}
func (s *Stream) Scan() bool {
	if s.P < len(s.b) {
		return true
	}
	buf := make([]byte, 1024)
	if n, e := s.r.Read(buf); e == nil && n > 0 {
		s.b = append(s.b, buf[:n]...)
	}
	return s.P < len(s.b)
}
func (s *Stream) Next() {
	s.P++
}
func (s *Stream) Char() byte {
	if s.Scan() {
		return s.b[s.P]
	}
	return 0
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

	trans map[byte][]byte
	mat   []map[byte]*State
}

func NewMatrix(m *ice.Message, nlang, ncell int) *Matrix {
	mat := &Matrix{nlang: nlang, ncell: ncell}
	mat.page = map[string]int{}
	mat.hand = map[int]string{}
	mat.hash = map[string]int{}
	mat.word = map[int]string{}

	mat.trans = map[byte][]byte{}
	for k, v := range map[byte]string{
		't': "\t", 'n': "\n", 'b': "\t ", 's': "\t \n",
		'd': "0123456789", 'x': "0123456789ABCDEFabcdef",
	} {
		mat.trans[k] = []byte(v)
	}

	mat.mat = make([]map[byte]*State, nlang)
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
func (mat *Matrix) Train(m *ice.Message, npage, nhash string, seed string) int {
	// m.Debug("%s %s page: %v hash: %v seed: %v", TRAIN, LEX, npage, nhash, seed)

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
		switch seed[i] { // 字符集
		case '[':
			set := true
			if i++; seed[i] == '^' { // 补集
				set, i = false, i+1
			}

			for ; seed[i] != ']'; i++ {
				if seed[i] == '\\' { // 转义
					i++
					for _, c := range mat.char(seed[i]) {
						cn[c] = true
					}
					continue
				}

				if seed[i+1] == '-' { // 区间
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

				cn[seed[i]] = true // 单个
			}

			for c := 1; c < len(cn); c++ { // 序列
				if (set && cn[c]) || (!set && !cn[c]) {
					cc = append(cc, byte(c))
				}
				cn[c] = false
			}

		case '.':
			for c := 1; c < len(cn); c++ { // 全集
				cc = append(cc, byte(c))
			}

		case '\\':
			i++
			for _, c := range mat.char(seed[i]) { // 转义
				cc = append(cc, c)
			}
		default:
			cc = append(cc, seed[i]) // 普通字符
		}

		// m.Debug("page: \033[31m%d %v\033[0m", len(ss), ss)
		// m.Debug("cell: \033[32m%d %v\033[0m", len(cc), cc)

		flag := '\000'
		if i+1 < len(seed) { // 次数
			switch flag = rune(seed[i+1]); flag {
			case '?', '+', '*':
				i++
			}
		}

		add := func(s int, c byte, cb func(*State)) { // 添加节点
			state := mat.mat[s][c]
			if state == nil {
				state = &State{}
			}
			// m.Debug("GET(%d,%d): %#v", s, c, state)

			if cb(state); state.next == 0 {
				sn = append(sn, true)
				state.next = len(mat.mat)
				mat.mat = append(mat.mat, make(map[byte]*State))
			} else {
				sn[state.next] = true
			}

			mat.mat[s][c] = state
			points = append(points, &Point{s, c})
			// m.Debug("SET(%d,%d): %#v", s, c, state)
		}

		for _, s := range ss {
			for _, c := range cc {
				add(s, c, func(state *State) {
					switch flag { // 次数
					case '+':
						state.star = true
						sn[s] = true
						break

						sn = append(sn, true)
						state.next = len(mat.mat)
						mat.mat = append(mat.mat, make(map[byte]*State))
						for _, c := range cc {
							add(state.next, c, func(state *State) { state.star = true })
						}
					case '*':
						state.star = true
						sn[s] = true
					case '?':
						sn[s] = true
					}
				})
			}
		}

		cc, ss = cc[:0], ss[:0]
		for s, b := range sn { // 迭代
			if sn[s] = false; b && s > 0 {
				ss = append(ss, s)
			}
		}
	}

	trans := map[int]int{page: page}
	for i := mat.nlang; i < len(mat.mat); i++ { // 去空
		if len(mat.mat[i]) > 0 {
			trans[i] = i
			continue
		}

		for j := i + 1; j < len(mat.mat); j++ {
			if len(mat.mat[j]) > 0 {
				mat.mat[i] = mat.mat[j]
				mat.mat[j] = nil
				trans[j] = i
				break
			}
		}
		if len(mat.mat[i]) == 0 {
			mat.mat = mat.mat[:i]
			break
		}
	}
	// m.Debug("DEL: %v", trans)

	for _, p := range points { // 去尾
		p.s = trans[p.s]
		state := mat.mat[p.s][p.c]
		if state.next = trans[state.next]; state.next == 0 {
			// m.Debug("GET(%d, %d): %#v", p.s, p.c, state)
			state.hash = hash
			// m.Debug("SET(%d, %d): %#v", p.s, p.c, state)
		}
	}

	// m.Debug("%s %s npage: %v nhash: %v", TRAIN, LEX, len(mat.page), len(mat.hash))
	return hash
}
func (mat *Matrix) Parse(m *ice.Message, npage string, stream *Stream) (hash int, word []byte) {
	// m.Debug("%s %s page: %v pos: %v", LEX, PARSE, npage, stream.P)
	page := mat.index(m, NPAGE, npage)

	pos := stream.P
	for star, s := 0, page; stream.Scan() && s != 0; stream.Next() {
		c := stream.Char()
		if c == '\\' { //跳过转义
			if stream.Next(); !stream.Scan() {
				break
			}
			c = mat.char(stream.Char())[0]
		}
		if c > 127 { //跳过中文
			word = append(word, c)
			continue
		}

		state := mat.mat[s][c]
		if state == nil {
			s, star, stream.P = star, 0, stream.P-1
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

	if hash == 0 {
		stream.P, word = pos, word[:0]
	}

	// m.Debug("%s %s hash: %v word: %v", LEX, PARSE, mat.word[hash], string(word))
	return
}
func (mat *Matrix) show(m *ice.Message) {
	showCol := map[int]bool{} // 有效列
	for j := 1; j < mat.ncell; j++ {
		for i := 1; i < len(mat.mat); i++ {
			if node := mat.mat[i][byte(j)]; node != nil {
				showCol[j] = true
			}
		}
	}

	for i := 1; i < len(mat.mat); i++ {
		if len(mat.mat[i]) == 0 { // 无效行
			continue
		}

		m.Push("00", kit.Select(kit.Format("%02d", i), mat.hand[i]))
		for j := 1; j < mat.ncell; j++ {
			if !showCol[j] { // 无效列
				continue
			}
			key, value := kit.Format("%c", j), []string{}
			if node := mat.mat[i][byte(j)]; node != nil {
				if node.star {
					value = append(value, "*")
				}
				if node.next > 0 {
					value = append(value, cli.ColorGreen(m, node.next))
				}
				if node.hash > 0 {
					value = append(value, cli.ColorRed(m, kit.Select(kit.Format("%d", node.hash), mat.word[node.hash])))
				}
			}
			m.Push(key, strings.Join(value, ","))
		}
	}

	m.Status(NLANG, mat.nlang, NCELL, mat.ncell, NPAGE, len(mat.page), NHASH, len(mat.hash))
}
func _lex_load(m *ice.Message) {
	m.Richs(m.Prefix(MATRIX), "", mdb.FOREACH, func(key string, value map[string]interface{}) {
		value = kit.GetMeta(value)

		mat := NewMatrix(m, kit.Int(kit.Select("32", value[NLANG])), kit.Int(kit.Select("256", value[NCELL])))
		m.Grows(m.Prefix(MATRIX), kit.Keys(mdb.HASH, key), "", "", func(index int, value map[string]interface{}) {
			mat.Train(m, kit.Format(value[NPAGE]), kit.Format(value[NHASH]), kit.Format(value[mdb.TEXT]))
		})
		value[MATRIX] = mat
	})
}

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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		MATRIX: {Name: MATRIX, Help: "魔方矩阵", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		MATRIX: {Name: "matrix hash npage text auto", Help: "魔方矩阵", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				// _lex_load(m.Load())
			}},
			mdb.CREATE: {Name: "create nlang=32 ncell=128", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				mat := NewMatrix(m, kit.Int(kit.Select("32", m.Option(NLANG))), kit.Int(kit.Select("128", m.Option(NCELL))))
				h := m.Rich(m.Prefix(MATRIX), "", kit.Data(mdb.TIME, m.Time(), MATRIX, mat, NLANG, mat.nlang, NCELL, mat.ncell))
				switch cb := m.Optionv(kit.Keycb(MATRIX)).(type) {
				case func(string, *Matrix):
					cb(h, mat)
				}
				m.Echo(h)
			}},
			mdb.INSERT: {Name: "insert hash npage=num nhash=num text=123", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(m.Prefix(MATRIX), "", m.Option(mdb.HASH), func(key string, value map[string]interface{}) {
					value = kit.GetMeta(value)

					mat, _ := value[MATRIX].(*Matrix)
					m.Echo("%d", mat.Train(m, m.Option(NPAGE), m.Option(NHASH), m.Option(mdb.TEXT)))
					m.Grow(m.Prefix(MATRIX), kit.Keys(mdb.HASH, key), kit.Dict(
						mdb.TIME, m.Time(), NPAGE, m.Option(NPAGE), NHASH, m.Option(NHASH), mdb.TEXT, m.Option(mdb.TEXT),
					))

					value[NPAGE] = len(mat.page)
					value[NHASH] = len(mat.hash)
				})
			}},
			mdb.REMOVE: {Name: "create", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.Prefix(MATRIX), "", mdb.HASH, mdb.HASH, m.Option(mdb.HASH))
			}},
			PARSE: {Name: "parse hash npage text=123", Help: "解析", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(m.Prefix(MATRIX), "", m.Option(mdb.HASH), func(key string, value map[string]interface{}) {
					value = kit.GetMeta(value)
					mat, _ := value[MATRIX].(*Matrix)

					stream := NewStream(bytes.NewBufferString(m.Option(mdb.TEXT)))
					hash, word := mat.Parse(m, m.Option(NPAGE), stream)
					m.Push(NHASH, kit.Select(kit.Format("%d", hash), mat.word[hash]))
					m.Push("word", string(word))
					m.Push("rest", string(stream.b[stream.P:]))
				})
				m.ProcessInner()
			}},
			"show": {Name: "show", Help: "矩阵", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(m.Prefix(MATRIX), "", kit.Select(m.Option(mdb.HASH), arg, 0), func(key string, value map[string]interface{}) {
					value = kit.GetMeta(value)
					value[MATRIX].(*Matrix).show(m)
				})
				m.ProcessInner()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if m.Action(mdb.CREATE); len(arg) == 0 { // 矩阵列表
				m.Fields(len(arg), "time,hash,npage,nhash")
				m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), "", mdb.HASH)
				m.PushAction(mdb.INSERT, "show", mdb.REMOVE)
				return
			}

			if m.Action(mdb.INSERT, "show"); len(arg) == 1 { // 词法列表
				m.Fields(len(arg[1:]), "time,npage,nhash,text")
				m.Cmdy(mdb.SELECT, m.Prefix(MATRIX), kit.Keys(mdb.HASH, arg[0]), mdb.LIST)
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

				hash, word := mat.Parse(m, arg[1], NewStream(bytes.NewBufferString(arg[2])))
				m.Push(mdb.TIME, m.Time())
				m.Push(mdb.HASH, mat.word[hash])
				m.Push("word", string(word))
			})
		}},
	}})
}
