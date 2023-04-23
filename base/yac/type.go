package yac

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
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
	sups  []string
	name  string
	stack *Stack
}
type Struct struct {
	index map[string]Any
	sups  []string
	name  string
	stack *Stack
	Position
}
type Fields []Field
type Field struct {
	types Any
	name  string
	tags  map[string]string
}
type Object struct {
	value Operater
	index Struct
}

func (s Fields) For(cb func(Field)) {
	for _, v := range s {
		cb(v)
	}
}
func (s Field) MarshalJSON() ([]byte, error) {
	return []byte(kit.Format("%q", s.Format())), nil
}
func (s Field) Format() string {
	if types := ""; len(s.tags) == 0 {
		switch t := s.types.(type) {
		case string:
			types = t
		default:
			types = Format(s.types)
		}
		return types
	} else {
		res := []string{}
		kit.For(s.tags, func(k, v string) { res = append(res, kit.Format("%s:\"%s\"", k, v)) })
		return kit.Format("%s `%s`", types, strings.Join(res, lex.SP))
	}
}
func (s Function) Operate(op string, v Any) Any {
	switch op {
	case "==":
		switch v := v.(type) {
		case Function:
			if len(s.arg) != len(v.arg) {
				return false
			}
			if len(s.res) != len(v.res) {
				return false
			}
			for i, field := range v.arg {
				if s.arg[i].types == field.types {
					continue
				}
				return false
			}
			for i, field := range v.res {
				if s.res[i].types == field.types {
					continue
				}
				return false
			}
			return true
		default:
			return ErrNotSupport(v)
		}
	default:
		return ErrNotImplement(op)
	}
}

func (s Struct) For(cb func(k string, v Any)) {
	kit.For(s.index, cb)
	kit.For(s.sups, func(sup string) {
		if sup, ok := s.stack.value(ice.Pulse, sup).(Struct); ok {
			sup.For(cb)
		}
	})
}
func (s Struct) Find(k string) Any {
	if v, ok := s.index[k]; ok {
		return v
	}
	for _, sup := range s.sups {
		if sup, ok := s.stack.value(ice.Pulse, sup).(Struct); ok {
			if v := sup.Find(k); v != nil {
				return v
			}
		}
	}
	return ErrNotFound(k)
}
func (s Struct) Operate(op string, v Any) Any {
	switch op {
	case "==":
		switch v := v.(type) {
		case Struct:
			return Boolean{s.name == v.name}
		default:
			return ErrNotSupport(v)
		}
	default:
		return ErrNotImplement(op)
	}
}
func (s Object) Operate(op string, v Any) Any {
	switch op {
	case "&", "*":
		return s
	case INSTANCEOF:
		if t, ok := v.(Struct); ok {
			return Value{list: []Any{s, Boolean{s.index.name == t.name}}}
		}
		return ErrNotSupport(v)
	case IMPLEMENTS:
		if t, ok := v.(Interface); ok {
			for k, v := range t.index {
				if v.Operate("==", s.index.Find(k)) == false {
					return Value{list: []Any{s, Boolean{false}}}
				}
			}
			return Value{list: []Any{s, Boolean{true}}}
		}
		return ErrNotSupport(v)
	case SUBS:
		switch v := s.index.Find(kit.Format(v)).(type) {
		case string:
			switch _v := s.value.Operate(op, v).(type) {
			case nil:
				return Object{Dict{kit.Dict()}, s.index.stack.value(ice.Pulse, v).(Struct)}
			default:
				return _v
			}
		case Function:
			v.object = s
			return v
		}
		fallthrough
	default:
		return s.value.Operate(op, v)
	}
}

const (
	MAP       = "map"
	SLICE     = "slice"
	STRUCT    = "struct"
	INTERFACE = "interface"
	STRING    = "string"
	BOOL      = "bool"
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
			pos := s.Position
			switch name := s.next(m); s.next(m) {
			case ASSIGN:
				s.next(m)
				fallthrough
			default:
				switch t := s.types(m).(type) {
				case Interface:
					t.name = name
					t.stack = s
					s.value(m, name, t)
				case Struct:
					t.name = name
					t.stack = s
					t.Position = pos
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
