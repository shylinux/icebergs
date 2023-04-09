package yac

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
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
	DEFER    = "defer"
	RETURN   = "return"

	CALL = "call"
	INIT = "init"
	MAIN = "main"
	INFO = "info"
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
			kit.If(res == ice.FALSE, func() { s.status_disable(f) })
		}},
		ELSE: {Name: "else if a = 1; a > 1 {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			if s.last.status == STATUS_DISABLE {
				s.status_normal(f)
			} else {
				f.status, f.defers = STATUS_DISABLE, append(f.defers, func() { s.status_normal(f) })
			}
			if s.next(m) == IF {
				res := s.expr(m)
				kit.If(s.token() == SPLIT, func() { res = s.expr(m) })
				kit.If(res == ice.FALSE, func() { s.status_disable(f) })
			}
		}},
		FOR: {Name: "for a = 1; a < 10; a++ {", Hand: func(m *ice.Message, arg ...string) {
			s, f := _parse_frame(m)
			if strings.Contains(s.list[s.line], RANGE) {
				pos, key, list := s.Position, []string{}, []Any{}
				kit.If(s.last != nil && s.last.line == s.line, func() { list, _ = s.last.value["_range"].([]Any) })
				for s.next(m) != BEGIN {
					switch s.token() {
					case RANGE:
						if obj, ok := s.cals(m).(Operater); ok {
							if list, ok := obj.Operate(RANGE, list).([]Any); ok {
								kit.For(key, func(i int, k string) { f.value[k] = list[i] })
								f.defers = append(f.defers, func() { kit.If(s.runable(), func() { s.pos(m, pos, -1) }) })
								f.value["_range"] = list
								return
							}
						}
						s.status_disable(f)
						return
					case ASSIGN:
					case DEFS:
					case FIELD:
					default:
						key = append(key, s.token())
					}
				}
				return
			}
			list, status := []Position{s.Position}, f.status
			for s.status_disable(f); s.token() != BEGIN; list = append(list, s.Position) {
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
			kit.If(res == ice.FALSE, func() { s.status_disable(f) })
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
				switch s.status_disable(f); f.key {
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
				switch s.status_disable(f); f.key {
				case FOR:
					f.defers = append(f.defers, func() { s.status_normal(f) })
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
			s.status_normal(f)
			v := s.cals(m)
			if s.status_disable(f); f.value["_case"] == "done" {
				return
			}
			if res, ok := v.(Operater); ok {
				if res, ok := res.Operate("==", Trans(s.value(m, "_switch"))).(Boolean); ok && res.value {
					f.value["_case"] = "done"
					s.status_normal(f)
				}
			}
		}},
		DEFAULT: {Name: "default:", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			f := s.peekf()
			if s.status_normal(f); f.value["_case"] == "done" {
				s.status_disable(f)
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
				kit.If(field.types == nil, func() { field.types = field.name })
				list = append(list, field)
				s.next(m)
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
			} else if !kit.IsIn(name, INIT, MAIN) {
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
				switch s.status_disable(f); f.key {
				case FUNC:
				case CALL:
					switch cb := f.value["_return"].(type) {
					case func(...Any):
						cb(args...)
					}
				case STACK:
					s.input = nil
				default:
					return false
				}
				return true
			})
		}},
		INFO: {Name: "info", Actions: ice.Actions{
			ERROR: {Hand: func(m *ice.Message, arg ...string) {
				for _, e := range _parse_stack(m).Error {
					m.EchoLine("  %s%s %s %s", e.key, e.detail, _parse_link(m, Format(e.Position)), _parse_link(m, e.fileline))
				}
			}},
			STACK: {Hand: func(m *ice.Message, arg ...string) {
				_parse_stack(m).stack(func(f *Frame, i int) bool {
					m.EchoLine("frame: %s %v:%v:%v", f.key, f.name, f.line, f.skip)
					show := func(p string) string { return _parse_link(m, p) }
					kit.For(f.value, func(k string, v Any) {
						switch v := v.(type) {
						case func(*ice.Message, string, ...Any) Any:
							m.EchoLine("  %s: %v", k, show(kit.FileLine(v, 100)))
						case Message:
							m.EchoLine("  %s: %v", k, show(kit.FileLine(v.Call, 100)))
						case Function:
							m.EchoLine("  %s: %v", k, show(v.Position.name+ice.DF+kit.Format(v.Position.line+1)))
						case Struct:
							m.EchoLine("  %s: %s", k, show(v.Position.name+ice.DF+kit.Format(v.Position.line+1)))
							break
							kit.For(v.index, func(k string, v Any) {
								switch v := v.(type) {
								case Function:
									m.EchoLine("  	%s: %v", k, show(v.Position.name+ice.DF+kit.Format(v.Position.line+1)))
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
			}},
			ctx.CONFIG: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(ctx.CONFIG, m.Option("__index")).EchoLine("")
			}},
			ctx.COMMAND: {Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(ctx.COMMAND, m.Option("__index"))
				m.EchoLine(msg.Append(mdb.LIST))
				m.EchoLine(msg.Append(mdb.META))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			kit.For([]string{ERROR, STACK, ctx.CONFIG, ctx.COMMAND}, func(k string) { m.EchoLine("%s: %s", k, arg[0]).Cmdy("", k) })
		}},
	})
}
