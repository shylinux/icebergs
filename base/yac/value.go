package yac

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

type Any = ice.Any
type List struct{ value []Any }
type Dict struct{ value map[string]Any }
type String struct{ value string }
type Number struct{ value string }
type Boolean struct{ value bool }
type Caller interface{ Call(string, ...Any) Any }
type Operater interface{ Operate(string, Any) Any }

func (s List) MarshalJSON() ([]byte, error)    { return json.Marshal(s.value) }
func (s Dict) MarshalJSON() ([]byte, error)    { return json.Marshal(s.value) }
func (s String) MarshalJSON() ([]byte, error)  { return json.Marshal(s.value) }
func (s Number) MarshalJSON() ([]byte, error)  { return json.Marshal(s.value) }
func (s Boolean) MarshalJSON() ([]byte, error) { return json.Marshal(s.value) }
func (s Function) MarshalJSON() ([]byte, error) {
	return []byte(kit.Format("%q", Format(s.Position))), nil
}

func wrap(v Any) Any {
	switch v := v.(type) {
	case []Any:
		return List{v}
	case map[string]Any:
		return Dict{v}
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
func Trans(v Any) Any {
	switch v := v.(type) {
	case Value:
		if len(v.list) > 0 {
			return Trans(v.list[0])
		} else {
			return nil
		}
	case List:
		return v.value
	case Dict:
		return v.value
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
		case "List":
			return kit.List(arg...)
		case "Dict":
			return kit.Dict(arg...)
		case "Format":
			return kit.Format(arg[0], arg[1:]...)
		case "Formats":
			return kit.Formats(arg[0])
		default:
			m.ErrorNotImplement(kit.Keys("kit", key))
			return nil
		}
	}
	f.value["ice.Cmd"] = func(m *ice.Message, key string, arg ...Any) Any {
		stack := m.Optionv(ice.YAC_STACK).(*Stack)
		command := &ice.Command{Name: "list hash auto", Help: "示例", Actions: ice.Actions{}}
		obj := arg[1].(Object)
		for k, v := range obj.index.index {
			switch v := v.(type) {
			case Function:
				v.object = obj
				if k == "List" {
					command.Hand = stack.Handler(v)
				} else {
					command.Actions[kit.LowerCapital(k)] = &ice.Action{Hand: stack.Handler(v)}
				}
			}
		}
		for k, v := range obj.index.index {
			switch v := v.(type) {
			case Field:
				if k == "list" {
					command.Name = v.tags[mdb.NAME]
				} else {
					command.Actions[k].Name = v.tags[mdb.NAME]
				}
			}
		}
		last, list := ice.Index, kit.Split(kit.Format(arg[0]), ice.PT)
		for i := 1; i < len(list); i++ {
			has := false
			if ice.Pulse.Search(strings.Join(list[:i], ice.PT)+ice.PT, func(p *ice.Context, s *ice.Context) { has, last = true, s }); !has {
				context := &ice.Context{Name: list[i-1], Caches: ice.Caches{ice.CTX_FOLLOW: &ice.Cache{Value: strings.Join(list[:i], ice.PT)}}}
				last = last.Register(context, &web.Frame{})
			}
			kit.If(i == len(list)-1, func() {
				last.Merge(&ice.Context{Commands: ice.Commands{list[i]: command}})
				last.Merge(last)
			})
		}
		link := ice.Render(m, ice.RENDER_ANCHOR, kit.Format(arg[0]), m.MergePodCmd("", kit.Format(arg[0])))
		s.frame[0].value["_index"] = kit.Format(arg[0])
		s.frame[0].value["_link"] = link
		return nil
	}
	f.value["ice.MergeActions"] = func(m *ice.Message, key string, arg ...Any) Any {
		_actions := ice.Actions{}
		for _, v := range arg {
			ice.MergeActions(_actions, TransActions(m, v))
		}
		res := Dict{value: kit.Dict()}
		for k, v := range _actions {
			res.value[k] = Object{value: Dict{kit.Dict("Name", v.Name, "Help", v.Help, "Hand", v.Hand)}}
		}
		return res
	}
	for k, v := range ice.Info.Stack {
		kit.If(strings.HasPrefix(k, "web.code."), func() { k = strings.TrimPrefix(k, "web.") })
		f.value[k] = v
	}
	f.value["m"] = Message{m}
	kit.If(cb != nil, func() { cb(f) })
	return s
}

func (m Message) Call(cmd string, arg ...Any) Any {
	str := func(v Any) string { return kit.Format(Trans(v)) }
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
	case "Table":
		s := _parse_stack(m.Message)
		m.Table(func(val ice.Maps) { s.calls(m.Message, arg[0], "", nil, Dict{kit.Dict(val)}) })
	case "Sleep":
		m.Sleep(str(arg[0]))
	case "Action":
		m.Action(arg...)
	case "Display":
		if len(arg) > 0 {
			m.ProcessDisplay(arg...)
		} else {
			m.ProcessDisplay(kit.Format("%s?_t=%d", Trans(_parse_stack(m.Message).value(m.Message, "_script")), time.Now().Unix()))
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
	default:
		m.ErrorNotImplement(cmd)
	}
	return m
}

type Message struct{ *ice.Message }

func TransContext(m *ice.Message, key string, arg ...Any) *ice.Context {
	s := &ice.Context{Caches: ice.Caches{ice.CTX_FOLLOW: &ice.Cache{}}}
	kit.For(arg[0], func(k string, v ice.Any) {
		switch k {
		case "Name":
			s.Name = kit.Format(Trans(v))
		case "Help":
			s.Help = kit.Format(Trans(v))
		case "Commands":
			s.Commands = TransCommands(m, v)
		}
	})
	s.Merge(s).Cap(ice.CTX_FOLLOW, kit.Keys(key, s.Name))
	return s
}
func TransCommands(m *ice.Message, arg ...Any) ice.Commands {
	commands := ice.Commands{}
	stack := m.Optionv(ice.YAC_STACK).(*Stack)
	kit.For(arg[0], func(k string, v ice.Any) {
		s := &ice.Command{}
		kit.For(v, func(k string, v ice.Any) {
			switch k {
			case "Name":
				s.Name = kit.Format(Trans(v))
			case "Help":
				s.Help = kit.Format(Trans(v))
			case "Actions":
				s.Actions = TransActions(m, v)
			case "Hand":
				switch v := v.(type) {
				case Function:
					s.Hand = stack.Handler(v)
				case ice.Handler:
					s.Hand = v
				}
			}
		})
		commands[k] = s
	})
	return commands
}
func TransActions(m *ice.Message, arg ...Any) ice.Actions {
	switch v := arg[0].(type) {
	case ice.Actions:
		return v
	}
	actions := ice.Actions{}
	stack := m.Optionv(ice.YAC_STACK).(*Stack)
	kit.For(arg[0], func(k string, v ice.Any) {
		s := &ice.Action{}
		switch k {
		case "Name":
			s.Name = kit.Format(Trans(v))
		case "Help":
			s.Help = kit.Format(Trans(v))
		case "Hand":
			switch v := v.(type) {
			case Function:
				s.Hand = stack.Handler(v)
			case ice.Handler:
				s.Hand = v
			}
		}
		actions[k] = s
	})
	return actions
}

func Format(arg ...Any) string {
	res := []string{}
	for _, v := range arg {
		switch v := v.(type) {
		case func(*ice.Message, string, ...Any) Any:
			res = append(res, kit.FileLine(v, 100))
		case Message:
			res = append(res, kit.FileLine(v.Call, 100))
		case Function:
			res = append(res, kit.Format("%s %s%s%s", Format(v.obj[:len(v.obj)-1]), v.obj[len(v.obj)-1].name, Format(v.arg), Format(v.res)), Format(v.Position))
		case *Stack:
			res = append(res, kit.Format("%d:%s", len(v.frame)-1, kit.Select(v.peekf().key, v.peekf().name)))
		case Position:
			if v.Buffer == nil {
				continue
			} else if v.skip == -1 {
				res = append(res, kit.Format("%s:%d", v.name, v.line+1))
			} else {
				res = append(res, kit.Format("%s:%d:%d", v.name, v.line+1, v.skip))
			}
		case Map:
			res = append(res, kit.Format("map[%s]%s", v.key, v.value))
		case Slice:
			res = append(res, kit.Format("[]%s", v.value))
		case Interface:
			res = append(res, kit.Format("interface%s", v.name))
		case Struct:
			res = append(res, kit.Format("struct%s", Format(v.index)))
		case Object:
			res = append(res, kit.Format("%s:%s", v.index.name, Format(v.value)))
		case []Field:
			for i, field := range v {
				res = append(res, kit.Select("", OPEN, i == 0)+field.name, kit.Format(field.types)+kit.Select(FIELD, CLOSE, i == len(v)-1))
			}
		case List:
			res = append(res, kit.Format(v.value))
		case Dict:
			res = append(res, kit.Format(v.value))
		case String:
			res = append(res, kit.Format("%q", v.value))
		case Number:
			res = append(res, kit.Format("%s", v.value))
		case Boolean:
			res = append(res, kit.Format("%t", v.value))
		case string:
			res = append(res, kit.Format("%q", v))
		default:
			res = append(res, kit.Format(v))
		}
	}
	return strings.Join(res, " ")
}
