package ice

import (
	"reflect"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func ref(obj interface{}) (reflect.Type, reflect.Value) {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t, v = t.Elem(), v.Elem()
	}
	return t, v
}
func val(m *Message, arg ...string) []reflect.Value {
	args := []reflect.Value{reflect.ValueOf(m)}
	for _, v := range arg {
		args = append(args, reflect.ValueOf(v))
	}
	return args
}
func transMethod(config *Config, command *Command, obj interface{}) {
	t, v := ref(obj)
	for i := 0; i < v.NumMethod(); i++ {
		method := v.Method(i)
		var h func(*Message, ...string)
		switch method.Interface().(type) {
		case func(*Message, ...string):
			h = func(m *Message, arg ...string) { method.Call(val(m, arg...)) }
		case func(*Message):
			h = func(m *Message, arg ...string) { method.Call(val(m)) }
		default:
			continue
		}

		if key := strings.ToLower(t.Method(i).Name); key == "list" {
			command.Hand = func(m *Message, c *Context, cmd string, arg ...string) { h(m, arg...) }
		} else {
			if action, ok := command.Action[key]; !ok {
				command.Action[key] = &Action{Hand: h}
			} else {
				action.Hand = h
			}
		}
	}

}
func transField(config *Config, command *Command, obj interface{}) {
	t, v := ref(obj)
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.Struct {
			if v.Field(i).CanInterface() {
				transField(config, command, v.Field(i).Interface())
			}
		}
	}

	meta := kit.Value(config.Value, kit.MDB_META)
	for i := 0; i < v.NumField(); i++ {
		key, tag := t.Field(i).Name, t.Field(i).Tag
		if data := tag.Get("data"); data != "" {
			kit.Value(meta, key, data)
		}

		if name := tag.Get("name"); name != "" {
			if help := tag.Get("help"); key == "list" {
				command.Name, command.Help = name, help
				config.Name, config.Help = name, help
			} else if action, ok := command.Action[key]; ok {
				action.Name, action.Help = name, help
			}
		}
	}
}
func Cmd(key string, obj interface{}) {
	if obj == nil {
		return
	}
	command := &Command{Action: map[string]*Action{}}
	config := &Config{Value: kit.Data()}
	transMethod(config, command, obj)
	transField(config, command, obj)

	last := Index
	list := strings.Split(key, ".")
	for i := 1; i < len(list); i++ {
		has := false
		Pulse.Search(strings.Join(list[:i], ".")+".", func(p *Context, s *Context) {
			has, last = true, s
		})
		if !has {
			context := &Context{Name: list[i-1]}
			last.Register(context, nil)
			last = context
		}
		if i < len(list)-1 {
			continue
		}

		last.Merge(&Context{Commands: map[string]*Command{
			CTX_INIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
				if action, ok := command.Action["init"]; ok {
					action.Hand(m, arg...)
				} else {
					m.Load()
				}
			}},
			CTX_EXIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
				m.Save()
			}},
			list[i]: command,
		}, Configs: map[string]*Config{list[i]: config}})
	}
}
