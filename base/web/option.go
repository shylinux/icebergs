package web

import (
	"net/url"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

type Message interface {
	Option(key string, arg ...ice.Any) string
	PrefixKey() string
}

func UserWeb(m Message) *url.URL { return kit.ParseURL(m.Option(ice.MSG_USERWEB)) }
func UserHost(m *ice.Message) string {
	if u := UserWeb(m); strings.Contains(u.Host, tcp.LOCALHOST) {
		return m.Option(ice.MSG_USERHOST, tcp.PublishLocalhost(m, u.Scheme+"://"+u.Host))
	} else {
		return m.Option(ice.MSG_USERHOST, u.Scheme+"://"+u.Host)
	}
}
func AgentIs(m Message, arg ...string) bool {
	for _, k := range arg {
		if strings.HasPrefix(strings.ToLower(m.Option(ice.MSG_USERUA)), k) {
			return true
		}
	}
	return false
}
func MergeURL2(m Message, url string, arg ...ice.Any) string {
	kit.If(m.Option(log.DEBUG) == ice.TRUE, func() { arg = append(arg, log.DEBUG, ice.TRUE) })
	kit.If(m.Option(ice.MSG_USERWEB) == "", func() {
		m.Option(ice.MSG_USERWEB, Domain(ice.Pulse.Cmdv(tcp.HOST, aaa.IP), ice.Pulse.Cmdv(SERVE, tcp.PORT)))
	})
	return kit.MergeURL2(m.Option(ice.MSG_USERWEB), url, arg...)
}
func MergeLink(m Message, url string, arg ...ice.Any) string {
	kit.If(m.Option(log.DEBUG) == ice.TRUE, func() { arg = append(arg, log.DEBUG, ice.TRUE) })
	return kit.MergeURL(strings.Split(MergeURL2(m, url), QS)[0], arg...)
}
func ProcessPodCmd(m *ice.Message, pod, cmd string, arg ...ice.Any) {
	m.ProcessOpen(m.MergePodCmd(pod, cmd, arg...))
}
func ProcessIframe(m *ice.Message, name, link string, arg ...string) {
	ctx.ProcessField(m, CHAT_IFRAME, func() []string {
		return []string{m.Cmdx(CHAT_IFRAME, mdb.CREATE, mdb.TYPE, LINK, mdb.NAME, name, LINK, link)}
	}, arg...)
}
func PushPodCmd(m *ice.Message, cmd string, arg ...string) {
	list := []string{}
	m.Cmds(SPACE, func(value ice.Maps) {
		kit.If(kit.IsIn(value[mdb.TYPE], WORKER), func() { list = append(list, value[mdb.NAME]) })
		// kit.If(kit.IsIn(value[mdb.TYPE], WORKER, SERVER), func() { list = append(list, value[mdb.NAME]) })
	})
	if len(list) == 0 {
		return
	}
	kit.If(m.Length() > 0 && len(m.Appendv(SPACE)) == 0, func() { m.Table(func(value ice.Maps) { m.Push(SPACE, "") }) })
	GoToast(m, "", func(toast func(string, int, int)) []string {
		kit.For(list, func(index int, space string) {
			toast(space, index, len(list))
			m.Cmd(SPACE, space, kit.Dict(ice.MSG_USERPOD, space), kit.Select(m.PrefixKey(), cmd), arg).Table(func(index int, val ice.Maps, head []string) {
				val[SPACE] = kit.Keys(space, val[SPACE])
				m.Push("", val, head)
			})
		})
		return nil
	})
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
	} else if m.Option(ice.MSG_USERPOD) == "" {
		m.Cmd(SPACE, m.Option(ice.MSG_DAEMON), arg, ice.Maps{ice.MSG_OPTION: "", ice.MSG_OPTS: ""})
	} else {
		m.Cmd(SPACE, kit.Keys(m.Option("__target"), m.Option(ice.MSG_DAEMON)), arg, ice.Maps{ice.MSG_OPTION: "", ice.MSG_OPTS: ""})
	}
}
func PushNoticeToast(m *ice.Message, arg ...ice.Any) { PushNotice(m, kit.List("toast", arg)...) }
func PushNoticeGrow(m *ice.Message, arg ...ice.Any)  { PushNotice(m, kit.List("grow", arg)...) }
func PushStream(m *ice.Message) {
	m.Options(cli.CMD_OUTPUT, file.NewWriteCloser(func(buf []byte) { PushNoticeGrow(m, string(buf)) }, nil)).ProcessHold(toastContent(m, ice.SUCCESS))
}
func init() { ice.Info.PushStream = PushStream }
func init() { ice.Info.PushNotice = PushNotice }

func Toast(m *ice.Message, text string, arg ...ice.Any) { // [title [duration [progress]]]
	if len(arg) > 1 {
		switch val := arg[1].(type) {
		case string:
			if value, err := time.ParseDuration(val); err == nil {
				arg[1] = int(value / time.Millisecond)
			}
		}
	}
	if len(arg) == 0 {
		arg = append(arg, m.PrefixKey())
	}
	if len(arg) > 0 {
		arg[0] = kit.Select(m.PrefixKey(), arg[0])
	}
	PushNoticeToast(m, text, arg)
}
func toastContent(m *ice.Message, state string) string {
	return kit.Join([]string{map[string]string{ice.PROCESS: "🕑", ice.FAILURE: "❌", ice.SUCCESS: "✅"}[state], state, m.ActionKey()}, " ")
}
func ToastSuccess(m *ice.Message, arg ...ice.Any) { Toast(m, toastContent(m, ice.SUCCESS), arg...) }
func ToastFailure(m *ice.Message, arg ...ice.Any) { Toast(m, toastContent(m, ice.FAILURE), arg...) }
func ToastProcess(m *ice.Message, arg ...ice.Any) func() {
	kit.If(len(arg) == 0, func() { arg = kit.List("", "-1") })
	kit.If(len(arg) == 1, func() { arg = append(arg, "-1") })
	Toast(m, toastContent(m, ice.PROCESS), arg...)
	return func() { Toast(m, toastContent(m, ice.SUCCESS)) }
}
func GoToast(m *ice.Message, title string, cb func(toast func(string, int, int)) []string) {
	_total := 0
	toast := func(name string, count, total int) {
		kit.If(total == 0, func() { total = 1 })
		Toast(m,
			kit.Format("%s %s/%s", name, strings.TrimSuffix(kit.FmtSize(int64(count)), "B"), strings.TrimSuffix(kit.FmtSize(int64(total)), "B")),
			kit.Format("%s %d%%", kit.Select(m.ActionKey(), title), count*100/total), kit.Select("1000", "30000", count < total), count*100/total,
		)
		_total = total
	}
	if list := cb(toast); len(list) > 0 {
		Toast(m, strings.Join(list, lex.NL), ice.FAILURE, "30s")
	} else {
		toast(ice.SUCCESS, _total, _total)
	}
}
