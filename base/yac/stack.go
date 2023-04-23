package yac

import (
	"bufio"
	"io"
	"path"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type Value struct {
	list []Any
}
type Function struct {
	obj Fields
	arg Fields
	res Fields
	Position
	object Object
}
type Frame struct {
	key    string
	name   string
	value  ice.Map
	defers []func()
	status int
	Position
}
type Stack struct {
	last    *Frame
	frame   []*Frame
	comment []string
	Error   []Error
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
func (s *Stack) pushf(m *ice.Message, arg ...string) *Frame {
	f := &Frame{key: kit.Select(m.CommandKey(), arg, 0), name: kit.Select("", arg, 1), value: kit.Dict(), Position: s.Position}
	kit.If(len(s.frame) > 0, func() { f.status = s.peekf().status })
	s.frame = append(s.frame, f)
	m.Debug("stack %s push %s", Format(s), kit.Select(s.show(), arg, 2))
	return f
}
func (s *Stack) popf(m *ice.Message) *Frame {
	f := s.peekf()
	for i := len(f.defers) - 1; i >= 0; i-- {
		f.defers[i]()
	}
	m.Debug("stack %s pop %s", Format(s), s.show())
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
	keys := strings.Split(key, nfs.PT)
	f, n := s.peekf(), len(s.frame)-1
	if len(arg) < 2 || arg[1] != DEFINE {
		s.stack(func(_f *Frame, i int) bool {
			if _f.value[key] != nil || _f.value[keys[0]] != nil {
				f, n = _f, i
				return true
			}
			return false
		})
	}
	kit.If(len(arg) > 0, func() {
		if f.value[key] != nil {
			f.value[key] = arg[0]
			return
		}
		var v Any = Dict{f.value}
		for i := 0; i < len(keys); i++ {
			switch k := keys[i]; _v := v.(type) {
			case Operater:
				if i == len(keys)-1 {
					m.Debug("value %d:%s set %s %s", n, f.key, key, Format(arg[0]))
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
	if v, ok := f.value[key]; ok {
		return v
	}
	var v Any = Dict{f.value}
	kit.For(keys, func(k string) {
		switch _v := v.(type) {
		case Operater:
			v = _v.Operate(SUBS, k)
		default:
			v = nil
		}
	})
	if v != nil {
		return v
	} else if v = _parse_const(m, key); v != "" {
		return v
	}
	return nil
}
func (s *Stack) status_disable(f *Frame) { f.status = STATUS_DISABLE }
func (s *Stack) status_normal(f *Frame) {
	kit.If(s.frame[len(s.frame)-2].status == STATUS_NORMAL, func() { f.status = STATUS_NORMAL })
}
func (s *Stack) runable() bool { return s.peekf().status > STATUS_DISABLE }
func (s *Stack) read(m *ice.Message) (text string, ok bool) {
	isvoid := func(text string) bool { return strings.TrimSpace(text) == "" }
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
		m.Debug("input %d read %q %s", len(s.list), text, s.show())
		return text, true
	}
	return
}
func (s *Stack) reads(m *ice.Message, cb func(k string) bool) {
	block, last := []string{}, 0
	comment := false
	for {
		if s.skip++; s.skip < len(s.rest) {
			if k, v := s.rest[s.skip], kit.Select("", s.rest, s.skip+1); k == "`" {
				if len(block) > 0 {
					kit.If(s.line != last, func() { block, last = append(block, lex.NL), s.line })
					block = append(block, k)
					cb(strings.Join(block, lex.SP))
					block = block[:0]
				} else {
					block = append(block, k)
				}
			} else if len(block) > 0 {
				kit.If(s.line != last, func() { block, last = append(block, lex.NL), s.line })
				block = append(block, k)
			} else if k == "*" && v == nfs.PS {
				comment = false
				s.skip++
			} else if comment {

			} else if k == nfs.PS && v == "*" {
				comment = true
				s.skip++
			} else if k == nfs.PS && v == nfs.PS {
				s.comment = append(s.comment, s.list[s.line])
				s.skip = len(s.rest)
			} else if s.skip == 0 && strings.HasPrefix(k, "#") {
				s.comment = append(s.comment, s.list[s.line])
				s.skip = len(s.rest)
			} else if cb(s.rest[s.skip]) {
				break
			} else {
				s.comment = s.comment[:0]
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
func (s *Stack) token() string { return kit.Select("", s.rest, s.skip) }
func (s *Stack) show() string  { return Format(s.Position) }
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
				m.Cmd(EXPR, s.rest)
			} else {
				m.Cmd(EXPR, kit.Slice(s.rest, s.skip))
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
		case MAP:
			s.skip += 2
			key := s.types(m)
			s.skip += 2
			return Map{key: key, value: s.types(m)}
		case SUBS:
			s.skip += 2
			return Slice{value: s.types(m)}
		case STRUCT:
			key, t := []string{}, Struct{index: map[string]Any{}, stack: s}
			for s.next(m); s.next(m) != END; {
				line := s.line
				kit.If(s.token() == "*", func() { s.next(m) })
				if key = append(key, s.token()); s.next(m) == FIELD {
					continue
				}
				if s.line != line {
					kit.For(key, func(key string) {
						field := Field{types: key, name: kit.Select("", kit.Split(key, nfs.PT), -1)}
						m.Debug("value %s field %s %#v", Format(s), key, field)
						t.index[field.name] = key
						t.sups = append(t.sups, key)
					})
					key, s.skip = key[:0], s.skip-1
					continue
				}
				types, tags := s.types(m), map[string]string{}
				kit.If(strings.HasPrefix(s.peek(m), "`"), func() {
					kit.For(kit.Split(strings.TrimPrefix(strings.TrimSuffix(s.next(m), "`"), "`"), ": "), func(k, v string) { tags[k] = v })
				})
				kit.For(key, func(key string) {
					field := Field{types: types, name: key, tags: tags}
					kit.If(field.types == nil, func() {
						t.sups = append(t.sups, field.name)
						field.types, field.name = field.name, kit.Select("", kit.Split(field.name, nfs.PT), -1)
					})
					m.Debug("value %s field %s %#v", Format(s), key, field)
					t.index[field.name] = field
				})
				key, s.skip = key[:0], len(s.rest)
			}
			return t
		case INTERFACE:
			t := Interface{index: map[string]Function{}}
			for s.next(m); s.next(m) != END; {
				name := s.token()
				s.rest[s.skip] = FUNC
				t.index[name] = s.types(m).(Function)
			}
			return t
		case FUNC:
			field, list := Field{}, [][]Field{}
			for s.skip++; s.skip < len(s.rest) && s.token() != BEGIN; s.skip++ {
				switch s.token() {
				case OPEN:
					list = append(list, []Field{})
				case FIELD, CLOSE:
					list[len(list)-1] = append(list[len(list)-1], field)
					field = Field{}
				case "*":
				default:
					switch t := s.types(m).(type) {
					case string:
						kit.If(field.name == "", func() { field.name = t }, func() { field.types = t })
					default:
						field.types = s.types(m)
					}
				}
			}
			rename := func(list []Field) []Field {
				for i := len(list) - 1; i > 0; i-- {
					field = list[i]
					if field.types != nil {
						continue
					}
					if i+1 < len(list) {
						field.types = list[i].types
					} else {
						field.types, field.name = field.name, ""
					}
					list[i] = field
				}
				return list
			}
			kit.If(len(list) == 1, func() { list = append(list, []Field{}) })
			return Function{arg: rename(list[0]), res: rename(list[1])}
		case "*":
			continue
		default:
			if strings.HasPrefix(s.token(), "`") {
				s.skip--
				return nil
			}
			return s.token()
		}
	}
	return nil
}
func (s *Stack) funcs(m *ice.Message, name string) Function {
	v := s.types(m).(Function)
	if f := s.pushf(m, FUNC, name); name == INIT {
		f.key = CALL
	} else {
		s.status_disable(f)
	}
	v.Position = s.Position
	s.run(m)
	return v
}
func (s *Stack) pusherr(err Error) {
	err.Position = s.Position
	err.Position.skip = -1
	s.Error = append(s.Error, err)
}
func (s *Stack) calls(m *ice.Message, obj Any, key string, cb func(*Frame, Function), arg ...Any) Any {
	m.Debug("calls %s %T %s(%s)", Format(s), obj, key, Format(arg...))
	m.Debug("calls %s %T %s(%#v)", Format(s), obj, key, arg)
	_obj, _key := obj, key
	switch v := obj.(type) {
	case *Stack:
		if _v := v.value(m, key); _v != nil {
			obj, key = _v, ""
		}
	}
	kit.For(kit.Split(key, nfs.PT), func(k string) {
		switch v := obj.(type) {
		case Operater:
			obj = v.Operate(SUBS, k)
		case *Stack:
			obj = v.value(m, k)
		default:
			return
		}
		key = strings.TrimPrefix(strings.TrimPrefix(key, k), nfs.PT)
	})
	m.Debug("calls %s %T %s(%s)", Format(s), obj, key, Format(arg...))
	if obj == nil {
		if _obj == s {
			s.pusherr(ErrNotFound(_key))
		} else {
			s.pusherr(ErrNotFound(_obj, _key))
		}
		return nil
	}
	switch obj := obj.(type) {
	case Function:
		name := kit.Format("%s%s", kit.Select("", kit.Format("%s.", obj.obj[0].types), len(obj.obj) > 1), obj.obj[len(obj.obj)-1].name)
		m.Debug("calls %s %s(%s) %s", Format(s), name, Format(arg...), Format(obj.Position))
		f := s.pushf(m, CALL, name, Format(obj.Position))
		obj.res.For(func(field Field) { f.value[field.name] = nil })
		obj.arg.For(func(field Field) { f.value[field.name] = nil })
		for i, field := range obj.arg {
			kit.If(i < len(arg), func() {
				if strings.HasPrefix(kit.Format(field.types), EXPAND) {
					f.value[field.name] = List{arg[i:]}
				} else {
					f.value[field.name] = arg[i]
				}
			})
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
			kit.If(len(obj.res) > 0 && len(value.list) == 0, func() { obj.res.For(func(field Field) { value.list = append(value.list, f.value[field.name]) }) })
			s.Position = pos
		})
		kit.If(cb != nil, func() { cb(f, obj) })
		s.run(m)
		return value
	case Caller:
		if msg, ok := m.Optionv(ice.YAC_MESSAGE).(*ice.Message); ok {
			m = msg
		}
		kit.For(arg, func(i int, v Any) { arg[i] = Trans(arg[i]) })
		m.Debug("calls %s %s.%s(%s)", Format(s), Format(obj), key, Format(arg...))
		return Wraps(obj.Call(kit.Format(key), arg...))
	case func(*ice.Message, string, ...Any) Any:
		if msg, ok := m.Optionv(ice.YAC_MESSAGE).(*ice.Message); ok {
			m = msg
		}
		m.Debug("calls %s %s %s %#v", Format(s), Format(obj), key, arg)
		kit.For(arg, func(i int, v Any) { arg[i] = Trans(arg[i]) })
		m.Debug("calls %s %s %s %s", Format(s), Format(obj), key, Format(arg...))
		m.Debug("calls %s %s %s %#v", Format(s), Format(obj), key, arg)
		return Wraps(obj(m, kit.Format(key), arg...))
	case func():
		obj()
		return nil
	default:
		if key == "" {
			s.pusherr(ErrNotSupport(obj))
			return nil
		}
		args := kit.List(key)
		kit.For(arg, func(i int, v Any) { args = append(args, Trans(v)) })
		m.Debug("calls %s %s", Format(s), Format(args...))
		return Message{m.Cmd(args...)}
	}
}
func (s *Stack) Action(m *ice.Message, obj Any, key string, arg ...string) *ice.Message {
	s.calls(m, obj, key, func(f *Frame, v Function) {
		n := 0
		for i, field := range v.arg {
			switch field.name {
			case "m", "msg":
				f.value[field.name] = Message{m}
			case ice.ARG:
				list := kit.List()
				kit.For(arg, func(v string) { list = append(list, String{v}) })
				f.value[field.name] = List{list}
			default:
				if strings.HasPrefix(kit.Format(field.types), EXPAND) {
					list := kit.List()
					kit.For(arg[i:], func(v string) { list = append(list, String{v}) })
					f.value[field.name] = List{list}
				} else {
					f.value[field.name], n = String{m.Option(field.name, kit.Select(m.Option(field.name), arg, n))}, n+1
				}
			}
		}
	})
	return m
}
func (s *Stack) Handler(obj Any) ice.Handler {
	return func(m *ice.Message, arg ...string) {
		m.Copy(s.Action(m.Options(ice.YAC_STACK, s, ice.YAC_MESSAGE, m).Spawn(Index).Spawn(m.Target()), obj, "", arg...))
		if m.Option(ice.DEBUG) == ice.TRUE && len(s.Error) > 0 {
			m.EchoLine("")
			for _, e := range s.Error {
				m.EchoLine("%s%s %s %s", e.key, e.detail, Format(e.Position), e.fileline)
			}
		}
	}
}
func (s *Stack) parse(m *ice.Message, name string, r io.Reader) *Stack {
	pos := s.Position
	defer func() { s.Position, s.peekf().Position = pos, pos }()
	s.Position = Position{Buffer: &Buffer{name: name, input: bufio.NewScanner(r)}}
	s.peekf().Position = s.Position
	m.Debug("stack %s parse %s", Format(s), s.show())
	s.run(m)
	return s
}
func (s *Stack) load(m *ice.Message, cb func(*Frame)) *Stack {
	f := s.peekf()
	for k, v := range ice.Info.Stack {
		kit.If(strings.HasPrefix(k, "web.code."), func() { k = strings.TrimPrefix(k, "web.") })
		f.value[k] = v
	}
	kit.If(cb != nil, func() { cb(f) })
	f.value["m"] = Message{m}
	return s
}
func NewStack(m *ice.Message, cb func(*Frame), arg ...string) *Stack {
	s := &Stack{}
	s.pushf(m.Options(ice.YAC_STACK, s, ice.YAC_MESSAGE, m), kit.Simple(STACK, arg)...)
	return s.load(m, cb)
}

func _parse_stack(m *ice.Message) *Stack { return m.Optionv(ice.YAC_STACK).(*Stack) }
func _parse_frame(m *ice.Message) (*Stack, *Frame) {
	return _parse_stack(m), _parse_stack(m).pushf(m, "")
}
func _parse_link(m *ice.Message, p string) string {
	ls := nfs.SplitPath(m, p)
	return ice.Render(m, ice.RENDER_ANCHOR, p, m.MergePodCmd("", "web.code.vimer", nfs.PATH, ls[0], nfs.FILE, ls[1], nfs.LINE, ls[2]))
}
func _parse_const(m *ice.Message, key string) string {
	if k := kit.Select(key, strings.Split(key, nfs.PT), -1); kit.IsUpper(k) {
		return strings.ToLower(k)
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
		STACK: {Name: "stack path auto", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 0, func() { arg = append(arg, nfs.SRC) })
			if strings.HasSuffix(arg[0], nfs.PS) {
				m.Cmdy(nfs.CAT, arg)
				return
			}
			nfs.Open(m, path.Join(arg...), func(r io.Reader, p string) {
				s := NewStack(m, nil, p, p).parse(m, p, r)
				m.Options("__index", kit.Format(s.value(m, "_index"))).Cmdy(INFO, arg)
				m.StatusTime(mdb.LINK, s.value(m, "_link"))
				ctx.AddFileCmd(kit.Path(p), m.Option("__index"))
			})
		}},
	})
	loaded := kit.Dict()
	ice.AddMergeAction(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) (init ice.Handler) {
		kit.IfNoKey(loaded, ice.SRC_SCRIPT+c.Prefix(key)+nfs.PS, func(p string) { kit.If(nfs.Exists(ice.Pulse, p), func() { init = StackHandler }) })
		return
	})
}

func StackHandler(m *ice.Message, arg ...string) {
	script := []string{}
	m = m.Spawn(Index).Spawn(m.Target())
	s := NewStack(m, nil, m.PrefixKey())
	nfs.Open(m, ice.SRC_SCRIPT+m.PrefixKey()+nfs.PS, func(r io.Reader, p string) {
		kit.If(kit.Ext(p) == nfs.SHY, func() {
			if strings.HasPrefix(path.Base(p), "on") {
				script = append(script, kit.Format("Volcanos(\"%s\", {", kit.TrimExt(path.Base(p), nfs.SHY)))
				kit.For(r, func(s string) {
					if strings.HasPrefix(s, FUNC) {
						script = append(script, lex.TB+strings.Replace(strings.TrimPrefix(s, FUNC+lex.SP), "(", ": function(", 1))
					} else if strings.HasPrefix(s, END) {
						script = append(script, lex.TB+"},")
					} else {
						script = append(script, lex.TB+s)
					}
				})
				script = append(script, "})")
			} else {
				s.parse(m, p, r)
			}
		})
	})
	if len(script) > 0 {
		p := ice.USR_SCRIPT + m.PrefixKey() + nfs.PS + "list.js"
		m.Cmd(nfs.SAVE, p, kit.Dict(nfs.CONTENT, strings.Join(script, lex.NL)))
		s.value(m, "_script", "/require/"+p)
	}
	cmd := m.Commands("")
	kit.For(s.peekf().value, func(k string, v Any) {
		switch k = kit.LowerCapital(k); v := v.(type) {
		case Function:
			list := kit.List()
			v.arg.For(func(field Field) {
				kit.If(!kit.IsIn(field.name, "m", "msg", ice.ARG), func() { list = append(list, kit.Dict(mdb.NAME, field.name, mdb.TYPE, mdb.TEXT, mdb.VALUE, "")) })
			})
			if k == mdb.LIST {
				cmd.Hand, cmd.List = s.Handler(v), append(list, kit.Dict(mdb.NAME, mdb.LIST, mdb.TYPE, "button", mdb.ACTION, ice.AUTO))
			} else {
				cmd.Actions[k], cmd.Meta[k] = &ice.Action{Hand: s.Handler(v)}, list
			}
		}
	})
}
func StackAction() ice.Actions { return ice.Actions{ice.CTX_INIT: {Hand: StackHandler}} }
