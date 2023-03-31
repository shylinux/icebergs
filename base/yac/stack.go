package yac

import (
	"encoding/json"
	"io"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type Value struct{ list []ice.Any }

func (s *Value) MarshalJSON() ([]byte, error) { return json.Marshal(s.list) }

type Func struct {
	obj []string
	arg []string
	res []string
	Position
}
type Frame struct {
	key    string
	value  ice.Map
	status int
	Position
	pop func()
}
type Stack struct {
	last  *Frame
	frame []*Frame
	input chan string
	list  []string
	rest  []string
	Position
}
type Position struct{ line, skip int }

func (s *Stack) peekf(m *ice.Message) *Frame {
	kit.If(len(s.frame) == 0, func() { s.pushf(m, "") })
	return s.frame[len(s.frame)-1]
}
func (s *Stack) pushf(m *ice.Message, key string) *Frame {
	f := &Frame{key: kit.Select(m.CommandKey(), key), value: kit.Dict(), Position: s.Position}
	kit.If(len(s.frame) > 0, func() { f.status = s.peekf(m).status })
	m.Debug("stack push %#v", f)
	s.frame = append(s.frame, f)
	return f
}
func (s *Stack) popf(m *ice.Message) *Frame {
	f := s.peekf(m)
	m.Debug("stack pop %#v", f)
	kit.If(len(s.frame) > 1, func() { s.frame = s.frame[:len(s.frame)-1] })
	return f
}
func (s *Stack) stack(cb ice.Any) {
	for i := len(s.frame) - 1; i >= 0; i-- {
		switch cb := cb.(type) {
		case func(*Frame, int) bool:
			if cb(s.frame[i], i) {
				return
			}
		case func(*Frame, int):
			cb(s.frame[i], i)
		case func(*Frame):
			cb(s.frame[i])
		}
	}
}
func (s *Stack) value(m *ice.Message, key string, arg ...ice.Any) ice.Any {
	f := s.peekf(m)
	s.stack(func(_f *Frame, i int) bool {
		if _f.value[key] != nil {
			f = _f
			return true
		}
		return false
	})
	kit.If(len(arg) > 0, func() {
		m.Debug("value set %v %#v", key, arg[0])
		switch v := arg[0].(type) {
		case *Value:
			f.value[key] = v.list[0]
		default:
			f.value[key] = arg[0]
		}
	})
	// m.Debug("value get %v %v", key, f.value[key])
	return f.value[key]
}
func (s *Stack) runable(m *ice.Message) bool { return s.peekf(m).status > DISABLE }
func (s *Stack) read(m *ice.Message) (text string, ok bool) {
	isvoid := func(text string) bool {
		return strings.TrimSpace(text) == "" || strings.HasPrefix(strings.TrimSpace(text), "#")
	}
	for s.line++; s.line < len(s.list); s.line++ {
		if isvoid(s.list[s.line]) {
			continue
		}
		return s.list[s.line], true
	}
	for {
		if text, ok = <-s.input; !ok {
			break
		}
		m.Debug("input read %v", text)
		if s.list = append(s.list, text); isvoid(text) {
			continue
		}
		s.line = len(s.list) - 1
		break
	}
	return
}
func (s *Stack) reads(m *ice.Message, cb func(i int, k string) bool) (arg []string) {
	i := 0
	for {
		if s.skip++; s.skip < len(s.rest) {
			arg = append(arg, s.rest[s.skip])
			if cb(i, s.rest[s.skip]) {
				break
			}
			i++
		} else if text, ok := s.read(m); ok {
			s.rest, s.skip = kit.Split(text, "\t ", OPS), -1
		} else {
			break
		}
	}
	return
}
func (s *Stack) call(m *ice.Message) {
	s.reads(m, func(i int, k string) bool {
		if k == SPLIT {

		} else if k == END {
			kit.If(s.peekf(m).pop != nil, func() { s.peekf(m).pop() })
			s.last = s.popf(m)
		} else if _, ok := m.Target().Commands[k]; ok {
			m.Cmdy(k)
		} else {
			s.skip--
			m.Cmdy(EXPR)
		}
		return false
	})
}
func (s *Stack) parse(m *ice.Message, r io.Reader) *ice.Message {
	s.input = make(chan string, 100)
	m.Options(STACK, s)
	kit.For(r, func(text string) { s.input <- text })
	close(s.input)

	// s.input = make(chan string)
	// m.Options(STACK, s).Go(func() {
	// 	defer func() { kit.If(s.input != nil, func() { close(s.input) }) }()
	// 	kit.For(r, func(text string) { s.input <- text })
	// })
	s.peekf(m)
	s.call(m)
	return m
}
func (s *Stack) token(m *ice.Message) string {
	if s.skip < len(s.rest) {
		return s.rest[s.skip]
	}
	return ""
}
func (s *Stack) expr(m *ice.Message, pos Position) string {
	s.Position = pos
	return m.Cmdx(EXPR)
}
func NewStack() *Stack                   { return &Stack{} }
func _parse_stack(m *ice.Message) *Stack { return m.Optionv(STACK).(*Stack) }
func _parse_frame(m *ice.Message) (*Stack, *Frame) {
	return _parse_stack(m), _parse_stack(m).pushf(m, "")
}

const (
	PWD    = "pwd"
	CMD    = "cmd"
	LET    = "let"
	IF     = "if"
	FOR    = "for"
	FUNC   = "func"
	CALL   = "call"
	RETURN = "return"
	EXPR   = "expr"
)
const STACK = "stack"

func init() {
	Index.MergeCommands(ice.Commands{
		PWD: {Name: "pwd", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			res := []string{kit.Format("%d:%d", s.line, s.skip)}
			s.stack(func(f *Frame, i int) {
				kit.If(i > 0, func() {
					res = append(res, kit.Format("%d:%d %s %v", f.line, f.skip, f.key, kit.Select(ice.FALSE, ice.TRUE, f.status > DISABLE)))
				})
			})
			m.Echo(strings.Join(res, " / ")).Echo(ice.NL)
		}},
		CMD: {Name: "cmd", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			kit.If(s.runable(m), func() {
				m.Cmdy(s.rest[s.skip+1:])
				m.EchoLine("")
			})
			s.skip = len(s.rest)
		}},
		LET: {Name: "let a, b = 1, 2", Hand: func(m *ice.Message, arg ...string) { m.Cmd(EXPR) }},
		IF: {Name: "if a = 1; a > 1 {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			res := m.Cmdx(EXPR)
			kit.If(s.token(m) == SPLIT, func() { res = m.Cmdx(EXPR) })
			kit.If(res == ice.FALSE, func() { f.status = DISABLE })
		}},
		FOR: {Name: "for a = 1; a < 10; a++ {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			list, status := []Position{s.Position}, f.status
			for f.status = DISABLE; s.token(m) != BEGIN; {
				m.Cmd(EXPR)
				list = append(list, s.Position)
			}
			f.status = status
			res := ice.TRUE
			if len(list) < 3 {
				res = s.expr(m, list[0])
			} else {
				if s.last == nil || s.last.line != s.line {
					res = s.expr(m, list[0])
				} else {
					kit.For(s.last.value, func(k string, v ice.Any) { f.value[k] = v })
				}
				res = s.expr(m, list[1])
			}
			kit.If(res == ice.FALSE, func() { f.status = DISABLE })
			s.Position, f.pop = list[len(list)-1], func() {
				if s.runable(m) {
					kit.If(len(list) > 3, func() { s.expr(m, list[2]) })
					s.Position = list[0]
					s.Position.skip--
				}
			}
		}},
		FUNC: {Name: "func show(a, b) (c, d)", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			list, key, kind := [][]string{[]string{}}, "", ""
			push := func() { kit.If(key, func() { list[len(list)-1], key, kind = append(list[len(list)-1], key), "", "" }) }
			s.reads(m, func(i int, k string) bool {
				switch k {
				case OPEN:
					defer kit.If(i > 0, func() { list = append(list, []string{}) })
				case FIELD, CLOSE:
				case BEGIN:
					return true
				default:
					kit.If(key, func() { kind = k }, func() { key = k })
					return false
				}
				push()
				return false
			})
			kit.If(len(list) < 2, func() { list = append(list, []string{}) })
			kit.If(len(list) < 3, func() { list = append(list, []string{}) })
			s.value(m, kit.Select("", list[0], -1), &Func{obj: list[0], arg: list[1], res: list[2], Position: s.Position})
			s.pushf(m, "").status = DISABLE
		}},
		CALL: {Name: "call show", Hand: func(m *ice.Message, arg ...string) {
			m.Echo("%v", NewExpr(m, _parse_stack(m)).call(m, arg[0]))
		}},
		RETURN: {Name: "return show", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			if len(s.frame) == 1 {
				close(s.input)
				s.input = nil
			}
			f := s.peekf(m)
			f.status = DISABLE
		}},
		STACK: {Name: "stack path auto parse", Hand: func(m *ice.Message, arg ...string) {
			if m.Options(nfs.DIR_ROOT, nfs.SRC).Cmdy(nfs.CAT, arg); len(m.Result()) == 0 {
				return
			}
			nfs.Open(m, path.Join(nfs.SRC, path.Join(arg...)), func(r io.Reader) {
				msg := NewStack().parse(m.Spawn(), r)
				if m.SetResult(); m.Option(log.DEBUG) != ice.TRUE {
					m.Copy(msg)
					return
				}
				s := _parse_stack(msg)
				m.EchoLine("script: %s", arg[0])
				span := func(s, k, t string) string {
					return strings.ReplaceAll(s, k, kit.Format("<span class='%s'>%s</span>", t, k))
				}
				kit.For(s.list, func(i int, s string) {
					kit.For([]string{"let", "if", "for", "func"}, func(k string) { s = span(s, k, "keyword") })
					kit.For([]string{"pwd", "cmd"}, func(k string) { s = span(s, k, "function") })
					m.EchoLine("%2d: %s", i, s)
				})
				m.EchoLine("").EchoLine("output: %s", arg[0]).Copy(msg)
			})
		}},
		EXPR: {Name: "expr a = 1", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			arg = s.rest
			if v := NewExpr(m, s).cals(m); !s.runable(m) {
				m.Resultv(v)
			} else if v != nil {
				m.Echo(kit.Format(v))
			} else if s.token(m) == BEGIN {
				m.Echo(ice.TRUE)
			}
			m.Debug("expr value %v %v", m.Result(), arg)
		}},
	})
}

