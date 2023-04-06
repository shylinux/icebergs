package yac

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

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
func (s String) MarshalJSON() ([]byte, error)  { return json.Marshal(s.value) }
func (s Number) MarshalJSON() ([]byte, error)  { return json.Marshal(s.value) }
func (s Boolean) MarshalJSON() ([]byte, error) { return json.Marshal(s.value) }

func wrap(v Any) Any {
	switch v := v.(type) {
	case map[string]Any:
		return Dict{v}
	case []Any:
		return List{v}
	case string:
		return String{v}
	case int:
		return Number{kit.Format(v)}
	case bool:
		return Boolean{v}
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
			return trans(v.list[0])
		} else {
			return nil
		}
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
	case RANGE:
		switch list := v.(type) {
		case []Any:
			if key := kit.SortedKey(s.value); list != nil && len(list) > 2 {
				if i := kit.Int(list[2]) + 1; i < len(key) {
					return []Any{key[i], s.value[key[i]], i}
				}
			} else {
				if len(key) > 0 {
					return []Any{key[0], s.value[key[0]], 0}
				}
			}
		}
		return nil
	case SUBS:
		return wrap(kit.Value(s.value, kit.Format(v)))
	default:
		s.value[op] = v
		return v
	}
	return nil
}
func (s List) Operate(op string, v Any) Any {
	switch op {
	case RANGE:
		switch list := v.(type) {
		case []Any:
			if list != nil && len(list) > 1 {
				if i := kit.Int(list[0]) + 1; i < len(s.value) {
					return []Any{i, s.value[i]}
				}
			} else {
				if len(s.value) > 0 {
					return []Any{0, s.value[0]}
				}
			}
		}
		return nil
	case SUBS:
		return wrap(kit.Value(s.value, kit.Format(v)))
	default:
		if i, e := strconv.ParseInt(op, 10, 32); e == nil {
			switch (int(i)+2+len(s.value)+2)%(len(s.value)+2) - 2 {
			case -1:
				s.value = append([]Any{v}, s.value...)
			case -2:
				s.value = append(s.value, v)
			default:
				s.value[i] = v
			}
			return v
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
	case "++":
		return Number{kit.Format(a + 1)}
	case "--":
		return Number{kit.Format(a - 1)}
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
func (s *Stack) load(m *ice.Message, cb func(*Frame)) *Stack {
	f := s.peekf()
	f.value["kit"] = func(m *ice.Message, key string, arg ...Any) Any {
		switch key {
		case "Dict":
			return kit.Dict(arg...)
		case "List":
			return kit.List(arg...)
		case "Format":
			return kit.Format(arg[0], arg[1:]...)
		case "Formats":
			return kit.Formats(arg[0])
		default:
			m.ErrorNotImplement(key)
			return nil
		}
	}
	f.value["ice.MergeActions"] = func(m *ice.Message, key string, arg ...Any) Any {
		s := _parse_stack(m)
		_actions := ice.Actions{}
		for _, v := range arg {
			actions := ice.Actions{}
			kit.For(v, func(k string, v Any) {
				action := &ice.Action{}
				kit.For(v, func(k string, v Any) {
					switch k {
					case "Name":
						action.Name = kit.Format(trans(v))
					case "Help":
						action.Help = kit.Format(trans(v))
					case "Hand":
						action.Hand = func(m *ice.Message, arg ...string) { s.action(m, v, nil, arg...) }
					}
				})
				actions[k] = action
			})
			ice.MergeActions(_actions, actions)
		}
		return _actions
	}
	for k, v := range ice.Info.Stack {
		f.value[k] = v
	}
	f.value["m"] = Message{m}
	kit.If(cb != nil, func() { cb(f) })
	return s
}

func (m Message) Call(cmd string, arg ...Any) Any {
	str := func(v Any) string { return kit.Format(trans(v)) }
	switch cmd {
	case "Option":
		return m.Option(str(arg[0]), arg[1:]...)
	case "Cmd":
		return Message{m.Cmd(arg...)}
	case "Cmdy":
		m.Cmdy(arg...)
	case "Copy":
		m.Copy(arg[0].(Message).Message, kit.Simple(arg[1:]...)...)
	case "Push":
		m.Push(str(arg[0]), arg[1], arg[2:]...)
	case "Echo":
		m.Echo(str(arg[0]), arg[1:]...)
	case "Action":
		m.Action(arg...)
	case "Display":
		if len(arg) > 0 {
			m.ProcessDisplay(arg...)
		} else {
			m.ProcessDisplay(kit.Format("%s?_t=%d", trans(_parse_stack(m.Message).value(m.Message, "_script")), time.Now().Unix()))
		}
	case "StatusTime":
		m.StatusTime(arg...)
	case "DebugStack":
		s := _parse_stack(m.Message)
		list := []string{}
		s.stack(func(f *Frame, i int) bool {
			list = append(list, kit.Format("stack: %s", f.key))
			kit.For(f.value, func(k string, v Any) {
				list = append(list, kit.Format("stack: %s %s:%#v", f.key, k, v))
			})
			return false
		})
		m.Debug(ice.NL + strings.Join(list, ice.NL))
	case "Sleep":
		m.Sleep(str(arg[0]))
	case "Table":
		s := _parse_stack(m.Message)
		m.Table(func(val ice.Maps) { s.calls(m.Message, arg[0], nil, nil, Dict{kit.Dict(val)}) })
	default:
		m.ErrorNotImplement(cmd)
	}
	return m
}

type Message struct{ *ice.Message }
