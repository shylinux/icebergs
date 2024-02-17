package web

import (
	"net/url"
	"strings"

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
func PushNoticeToast(m *ice.Message, arg ...ice.Any) {
	PushNotice(m, kit.List(TOAST, arg)...)
}
func PushNoticeRefresh(m *ice.Message, arg ...ice.Any) {
	PushNotice(m, kit.List("refresh", arg)...)
}
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
func init() { ice.Info.PushNotice = PushNotice }
func init() { ice.Info.PushStream = PushStream }
