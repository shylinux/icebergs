package ctx

import (
	"reflect"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/toolkits/logs"
)

const MESSAGE = "message"
const OPTION = "option"

func init() {
	Index.MergeCommands(ice.Commands{
		MESSAGE: {Name: "message", Help: "消息", Hand: func(m *ice.Message, arg ...string) {
			t := reflect.TypeOf(m)
			for i := 0; i < t.NumMethod(); i++ {
				method := t.Method(i)
				p := logs.FileLine(method.Func.Interface())
				m.Push(mdb.NAME, method.Name)
				m.Push(mdb.TEXT, strings.Split(p, ice.ICEBERGS+"/")[1])
			}
		}},
		OPTION: {Name: "option", Help: "选项", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 1 {
				if msg, ok := m.Optionv("message").(*ice.Message); ok {
					msg.Option(arg[0], arg[1])
				}
			}
		}},
	})
}
