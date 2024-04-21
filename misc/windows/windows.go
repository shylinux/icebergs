package windows

import (
	"strings"
	"time"

	"shylinux.com/x/ice"
	kit "shylinux.com/x/toolkits"
)

type windows struct {
	cmds string `data:"logged,process,service,installed"`
	list string `name:"list name auto"`
}

func (s windows) List(m *ice.Message, arg ...string) { m.DisplayStudio() }

func init() { ice.ChatCtxCmd(windows{}) }

func ListPush(m *ice.Message, list ice.Any, err error, arg ...string) *ice.Message {
	if !m.Warn(err) {
		kit.For(kit.UnMarshal(kit.Format(list)), func(value ice.Map) { m.PushRecord(value, arg...) })
		kit.If(m.IsDebug(), func() { m.Echo(kit.Formats(list)) })
	}
	return m
}
func ParseTime(m *ice.Message, value string) string {
	if t, e := time.Parse(time.RFC3339, value); e == nil {
		value = strings.TrimSuffix(t.Local().Format(ice.MOD_TIME), ".000")
	}
	return value
}
