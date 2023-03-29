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
	arg  []string
	res  []string
}
type Frame struct {
	key    string
	value  ice.Map
	status int
	line   int
	pop    func()
}
type Stack struct {
	key   string
	frame []*Frame
	last  *Frame
	rest  []string
	list  []string
	line  int
}

func (s *Stack) peekf() *Frame {
	kit.If(len(s.frame) == 0, func() { s.pushf("") })
	return s.frame[len(s.frame)-1]
}
func (s *Stack) pushf(key string) *Frame {
	f := &Frame{key: key, value: kit.Dict(), line: s.line}
	kit.If(len(s.frame) > 0, func() { f.status = s.peekf().status })
	s.frame = append(s.frame, f)
	return f
}
func (s *Stack) popf() *Frame {
	f := s.peekf()
	kit.If(len(s.frame) > 1, func() { s.frame = s.frame[:len(s.frame)-1] })
	return f
}
func (s *Stack) stack(cb ice.Any) {
	for i := len(s.frame) - 1; i >= 0; i-- {
		switch cb := cb.(type) {
		case func(*Frame, int):
			cb(s.frame[i], i)
		case func(*Frame):
			cb(s.frame[i])
		}
	}
}
func (s *Stack) value(key string, arg ...ice.Any) ice.Any {
	for i := len(s.frame) - 1; i >= 0; i-- {
		if f := s.frame[i]; f.value[key] != nil {
			kit.If(len(arg) > 0, func() { f.value[key] = arg[0] })
			return f.value[key]
		}
	}
	f := s.peekf()
	kit.If(len(arg) > 0, func() { f.value[key] = arg[0] })
	return f.value[key]
}
func (s *Stack) runable() bool { return s.peekf().status > DISABLE }
func (s *Stack) parse(m *ice.Message, p string) *Stack {
	nfs.Open(m, p, func(r io.Reader) {
		s.key, s.peekf().key = p, p
		kit.For(r, func(text string) {
			s.list = append(s.list, text)
			for s.line = len(s.list) - 1; s.line < len(s.list); s.line++ {
				if text = s.list[s.line]; text == "" || strings.HasPrefix(text, "#") {
					continue
				}
				for s.rest = kit.Split(text, "\t ", OPS); len(s.rest) > 0; {
					ls, rest := _parse_rest(BEGIN, s.rest...)
					if ls[0] == END {
						if f := s.peekf(); f.pop == nil {
							s.last = s.popf()
						} else {
							f.pop()
						}
						s.rest = ls[1:]
						continue
					}
					kit.If(len(rest) == 0, func() { ls, rest = _parse_rest(END, ls...) })
					switch s.rest = []string{}; v := s.value(ls[0]).(type) {
					case *Func:
						f, line := s.pushf(ls[0]), s.line
						f.pop, s.line = func() { s.rest, s.line, _ = nil, line+1, s.popf() }, v.line
					default:
						if _, ok := m.Target().Commands[ls[0]]; ok {
							m.Options(STACK, s).Cmdy(ls)
						} else {
							m.Options(STACK, s).Cmdy(CMD, ls)
						}
					}
					s.rest = append(s.rest, rest...)
				}
			}
		})
	})
	return s
}
func (s *Stack) show() (res []string) {
	for i, l := range s.list {
		res = append(res, kit.Format("%2d: ", i)+l)
	}
	res = append(res, "")
	for i, f := range s.frame {
		res = append(res, kit.Format("frame: %v line: %v %v %v", i, f.line, f.key, f.status))
		kit.For(f.value, func(k string, v ice.Any) { res = append(res, kit.Format("       %v %v: %v", i, k, v)) })
	}
	return
}
func NewStack() *Stack { return &Stack{} }

func _parse_stack(m *ice.Message) *Stack           { return m.Optionv(STACK).(*Stack) }
func _parse_frame(m *ice.Message) (*Stack, *Frame) { return _parse_stack(m), _parse_stack(m).peekf() }
func _parse_rest(split string, arg ...string) ([]string, []string) {
	if i := kit.IndexOf(arg, split); i == -1 {
		return arg, nil
	} else if split == END {
		if i == 0 {
			return arg, nil
		}
		return arg[:i], arg[i:]
	} else {
		return arg[:i], arg[i+1:]
	}
}

const (
	OPS     = "<!=>+-*/;"
	EXPR    = "expr"
	BEGIN   = "{"
	END     = "}"
	DISABLE = -1

	PWD    = "pwd"
	CMD    = "cmd"
	LET    = "let"
	IF     = "if"
	FOR    = "for"
	FUNC   = "func"
	TYPE   = "type"
	SOURCE = "source"
	RETURN = "return"
)
const STACK = "stack"

