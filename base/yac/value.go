package yac

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
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

func Wraps(v Any) Any {
	switch v := v.(type) {
	case *ice.Message:
		return Message{v}
	case []Any:
		list := []Any{}
		kit.For(v, func(v Any) { list = append(list, Wraps(v)) })
		return List{list}
	case map[string]Any:
		list := map[string]Any{}
		kit.For(v, func(k string, v Any) { list[k] = Wraps(v) })
		return Dict{list}
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
	case Message:
		return v.Message
	case Value:
		list := []Any{}
		kit.For(v.list, func(v Any) { list = append(list, Trans(v)) })
		return list
	case List:
		list := []Any{}
		kit.For(v.value, func(v Any) { list = append(list, Trans(v)) })
		return list
	case Dict:
		list := map[string]Any{}
		kit.For(v.value, func(k string, v Any) { list[k] = Trans(v) })
		return list
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
	case LEN:
		return len(s.value)
	case RANGE:
		switch list := v.(type) {
		case []Any:
			if len(list) > 1 {
				if i := kit.Int(list[0]) + 1; i < len(s.value) {
					return []Any{i, s.value[i]}
				}
			} else if len(s.value) > 0 {
				return []Any{0, s.value[0]}
			}
		default:
			return ErrNotSupport(list)
		}
		return nil
	case EXPAND:
		return s.value
	case SUBS:
		return Wraps(kit.Value(s.value, kit.Format(v)))
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
		} else {
			return ErrNotValid(op)
		}
	}
}
func (s Dict) Operate(op string, v Any) Any {
	switch op {
	case LEN:
		return len(s.value)
	case RANGE:
		switch list := v.(type) {
		case []Any:
			if key := kit.SortedKey(s.value); len(list) > 2 {
				if i := kit.Int(list[2]) + 1; i < len(key) {
					return []Any{key[i], s.value[key[i]], i}
				}
			} else if len(key) > 0 {
				return []Any{key[0], s.value[key[0]], 0}
			}
		}
		return ErrNotSupport(v)
	case EXPAND:
		list := []Any{}
		kit.For(kit.KeyValue(nil, "", s.value), func(k string, v Any) { list = append(list, k, v) })
		return list
	case SUBS:
		return Wraps(kit.Value(s.value, kit.Format(v)))
	default:
		s.value[op] = v
		return v
	}
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
		return ErrNotImplement(op)
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
		return ErrNotImplement(op)
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
		return ErrNotImplement(op)
	}
}

const (
	LEN = "len"
	KIT = "kit"
)

func init() {
	ice.Info.Stack[LEN] = func(m *ice.Message, key string, arg ...Any) Any {
		if len(arg) == 0 {
			return ErrNotValid()
		}
		switch v := arg[0].(type) {
		case Operater:
			return v.Operate(LEN, nil)
		case map[string]Any:
			return len(v)
		case []Any:
			return len(v)
		default:
			return ErrNotSupport(v)
		}
	}
}
func init() {
	ice.Info.Stack[KIT] = func(m *ice.Message, key string, arg ...Any) Any {
		switch key {
		case "Dict":
			return kit.Dict(arg...)
		case "List":
			return kit.List(arg...)
		case "Slice":
			args := []int{}
			kit.For(arg, func(v string) { args = append(args, kit.Int(v)) })
			return kit.Slice(kit.Simple(arg[0]), args...)
		case "Select":
			return kit.Select(kit.Format(arg[0]), arg[1:]...)
		case "Format":
			return kit.Format(arg[0], arg[1:]...)
		case "Formats":
			return kit.Formats(arg[0])
		default:
			return ErrNotImplement(kit.Keys(KIT, key))
		}
	}
}
func init() {
	ice.Info.Stack["ice.Cmd"] = func(m *ice.Message, key string, arg ...Any) Any {
		if len(arg) < 2 {
			return ErrNotValid(arg)
		}
		obj, ok := arg[1].(Object)
		if !ok {
			return ErrNotSupport(arg[1])
		}
		stack, ok := m.Optionv(ice.YAC_STACK).(*Stack)
		if !ok {
			return ErrNotFound(ice.YAC_STACK)
		}
		if key = kit.Format(arg[0]); key == "" {
			ls := kit.Split(stack.Position.name, nfs.PS)
			key = kit.Keys("web.code", kit.Select("", ls, -2), kit.TrimExt(kit.Select("", ls, -1)))
		}
		command, config := &ice.Command{Actions: ice.Actions{}}, &ice.Config{Value: kit.Data()}
		obj.index.For(func(k string, v Any) {
			switch k = kit.LowerCapital(k); v := v.(type) {
			case Function:
				if v.object = obj; k == mdb.LIST {
					kit.If(command.Hand == nil, func() { command.Hand = stack.Handler(v) })
				} else {
					kit.If(command.Actions[k] == nil, func() { command.Actions[k] = &ice.Action{Hand: stack.Handler(v)} })
				}
			}
		})
		obj.index.For(func(k string, v Any) {
			switch v := v.(type) {
			case Field:
				if k == mdb.LIST {
					kit.If(command.Name == "", func() { command.Name, command.Help = v.tags[mdb.NAME], v.tags[mdb.HELP] })
				} else if action, ok := command.Actions[k]; ok {
					kit.If(action.Name == "", func() { action.Name, action.Help = v.tags[mdb.NAME], v.tags[mdb.HELP] })
				}
				kit.If(v.tags[mdb.DATA] != "", func() { kit.Value(config.Value, kit.Keym(v.name), v.tags[mdb.DATA]) })
			}
		})
		last, list := ice.Index, kit.Split(key, nfs.PT)
		for i := 1; i < len(list); i++ {
			has := false
			if ice.Pulse.Search(strings.Join(list[:i], nfs.PT)+nfs.PT, func(p *ice.Context, s *ice.Context) { has, last = true, s }); !has {
				last = last.Register(&ice.Context{Name: list[i-1], Caches: ice.Caches{ice.CTX_FOLLOW: &ice.Cache{Value: kit.Keys(list[i-1])}}}, &web.Frame{})
			}
			if i == len(list)-1 {
				last.Merge(&ice.Context{Commands: ice.Commands{list[i]: command}, Configs: ice.Configs{list[i]: config}})
			}
		}
		stack.frame[0].value["_link"] = _parse_link(m, key)
		stack.frame[0].value["_index"] = key
		return key
	}
	return
	ice.Info.Stack["ice.MergeActions"] = func(m *ice.Message, key string, arg ...Any) Any {
		list := ice.Actions{}
		kit.For(arg, func(v Any) { ice.MergeActions(list, TransActions(m, v)) })
		res := Dict{kit.Dict()}
		for k, v := range list {
			res.value[k] = Dict{kit.Dict("Name", v.Name, "Help", v.Help, "Hand", v.Hand)}
		}
		return res
	}
}

