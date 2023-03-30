package yac

import (
	"io"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	EXPR    = "expr"
	OPS     = "(,){;}!=<>+-*/"
	OPEN    = "("
	FIELD   = ","
	CLOSE   = ")"
	BEGIN   = "{"
	SPLIT   = ";"
	END     = "}"
	DISABLE = -1
)

var level = map[string]int{
	"++": 100,
	"+":  20, "-": 20, "*": 30, "/": 30,
	"<": 10, ">": 10, "<=": 10, ">=": 10, "==": 10, "!=": 10,
	"(": 2, ")": 2,
	"=": 1,
}

type Expr struct {
	list ice.List

	s *Stack
}

func (s *Expr) push(v ice.Any) { s.list = append(s.list, v) }
func (s *Expr) pop(n int)      { s.list = s.list[:len(s.list)-n] }
func (s *Expr) get(p int) (v ice.Any) {
	kit.If(p+len(s.list) >= 0, func() { v = s.list[p+len(s.list)] })
	return
}
func (s *Expr) gets(p int) string { return kit.Format(s.get(p)) }
func (s *Expr) getl(p int) int    { return level[kit.Format(s.get(p))] }
func (s *Expr) getv(p int) (v ice.Any) {
	k := s.get(p)
	if v = s.s.value(kit.Format(k)); v != nil {
		return v
	} else {
		return k
	}
}
func (s *Expr) setv(k string, v ice.Any) {
	kit.If(s.s.runable(), func() { s.s.value(k, v) })
}
func (s *Expr) ops() {
	switch s.gets(-2) {
	case "=":
		s.setv(s.gets(-3), s.getv(-1))
		s.pop(2)
		return
	}
	bin := func(v ice.Any) {
		s.pop(3)
		s.push(v)
	}
	switch a, b := kit.Int(s.getv(-3)), kit.Int(s.getv(-1)); s.gets(-2) {
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
func (s *Expr) Cals(arg ...string) (ice.Any, int) {
	ice.Pulse.Debug("what %d %v", s.s.line, arg)
	for i := 0; i < len(arg); i++ {
		k := arg[i]
		switch k {
		case OPEN:
			if s.getl(-1) == 0 {
				list := kit.List(s.gets(-1))
				for {
					v, n := NewExpr(s.s).Cals(arg[i+1:]...)
					list = append(list, v)
					if i += 1 + n; arg[i] == CLOSE {
						i++
						break
					} else if arg[i] == FIELD {

					} else {
						i--
					}
				}
				s.pop(1)
				s.push(ice.Pulse.Cmdx(list...))
				continue
			}
		case CLOSE:
			if len(s.list) > 1 {
				break
			}
			fallthrough
		case BEGIN, SPLIT, FIELD:
			s.s.rest = arg[i:]
			return s.End(arg[:i]...), i
		}
		if op := s.gets(-1) + k; level[op] > 0 {
			if op == "++" {
				s.setv(s.gets(-2), kit.Int(s.s.value(s.gets(-2)))+1)
				s.pop(1)
			} else {
				s.pop(1)
				s.push(op)
			}
		} else {
			if level[k] > 0 {
				for level[k] <= s.getl(-2) {
					s.ops()
				}
			} else if len(s.list) > 0 && s.getl(-1) == 0 {
				return s.End(arg[:i]...), i
			}
			s.push(k)
		}
	}
	return s.End(arg...), len(arg)
}
func (s *Expr) End(arg ...string) ice.Any {
	ice.Pulse.Debug("what %v", s.list)
	for len(s.list) > 1 {
		s.ops()
	}
	if s.s.runable() {
		if len(s.list) > 0 {
			return s.list[0]
		}
		return nil
	}
	return arg
}
func NewExpr(s *Stack) *Expr { return &Expr{kit.List(), s} }

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
	last  *Frame
	frame []*Frame
	rest  []string
	list  []string
	line  int

	input chan string
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

var n = 0

func (s *Stack) value(key string, arg ...ice.Any) ice.Any {

	f := s.peekf()
	for i := len(s.frame) - 1; i >= 0; i-- {
		if _f := s.frame[i]; _f.value[key] != nil {
			f = _f
		}
	}
	n++
	if n > 1000 {
		panic(n)
	}
	kit.If(len(arg) > 0, func() {
		f.value[key] = arg[0]
		ice.Pulse.Debug("set value %v %v %v", f.key, key, arg[0])
	}, func() {
		ice.Pulse.Debug("get value %v %v", f.key, key)
	})
	return f.value[key]
}
func (s *Stack) runable() bool { return s.peekf().status > DISABLE }
func (s *Stack) read() (text string, ok bool) {
	for {
		if text, ok = <-s.input; !ok {
			return
		}
		if s.list = append(s.list, text); text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		return
	}
}
func (s *Stack) parse(m *ice.Message, p string) *Stack {
	s.input = make(chan string)
	m.Options(STACK, s).Go(func() {
		nfs.Open(m, p, func(r io.Reader) {
			defer close(s.input)
			kit.For(r, func(text string) { s.input <- text })
		})
	})
	s.key, s.peekf().key = p, p
	for {
		if _, ok := s.read(); !ok {
			break
		}
		for s.line = len(s.list) - 1; s.line < len(s.list); s.line++ {
			for s.rest = kit.Split(s.list[s.line], "\t ", OPS); len(s.rest) > 0; {
				if s.rest[0] == END {
					rest := s.rest[1:]
					kit.If(s.peekf().pop != nil, func() { s.peekf().pop() })
					s.last, s.rest = s.popf(), rest
					continue
				}
				ls, rest := _parse_rest("", s.rest...)
				if len(ls) == 0 {
					s.rest = rest
					continue
				}
				switch s.rest = []string{}; v := s.value(ls[0]).(type) {
				case *Func:
					f, line := s.pushf(ls[0]), s.line
					f.pop, s.line = func() { s.rest, s.line = rest, line+1 }, v.line
					continue
				default:
					if _, ok := m.Target().Commands[ls[0]]; ok {
						m.Cmdy(ls)
					} else if s.runable() {
						m.Cmdy(CMD, ls)
					}
				}
				s.rest = append(s.rest, rest...)
			}

		}
	}
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
	for i, k := range arg {
		if split != "" && split == k {
			return arg[:i], arg[i+1:]
		}
		switch k {
		case BEGIN:
			return arg[:i+1], arg[i+1:]
		case END:
			return arg[:i], arg[i:]
		}
	}
	return arg, nil
}

const (
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
			m.Cmdy(arg)
		}},
		LET: {Name: "let a = 1", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			kit.If(s.runable(), func() { s.value(arg[0], m.Cmdx(EXPR, arg[2:])) })
		}},
		IF: {Name: "if a = 1; a > 1 {", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			f := s.pushf(m.CommandKey())
			res := m.Cmdx(EXPR, arg)
			kit.If(s.rest[0] == SPLIT, func() { res = m.Cmdx(EXPR, s.rest[1:]) })
			kit.If(res == ice.FALSE, func() { f.status = DISABLE })
			s.rest = s.rest[1:]
		}},
		FOR: {Name: "for a = 1; a < 10; a++ {", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			f := s.pushf(m.CommandKey())
			list, status := [][]string{}, f.status
			for f.status, s.rest = DISABLE, arg; s.rest[0] != BEGIN; {
				if list = append(list, m.Cmd(EXPR, s.rest).Resultv()); s.rest[0] == SPLIT {
					s.rest = s.rest[1:]
				}
			}
			f.status, s.rest = status, s.rest[1:]
			res := ice.TRUE
			if len(list) == 1 {
				res = m.Cmdx(EXPR, list[0])
			} else if len(list) > 1 {
				if s.last == nil || s.last.line != s.line {
					res = m.Cmdx(EXPR, list[0])
				} else {
					kit.For(s.last.value, func(k string, v ice.Any) { f.value[k] = v })
				}
				res = m.Cmdx(EXPR, list[1])
			}
			kit.If(res == ice.FALSE, func() { f.status = DISABLE })
			f.pop = func() {
				if s.runable() {
					kit.If(len(list) > 2, func() { m.Cmd(EXPR, list[2]) })
					s.line = f.line - 1
				}
			}
		}},
		FUNC: {Name: "func show", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			s.value(arg[0], &Func{line: s.line, arg: arg[1:]})
			f := s.pushf(m.CommandKey())
			f.status = DISABLE
		}},
		RETURN: {Name: "return show", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			if len(s.frame) == 1 {
				close(s.input)
				s.line = len(s.list)
			}
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
			v, n := NewExpr(s).Cals(arg...)
			s.rest = arg[n:]
			if s.runable() {
				m.Echo(kit.Format(v))
			} else {
				m.Resultv(v)
			}
			m.Debug("expr %s %v", m.Result(), s.value(m.Result()))
		}},
	})
}
