package yac

import (
	"io"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type Func struct {
	line int
}
type Frame struct {
	key    string
	value  ice.Map
	status int
	line   int
	pop    func()
}
type Stack struct {
	frame []*Frame
	last  *Frame
	rest  []string
	list  []string
	line  int
}

func (s *Stack) peekf(m *ice.Message) *Frame { return s.frame[len(s.frame)-1] }
func (s *Stack) pushf(m *ice.Message, key string) *Frame {
	f := &Frame{key: key, value: kit.Dict(), status: s.peekf(m).status, line: s.line}
	s.frame = append(s.frame, f)
	return f
}
func (s *Stack) popf(m *ice.Message) *Frame {
	f := s.frame[len(s.frame)-1]
	kit.If(len(s.frame) > 1, func() { s.frame = s.frame[:len(s.frame)-1] })
	return f
}
func (s *Stack) value(m *ice.Message, key string, arg ...ice.Any) ice.Any {
	for i := len(s.frame) - 1; i >= 0; i-- {
		if f := s.frame[i]; f.value[key] != nil {
			kit.If(len(arg) > 0, func() { f.value[key] = arg[0] })
			return f.value[key]
		}
	}
	f := s.frame[len(s.frame)-1]
	kit.If(len(arg) > 0, func() { f.value[key] = arg[0] })
	return f.value[key]
}
func (s *Stack) parse(m *ice.Message, p string) *Stack {
	nfs.Open(m, p, func(r io.Reader) {
		s.peekf(m).key = p
		kit.For(r, func(text string) {
			s.list = append(s.list, text)
			for s.line = len(s.list) - 1; s.line < len(s.list); s.line++ {
				if text = s.list[s.line]; text == "" || strings.HasPrefix(text, "#") {
					continue
				}
				for s.rest = kit.Split(text, "\t ", "<=>+-*/;"); len(s.rest) > 0; {
					ls := s.rest
					switch s.rest = []string{}; v := s.value(m, ls[0]).(type) {
					case *Func:
						f, line := s.pushf(m, ls[0]), s.line
						f.pop, s.line = func() { s.rest, s.line, _ = nil, line+1, s.popf(m) }, v.line
					default:
						if _, ok := m.Target().Commands[ls[0]]; ok {
							m.Options(STACK, s).Cmdy(ls)
						} else {
							m.Options(STACK, s).Cmdy(CMD, ls)
						}
					}
				}
			}
		})
	})
	return s
}
func (s *Stack) show(m *ice.Message) (res []string) {
	for i, l := range s.list {
		res = append(res, kit.Format("%2d: ", i)+l)
	}
	res = append(res, "")
	for i, f := range s.frame {
		res = append(res, kit.Format("frame: %v line: %v %v %v", i, f.line, f.key, f.status))
		kit.For(f.value, func(k string, v ice.Any) { res = append(res, kit.Format("frame: %v %v: %v", i, k, v)) })
	}
	return
}
func _parse_split(m *ice.Message, split string, arg ...string) ([]string, []string) {
	if i := kit.IndexOf(arg, split); i == -1 {
		return arg, nil
	} else {
		return arg[:i], arg[i+1:]
	}
}
func NewStack(m *ice.Message) *Stack { return &Stack{frame: []*Frame{&Frame{value: kit.Dict()}}} }

const (
	PARSE  = "parse"
	CMD    = "cmd"
	LET    = "let"
	IF     = "if"
	FOR    = "for"
	FUNC   = "func"
	TYPE   = "type"
	RETURN = "return"
	EXPR   = "expr"
)
const STACK = "stack"

func init() {
	Index.MergeCommands(ice.Commands{
		STACK: {Name: "stack path auto parse", Actions: ice.Actions{
			PARSE: {Hand: func(m *ice.Message, arg ...string) {}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Options(nfs.DIR_ROOT, nfs.SRC).Cmdy(nfs.CAT, arg)
			if len(m.Result()) == 0 {
				return
			}
			m.Echo(ice.NL).Echo("output:" + ice.NL)
			s := NewStack(m).parse(m, path.Join(nfs.SRC, path.Join(arg...)))
			m.Echo(ice.NL).Echo("script:" + ice.NL)
			m.Echo(strings.Join(s.show(m), ice.NL))
		}},
		CMD: {Name: "cmd", Hand: func(m *ice.Message, arg ...string) {
			s := m.Optionv(STACK).(*Stack)
			arg, s.rest = _parse_split(m, "}", arg...)
			if f := s.peekf(m); f.status >= 0 {
				m.Cmdy(arg)
				m.Echo(ice.NL)
			}
		}},
		LET: {Name: "let a = 1", Hand: func(m *ice.Message, arg ...string) {
			s := m.Optionv(STACK).(*Stack)
			f := s.peekf(m)
			kit.If(f.status > -1, func() { s.value(m, arg[0], m.Cmdx(EXPR, arg[2:])) })
		}},
		IF: {Name: "if a > 1", Hand: func(m *ice.Message, arg ...string) {
			s := m.Optionv(STACK).(*Stack)
			f := s.pushf(m, m.CommandKey())
			arg, s.rest = _parse_split(m, "{", arg...)
			kit.If(f.status < 0 || m.Cmdx(EXPR, arg) == ice.FALSE, func() { f.status = -1 })
		}},
		FOR: {Name: "for a > 1", Hand: func(m *ice.Message, arg ...string) {
			s := m.Optionv(STACK).(*Stack)
			arg, s.rest = _parse_split(m, "{", arg...)
			f := s.pushf(m, m.CommandKey())
			kit.If(f.status < 0 || m.Cmdx(EXPR, arg) == ice.FALSE, func() { f.status = -1 })
			line := s.line
			f.pop = func() {
				kit.If(f.status > -1, func() { s.line = line - 1 })
				s.popf(m)
			}
		}},
		FUNC: {Name: "func show", Hand: func(m *ice.Message, arg ...string) {
			s := m.Optionv(STACK).(*Stack)
			arg, s.rest = _parse_split(m, "{", arg...)
			fun := &Func{line: s.line}
			s.value(m, arg[0], fun)
			f := s.pushf(m, m.CommandKey())
			f.status = -1
		}},
		RETURN: {Name: "return show", Hand: func(m *ice.Message, arg ...string) {
			s := m.Optionv(STACK).(*Stack)
			f := s.peekf(m)
			f.status = -2
		}},
		"}": {Name: "}", Hand: func(m *ice.Message, arg ...string) {
			s := m.Optionv(STACK).(*Stack)
			if f := s.peekf(m); f.pop == nil {
				s.last = s.peekf(m)
				s.popf(m)
			} else {
				f.pop()
			}
		}},
		EXPR: {Name: "expr a = 1", Hand: func(m *ice.Message, arg ...string) {
			s := m.Optionv(STACK).(*Stack)
			level := map[string]int{
				"*": 30, "/": 30,
				"+": 20, "-": 20,
				"<": 10, ">": 10, "<=": 10, ">=": 10, "==": 10, "!=": 10,
			}
			list := kit.List()
			get := func(p int) ice.Any {
				if p+len(list) >= 0 {
					return list[p+len(list)]
				}
				return ""
			}
			gets := func(p int) string {
				k := kit.Format(get(p))
				if v := s.value(m, k); v != nil {
					return kit.Format(v)
				}
				return k
			}
			getl := func(p int) int { return level[kit.Format(get(p))] }
			push := func(v ice.Any) { list = append(list, v) }
			pop := func(n int) { list = list[:len(list)-n] }
			ops := func() {
				bin := func(v ice.Any) {
					pop(3)
					push(kit.Format(v))
				}
				switch a, b := kit.Int(gets(-3)), kit.Int(gets(-1)); gets(-2) {
				case "==":
					bin(a == b)
				case "!=":
					bin(a != b)
				case "<=":
					bin(a <= b)
				case ">=":
					bin(a >= b)
				case ">":
					bin(a > b)
				case "<":
					bin(a < b)
				case "+":
					bin(a + b)
				case "-":
					bin(a - b)
				case "*":
					bin(a * b)
				case "/":
					bin(a / b)
				}
			}
			kit.For(arg, func(k string) {
				if level[k] > 0 {
					for level[k] <= getl(-2) {
						ops()
					}
				}
				push(k)
			})
			for len(list) > 1 {
				ops()
			}
			m.Echo(kit.Format(list[0]))
			m.Debug("expr %s", m.Result())
		}},
	})
}