func init() {
	Index.MergeCommands(ice.Commands{
		CMD: {Name: "cmd", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			kit.If(s.runable(), func() { m.Cmdy(arg) })
		}},
		LET: {Name: "let a = 1", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			kit.If(s.runable(), func() { s.value(arg[0], m.Cmdx(EXPR, arg[2:])) })
		}},
		IF: {Name: "if a > 1", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			f := s.pushf(m.CommandKey())
			kit.If(s.runable() && m.Cmdx(EXPR, arg) == ice.FALSE, func() { f.status = DISABLE })
		}},
		FOR: {Name: "for a = 1; a < 10; a++ {", Hand: func(m *ice.Message, arg ...string) {
			init, next := []string{}, []string{}
			switch strings.Count(strings.Join(arg, ""), ";") {
			case 2:
				init, arg = _parse_rest(";", arg...)
				arg, next = _parse_rest(";", arg...)
			case 1:
				arg, next = _parse_rest(";", arg...)
			}
			s := _parse_stack(m)
			m.Debug("what %#v", s.last)
			m.Debug("what %+v", s.last)
			m.Debug("what %v", s.last)
			kit.If(s.last != nil && s.last.line == s.line, func() { init = init[:0] })
			kit.If(len(init) > 0, func() { m.Cmd(EXPR, init) })
			f, line := s.pushf(m.CommandKey()), s.line
			f.pop = func() {
				kit.If(s.runable(), func() {
					s.line, s.last = line-1, s.popf()
					kit.If(len(next) > 0, func() { m.Cmd(EXPR, next) })
				}, func() {
					s.popf()
				})
			}
			kit.If(s.runable() && len(arg) > 0 && m.Cmdx(EXPR, arg) == ice.FALSE, func() { f.status = DISABLE })
		}},
		FUNC: {Name: "func show", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			s.value(arg[0], &Func{line: s.line, arg: arg[1:]})
			f := s.pushf(m.CommandKey())
			f.status = DISABLE
		}},
		RETURN: {Name: "return show", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			f := s.peekf()
			f.status = DISABLE
		}},
		PWD: {Name: "pwd", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			kit.If(s.runable(), func() {
				res := []string{kit.Format(s.line)}
				s.stack(func(f *Frame, i int) { kit.If(i > 0, func() { res = append(res, kit.Format("%v:%v", f.line, f.key)) }) })
				m.Echo(strings.Join(res, " / ")).Echo(ice.NL)
			})
		}},
		STACK: {Name: "stack path auto parse", Hand: func(m *ice.Message, arg ...string) {
			if m.Options(nfs.DIR_ROOT, nfs.SRC).Cmdy(nfs.CAT, arg); len(m.Result()) == 0 {
				return
			}
			m.SetResult()
			m.Echo(ice.NL).Echo("output: %s\n", arg[0])
			s := NewStack().parse(m, path.Join(nfs.SRC, path.Join(arg...)))
			m.Echo(ice.NL).Echo("script: %s\n", arg[0])
			m.Echo(strings.Join(s.show(), ice.NL))
		}},
		EXPR: {Name: "expr a = 1", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			level := map[string]int{
				"++": 40,
				"*":  30, "/": 30,
				"+": 20, "-": 20,
				"<": 10, ">": 10, "<=": 10, ">=": 10, "==": 10, "!=": 10,
				"=": 1,
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
				if v := s.value(k); v != nil {
					return kit.Format(v)
				}
				return k
			}
			getl := func(p int) int { return level[kit.Format(get(p))] }
			push := func(v ice.Any) { list = append(list, v) }
			pop := func(n int) { list = list[:len(list)-n] }
			ops := func() {
				switch gets(-2) {
				case "=":
					s.value(kit.Format(get(-3)), gets(-1))
					pop(2)
					return
				}
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
				if op := k + kit.Format(get(-1)); level[op] > 0 {
					if op == "++" {
						v := kit.Int(s.value(kit.Format(get(-2)))) + 1
						s.value(kit.Format(get(-2)), v)
						pop(1)
						return
					}
					pop(1)
					push(op)
					return
				}
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
			kit.If(len(list) > 0, func() { m.Echo(kit.Format(list[0])) })
			m.Debug("expr %s %v", m.Result(), s.value(m.Result()))
		}},
	})
}
