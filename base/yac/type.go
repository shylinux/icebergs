package yac

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

type Map struct {
	value Any
	key   Any
}
type Slice struct {
	value Any
}
type Interface struct {
	index map[string]Function
	name  string
}
type Struct struct {
	index map[string]Any
	name  string
}
type Field struct {
	types Any
	tags  map[string]string
	name  string
}

func (s Field) MarshalJSON() ([]byte, error) {
	return []byte(kit.Format("%q", s.Format())), nil
}
func (s Field) Format() string {
	if len(s.tags) == 0 {
		return kit.Format("%s", s.types)
	}
	res := []string{}
	kit.For(s.tags, func(k, v string) { res = append(res, kit.Format("%s:\"%s\"", k, v)) })
	return kit.Format("%s `%s`", s.types, strings.Join(res, ice.SP))
}

type Object struct {
	value Operater
	index Struct
}

func (s Object) Operate(op string, v Any) Any {
	switch op {
	case "&", "*":
		return s
	case INSTANCEOF:
		if t, ok := v.(Struct); ok {
			return Value{list: []Any{s, s.index.name == t.name}}
		}
		return Value{list: []Any{s, false}}
	case IMPLEMENTS:
		if t, ok := v.(Interface); ok {
			for k, v := range t.index {
				if _v, ok := s.index.index[k].(Function); ok {
					for i, field := range v.arg {
						if i < len(_v.arg) && _v.arg[i].types == field.types {
							continue
						}
						return Value{list: []Any{s, false}}
					}
					for i, field := range v.res {
						if i < len(_v.res) && _v.res[i].types == field.types {
							continue
						}
						return Value{list: []Any{s, false}}
					}
				} else {
					return Value{list: []Any{s, false}}
				}
			}
		}
		return Value{list: []Any{s, true}}
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

const (
	MAP       = "map"
	SLICE     = "slice"
	STRUCT    = "struct"
	INTERFACE = "interface"
	STRING    = "string"
	INT       = "int"

	INSTANCEOF = "instanceof"
	IMPLEMENTS = "implements"

	CONST = "const"
	TYPE  = "type"
	VAR   = "var"
)

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
				switch t := s.types(m).(type) {
				case Interface:
					t.name = name
					s.value(m, name, t)
				case Struct:
					t.name = name
					s.value(m, name, t)
				default:
					s.value(m, name, t)
				}
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
