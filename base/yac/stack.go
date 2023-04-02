package yac

import (
	"bufio"
	"io"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type Function struct {
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
	Position
	n int
}
type Position struct {
	rest []string
	skip int
	line int
	*Buffer
}
type Buffer struct {
	name  string
	list  []string
	input *bufio.Scanner
}

func (s *Stack) peekf() *Frame { return s.frame[len(s.frame)-1] }
func (s *Stack) pushf(m *ice.Message, key string) *Frame {
	f := &Frame{key: kit.Select(m.CommandKey(), key), value: kit.Dict(), Position: s.Position}
	kit.If(len(s.frame) > 0, func() { f.status = s.peekf().status })
	m.Debug("stack push %d %v %s:%d", len(s.frame), f.key, f.name, f.line)
	s.frame = append(s.frame, f)
	return f
}
func (s *Stack) popf(m *ice.Message) *Frame {
	f := s.peekf()
	list := kit.List(f.value["_defer"])
	delete(f.value, "_defer")
	for i := len(list) - 1; i >= 0; i-- {
		list[i].(func())()
	}
	kit.If(f.pop != nil, func() { f.pop() })
	m.Debug("stack pop %d %v %s:%d", len(s.frame)-1, f.key, f.name, f.line)
	kit.If(len(s.frame) > 0, func() { s.frame = s.frame[:len(s.frame)-1] })
	s.last = f
	return f
}
func (s *Stack) stack(cb func(*Frame, int) bool) {
	for i := len(s.frame) - 1; i >= 0; i-- {
		if cb(s.frame[i], i) {
			return
		}
	}
}
func (s *Stack) value(m *ice.Message, key string, arg ...Any) Any {
	f, n := s.peekf(), len(s.frame)-1
	if len(arg) < 2 || arg[1] != DEFINE {
		s.stack(func(_f *Frame, i int) bool {
			if _f.value[key] != nil {
				f, n = _f, i
				return true
			}
			return false
		})
	}
	kit.If(len(arg) > 0, func() {
		m.Debug("value set %d %v %#v", n, key, arg[0])
		f.value[key] = arg[0]
	})
	return f.value[key]
}
func (s *Stack) runable() bool { return s.peekf().status > STATUS_DISABLE }
func (s *Stack) token() string { return kit.Select("", s.rest, s.skip) }
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
	for s.input != nil && s.input.Scan() {
		text = s.input.Text()
		m.Debug("input read %d \"%s\" %s:%d", len(s.list), text, s.name, len(s.list))
		if s.line, s.list = len(s.list), append(s.list, text); isvoid(text) {
			continue
		}
		return text, true
	}
	return
}
func (s *Stack) reads(m *ice.Message, cb func(k string) bool) {
	for {
		if s.skip++; s.skip < len(s.rest) {
			if cb(s.rest[s.skip]) {
				break
			}
		} else if text, ok := s.read(m); ok {
			s.rest, s.skip = kit.Split(text, SPACE, BLOCK, QUOTE, TRANS, ice.TRUE), -1
		} else {
			cb(SPLIT)
			break
		}
	}
}
func (s *Stack) run(m *ice.Message) {
	begin := len(s.frame)
	s.reads(m, func(k string) bool {
		if s.n++; s.n > 300 {
			panic(s.n)
		}
		if k == SPLIT {

		} else if k == END {
			if s.last = s.popf(m); len(s.frame) < begin {
				return true
			}
		} else if _, ok := m.Target().Commands[k]; ok {
			m.Cmdy(k, kit.Slice(s.rest, s.skip+1))
		} else {
			s.skip--
			m.Cmd(EXPR, kit.Slice(s.rest, s.skip))
		}
		return false
	})
}
func (s *Stack) call(m *ice.Message, obj Any, key Any, cb func(*Frame, Function), arg ...Any) Any {
	if _k, ok := key.(string); ok {
		kit.For(kit.Split(_k, ice.PT), func(k string) {
			switch v := obj.(type) {
			case *Stack:
				if v := v.value(m, _k); v != nil {
					obj = v
					break
				}
				obj, key = v.value(m, k), strings.TrimPrefix(_k, k+ice.PT)
			}
		})
	}
	m.Debug("expr call %T %s %v", obj, key, kit.Format(arg))
	switch obj := obj.(type) {
	case Function:
		f := s.pushf(m, CALL)
		kit.For(obj.res, func(k string) { f.value[k] = "" })
		kit.For(obj.arg, func(i int, k string) {
			kit.If(i < len(arg), func() { f.value[k] = arg[i] }, func() { f.value[k] = "" })
		})
		value, pos := Value{list: kit.List()}, s.Position
		f.value["_return"] = func(arg ...Any) {
			if len(obj.res) > 0 {
				kit.For(obj.res, func(i int, k string) {
					kit.If(i < len(arg), func() { value.list = append(value.list, f.value[k]) }, func() { value.list = append(value.list, "") })
				})
			} else {
				value.list = arg
			}
		}
		f.pop, s.Position = func() {
			if len(obj.res) > 0 && len(value.list) == 0 {
				kit.For(obj.res, func(i int, k string) { value.list = append(value.list, f.value[k]) })
			}
			s.Position = pos
		}, obj.Position
		kit.If(cb != nil, func() { cb(f, obj) })
		s.run(m.Options(STACK, s))
		return value
	case Caller:
		return obj.Call(kit.Format(key), arg...)
	case func(string, ...Any) Any:
		return obj(kit.Format(key), arg...)
	case func():
		obj()
		return nil
	default:
		args := kit.List(key)
		for _, v := range arg {
			switch v := v.(type) {
			case Value:
				args = append(args, v.list...)
			default:
				args = append(args, trans(v))
			}
		}
		return Message{m.Cmd(args...)}
	}
}
func (s *Stack) cals(m *ice.Message) Any { return NewExpr(s).cals(m) }
func (s *Stack) expr(m *ice.Message, pos ...Position) string {
	kit.If(len(pos) > 0, func() { s.Position = pos[0] })
	return m.Cmdx(EXPR, kit.Slice(s.rest, s.skip))
}
func (s *Stack) funcs(m *ice.Message) string {
	name := kit.Format("%s:%d:%d", s.name, s.line, s.skip)
	s.rest[s.skip] = name
	s.skip--
	m.Cmd(FUNC, name)
	f := s.peekf()
	status := f.status
	defer func() { f.status = status }()
	f.status = STATUS_DISABLE
	s.run(m)
	return name
}
func (s *Stack) parse(m *ice.Message, name string, r io.Reader, cb func(*Frame)) *Stack {
	s.Buffer = &Buffer{name: name, input: bufio.NewScanner(r)}
	s.load(m)
	kit.If(cb != nil, func() { cb(s.peekf()) })
	s.run(m)
	return s
}
func NewStack() *Stack { return &Stack{} }

