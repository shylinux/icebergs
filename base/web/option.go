package web

import (
	"net/url"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

func UserWeb(m *ice.Message) *url.URL {
	return kit.ParseURL(m.Option(ice.MSG_USERWEB))
}
func UserHost(m *ice.Message) string {
	if p := m.Option(ice.MSG_USERHOST); p != "" {
		return p
	} else if u := UserWeb(m); u.Hostname() == tcp.LOCALHOST {
		return m.Option(ice.MSG_USERHOST, tcp.PublishLocalhost(m, u.Scheme+"://"+u.Host))
	} else {
		return m.Option(ice.MSG_USERHOST, u.Scheme+"://"+u.Host)
	}
}
func AgentIs(m *ice.Message, arg ...string) bool {
	for _, k := range arg {
		if strings.HasPrefix(strings.ToLower(m.Option(ice.MSG_USERUA)), k) {
			return true
		}
	}
	return false
}
func ParseLink(m *ice.Message, url string) ice.Maps {
	list := ice.Maps{}
	u := kit.ParseURL(url)
	switch arg := strings.Split(strings.TrimPrefix(u.Path, nfs.PS), nfs.PS); arg[0] {
	case CHAT:
		kit.For(arg[1:], func(k, v string) { list[k] = v })
	case SHARE:
		list[arg[0]] = arg[1]
	}
	kit.For(u.Query(), func(k string, v []string) { list[k] = v[0] })
	return list
}
func PushPodCmd(m *ice.Message, cmd string, arg ...string) *ice.Message {
	list := []string{}
	m.Cmds(SPACE, func(value ice.Maps) {
		kit.If(kit.IsIn(value[mdb.TYPE], WORKER), func() { list = append(list, value[mdb.NAME]) })
	})
	if len(list) == 0 {
		return m
	}
	kit.If(m.Length() > 0 && len(m.Appendv(SPACE)) == 0, func() { m.Table(func(value ice.Maps) { m.Push(SPACE, "") }) })
	GoToast(m, "", func(toast func(string, int, int)) []string {
		kit.For(list, func(index int, space string) {
			toast(space, index, len(list))
			m.Cmd(SPACE, space, kit.Dict(ice.MSG_USERPOD, space), kit.Select(m.PrefixKey(), cmd), arg).Table(func(index int, val ice.Maps, head []string) {
				kit.If(!kit.IsIn(SPACE, head...), func() { head = append(head, SPACE) })
				val[SPACE] = kit.Keys(space, val[SPACE])
				m.Push("", val, head)
			})
		})
		return nil
	})
	return m
}
func PushImages(m *ice.Message, name string) {
	if kit.ExtIsImage(name) {
		m.PushImages(IMAGE, name)
	} else if kit.ExtIsVideo(name) {
		m.PushVideos(VIDEO, name)
	}
}
func PushNotice(m *ice.Message, arg ...ice.Any) {
	if m.Option(ice.MSG_DAEMON) == "" {
		return
	}
	opts := ice.Map{ice.MSG_OPTION: []string{}, ice.MSG_OPTS: []string{}}
	kit.For([]string{ice.MSG_TITLE, ctx.DISPLAY, ctx.STYLE, cli.DELAY, ice.MSG_STATUS, ice.LOG_DEBUG, ice.LOG_TRACEID}, func(key string) {
		opts[ice.MSG_OPTION] = kit.Simple(opts[ice.MSG_OPTION], key)
		opts[key] = m.Option(key)
	})
	m.Cmd(SPACE, m.Option(ice.MSG_DAEMON), arg, opts)
}
func PushNoticeRefresh(m *ice.Message, arg ...ice.Any) { PushNotice(m, kit.List("refresh", arg)...) }
func PushNoticeToast(m *ice.Message, arg ...ice.Any)   { PushNotice(m, kit.List("toast", arg)...) }
func PushNoticeGrow(m *ice.Message, arg ...ice.Any) {
	PushNotice(m.StatusTimeCount(), kit.List("grow", arg)...)
}
func PushNoticeRich(m *ice.Message, arg ...ice.Any) {
	PushNotice(m.StatusTimeCount(), kit.Simple("rich", arg))
}
func PushStream(m *ice.Message) *ice.Message {
	m.Options(cli.CMD_OUTPUT, file.NewWriteCloser(func(buf []byte) { PushNoticeGrow(m, string(buf)) }, nil)).ProcessHold(toastContent(m, ice.SUCCESS))
	return m
}
func init() { ice.Info.PushStream = PushStream }
func init() { ice.Info.PushNotice = PushNotice }

func Toast(m *ice.Message, text string, arg ...ice.Any) *ice.Message { // [title [duration [progress]]]
	if len(arg) > 1 {
		switch val := arg[1].(type) {
		case string:
			if value, err := time.ParseDuration(val); err == nil {
				arg[1] = int(value / time.Millisecond)
			}
		}
	}
	kit.If(len(arg) == 0, func() { arg = append(arg, m.PrefixKey()) })
	kit.If(len(arg) > 0, func() { arg[0] = kit.Select(m.PrefixKey(), arg[0]) })
	PushNoticeToast(m, text, arg)
	return m
}

var Icons = map[string]string{ice.PROCESS: "🕑", ice.FAILURE: "❌", ice.SUCCESS: "✅"}

func toastContent(m *ice.Message, state string, arg ...ice.Any) string {
	if len(arg) == 0 {
		return kit.JoinWord(kit.Simple(Icons[state], kit.Select("", m.ActionKey(), m.ActionKey() != ice.LIST), state)...)
	} else {
		return kit.JoinWord(kit.Simple(Icons[state], arg)...)
	}
}
func ToastSuccess(m *ice.Message, arg ...ice.Any) {
	Toast(m, toastContent(m, ice.SUCCESS, arg...), "", cli.TIME_3s)
}
func ToastFailure(m *ice.Message, arg ...ice.Any) {
	Toast(m, toastContent(m, ice.FAILURE, arg...), "", m.OptionDefault(ice.TOAST_DURATION, cli.TIME_3s)).Sleep(m.OptionDefault(ice.TOAST_DURATION, cli.TIME_3s))
}
func ToastProcess(m *ice.Message, arg ...ice.Any) func(...ice.Any) {
	Toast(m, toastContent(m, ice.PROCESS, arg...), "", cli.TIME_30s)
	return func(arg ...ice.Any) { Toast(m, toastContent(m, ice.SUCCESS, arg...), "", "1s") }
}
func GoToast(m *ice.Message, title string, cb func(toast func(name string, count, total int)) []string) *ice.Message {
	_total := 0
	icon := Icons[ice.PROCESS]
	toast := func(name string, count, total int) {
		kit.If(total == 0, func() { total = 1 })
		Toast(m, kit.Format("%s %s %s", icon, kit.JoinWord(kit.Select(kit.Select("", m.ActionKey(), m.ActionKey() != ice.LIST), title, m.Option(ice.MSG_TITLE)), name), strings.ReplaceAll(kit.FmtSize(count, total), "B", "")),
			"", m.OptionDefault(ice.TOAST_DURATION, cli.TIME_30s), count*100/total)
		_total = total
	}
	if list := cb(toast); len(list) > 0 {
		icon = Icons[ice.FAILURE]
		m.Option(ice.TOAST_DURATION, cli.TIME_30s)
		toast(kit.JoinWord(list...), len(list), _total)
	} else {
		icon = Icons[ice.SUCCESS]
		m.Option(ice.TOAST_DURATION, cli.TIME_1s)
		toast(ice.SUCCESS, _total, _total)
	}
	m.Sleep(m.Option(ice.TOAST_DURATION))
	return m
}
