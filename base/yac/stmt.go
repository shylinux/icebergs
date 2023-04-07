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
	MAIN     = "main"
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
			if s.next(m) == IF {
				res := s.expr(m)
				kit.If(s.token() == SPLIT, func() { res = s.expr(m) })
				kit.If(res == ice.FALSE, func() { f.status = STATUS_DISABLE })
			}
		}},
		FOR: {Name: "for a = 1; a < 10; a++ {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			if strings.Contains(s.list[s.line], RANGE) {
				pos, key, list := s.Position, []string{}, []Any{}
				kit.If(s.last != nil && s.last.line == s.line, func() { list, _ = s.last.value["_range"].([]Any) })
				for {
					if k := s.cals0(m, FIELD, DEFS, ASSIGN); k == RANGE {
						if obj, ok := s.cals(m).(Operater); ok {
							if list, ok := obj.Operate(RANGE, list).([]Any); ok {
								kit.For(key, func(i int, k string) { f.value[k] = list[i] })
								f.value["_range"] = list
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
					s.pos(m, list[0], -1)
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
			if f.value["_case"] == "done" {
				return
			}
			if res, ok := v.(Operater); ok {
				if res, ok := res.Operate("==", Trans(s.value(m, "_switch"))).(Boolean); ok && res.value {
					f.status, f.value["_case"] = STATUS_NORMAL, "done"
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
			field, list := Field{}, []Field{}
			if s.next(m) == OPEN {
				for s.next(m) != CLOSE {
					kit.If(field.name == "", func() { field.name = s.token() }, func() { field.types = s.token() })
				}
				s.next(m)
				list = append(list, field)
			}
			name := s.token()
			list = append(list, Field{name: name})
			s.rest[s.skip] = FUNC
			v := s.funcs(m, name)
			if v.obj = list; field.types != nil {
				if t, ok := s.value(m, kit.Format(field.types)).(Struct); ok {
					m.Debug("value %s set %s.%s %s", Format(s), field.types, name, Format(v))
					t.index[name] = v
				}
			} else if name != INIT {
				s.value(m, name, v)
			}
		}},
		DEFER: {Name: "defer func() {} ()", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			obj, k := Any(nil), ""
			obj, k = s, s.next(m)
			kit.If(k == FUNC, func() { obj, k = s.funcs(m, ""), "" })
			s.skip++
			args := _parse_res(m, s.cals(m))
			if !s.runable() {
				return
			}
			s.stack(func(f *Frame, i int) bool {
				if f.key == CALL {
					f.defers = append(f.defers, func() { s.calls(m, obj, k, nil, args...) })
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
				show := func(p string) string {
					ls := nfs.SplitPath(m, p)
					return ice.Render(m, ice.RENDER_ANCHOR, p, m.MergePodCmd("", "web.code.vimer", nfs.PATH, ls[0], nfs.FILE, ls[1], nfs.LINE, ls[2]))
				}
				kit.For(f.value, func(k string, v Any) {
					switch v := v.(type) {
					case func(*ice.Message, string, ...Any) Any:
						m.EchoLine("  %s: %v", k, show(kit.FileLine(v, 100)))
					case Message:
						m.EchoLine("  %s: %v", k, show(kit.FileLine(v.Call, 100)))
					case Function:
						m.EchoLine("  %s: %v", k, show(v.Position.name+":"+kit.Format(v.Position.line+1)))
					case Struct:
						m.EchoLine("  %s: struct", k)
						kit.For(v.index, func(k string, v Any) {
							switch v := v.(type) {
							case Function:
								m.EchoLine("  	%s: %v", k, show(v.Position.name+":"+kit.Format(v.Position.line+1)))
							case Field:
								m.EchoLine("  	%s: %v", k, v.Format())
							}
						})
					case string:
						m.EchoLine("  %s: %v", k, v)
					default:
						m.EchoLine("  %s: %v", k, Format(v))
					}
				})
				return false
			})
			m.EchoLine("stmt: %s", arg[0])
			kit.For(kit.SortedKey(m.Target().Commands), func(key string) {
				if strings.HasPrefix(key, "_") || strings.HasPrefix(key, ice.PS) {
					return
				}
				cmd := m.Target().Commands[key]
				m.EchoLine("  %s: %#v", key, cmd.Name)
			})
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