const (
	OPS     = "(,){;}*/+-!=<>&|"
	OPEN    = "("
	FIELD   = ","
	CLOSE   = ")"
	BEGIN   = "{"
	SPLIT   = ";"
	END     = "}"
	DISABLE = -1
)

var level = map[string]int{
	"++": 100, "--": 100, "!": 90,
	"*": 40, "/": 40, "+": 30, "-": 30,
	"==": 20, "!=": 20, "<=": 20, ">=": 20, "<": 20, ">": 20, "&&": 10, "||": 10,
	"=": 2, ",": 2, "(": 1, ")": 1,
}

type Expr struct {
	list ice.List
	s    *Stack
	m    *ice.Message
}

func (s *Expr) push(v ice.Any) *Expr {
	s.list = append(s.list, v)
	return s
}
func (s *Expr) pop(n int) *Expr {
	s.list = s.list[:len(s.list)-n]
	return s
}
func (s *Expr) get(p int) (v ice.Any) {
	kit.If(0 <= p+len(s.list) && p+len(s.list) < len(s.list), func() { v = s.list[p+len(s.list)] })
	kit.If(0 <= p && p < len(s.list), func() { v = s.list[p] })
	return
}
func (s *Expr) gets(p int) string { return kit.Format(s.get(p)) }
func (s *Expr) getl(p int) int    { return level[s.gets(p)] }
func (s *Expr) getv(m *ice.Message, p int) (v ice.Any) {
	if !s.s.runable(m) {
		return nil
	}
	k := s.get(p)
	if v = s.s.value(m, kit.Format(k)); v != nil {
		return v
	}
	return k
}
func (s *Expr) setv(m *ice.Message, k string, v ice.Any) *Expr {
	kit.If(s.s.runable(m), func() {
		switch v := v.(type) {
		case *Value:
			s.s.value(m, k, v.list[0])
		default:
			s.s.value(m, k, v)
		}
	})
	return s
}
func (s *Expr) ops(m *ice.Message) {
	m.Debug("expr ops %v", s.list)
	bin := func(v ice.Any) { s.pop(3).push(v) }
	switch a, b := kit.Int(s.getv(m, -3)), kit.Int(s.getv(m, -1)); s.gets(-2) {
	case "*":
		bin(a * b)
	case "/":
		bin(a / b)
	case "+":
		bin(a + b)
	case "-":
		bin(a - b)
	case ">":
		bin(a > b)
	case "<":
		bin(a < b)
	case "<=":
		bin(a <= b)
	case ">=":
		bin(a >= b)
	case "!=":
		bin(a != b)
	case "==":
		bin(a == b)
	}
}
func (s *Expr) end(m *ice.Message, arg ...string) ice.Any {
	if s.gets(-1) == CLOSE {
		s.pop(1)
	}
	for len(s.list) > 1 {
		switch s.ops(m); s.gets(-2) {
		case ",":
			list := kit.List()
			for i := len(s.list) - 2; i > 0; i -= 2 {
				if s.list[i] == "=" {
					for j := 0; j < i; j += 2 {
						s.setv(m, s.gets(j), s.getv(m, j+i+1))
						list = append(list, s.getv(m, j))
					}
					s.list = kit.List(&Value{list: list})
					break
				}
			}
			if len(s.list) == 1 {
				break
			}
			for i := 0; i < len(s.list); i += 2 {
				list = append(list, s.getv(m, i))
			}
			s.list = kit.List(&Value{list: list})
		case "=":
			if len(s.list) == 3 {
				s.setv(m, s.gets(-3), s.getv(m, -1)).pop(2)
				break
			}
			list := kit.List()
			switch v := s.getv(m, -1).(type) {
			case *Value:
				for i := 0; i < len(s.list)-1; i += 2 {
					if i/2 < len(v.list) {
						s.setv(m, s.gets(i), v.list[i/2])
						list = append(list, s.getv(m, i))
					}
				}
			}
			s.list = kit.List(&Value{list})
		}
	}
	if !s.s.runable(m) {
		return arg
	}
	if len(s.list) > 0 {
		return s.list[0]
	}
	return nil
}

