package ctx

import (
	"reflect"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/toolkits/logs"
)

const OPTION = "option"
const MESSAGE = "message"

func init() {
	Index.MergeCommands(ice.Commands{
		OPTION: {Name: "option", Help: "选项", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 1 {
				if msg, ok := m.Optionv("message").(*ice.Message); ok {
					msg.Option(arg[0], arg[1])
				}
			}
		}},
		MESSAGE: {Name: "message auto", Help: "消息", Hand: func(m *ice.Message, arg ...string) {
			t := reflect.TypeOf(m)
			for i := 0; i < t.NumMethod(); i++ {
				method := t.Method(i)
				p := logs.FileLine(method.Func.Interface())
				m.Push(mdb.NAME, method.Name)
				m.Push(mdb.TEXT, strings.Split(p, ice.ICEBERGS+ice.PS)[1])
			}
		}},
	})
}
