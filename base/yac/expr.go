package yac

import (
	"regexp"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	SPACE  = "\t "
	QUOTE  = "\"'`"
	TRANS  = " "
	BLOCK  = "[:](,){;}*/+-<>!=&|"
	EXPAND = "..."
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

var level = map[string]int{
	"//": 200, "/*": 200, "*/": 200,
	"!": 100, "&": 100, // "*": 100,
	"[": 100, "]": 100, "++": 100, "--": 100,
	"*": 40, "/": 40, "+": 30, "-": 30,
	"<": 20, ">": 20, ">=": 20, "<=": 20, "!=": 20, "==": 20, "&&": 10, "||": 10,
	DEFS: 2, DEFINE: 2, ASSIGN: 2, FIELD: 2, OPEN: 1, CLOSE: 1,
}
var keyword = regexp.MustCompile(`^[_a-zA-Z.][_a-zA-Z0-9.]*$`)

type Expr struct {
	list ice.List
	p    string
	t    Any
	n    int
	*Stack
}

func (s *Expr) push(v Any) *Expr {
	s.list = append(s.list, v)
	return s
}
func (s *Expr) pop(n int) *Expr {
	if n <= len(s.list) {
		s.list = s.list[:len(s.list)-n]
	} else {
		s.list = s.list[:0]
	}
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
		return s.value(m, v)
	default:
		return v
	}
}
func (s *Expr) setv(m *ice.Message, p int, op string, v Any) Any {
	if !s.runable() {
		return nil
	}
	switch k := s.gets(p); v := v.(type) {
	case string:
		return s.value(m, k, s.value(m, v), op)
	case Value:
		if len(v.list) > 0 {
			return s.value(m, k, v.list[0], op)
		} else {
			return s.value(m, k, nil, op)
		}
	default:
		return s.value(m, k, v, op)
	}
}
func (s *Expr) isop(k Any) bool {
	switch k := k.(type) {
	case int:
		return level[s.gets(k)] > 0
	case string:
		return level[k] > 0
	}
	return false
}
func (s *Expr) opv(m *ice.Message, p int, op string, v Any) Any {
	if !s.runable() {
		return s.getv(m, p)
	}
	if obj, ok := s.getv(m, p).(Operater); ok {
		return obj.Operate(op, Trans(v))
	}
	return s.getv(m, p)
}
func (s *Expr) ops(m *ice.Message) {
	if !s.runable() || s.getl(-2) < 10 {
		return
	}
	if s.getl(-3) > 0 {
		s.pops(3, s.opv(m, -1, s.gets(-2), nil))
	} else {
		s.pops(3, s.opv(m, -3, s.gets(-2), s.get(-1)))
	}
}
func (s *Expr) end(m *ice.Message) Any {
	if !s.runable() || len(s.list) == 0 {
		return nil
	} else if len(s.list) == 1 {
		return s.getv(m, 0)
	}
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
				s.list = append(s.list[:0], s.setv(m, -3, s.gets(-2), s.get(-1)))
			}
		case FIELD:
			list := kit.List()
			for i := len(s.list) - 2; i > 0; i -= 2 {
				if s.gets(i) == DEFINE || s.gets(i) == ASSIGN {
					for j := 0; j < i; j += 2 {
						list = append(list, s.setv(m, j, s.gets(i), s.get(j+i+1)))
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
	return s.getv(m, 0)
}
func (s *Expr) sub(m *ice.Message) *Expr {
	sub := NewExpr(s.Stack)
	sub.n = s.n + 1
	return sub
}
func (s *Expr) ktv(m *ice.Message, t Any, p string) map[string]Any {
	var kt, vt Any
	switch t := t.(type) {
	case Map:
		kt, vt = s.types(m, t.key, p), s.types(m, t.value, p)
	}
	data := kit.Dict()
	for s.token() != END {
		k := ""
		kit.If(t == nil || kt != nil, func() {
			sub := s.sub(m)
			sub.t, sub.p = kt, p
			k = kit.Format(Trans(sub.cals(m, DEFS, END)))
		}, func() {
			k = s.next(m)
			kit.If(s.token() != END, func() { s.next(m) })
		})
		kit.If(s.token() == DEFS, func() {
			sub := s.sub(m)
			if sub.p = p; t == nil || kt != nil {
				sub.t = vt
			} else {
				switch t := t.(type) {
				case Struct:
					sub.t = t.index[k].(Field)
				}
			}
			m.Debug("field %d %s %s", sub.n, k, Format(sub.t))
			data[k] = sub.cals(m, FIELD, END)
			m.Debug("field %d %s %s", sub.n, k, Format(data[k]))
		})
	}
	return data
}
func (s *Expr) ntv(m *ice.Message, t Any, p string) []Any {
	data := kit.List()
	for !kit.IsIn(s.token(), SUPS, END) {
		sub := s.sub(m)
		sub.t, sub.p = t, p
		m.Debug("field %d %d %s", sub.n, len(data), Format(sub.t))
		if v := sub.cals(m, FIELD, SUPS, END); v != nil {
			m.Debug("field %d %d %s", sub.n, len(data), Format(v))
			data = append(data, v)
		}
	}
	return data
}
func (s *Expr) cals(m *ice.Message, arg ...string) Any {
	if s.skip == -1 {
		m.Debug("calcs %d %v %v", s.n, s.rest, arg)
	} else {
		m.Debug("calcs %d %v %v", s.n, s.rest[s.skip:], arg)
	}
	line := s.line
	s.reads(m, func(k string) bool {
		switch s.get(-1).(type) {
		case string:
			if op := s.gets(-1) + k; s.isop(op) {
				s.pop(1)
				k = op
			}
		}
		if kit.IsIn(k, arg...) {
			return true
		}
		switch k {
		case SPLIT:
			return true
		case BEGIN:
			p := ""
			kit.If(strings.Contains(s.gets(-1), nfs.PT), func() { p = kit.Split(s.gets(-1), nfs.PT)[0] })
			switch t := s.getv(m, -1).(type) {
			case Map:
				s.pops(1, Dict{s.ktv(m, t, p)})
				return false
			case Slice:
				s.pops(1, List{s.ntv(m, t, p)})
				return false
			case Struct:
				s.pops(1, Object{Dict{s.ktv(m, t, p)}, t})
				return false
			}
			switch t := s.t.(type) {
			case Map:
				s.pops(0, Dict{s.ktv(m, t.value, s.p)})
				return false
			case Slice:
				s.pops(0, List{s.ntv(m, t, s.p)})
				return false
			case Struct:
				s.pops(0, Object{Dict{s.ktv(m, t, s.p)}, t})
				return false
			}
			if kit.IsIn(s.gets(-1), DEFINE) || len(s.list) == 0 && len(arg) > 0 {
				s.pops(0, Dict{s.ktv(m, s.t, s.p)})
				return false
			}
			return true
		case END:
			s.skip--
			return true
		case MAP:
			s.push(s.Stack.types(m))
			return false
		case SUBS:
			if s.peek(m) == SUPS && !kit.IsIn(kit.Select("", s.rest, s.skip+2), FIELD, SUPS, END, "") {
				s.push(s.Stack.types(m))
				return false
			}
			if kit.IsIn(s.gets(-1), DEFINE) || len(s.list) == 0 && len(arg) > 0 {
				s.push(List{s.ntv(m, s.t, s.p)})
				return false
			}
		case STRUCT, INTERFACE:
			s.push(s.Stack.types(m))
			return false
		case FUNC:
			if s.skip > 0 {
				s.push(s.funcs(m, ""))
				return false
			}
			s.skip--
			return true
		case CLOSE:
			if s.gets(-2) == OPEN {
				s.pops(2, s.get(-1))
				return false
			}
			return true
		}
		if len(s.list) > 0 && !s.isop(-1) {
			switch k {
			case OPEN:
				if strings.HasSuffix(s.gets(-1), nfs.PT) {
					if s.peek(m) == TYPE {
						switch v := s.getv(m, -1).(type) {
						case Object:
							s.pops(1, v.index)
						default:
							s.pops(1, kit.Format("%T", v))
						}
					} else {
						switch t := s.sub(m).cals(m, CLOSE).(type) {
						case Struct:
							if v, ok := s.get(-1).(Operater); ok {
								s.pops(1, Value{list: []Any{v, v.Operate(INSTANCEOF, t)}})
							}
						case Interface:
							if v, ok := s.get(-1).(Operater); ok {
								s.pops(1, Value{list: []Any{v, v.Operate(IMPLEMENTS, t)}})
							}
						}
					}
					return false
				}
				switch k := s.get(-1).(type) {
				case string:
					s.pops(1, s.call(m, s.Stack, k))
				default:
					s.pops(1, s.call(m, k, ""))
				}
				return false
			case SUBS:
				switch v := s.sub(m).cals(m, SUPS); s.get(-1).(type) {
				case string:
					s.pops(1, kit.Keys(s.gets(-1), kit.Format(Trans(v))))
				default:
					s.pops(1, s.opv(m, -1, SUBS, v))
				}
				return false
			case "++", "--":
				s.pops(1, s.setv(m, -1, ASSIGN, s.opv(m, -1, k, nil)))
				return false
			}
			if !s.isop(k) {
				if strings.HasPrefix(k, nfs.PT) {
					if s.peek(m) == OPEN {
						s.skip++
						s.pops(1, s.call(m, s.getv(m, -1), strings.TrimPrefix(k, nfs.PT)))
						return false
					} else if !s.isop(-1) && len(s.list) > 0 {
						s.pops(1, s.gets(-1)+k)
						return false
					}
				}
				s.skip--
				return true
			}
		}
		if s.isop(k) {
			for 9 <= level[k] && level[k] <= s.getl(-2) && level[k] < 100 {
				s.ops(m)
			}
			s.push(k)
		} else {
			if strings.HasSuffix(k, EXPAND) {
				if v, ok := s.Stack.value(m, strings.TrimSuffix(k, EXPAND)).(Operater); ok {
					if list, ok := v.Operate(EXPAND, nil).([]Any); ok && len(list) > 0 {
						kit.For(list, func(v Any) {
							s.list = append(s.list, v, FIELD)
						})
					}
				}
				kit.If(s.gets(-1) == FIELD, func() { s.pop(1) })
			} else {
				s.push(s.trans(m, k))
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
func (s *Expr) types(m *ice.Message, t Any, p string) Any {
	switch t := t.(type) {
	case string:
		if p == "" || kit.IsIn(t, STRING, INT) {
			return t
		}
		return s.value(m, kit.Keys(p, t))
	default:
		return t
	}
}
func (s *Expr) cmds(m *ice.Message, line int) (done bool) {
	if len(s.list) == 1 && s.skip < 2 {
		m.Search(s.gets(0), func(key string, cmd *ice.Command) {
			args := kit.List(s.gets(0))
			for done = true; s.line == line; {
				args = append(args, kit.Format(Trans(s.sub(m).cals(m))))
			}
			kit.If(s.runable(), func() { m.Cmdy(args...) })
		})
	}
	return
}
func (s *Expr) call(m *ice.Message, obj Any, key string) Any {
	if arg := _parse_res(m, s.sub(m).cals(m, CLOSE)); s.runable() {
		return s.calls(m, obj, key, nil, arg...)
	} else {
		return nil
	}
}
func NewExpr(s *Stack) *Expr { return &Expr{list: kit.List(), Stack: s} }

const EXPR = "expr"

func init() {
	Index.MergeCommands(ice.Commands{
		EXPR: {Name: "expr a, b = 1, 2", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			if v := s.cals(m); !s.runable() {
				return
			} else if v != nil {
				m.Debug("value %s %s <- %v", Format(s), Format(v), arg)
				switch v := Trans(v).(type) {
				case Message:
				case Value:
					kit.If(len(v.list) > 0, func() { m.Echo(kit.Format(Trans(v.list[0]))) })
				default:
					m.Echo(kit.Format(Trans(v)))
				}
			} else if s.token() == BEGIN {
				m.Echo(ice.TRUE)
			}
		}},
	})
}
