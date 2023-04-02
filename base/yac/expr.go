package yac

import (
	"regexp"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const (
	SPACE  = "\t "
	QUOTE  = "\""
	TRANS  = " "
	BLOCK  = "[:](,){;}*/+-<>!=&|"
	DEFINE = ":="
	ASSIGN = "="
	SUBS   = "["
	DEFS   = ":"
	SUPS   = "]"
	OPEN   = "("
	FIELD  = ","
	CLOSE  = ")"
	BEGIN  = "{"
	SPLIT  = ";"
	END    = "}"
)

var keyword = regexp.MustCompile("[_a-zA-Z][_a-zA-Z0-9]*")

var level = map[string]int{
	"//": 200, "/*": 200, "*/": 200,
	"!": 100, "++": 100, "--": 100, "[": 100, "]": 100,
	"*": 40, "/": 40, "+": 30, "-": 30,
	"<": 20, ">": 20, ">=": 20, "<=": 20, "!=": 20, "==": 20, "&&": 10, "||": 10,
	DEFINE: 2, ASSIGN: 2, FIELD: 2, OPEN: 1, CLOSE: 1,
}

type Expr struct {
	list ice.List
	s    *Stack
}

func (s *Expr) push(v Any) *Expr {
	s.list = append(s.list, v)
	return s
}
func (s *Expr) pop(n int) *Expr {
	s.list = s.list[:len(s.list)-n]
	return s
}
func (s *Expr) pops(n int, v Any) *Expr { return s.pop(n).push(v) }
func (s *Expr) get(p int) (v Any) {
	kit.If(0 <= p+len(s.list) && p+len(s.list) < len(s.list), func() { v = s.list[p+len(s.list)] })
	kit.If(0 <= p && p < len(s.list), func() { v = s.list[p] })
	return
}
func (s *Expr) gets(p int) string { return kit.Format(s.get(p)) }
func (s *Expr) getl(p int) int    { return level[s.gets(p)] }
func (s *Expr) getv(m *ice.Message, p int) (v Any) {
	switch v := s.get(p); v := v.(type) {
	case string:
		return s.s.value(m, v)
	default:
		return v
	}
}
func (s *Expr) setv(m *ice.Message, p int, op string, v Any) Any {
	if !s.s.runable() {
		return nil
	}
	switch k := s.gets(p); v := v.(type) {
	case string:
		return s.s.value(m, k, s.s.value(m, v), op)
	case Value:
		if len(v.list) > 0 {
			return s.s.value(m, k, v.list[0], op)
		} else {
			return s.s.value(m, k)
		}
	default:
		return s.s.value(m, k, v, op)
	}
}
func (s *Expr) opv(m *ice.Message, p int, op string, v Any) Any {
	if !s.s.runable() {
		return s.getv(m, p)
	}
	if obj, ok := s.getv(m, p).(Operater); ok {
		return obj.Operate(op, trans(v))
	}
	return s.getv(m, p)
}
func (s *Expr) ops(m *ice.Message) {
	if !s.s.runable() || s.getl(-2) < 10 {
		return
	}
	s.pops(3, s.opv(m, -3, s.gets(-2), s.getv(m, -1)))
}
func (s *Expr) end(m *ice.Message) Any {
	if !s.s.runable() || len(s.list) == 0 {
		return nil
	} else if len(s.list) == 1 {
		return s.getv(m, 0)
	}
	m.Debug("expr ops %v", s.list)
	for i := 0; i < 100 && len(s.list) > 1; i++ {
		switch s.ops(m); s.gets(-2) {
		case DEFINE, ASSIGN:
			switch v := s.getv(m, -1).(type) {
			case Value:
				list := kit.List()
				for i := 0; i < len(s.list)-1; i += 2 {
					kit.If(i/2 < len(v.list), func() { list = append(list, s.setv(m, i, s.gets(-2), v.list[i/2])) })
				}
				s.list = append(s.list[:0], Value{list})
			default:
				s.list = append(s.list[:0], s.setv(m, -3, s.gets(-2), s.getv(m, -1)))
			}
		case FIELD:
			list := kit.List()
			for i := len(s.list) - 2; i > 0; i -= 2 {
				if s.gets(i) == DEFINE || s.gets(i) == ASSIGN {
					for j := 0; j < i; j += 2 {
						list = append(list, s.setv(m, j, s.gets(i), s.getv(m, j+i+1)))
					}
					s.list = append(s.list[:0], Value{list})
					break
				} else if i == 1 {
					for i := 0; i < len(s.list); i += 2 {
						list = append(list, s.getv(m, i))
					}
					s.list = append(s.list[:0], Value{list})
					break
				}
			}
		}
	}
	m.Debug("expr ops %v", s.list)
	return s.getv(m, 0)
}
func (s *Expr) cals(m *ice.Message) Any {
	line := s.s.line
	if s.s.skip == -1 {
		m.Debug("expr calcs %v %s:%d", s.s.rest, s.s.name, s.s.line)
	} else {
		m.Debug("expr calcs %v %s:%d", s.s.rest[s.s.skip:], s.s.name, s.s.line)
	}
	s.s.reads(m, func(k string) bool {
		if op := s.gets(-1) + k; level[op] > 0 && s.getl(-1) > 0 {
			s.pop(1)
			k = op
		}
		switch k {
		case DEFS, SUPS, BEGIN, SPLIT:
			return true
		case END:
			s.s.skip--
			return true
		case CLOSE:
			if s.gets(-2) == OPEN {
				s.pops(2, s.get(-1))
				return false
			}
			return true
		}
		if len(s.list) > 0 && s.getl(-1) == 0 {
			switch k {
			case "++", "--":
				s.pops(1, s.setv(m, -1, ASSIGN, s.opv(m, -1, k, nil)))
				return false
			case SUBS:
				s.pops(1, s.opv(m, -1, SUBS, s.s.cals(m)))
				return false
			case OPEN:
				if s.gets(-1) == FUNC && s.s.skip > 1 {
					s.s.skip--
					s.pops(1, s.s.value(m, s.s.funcs(m)))
				} else {
					switch k := s.get(-1).(type) {
					case string:
						s.pops(1, s.call(m, s.s, k))
					default:
						s.pops(1, s.call(m, k, ""))
					}
				}
				return false
			}
			if level[k] == 0 {
				if strings.HasPrefix(k, ice.PT) && kit.Select("", s.s.rest, s.s.skip+1) == OPEN {
					s.s.skip++
					s.pops(1, s.call(m, s.getv(m, -1), strings.TrimPrefix(k, ice.PT)))
					return false
				} else {
					s.s.skip--
					return true
				}
			}
		}
		if level[k] > 0 {
			for 9 <= level[k] && level[k] <= s.getl(-2) && level[k] < 100 {
				s.ops(m)
			}
			s.push(k)
		} else {
			if s.push(s.trans(m, k)); s.gets(-2) == "!" {
				s.pops(2, s.opv(m, -1, "!", nil))
			}
		}
		return false
	})
	if s.cmds(m, line) {
		return nil
	}
	return s.end(m)
}
func (s *Expr) trans(m *ice.Message, k string) Any {
	if strings.HasPrefix(k, "\"") {
		return String{value: k[1 : len(k)-1]}
	} else if k == ice.TRUE {
		return Boolean{value: true}
	} else if k == ice.FALSE {
		return Boolean{value: false}
	} else if keyword.MatchString(k) {
		return k
	} else {
		return Number{value: k}
	}
}
func (s *Expr) cmds(m *ice.Message, line int) bool {
	if cmds := false; len(s.list) == 1 && s.s.skip < 2 {
		m.Search(s.gets(0), func(key string, cmd *ice.Command) {
			if cmds = true; s.s.line == line {
				args := kit.List(s.gets(0))
				for {
					s := NewExpr(s.s)
					s.cals(m)
					if v := s.getv(m, 0); v != nil {
						args = append(args, trans(v))
					} else {
						args = append(args, v)
					}
					if s.s.line != line {
						break
					}
				}
				m.Cmdy(args...)
			} else {
				m.Cmdy(s.gets(0))
			}
		})
		if cmds {
			return true
		}
	}
	return false
}
func (s *Expr) call(m *ice.Message, obj Any, key string) Any {
	list := kit.List()
	switch v := s.s.cals(m).(type) {
	case Value:
		list = append(list, v.list...)
	default:
		list = append(list, v)
	}
	if !s.s.runable() {
		return list
	}
	return s.s.call(m, obj, key, nil, list...)
}
func NewExpr(s *Stack) *Expr { return &Expr{kit.List(), s} }

const EXPR = "expr"

func init() {
	Index.MergeCommands(ice.Commands{
		EXPR: {Name: "expr a, b = 1, 2", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			arg = s.rest
			if v := s.cals(m); !s.runable() {
				return
			} else if v != nil {
				m.Echo(kit.Format(trans(v)))
			} else if s.token() == BEGIN {
				m.Echo(ice.TRUE)
			}
			m.Debug("expr value %s %v %s:%d", m.Result(), arg, s.name, s.line)
		}},
	})
}
