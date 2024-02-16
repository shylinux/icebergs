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
	msg := m.Spawn()
	m.Cmds(SPACE, func(value ice.Maps) {
		kit.If(kit.IsIn(value[mdb.TYPE], WORKER), func() { msg.Push(SPACE, value[mdb.NAME]) })
	})
	kit.If(m.Length() > 0 && len(m.Appendv(SPACE)) == 0, func() { m.Table(func(value ice.Maps) { m.Push(SPACE, "") }) })
	GoToastTable(msg, SPACE, func(value ice.Maps) {
		m.Cmd(SPACE, value[SPACE], kit.Dict(ice.MSG_USERPOD, value[SPACE]), kit.Select(m.PrefixKey(), cmd), arg).Table(func(val ice.Maps, index int, head []string) {
			kit.If(!kit.IsIn(SPACE, head...), func() { head = append(head, SPACE) })
			val[SPACE] = kit.Keys(value[SPACE], val[SPACE])
			m.Push("", val, head)
		})
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
	kit.For([]string{ctx.DISPLAY, ctx.STYLE, cli.DELAY, ice.MSG_TITLE, ice.MSG_STATUS, ice.LOG_DEBUG, ice.LOG_TRACEID}, func(key string) {
		opts[ice.MSG_OPTION], opts[key] = kit.Simple(opts[ice.MSG_OPTION], key), m.Option(key)
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

func Toast(m *ice.Message, text string, arg ...ice.Any) *ice.Message { // [title [duration [progress [hash]]]]
	if len(arg) > 1 {
		switch val := arg[1].(type) {
		case string:
			if value, err := time.ParseDuration(val); err == nil {
				arg[1] = int(value / time.Millisecond)
			}
		}
	}
	kit.If(len(arg) == 0, func() { arg = append(arg, m.PrefixKey()) })
	kit.If(len(arg) > 0 && arg[0] == "", func() {
		arg[0] = kit.Keys(m.Option(ice.MSG_USERPOD0), m.Option(ice.MSG_USERPOD), ctx.ShortCmd(m.PrefixKey()))
	})
	PushNoticeToast(m, text, arg)
	return m
}

var Icons = map[string]string{ice.PROCESS: "🕑", ice.FAILURE: "❌", ice.SUCCESS: "✅"}

func toastContent(m *ice.Message, state string, arg ...ice.Any) string {
	if len(arg) == 0 {
		return kit.JoinWord(kit.Simple(Icons[state], kit.Select(ice.LIST, m.ActionKey()), state)...)
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
	h := kit.HashsUniq()
	Toast(m, toastContent(m, ice.PROCESS, arg...), "", "-1", "", h)
	return func(arg ...ice.Any) { Toast(m, toastContent(m, ice.SUCCESS, arg...), "", cli.TIME_3s, "", h) }
}
func GoToastTable(m *ice.Message, key string, cb func(value ice.Maps)) *ice.Message {
	if m.Length() == 0 {
		return m
	}
	return GoToast(m, func(toast func(string, int, int)) []string {
		m.Table(func(value ice.Maps, index, total int) { toast(value[key], index, total); cb(value) })
		return nil
	})
}
func GoToast(m *ice.Message, cb func(toast func(name string, count, total int)) []string) *ice.Message {
	icon, _total, h := Icons[ice.PROCESS], 1, kit.HashsUniq()
	toast := func(name string, count, total int) {
		kit.If(total == 0, func() { total = 1 })
		Toast(m, kit.Format("%s %s %s", icon, kit.JoinWord(kit.Select(ice.LIST, m.ActionKey()), name), strings.ReplaceAll(kit.FmtSize(count, total), "B", "")),
			m.Option(ice.MSG_TITLE), m.OptionDefault(ice.TOAST_DURATION, "-1"), count*100/total, h)
		_total = total
	}
	if list := cb(toast); len(list) > 0 {
		icon = Icons[ice.FAILURE]
		m.Option(ice.TOAST_DURATION, cli.TIME_30s)
		toast(kit.JoinWord(list...), len(list), _total)
	} else {
		icon = Icons[ice.SUCCESS]
		m.Option(ice.TOAST_DURATION, cli.TIME_3s)
		toast(ice.SUCCESS, _total, _total)
	}
	m.Sleep(m.Option(ice.TOAST_DURATION))
	return m
}
