package web

import (
	"net/url"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

func Upload(m *ice.Message) []string { // hash name size
	up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
	if len(up) < 2 {
		msg := m.Cmdy(CACHE, UPLOAD)
		up = kit.Simple(msg.Append(mdb.HASH), msg.Append(mdb.NAME), msg.Append(nfs.SIZE))
	}
	return up
}
func PushNotice(m *ice.Message, arg ...ice.Any) {
	m.Optionv(ice.MSG_OPTS, m.Optionv(ice.MSG_OPTION))
	if m.Option(ice.MSG_USERPOD) == "" {
		m.Cmd(SPACE, m.Option(ice.MSG_DAEMON), arg)
	} else {
		opts := kit.Dict(ice.POD, m.Option(ice.MSG_DAEMON), "cmds", kit.Simple(arg...))
		for _, k := range kit.Simple(m.Optionv(ice.MSG_OPTS)) {
			opts[k] = m.Option(k)
		}
		m.Cmd("web.spide", ice.OPS, MergeURL2(m, SHARE_TOAST), kit.Format(opts))
	}
}
func PushNoticeGrow(m *ice.Message, arg ...ice.Any) {
	PushNotice(m, kit.List("grow", arg)...)
}
func PushNoticeToast(m *ice.Message, arg ...ice.Any) {
	PushNotice(m, kit.List("toast", arg)...)
}
func PushNoticeRefresh(m *ice.Message, arg ...ice.Any) {
	PushNotice(m, kit.List("refresh")...)
}

func ToastProcess(m *ice.Message, arg ...ice.Any) func() {
	if len(arg) == 0 {
		arg = kit.List("", "-1")
	}
	if len(arg) == 1 {
		arg = append(arg, "-1")
	}
	Toast(m, ice.PROCESS, arg...)
	return func() { Toast(m, ice.SUCCESS) }
}
func ToastRestart(m *ice.Message, arg ...ice.Any) { Toast(m, gdb.RESTART, arg...) }
func ToastFailure(m *ice.Message, arg ...ice.Any) { Toast(m, ice.FAILURE, arg...) }
func ToastSuccess(m *ice.Message, arg ...ice.Any) { Toast(m, ice.SUCCESS, arg...) }
func Toast(m *ice.Message, text string, arg ...ice.Any) { // [title [duration [progress]]]
	if len(arg) > 1 {
		switch val := arg[1].(type) {
		case string:
			if value, err := time.ParseDuration(val); err == nil {
				arg[1] = int(value / time.Millisecond)
			}
		}
	}

	PushNoticeToast(m, "", text, arg)
}
func Toast3s(m *ice.Message, text string, arg ...ice.Any) {
	Toast(m, text, kit.List(kit.Select("", arg, 0), kit.Select("3s", arg, 1))...)
}
func Toast30s(m *ice.Message, text string, arg ...ice.Any) {
	Toast(m, text, kit.List(kit.Select("", arg, 0), kit.Select("30s", arg, 1))...)
}
func GoToast(m *ice.Message, title string, cb func(toast func(string, int, int))) {
	m.Go(func() {
		cb(func(name string, count, total int) {
			Toast(m,
				kit.Format("%s %s/%s", name, strings.TrimSuffix(kit.FmtSize(int64(count)), "B"), strings.TrimSuffix(kit.FmtSize(int64(total)), "B")),
				kit.Format("%s %d%%", title, count*100/total),
				kit.Select("3000", "30000", count < total),
				count*100/total,
			)
		})
	})
}
func PushStream(m *ice.Message) {
	m.Option(cli.CMD_OUTPUT, file.NewWriteCloser(func(buf []byte) (int, error) {
		PushNoticeGrow(m, string(buf))
		return len(buf), nil
	}, func() error { PushNoticeToast(m, "done"); return nil }))
}
func PushPodCmd(m *ice.Message, cmd string, arg ...string) {
	if m.Length() > 0 && len(m.Appendv(ice.POD)) == 0 {
		m.Tables(func(value ice.Maps) { m.Push(ice.POD, m.Option(ice.MSG_USERPOD)) })
	}

	m.Cmd(SPACE, ice.OptionFields(mdb.TYPE, mdb.NAME)).Tables(func(value ice.Maps) {
		switch value[mdb.TYPE] {
		case SERVER, WORKER:
			if value[mdb.NAME] == ice.Info.HostName {
				break
			}
			m.Cmd(SPACE, value[mdb.NAME], m.Prefix(cmd), arg).Table(func(index int, val ice.Maps, head []string) {
				val[ice.POD] = kit.Keys(value[mdb.NAME], val[ice.POD])
				m.Push("", val, head)
			})
		}
	})
}

func OptionAgentIs(m *ice.Message, arg ...string) bool {
	for _, k := range arg {
		if strings.HasPrefix(m.Option(ice.MSG_USERUA), k) {
			return true
		}
	}
	return false
}
func OptionUserWeb(m *ice.Message) *url.URL {
	return kit.ParseURL(m.Option(ice.MSG_USERWEB))
}
func MergeURL2(m *ice.Message, url string, arg ...ice.Any) string {
	return kit.MergeURL2(m.Option(ice.MSG_USERWEB), url, arg...)
}
func MergeLink(m *ice.Message, url string, arg ...ice.Any) string {
	return strings.Split(MergeURL2(m, url, arg...), "?")[0]
}
func MergePod(m *ice.Message, pod string, arg ...ice.Any) string {
	return kit.MergePOD(kit.Select(ice.Info.Domain, m.Option(ice.MSG_USERWEB)), pod, arg...)
}