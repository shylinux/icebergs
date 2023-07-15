package web

import (
	"net/url"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

type Message interface {
	Option(key string, arg ...ice.Any) string
	PrefixKey(...string) string
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
	if m.Option(log.DEBUG) == ice.TRUE {
		arg = append([]ice.Any{log.DEBUG, ice.TRUE}, arg)
	}
	if m.Option(ice.MSG_USERWEB) == "" {
		return kit.MergeURL2(Domain(ice.Pulse.Cmdv(tcp.HOST, aaa.IP), ice.Pulse.Cmdv(SERVE, tcp.PORT)), url, arg...)
	}
	return kit.MergeURL2(m.Option(ice.MSG_USERWEB), url, arg...)
}
func MergeLink(m Message, url string, arg ...ice.Any) string {
	if m.Option(log.DEBUG) == ice.TRUE {
		arg = append(arg, log.DEBUG, ice.TRUE)
	}
	return kit.MergeURL(strings.Split(MergeURL2(m, url), mdb.QS)[0], arg...)
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
	kit.If(m.Length() > 0 && len(m.Appendv(ice.POD)) == 0, func() { m.Table(func(value ice.Maps) { m.Push(ice.POD, m.Option(ice.MSG_USERPOD)) }) })
	m.Cmds(SPACE, func(value ice.Maps) {
		kit.Switch(value[mdb.TYPE], []string{SERVER, WORKER}, func() {
			m.Cmd(SPACE, value[mdb.NAME], kit.Select(m.PrefixKey(), cmd), arg).Table(func(index int, val ice.Maps, head []string) {
				val[ice.POD] = kit.Keys(value[mdb.NAME], val[ice.POD])
				m.Push("", val, head)
			})
		})
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
	m.Options(cli.CMD_OUTPUT, file.NewWriteCloser(func(buf []byte) { PushNoticeGrow(m, string(buf)) }, func() { PushNoticeToast(m, "done") })).ProcessHold()
}

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
func ToastFailure(m *ice.Message, arg ...ice.Any) { Toast(m, ice.FAILURE, arg...) }
func ToastSuccess(m *ice.Message, arg ...ice.Any) { Toast(m, ice.SUCCESS, arg...) }
func ToastProcess(m *ice.Message, arg ...ice.Any) func() {
	kit.If(len(arg) == 0, func() { arg = kit.List("", "-1") })
	kit.If(len(arg) == 1, func() { arg = append(arg, "-1") })
	Toast(m, ice.PROCESS, arg...)
	return func() { Toast(m, ice.SUCCESS) }
}
func GoToast(m *ice.Message, title string, cb func(toast func(string, int, int))) {
	cb(func(name string, count, total int) {
		kit.If(total == 0, func() { total = 1 })
		Toast(m,
			kit.Format("%s %s/%s", name, strings.TrimSuffix(kit.FmtSize(int64(count)), "B"), strings.TrimSuffix(kit.FmtSize(int64(total)), "B")),
			kit.Format("%s %d%%", title, count*100/total), kit.Select("1000", "30000", count < total), count*100/total,
		)
	})
}
