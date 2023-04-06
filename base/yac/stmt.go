package yac

import (
	"io"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	IF       = "if"
	ELSE     = "else"
	FOR      = "for"
	RANGE    = "range"
	BREAK    = "break"
	CONTINUE = "continue"
	SWITCH   = "switch"
	CASE     = "case"
	DEFAULT  = "default"
	FUNC     = "func"
	INIT     = "init"
	CALL     = "call"
	DEFER    = "defer"
	RETURN   = "return"
	SOURCE   = "source"
	INFO     = "info"
	PWD      = "pwd"
)
const (
	KEYWORD  = "keyword"
	FUNCTION = "function"
)

func init() {
	Index.MergeCommands(ice.Commands{
		IF: {Name: "if a = 1; a > 1 {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			res := s.expr(m)
			kit.If(s.token() == SPLIT, func() { res = s.expr(m) })
			kit.If(res == ice.FALSE, func() { f.status = STATUS_DISABLE })
		}},
		ELSE: {Name: "else if a = 1; a > 1 {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			if s.last.status == STATUS_DISABLE {
				f.status = STATUS_NORMAL
			} else {
				f.status, f.defers = STATUS_DISABLE, append(f.defers, func() { f.status = STATUS_NORMAL })
			}
			s.reads(m, func(k string) bool {
				if k == IF {
					res := s.expr(m)
					kit.If(s.token() == SPLIT, func() { res = s.expr(m) })
					kit.If(res == ice.FALSE, func() { f.status = STATUS_DISABLE })
				}
				return true
			})
		}},
		FOR: {Name: "for a = 1; a < 10; a++ {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			if strings.Contains(s.list[s.line], RANGE) {
				pos, key, list := s.Position, []string{}, []Any{}
				kit.If(s.last != nil && s.last.line == s.line, func() { list, _ = s.last.value["_range"].([]Any) })
				for { // for k, v := range value {
					if k := s.cals0(m, FIELD, DEFS, ASSIGN); k == RANGE {
						if obj, ok := s.cals(m).(Operater); ok {
							if _list, ok := obj.Operate(RANGE, list).([]Any); ok {
								kit.For(key, func(i int, k string) { f.value[k] = _list[i] })
								f.value["_range"] = _list
								break
							}
						}
						f.status = STATUS_DISABLE
						break
					} else if k != "" {
						key = append(key, k)
					}
				}
				f.defers = append(f.defers, func() { kit.If(s.runable(), func() { s.pos(m, pos, -1) }) })
				return
			}
			list, status := []Position{s.Position}, f.status
			for f.status = STATUS_DISABLE; s.token() != BEGIN; list = append(list, s.Position) {
				s.expr(m)
			}
			f.status = status
			res := ice.TRUE
			if len(list) == 1 {

			} else if len(list) == 2 {
				res = s.expr(m, list[0])
			} else {
				if s.last == nil || s.last.line != s.line {
					res = s.expr(m, list[0])
				} else {
					kit.For(s.last.value, func(k string, v Any) { f.value[k] = v })
				}
				res = s.expr(m, list[1])
			}
			kit.If(res == ice.FALSE, func() { f.status = STATUS_DISABLE })
			s.Position, f.defers = list[len(list)-1], append(f.defers, func() {
				if s.runable() {
					kit.If(len(list) > 3, func() { s.expr(m, list[2]) })
					s.Position = list[0]
					s.Position.skip--
				}
			})
		}},
		BREAK: {Name: "break", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			if !s.runable() {
				return
			}
			s.stack(func(f *Frame, i int) bool {
				f.status = STATUS_DISABLE
				switch s.popf(m); f.key {
				case FOR, SWITCH:
					return true
				default:
					return false
				}
			})
		}},
		CONTINUE: {Name: "continue", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			if !s.runable() {
				return
			}
			s.stack(func(f *Frame, i int) bool {
				switch s.popf(m); f.key {
				case FOR:
					return true
				default:
					return false
				}
			})
		}},
		SWITCH: {Name: "switch a = 1; a {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			res := s.cals(m)
			kit.If(s.token() == SPLIT, func() { res = s.cals(m) })
			f.value["_switch"], f.value["_case"] = res, ""
		}},
		CASE: {Name: "case b:", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			f := s.peekf()
			f.status = STATUS_NORMAL
			v := s.cals(m)
			f.status = STATUS_DISABLE
			if res, ok := v.(Operater); ok {
				if res, ok := res.Operate("==", trans(s.value(m, "_switch"))).(Boolean); ok && res.value {
					f.status, f.value["_case"] = 0, "done"
				}
			}
		}},
		DEFAULT: {Name: "default:", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			f := s.peekf()
			if f.status = STATUS_NORMAL; f.value["_case"] == "done" {
				f.status = STATUS_DISABLE
			}
			s.skip++
		}},
		FUNC: {Name: "func show(a, b) (c, d)", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			list, key, kind := [][]Field{[]Field{}}, "", ""
			push := func() {
				kit.If(key, func() { list[len(list)-1], key, kind = append(list[len(list)-1], Field{name: key, kind: kind}), "", "" })
			}
			s.reads(m, func(k string) bool {
				switch k {
				case OPEN:
					defer kit.If(key != "" || len(list) > 1, func() { list = append(list, []Field{}) })
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
			kit.If(len(list) < 2, func() { list = append(list, []Field{}) })
			kit.If(len(list) < 3, func() { list = append(list, []Field{}) })
			name, fun := list[0][len(list[0])-1].name, Function{obj: list[0], arg: list[1], res: list[2], Position: s.Position}
			if len(list[0]) > 1 {
				st := list[0][0].kind.(Struct)
				st.method = append(st.method, fun)
				st.index[name] = fun
			} else {
				kit.If(name != INIT, func() { s.value(m, name, fun) })
			}
			if f := s.pushf(m, ""); name == INIT {
				f.key = CALL
			} else {
				f.status = STATUS_DISABLE
			}
		}},
		DEFER: {Name: "defer func() {} ()", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			k := s.next(m)
			kit.If(k == FUNC, func() { k = s.funcs(m) })
			s.skip++
			args := _parse_res(m, s.cals(m))
			if !s.runable() {
				return
			}
			s.stack(func(f *Frame, i int) bool {
				if f.key == CALL {
					f.defers = append(f.defers, func() { s.calls(m, s, k, nil, args...) })
					return true
				}
				return false
			})
		}},
		RETURN: {Name: "return show", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			args := _parse_res(m, s.cals(m))
			s.stack(func(f *Frame, i int) bool {
				f.status = STATUS_DISABLE
				switch f.key {
				case FUNC:
				case CALL:
					switch cb := f.value["_return"].(type) {
					case func(...Any):
						cb(args...)
					}
				case SOURCE:
					s.input = nil
				case STACK:
					s.input = nil
				default:
					return false
				}
				return true
			})
		}},
		SOURCE: {Name: "source", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			u := kit.ParseURL(s.expr(m))
			nfs.Open(m, u.Path, func(r io.Reader, p string) {
				s.parse(m, p, r)
				s.skip = len(s.rest)
			})
		}},
		INFO: {Name: "info", Hand: func(m *ice.Message, arg ...string) {
			m.EchoLine("").EchoLine("stack: %s", arg[0])
			_parse_stack(m).stack(func(f *Frame, i int) bool {
				m.EchoLine("frame: %s %v:%v:%v", f.key, f.name, f.line, f.skip)
				kit.For(f.value, func(k string, v Any) { m.EchoLine("  %s: %#v", k, v) })
				return false
			})
			m.EchoLine("stmt: %s", arg[0])
			for key, cmd := range m.Target().Commands {
				if strings.HasPrefix(key, "_") || strings.HasPrefix(key, "/") {
					continue
				}
				m.EchoLine("  %s: %#v", key, cmd.Name)
			}
		}},
		PWD: {Name: "pwd", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			res := []string{kit.Format("%d:%d", s.line, s.skip)}
			s.stack(func(f *Frame, i int) bool {
				kit.If(i > 0, func() {
					res = append(res, kit.Format("%s %s %s:%d:%d", f.key, kit.Select(ice.FALSE, ice.TRUE, f.status > STATUS_DISABLE), f.name, f.line, f.skip))
				})
				return false
			})
			m.Echo(strings.Join(res, " / ")).Echo(ice.NL)
		}},
	})
}