func _parse_stack(m *ice.Message) *Stack { return m.Optionv(STACK).(*Stack) }
func _parse_frame(m *ice.Message) (*Stack, *Frame) {
	return _parse_stack(m), _parse_stack(m).pushf(m, "")
}

const (
	STATUS_NORMAL  = 0
	STATUS_DISABLE = -1
)
const STACK = "stack"

func init() {
	Index.MergeCommands(ice.Commands{
		STACK: {Name: "stack path auto parse", Actions: ice.Actions{
			ice.CMD: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Open(m, existsFile(m, arg[0]), func(r io.Reader, p string) {
					meta := kit.Dict()
					kit.For(NewStack().parse(m.Spawn(), p, r, nil).peekf().value, func(k string, v Any) {
						switch v := v.(type) {
						case Function:
							list := kit.List()
							kit.For(v.arg, func(k string) {
								switch k {
								case "m", ice.ARG:
								default:
									list = append(list, kit.Dict(mdb.NAME, k, mdb.TYPE, mdb.TEXT, mdb.VALUE, ""))
								}
							})
							kit.If(k == mdb.LIST, func() { list = append(list, kit.Dict(mdb.NAME, mdb.LIST, mdb.TYPE, "button", mdb.ACTION, ice.AUTO)) })
							meta[k] = list
						}
					})
					m.Push(ice.INDEX, arg[0])
					m.Push(mdb.NAME, arg[0])
					m.Push(mdb.HELP, arg[0])
					m.Push(mdb.LIST, kit.Format(meta[mdb.LIST]))
					m.Push(mdb.META, kit.Format(meta))
				})
			}},
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Open(m, existsFile(m, arg[0]), func(r io.Reader, p string) {
					s := NewStack().parse(m.Spawn(), p, r, nil)
					action := mdb.LIST
					if len(arg) > 2 && arg[1] == ice.ACTION && s.value(m, arg[2]) != nil {
						action, arg = arg[2], arg[3:]
					} else {
						arg = arg[1:]
					}
					i := 0
					s.call(m, s, action, func(f *Frame, v Function) {
						kit.For(v.arg, func(k string) {
							switch k {
							case "m":
								f.value[k] = Message{m}
							case ice.ARG:
								list := kit.List()
								kit.For(arg, func(v string) { list = append(list, String{v}) })
								f.value[k] = Value{list}
							default:
								f.value[k] = String{m.Option(k, kit.Select(m.Option(k), arg, i))}
								i++
							}
						})
					})
				})
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Options(nfs.DIR_ROOT, nfs.SRC).Cmdy(nfs.CAT, arg); len(m.Result()) == 0 {
				return
			}
			nfs.Open(m, path.Join(nfs.SRC, path.Join(arg...)), func(r io.Reader, p string) {
				if NewStack().parse(m, p, r, nil); m.Option(ice.DEBUG) != ice.TRUE {
					return
				}
				m.EchoLine("").EchoLine("stack: %s", arg[0]).Cmdy(INFO)

			})
		}},
	})
}
func existsFile(m *ice.Message, p string) string {
	return nfs.SRC + strings.Replace(p, ice.PT, ice.PS, -1) + ice.PT + nfs.SHY
}
func ExistsFile(m *ice.Message, p string) bool { return nfs.Exists(m, existsFile(m, p)) }