func TransContext(m *ice.Message, key string, arg ...Any) *ice.Context {
	s := &ice.Context{Caches: ice.Caches{ice.CTX_FOLLOW: &ice.Cache{}}}
	defer func() { s.Merge(s).Cap(ice.CTX_FOLLOW, kit.Keys(key, s.Name)) }()
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
	return s
}
func TransCommands(m *ice.Message, arg ...Any) ice.Commands {
	commands := ice.Commands{}
	stack := m.Optionv(ice.YAC_STACK).(*Stack)
	kit.For(arg[0], func(k string, v ice.Any) {
		s := &ice.Command{}
		defer func() { commands[k] = s }()
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
		defer func() { actions[k] = s }()
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
				res = append(res, kit.Format("%s:%d:%d", v.name, v.line+1, v.skip+1))
			}
		case Map:
			res = append(res, kit.Format("map[%s]%s", v.key, v.value))
		case Slice:
			res = append(res, kit.Format("[]%s", v.value))
		case Interface:
			res = append(res, kit.Format("interface:%s", v.name))
		case Struct:
			res = append(res, kit.Format("struct%s", Format(v.index)))
		case Object:
			res = append(res, kit.Format("%s:%s", v.index.name, Format(v.value)))
		case Fields:
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
	return strings.Join(res, ice.SP)
}

type Message struct{ *ice.Message }

func (m Message) Call(cmd string, arg ...Any) Any {
	switch cmd {
	case "Table":
		s := _parse_stack(m.Message)
		m.Table(func(val ice.Maps) { s.calls(m.Message, arg[0], "", nil, Dict{kit.Dict(val)}) })
	case "Display":
		if len(arg) > 0 {
			m.ProcessDisplay(arg...)
		} else {
			m.ProcessDisplay(kit.Format("%s?_t=%d", Trans(_parse_stack(m.Message).value(m.Message, "_script")), time.Now().Unix()))
		}
	case "DebugStack":
		list := []string{}
		s := _parse_stack(m.Message)
		s.stack(func(f *Frame, i int) bool {
			list = append(list, kit.Format("stack: %s", f.key))
			kit.For(f.value, func(k string, v Any) { list = append(list, kit.Format("stack: %s %s:%#v", f.key, k, v)) })
			return false
		})
		m.Debug(ice.NL + strings.Join(list, ice.NL))
	default:
		msg, args := reflect.ValueOf(m), []reflect.Value{}
		kit.For(arg, func(v Any) { args = append(args, reflect.ValueOf(v)) })
		if res := msg.MethodByName(cmd).Call(args); len(res) > 0 && res[0].CanInterface() {
			return res[0].Interface()
		}
	}
	return m
}
