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
	if len(arg) < 2 || arg[1] == ASSIGN {
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
func (s *Stack) runable() bool { return s.peekf().status > DISABLE }
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
	s.reads(m, func(k string) bool {
		if s.n++; s.n > 100 {
			return true
		}

		if k == SPLIT {

		} else if k == END {
			s.last = s.popf(m)
		} else if _, ok := m.Target().Commands[k]; ok {
			m.Cmdy(k, kit.Slice(s.rest, s.skip+1))
		} else {
			s.skip--
			m.Cmd(EXPR, kit.Slice(s.rest, s.skip))
		}
		if len(s.frame) == 0 {
			return true
		}
		return false
	})
}
func (s *Stack) call(m *ice.Message, v Any, cb func(*Frame, Function), arg ...Any) Any {
	switch v := v.(type) {
	case Function:
		f := s.pushf(m, CALL)
		kit.For(v.res, func(k string) { f.value[k] = "" })
		kit.For(v.arg, func(i int, k string) {
			if i < len(arg) {
				f.value[k] = arg[i]
			} else {
				f.value[k] = ""
			}
		})
		f.value["_return"], f.value["_res"] = "", v.res
		value, pos := Value{list: kit.List()}, s.Position
		f.pop, s.Position = func() {
			if len(v.res) == 0 {
				value.list = append(value.list, f.value["_return"])
			} else {
				kit.For(v.res, func(k string) { value.list = append(value.list, f.value[k]) })
			}
			s.Position = pos
		}, v.Position
		cb(f, v)
		s.run(m.Options(STACK, s))
		return value
	default:
		return nil
	}
}
func (s *Stack) cals(m *ice.Message) Any { return NewExpr(m, s).cals(m) }
func (s *Stack) expr(m *ice.Message, pos ...Position) string {
	kit.If(len(pos) > 0, func() { s.Position = pos[0] })
	return m.Cmdx(EXPR, kit.Slice(s.rest, s.skip))
}
func (s *Stack) load(m *ice.Message) *Stack {
	f := s.pushf(m.Options(STACK, s), "")
	f.value["kit"] = func(key string, arg ...Any) Any {
		kit.For(arg, func(i int, v Any) { arg[i] = trans(v) })
		switch key {
		case "Dict":
			return Dict{kit.Dict(arg...)}
		case "List":
			return List{kit.List(arg...)}
		default:
			m.ErrorNotImplement(key)
			return nil
		}
	}
	f.value["m"] = Message{m}
	return s
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
					s.call(m, s.value(m, action), func(f *Frame, v Function) {
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
				s := NewStack().parse(m, p, r, nil)
				if m.Option(ice.DEBUG) != ice.TRUE {
					return
				}
				m.EchoLine("").EchoLine("script: %s", arg[0])
				span := func(s, k, t string) string {
					return strings.ReplaceAll(s, k, kit.Format("<span class='%s'>%s</span>", t, k))
				}
				kit.For(s.list, func(i int, s string) {
					kit.For([]string{LET, IF, FOR, FUNC}, func(k string) { s = span(s, k, KEYWORD) })
					kit.For([]string{PWD, INFO, SOURCE}, func(k string) { s = span(s, k, FUNCTION) })
					m.EchoLine("%2d: %s", i, s)
				})
				m.EchoLine("").EchoLine("stack: %s", arg[0]).Cmdy(INFO)

			})
		}},
	})
}
func existsFile(m *ice.Message, p string) string {
	return nfs.SRC + strings.Replace(p, ice.PT, ice.PS, -1) + ice.PT + nfs.SHY
}
func ExistsFile(m *ice.Message, p string) bool { return nfs.Exists(m, existsFile(m, p)) }
