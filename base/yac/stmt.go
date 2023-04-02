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
	BREAK    = "break"
	CONTINUE = "continue"
	SWITCH   = "switch"
	CASE     = "case"
	DEFAULT  = "default"
	FUNC     = "func"
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
				f.status = 0
			} else {
				f.status, f.defers = STATUS_DISABLE, append(f.defers, func() { f.status = 0 })
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
				defer s.popf(m)
				switch f.key {
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
				defer s.popf(m)
				switch f.key {
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
			f.status = 0
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
			if f.status = 0; f.value["_case"] == "done" {
				f.status = STATUS_DISABLE
			}
			s.skip++
		}},
		FUNC: {Name: "func show(a, b) (c, d)", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			list, key, kind := [][]string{[]string{}}, "", ""
			push := func() { kit.If(key, func() { list[len(list)-1], key, kind = append(list[len(list)-1], key), "", "" }) }
			s.reads(m, func(k string) bool {
				switch k {
				case OPEN:
					defer kit.If(key != "" || len(list) > 1, func() { list = append(list, []string{}) })
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
			s.value(m, kit.Select("", list[0], -1), Function{obj: list[0], arg: list[1], res: list[2], Position: s.Position})
			s.pushf(m, "").status = STATUS_DISABLE
		}},
		DEFER: {Name: "defer func() {} ()", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			s.reads(m, func(k string) bool {
				kit.If(k == FUNC, func() { k = s.funcs(m) })
				s.skip++
				v := s.cals(m)
				if !s.runable() {
					return true
				}
				if vv, ok := v.(Value); ok {
					v = vv.list
				} else {
					v = []Any{v}
				}
				s.stack(func(f *Frame, i int) bool {
					if f.key == CALL {
						f.defers = append(f.defers, func() { s.call(m, s, k, nil, v.([]Any)...) })
						return true
					}
					return false
				})
				return true
			})
		}},
		RETURN: {Name: "return show", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			v := s.cals(m)
			s.stack(func(f *Frame, i int) bool {
				f.status = STATUS_DISABLE
				switch f.key {
				case FUNC:

				case CALL:
					switch cb := f.value["_return"].(type) {
					case func(...Any):
						cb(_parse_res(m, v)...)
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
				f := s.parse(m, p, r, func(f *Frame) { kit.For(u.Query(), func(k string, v []string) { f.value[k] = v[0] }) }).popf(m)
				s.Position = f.Position
				s.skip = len(s.rest)
			})
		}},
		INFO: {Name: "info", Hand: func(m *ice.Message, arg ...string) {
			_parse_stack(m).stack(func(f *Frame, i int) bool {
				m.EchoLine("frame: %s %v:%v:%v", f.key, f.name, f.line, f.skip)
				kit.For(f.value, func(k string, v Any) { m.EchoLine("  %s: %#v", k, v) })
				return false
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