func (s *Expr) cals(m *ice.Message) ice.Any {
	m.Debug("expr calcs %d %v", s.s.line, s.s.rest[s.s.skip:])
	arg := s.s.reads(m, func(i int, k string) bool {
		switch k {
		case SPLIT, BEGIN:
			return true
		case END:
			s.s.skip--
			return true
		}
		if level[k] == 0 && len(s.list) > 0 && s.getl(-1) == 0 {
			s.s.skip--
			return true
		} else if k == OPEN && len(s.list) > 0 && s.getl(-1) == 0 {
			value := s.call(m, s.get(-1))
			s.pop(1).push(value)
		} else if op := s.gets(-1) + k; level[op] > 0 && s.getl(-1) > 0 {
			if op == "++" {
				s.setv(m, s.gets(-2), kit.Int(s.s.value(m, s.gets(-2)))+1).pop(1)
			} else {
				s.pop(1).push(op)
			}
		} else if level[k] > 0 {
			for level[k] >= 9 && level[k] <= s.getl(-2) {
				s.ops(m)
			}
			if k == CLOSE {
				if s.gets(-2) == OPEN {
					v := s.get(-1)
					s.pop(2).push(v)
					return false
				}
				return true
			}
			s.push(k)
		} else {
			s.push(k)
		}
		return false
	})
	return s.end(m, arg...)
}
func (s *Expr) call(m *ice.Message, name ice.Any) (v ice.Any) {
	m.Debug("call %v", name)
	list := kit.List(name)
	switch v = NewExpr(m, s.s).cals(m); v := v.(type) {
	case *Value:
		list = append(list, v.list...)
	default:
		list = append(list, v)
	}
	m.Debug("call %v", list)
	switch v := s.s.value(m, kit.Format(name)).(type) {
	case *Func:
		f := s.s.pushf(m, "")
		kit.For(v.res, func(k string) { f.value[k] = "" })
		kit.For(v.arg, func(k string) { f.value[k] = "" })
		kit.For(v.arg, func(i int, k string) { kit.If(i+1 < len(list), func() { f.value[k] = list[i+1] }) })
		value, pos := &Value{list: kit.List()}, s.s.Position
		f.pop, s.s.Position = func() {
			kit.For(v.res, func(k string) { value.list = append(value.list, f.value[k]) })
			s.s.Position = pos
		}, v.Position
		m.Debug("call %#v", f)
		s.s.call(m)
		m.Debug("call %#v", value)
		return value
	default:
		if s.s.runable(m) {
			return m.Cmdx(list...)
		}
		return nil
	}
}
func NewExpr(m *ice.Message, s *Stack) *Expr { return &Expr{kit.List(), s, m} }
