package yac

import (
	"encoding/json"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

type Any = ice.Any
type Dict struct{ value map[string]Any }
type List struct{ value []Any }
type Value struct{ list []Any }
type String struct{ value string }
type Number struct{ value string }
type Boolean struct{ value bool }
type Caller interface{ Call(string, ...Any) Any }
type Operater interface{ Operate(string, Any) Any }

func (s Dict) MarshalJSON() ([]byte, error)    { return json.Marshal(s.value) }
func (s List) MarshalJSON() ([]byte, error)    { return json.Marshal(s.value) }
func (s Value) MarshalJSON() ([]byte, error)   { return json.Marshal(s.list) }
func (s String) MarshalJSON() ([]byte, error)  { return json.Marshal(s.value) }
func (s Number) MarshalJSON() ([]byte, error)  { return json.Marshal(s.value) }
func (s Boolean) MarshalJSON() ([]byte, error) { return json.Marshal(s.value) }

func wrap(v Any) Any {
	switch v := v.(type) {
	case string:
		return String{v}
	default:
		return v
	}
}
func trans(v Any) Any {
	switch v := v.(type) {
	case Dict:
		return v.value
	case List:
		return v.value
	case Value:
		if len(v.list) > 0 {
			return v.list[0]
		}
		return nil
	case String:
		return v.value
	case Number:
		return v.value
	case Boolean:
		return v.value
	default:
		return v
	}
}
func (s Dict) Operate(op string, v Any) Any {
	switch op {
	case SUBS:
		return wrap(kit.Value(s.value, kit.Format(v)))
	}
	return nil
}
func (s List) Operate(op string, v Any) Any {
	switch op {
	case SUBS:
		return wrap(kit.Value(s.value, kit.Format(v)))
	}
	return nil
}
func (s Value) Operate(op string, v Any) Any {
	switch op {
	case SUBS:
		if i := kit.Int(v); i < len(s.list) {
			return s.list[i]
		}
	}
	return nil
}
func (s String) Operate(op string, v Any) Any {
	switch a, b := s.value, kit.Format(v); op {
	case "+":
		return String{a + b}
	case "-":
		return String{strings.Replace(a, b, "", -1)}
	case "<":
		return Boolean{a < b}
	case ">":
		return Boolean{a > b}
	case ">=":
		return Boolean{a >= b}
	case "<=":
		return Boolean{a <= b}
	case "!=":
		return Boolean{a != b}
	case "==":
		return Boolean{a == b}
	default:
		return nil
	}
}
func (s Number) Operate(op string, v Any) Any {
	switch a, b := kit.Int(s.value), kit.Int(v); op {
	case "*":
		return Number{kit.Format(a * b)}
	case "/":
		return Number{kit.Format(a / b)}
	case "+":
		return Number{kit.Format(a + b)}
	case "-":
		return Number{kit.Format(a - b)}
	case "<":
		return Boolean{a < b}
	case ">":
		return Boolean{a > b}
	case ">=":
		return Boolean{a >= b}
	case "<=":
		return Boolean{a <= b}
	case "!=":
		return Boolean{a != b}
	case "==":
		return Boolean{a == b}
	default:
		return nil
	}
}
func (s Boolean) Operate(op string, v Any) Any {
	switch a, b := s.value, !kit.IsIn(kit.Format(v), "", "0", ice.FALSE); op {
	case "&&":
		return Boolean{a && b}
	case "||":
		return Boolean{a || b}
	case "!":
		return Boolean{!a}
	default:
		return nil
	}
}

type Message struct{ *ice.Message }

func (m Message) Call(cmd string, arg ...Any) Any {
	str := func(v Any) string { return kit.Format(trans(v)) }
	args := []Any{}
	for _, v := range arg {
		args = append(args, trans(v))
	}
	switch cmd {
	case "Option":
		return String{m.Option(str(args[0]), args[1:]...)}
	case "Cmd":
		return Message{m.Cmd(args...)}
	case "Action":
		m.Action(args...)
	case "Cmdy":
		m.Cmdy(args...)
	case "Copy":
		m.Copy(args[0].(Message).Message, kit.Simple(args[1:]...)...)
	case "Push":
		m.Push(str(args[0]), args[1], args[2:]...)
	case "Echo":
		m.Echo(str(args[0]), args[1:]...)
	case "Sleep":
		m.Sleep(str(args[0]))
	case "Table":
		s := _parse_stack(m.Message)
		var value Any
		m.Table(func(v ice.Maps) { value = s.call(m.Message, v, nil, Dict{kit.Dict(v)}) })
		return value
	default:
		m.ErrorNotImplement(cmd)
	}
	return m
}
