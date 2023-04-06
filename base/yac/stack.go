package yac

import (
	"bufio"
	"io"
	"path"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type Function struct {
	obj []Field
	arg []Field
	res []Field
	Position
	object Object
}
type Frame struct {
	key    string
	value  ice.Map
	defers []func()
	status int
	Position
}
type Stack struct {
	last  *Frame
	frame []*Frame
	Position
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
	m.Debug("stack %d push %s %s", len(s.frame), f.key, s.show())
	s.frame = append(s.frame, f)
	return f
}
func (s *Stack) popf(m *ice.Message) *Frame {
	f := s.peekf()
	for i := len(f.defers) - 1; i >= 0; i-- {
		f.defers[i]()
	}
	m.Debug("stack %d pop %s %s", len(s.frame)-1, f.key, s.show())
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
	keys := strings.Split(key, ice.PT)
	kit.If(strings.Contains(key, ice.PS), func() { keys = []string{key} })
	kit.If(len(arg) > 0, func() {
		var v Any = Dict{f.value}
		for i := 0; i < len(keys); i++ {
			switch k := keys[i]; _v := v.(type) {
			case Operater:
				if i == len(keys)-1 {
					m.Debug("value %d:%s set %v %#v", n, f.key, key, arg[0])
					_v.Operate(k, arg[0])
				} else {
					if v = _v.Operate(SUBS, k); v == nil {
						if _, e := strconv.ParseInt(keys[i+1], 10, 32); e == nil {
							v = _v.Operate(k, List{})
						} else {
							v = _v.Operate(k, Dict{kit.Dict()})
						}
					}
				}
			default:
				v = nil
			}
		}
	})
	v, ok := f.value[key]
	if ok {
		return v
	} else {
		if s.stack(func(_f *Frame, i int) bool {
			v, ok = _f.value[key]
			return ok
		}); ok {
			return v
		}
	}
	v = s
	kit.For(keys, func(k string) {
		switch _v := v.(type) {
		case Operater:
			v = _v.Operate(SUBS, k)
		case *Stack:
			v = nil
			_v.stack(func(_f *Frame, i int) bool {
				v, ok = _f.value[k]
				return ok
			})
		default:
			v = nil
		}
	})
	if v != nil {
		return v
	}
	return _parse_const(m, key)
}
func (s *Stack) runable() bool { return s.peekf().status > STATUS_DISABLE }
func (s *Stack) token() string { return kit.Select("", s.rest, s.skip) }
func (s *Stack) show() string {
	if s.Buffer == nil {
		return ""
	} else if s.skip == -1 {
		return kit.Format("%s:%d", s.name, s.line+1)
	} else {
		return kit.Format("%s:%d:%d", s.name, s.line+1, s.skip)
	}
}
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
		if s.line, s.list = len(s.list), append(s.list, text); isvoid(text) {
			continue
		}
		m.Debug("input %d read \"%s\" %s", len(s.list), text, s.show())
		return text, true
	}
	return
}
func (s *Stack) reads(m *ice.Message, cb func(k string) bool) {
	block, last := []string{}, 0
	for {
		if s.skip++; s.skip < len(s.rest) {
			if k := s.rest[s.skip]; k == "`" {
				if len(block) > 0 {
					kit.If(s.line != last, func() { block, last = append(block, ice.NL), s.line })
					block = append(block, k)
					cb(strings.Join(block, ice.SP))
					block = block[:0]
				} else {
					block = append(block, k)
				}
				continue
			} else if len(block) > 0 {
				kit.If(s.line != last, func() { block, last = append(block, ice.NL), s.line })
				block = append(block, k)
				continue
			}
			if s.rest[s.skip] == ice.PS && kit.Select("", s.rest, s.skip+1) == ice.PS {
				s.skip = len(s.rest)
				continue
			}
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
func (s *Stack) nextLine(m *ice.Message) []string {
	s.skip = len(s.rest)
	s.reads(m, func(string) bool { return true })
	return s.rest
}
func (s *Stack) next(m *ice.Message) string {
	s.reads(m, func(k string) bool { return true })
	return s.token()
}
func (s *Stack) peek(m *ice.Message) string {
	pos := s.Position
	defer func() { s.Position = pos }()
	s.reads(m, func(k string) bool { return true })
	return s.token()
}
func (s *Stack) pos(m *ice.Message, pos Position, n int) {
	s.Position = pos
	s.skip += n
}
func (s *Stack) run(m *ice.Message) {
	begin := len(s.frame)
	s.reads(m, func(k string) bool {
		if k == SPLIT {

		} else if k == END {
			if s.last = s.popf(m); len(s.frame) < begin {
				return true
			}
		} else if _, ok := m.Source().Commands[k]; ok {
			m.Cmdy(k, kit.Slice(s.rest, s.skip+1))
		} else if _, ok := m.Target().Commands[k]; ok {
			m.Cmdy(k, kit.Slice(s.rest, s.skip+1))
		} else {
			if s.skip--; s.skip == -1 {
				m.Cmd(EXPR, s.rest, ice.SP, s.show())
			} else {
				m.Cmd(EXPR, kit.Slice(s.rest, s.skip), ice.SP, s.show())
			}
		}
		return false
	})
}
func (s *Stack) expr(m *ice.Message, pos ...Position) string {
	kit.If(len(pos) > 0, func() { s.Position = pos[0] })
	return m.Cmdx(EXPR, kit.Slice(s.rest, s.skip))
}
func (s *Stack) cals(m *ice.Message, arg ...string) Any {
	sub := NewExpr(s)
	return sub.cals(m, arg...)
}
func (s *Stack) cals0(m *ice.Message, arg ...string) string {
	sub := NewExpr(s)
	sub.cals(m, arg...)
	return sub.gets(0)
}
func (s *Stack) types(m *ice.Message) Any {
	for ; s.skip < len(s.rest); s.skip++ {
		switch s.token() {
		case "*":
		case MAP:
			s.skip += 2
			key := s.types(m)
			s.skip += 2
			return Map{key: key, value: s.types(m)}
		case SUBS:
			s.skip += 2
			return Slice{value: s.types(m)}
		case STRUCT:
			key, t := []string{}, Struct{index: map[string]Any{}}
			for s.next(m); s.next(m) != END; {
				if key = append(key, s.token()); s.next(m) == FIELD {
					continue
				}
				types := s.types(m)
				kit.For(key, func(key string) {
					field := Field{key, types}
					t.field = append(t.field, field)
					t.index[key] = field
				})
				key, s.skip = key[:0], len(s.rest)
			}
			return t
		case INTERFACE:
			t := Interface{index: map[string]Function{}}
			for s.next(m); s.next(m) != END; {
				name := s.token()
				field, list := Field{}, [][]Field{}
				for s.skip++; s.skip < len(s.rest); s.skip++ {
					switch s.token() {
					case OPEN:
						list = append(list, []Field{})
					case "*":
					case FIELD, CLOSE:
						list[len(list)-1] = append(list[len(list)-1], field)
						field = Field{}
					default:
						switch t := s.types(m).(type) {
						case string:
							kit.If(field.name == "", func() { field.name = t }, func() { field.kind = t })
						default:
							field.kind = s.types(m)
						}
					}
				}
				kit.If(len(list) == 1, func() { list = append(list, []Field{}) })
				t.index[name] = Function{arg: list[0], res: list[1]}
				s.skip = len(s.rest)
			}
			return t
		case FUNC:
			field, list := Field{}, [][]Field{}
			for s.skip++; s.skip < len(s.rest); s.skip++ {
				switch s.token() {
				case OPEN:
					list = append(list, []Field{})
				case "*":
				case FIELD, CLOSE:
					list[len(list)-1] = append(list[len(list)-1], field)
					field = Field{}
				default:
					switch t := s.types(m).(type) {
					case string:
						kit.If(field.name == "", func() { field.name = t }, func() { field.kind = t })
					default:
						field.kind = s.types(m)
					}
				}
			}
			kit.If(len(list) == 1, func() { list = append(list, []Field{}) })
			return Function{arg: list[0], res: list[1]}
		default:
			// if t := s.value(m, s.token()); t != nil && t != "" {
			// 	return t
			// }
			return s.token()
		}
	}
	return ""
}
func (s *Stack) funcs(m *ice.Message) string {
	name := s.show()
	s.rest[s.skip], s.skip = name, s.skip-1
	m.Cmd(FUNC, name)
	f := s.peekf()
	status := f.status
	defer func() { f.status = status }()
	f.status = STATUS_DISABLE
	s.run(m)
	return name
}
func (s *Stack) calls(m *ice.Message, obj Any, key Any, cb func(*Frame, Function), arg ...Any) Any {
	if _k, ok := key.(string); ok && _k != "" {
		kit.For(kit.Split(_k, ice.PT), func(k string) {
			switch v := obj.(type) {
			case Operater:
				obj = v.Operate(SUBS, k)
			case *Stack:
				if _v := v.value(m, _k); _v != nil {
					obj, key = _v, ""
				} else {
					obj, key = v.value(m, k), strings.TrimPrefix(_k, k+ice.PT)
				}
			}
		})
	}
	switch obj := obj.(type) {
	case Function:
		m.Debug("stack %d call %T %s %#v", len(s.frame)-1, obj, kit.Select("", obj.obj, -1), arg)
		f := s.pushf(m, CALL)
		for _, field := range obj.res {
			f.value[field.name] = nil
		}
		for i, field := range obj.arg {
			kit.If(i < len(arg), func() { f.value[field.name] = arg[i] }, func() { f.value[field.name] = nil })
		}
		kit.If(len(obj.obj) > 1, func() { f.value[obj.obj[0].name] = obj.object })
		value, pos := Value{list: kit.List()}, s.Position
		f.value["_return"] = func(arg ...Any) {
			if len(obj.res) > 0 {
				for i, field := range obj.res {
					kit.If(i < len(arg), func() { value.list = append(value.list, f.value[field.name]) }, func() { value.list = append(value.list, nil) })
				}
			} else {
				value.list = arg
			}
		}
		s.Position, f.defers = obj.Position, append(f.defers, func() {
			if len(obj.res) > 0 && len(value.list) == 0 {
				for _, field := range obj.res {
					value.list = append(value.list, f.value[field.name])
				}
			}
			s.Position = pos
		})
		kit.If(cb != nil, func() { cb(f, obj) })
		s.run(m.Options(STACK, s))
		return value
	case Caller:
		m.Debug("stack %d call %T %s %#v", len(s.frame)-1, obj, key, arg)
		kit.For(arg, func(i int, v Any) { arg[i] = trans(arg[i]) })
		return wrap(obj.Call(kit.Format(key), arg...))
	case func(*ice.Message, string, ...Any) Any:
		m.Debug("stack %d call %T %s %#v", len(s.frame)-1, obj, key, arg)
		kit.For(arg, func(i int, v Any) { arg[i] = trans(arg[i]) })
		return wrap(obj(m, kit.Format(key), arg...))
	case func():
		obj()
		return nil
	default:
		args := kit.List(key)
		kit.For(arg, func(i int, v Any) { args = append(args, trans(v)) })
		return Message{m.Cmd(args...)}
	}
}
func (s *Stack) action(m *ice.Message, obj Any, key Any, arg ...string) *ice.Message {
	s.calls(m, obj, key, func(f *Frame, v Function) {
		i := 0
		for _, field := range v.arg {
			switch field.name {
			case "m", "msg":
				f.value[field.name] = Message{m}
			case ice.ARG:
				list := kit.List()
				kit.For(arg, func(v string) { list = append(list, String{v}) })
				f.value[field.name] = Value{list}
			default:
				f.value[field.name], i = String{m.Option(field.name, kit.Select(m.Option(field.name), arg, i))}, i+1
			}
		}
	})
	return m
}
func (s *Stack) parse(m *ice.Message, name string, r io.Reader) *Stack {
	pos := s.Position
	defer func() { s.Position = pos }()
	s.Position = Position{Buffer: &Buffer{name: name, input: bufio.NewScanner(r)}}
	s.peekf().Position = s.Position
	m.Debug("stack %d parse %s", len(s.frame)-1, s.show())
	s.run(m)
	return s
}
func NewStack(m *ice.Message, cb func(*Frame)) *Stack {
	s := &Stack{}
	s.pushf(m.Options(STACK, s), STACK)
	s.load(m, cb)
	return s
}

func _parse_stack(m *ice.Message) *Stack { return m.Optionv(STACK).(*Stack) }
func _parse_frame(m *ice.Message) (*Stack, *Frame) {
	return _parse_stack(m), _parse_stack(m).pushf(m, "")
}
func _parse_const(m *ice.Message, key string) string {
	if k := kit.Select(key, strings.Split(key, ice.PT), -1); kit.IsUpper(k) {
		if c, ok := ice.Info.Index[strings.ToLower(k)].(*ice.Context); ok && (key == k || key == c.Prefix(k)) {
			return strings.ToLower(k)
		}
	}
	return ""
}
func _parse_res(m *ice.Message, v Any) []Any {
	switch v := v.(type) {
	case nil:
		return nil
	case Value:
		return v.list
	default:
		return []Any{v}
	}
}

const (
	STATUS_NORMAL  = 0
	STATUS_DISABLE = -1
)
const STACK = "stack"

func init() {
	Index.MergeCommands(ice.Commands{
		STACK: {Name: "stack path auto parse", Actions: ice.Actions{
			"start": {Hand: func(m *ice.Message, arg ...string) {}},
			ice.CMD: {Hand: func(m *ice.Message, arg ...string) {}},
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
				m.Options(nfs.DIR_ROOT, nfs.SRC).Cmdy(nfs.CAT, arg)
				return
			}
			nfs.Open(m, path.Join(nfs.SRC, strings.TrimPrefix(path.Join(arg...), nfs.SRC)), func(r io.Reader, p string) {
				if NewStack(m, nil).parse(m, p, r); m.Option(ice.DEBUG) == ice.TRUE {
					m.Cmdy(INFO, arg)
				}
			})
		}},
	})
	loaded := kit.Dict()
	ice.AddMergeAction(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) (init ice.Handler) {
		kit.IfNoKey(loaded, ice.SRC_SCRIPT+c.Prefix(key)+ice.PS, func(p string) { kit.If(nfs.Exists(ice.Pulse, p), func() { init = StackHandler }) })
		return
	})
}

func StackHandler(m *ice.Message, arg ...string) {
	s := NewStack(m, nil)
	script := []string{}
	nfs.Open(m, ice.SRC_SCRIPT+m.PrefixKey()+ice.PS, func(r io.Reader, p string) {
		kit.If(kit.Ext(p) == nfs.SHY, func() {
			if strings.HasPrefix(path.Base(p), "on") {
				script = append(script, kit.Format("Volcanos(\"%s\", {", kit.TrimExt(path.Base(p), nfs.SHY)))
				kit.For(r, func(s string) {
					if strings.HasPrefix(s, FUNC) {
						script = append(script, ice.TB+strings.Replace(strings.TrimPrefix(s, FUNC+ice.SP), "(", ": function(", 1))
					} else if strings.HasPrefix(s, END) {
						script = append(script, ice.TB+"},")
					} else {
						script = append(script, ice.TB+s)
					}
				})
				script = append(script, "})")
			} else {
				s.parse(m.Spawn(Index).Spawn(m.Target()), p, r)
			}
		})
	})
	if len(script) > 0 {
		p := ice.USR_SCRIPT + m.PrefixKey() + ice.PS + "list.js"
		s.value(m, "_script", "/require/"+p)
		m.Cmd(nfs.SAVE, p, kit.Dict(nfs.CONTENT, strings.Join(script, ice.NL)))
	}
	cmd := m.Commands("")
	kit.For(s.peekf().value, func(k string, v Any) {
		switch v := v.(type) {
		case Function:
			list := kit.List()
			for _, field := range v.arg {
				kit.If(!kit.IsIn(field.name, "m", "msg", ice.ARG), func() { list = append(list, kit.Dict(mdb.NAME, field.name, mdb.TYPE, mdb.TEXT, mdb.VALUE, "")) })
			}
			kit.If(k == mdb.LIST, func() { list = append(list, kit.Dict(mdb.NAME, mdb.LIST, mdb.TYPE, "button", mdb.ACTION, ice.AUTO)) })
			h := func(m *ice.Message, arg ...string) { m.Copy(s.action(m.Spawn(Index).Spawn(m.Target()), s, k, arg...)) }
			if k == mdb.LIST {
				cmd.Hand, cmd.List = h, list
			} else {
				cmd.Actions[k], cmd.Meta[k] = &ice.Action{Hand: h}, list
			}
		}
	})
}
func StackAction() ice.Actions { return ice.Actions{ice.CTX_INIT: {Hand: StackHandler}} }
