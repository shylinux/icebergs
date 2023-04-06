package yac

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const (
	CONST = "const"
	TYPE  = "type"
	VAR   = "var"

	STRING    = "string"
	INT       = "int"
	MAP       = "map"
	SLICE     = "slice"
	STRUCT    = "struct"
	INTERFACE = "interface"
)

type Map struct {
	key   Any
	value Any
}
type Slice struct {
	value Any
}
type Interface struct {
	index map[string]Function
}
type Struct struct {
	index  map[string]Any
	method []Function
	field  []Field
}
type Field struct {
	name string
	kind Any
}
type Object struct {
	value Operater
	index Struct
}

func (s Object) Operate(op string, v Any) Any {
	switch op {
	case "&", "*":
		return s
	case SUBS:
		switch v := s.index.index[kit.Format(v)].(type) {
		case Function:
			v.object = s
			return v
		}
		fallthrough
	default:
		return s.value.Operate(op, v)
	}
	return nil
}

func init() {
	Index.MergeCommands(ice.Commands{
		CONST: {Name: "const a = 1", Hand: func(m *ice.Message, arg ...string) {
			if s := _parse_stack(m); s.next(m) == OPEN {
				for s.token() != CLOSE {
					s.nextLine(m)
					s.skip--
					s.cals(m, CLOSE)
				}
			} else {
				s.skip--
				s.cals(m)
			}
		}},
		TYPE: {Name: "type student struct {", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			switch name := s.next(m); s.next(m) {
			case ASSIGN:
				s.next(m)
				fallthrough
			default:
				s.value(m, name, s.types(m))
			}
		}},
		VAR: {Name: "var a = 1", Hand: func(m *ice.Message, arg ...string) {
			if s := _parse_stack(m); s.next(m) == OPEN {
				for s.token() != CLOSE {
					s.nextLine(m)
					s.skip--
					s.cals(m, CLOSE)
				}
			} else {
				s.skip--
				s.cals(m)
			}
		}},
	})
}
