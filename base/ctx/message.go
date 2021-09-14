package ctx

import (
	"reflect"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const MESSAGE = "message"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		MESSAGE: {Name: "message", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			t := reflect.TypeOf(m)
			for i := 0; i < t.NumMethod(); i++ {
				method := t.Method(i)
				p := kit.FileLine(method.Func.Interface(), 4)
				m.Push(kit.MDB_NAME, method.Name)
				m.Push(kit.MDB_TEXT, strings.Split(p, ice.ICEBERGS+"/")[1])
			}
		}},
	}})
}