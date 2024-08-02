package web

import (
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	TOAST_INIT = "init"
	TOAST_DONE = "done"
)
const TOAST = "toast"

func init() {
	Index.MergeCommands(ice.Commands{
		TOAST: {Help: "进度条", Actions: ice.MergeActions(ice.Actions{
			ctx.PREVIEW: &ice.Action{Hand: func(m *ice.Message, arg ...string) { ProcessHashPodCmd(m, arg...) }},
			mdb.PRUNES:  &ice.Action{Name: "prunes status=done"},
		}, mdb.ClearOnExitHashAction(), mdb.StatusHashAction(html.CHECKBOX, ice.TRUE,
			mdb.FIELD, "time,hash,type,name,text,cost,status,index,icons,agent,system,ip,ua",
		)), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(ctx.PREVIEW, mdb.REMOVE).Action(mdb.PRUNES)
			m.Sort("status,cost", []string{"", TOAST_INIT, TOAST_DONE}, func(value string) int { return -int(kit.Duration(value)) })
		}},
	})
}
func toastCreate(m *ice.Message, arg ...string) (string, time.Time) {
	return m.Cmdx(TOAST, mdb.CREATE, mdb.TYPE, kit.FuncName(2), mdb.NAME, toastTitle(m), mdb.STATUS, TOAST_INIT, ctx.INDEX, m.ShortKey(), ParseUA(m), arg, ice.OptionSilent()), time.Now()
}
func toastUpdate(m *ice.Message, h string, begin time.Time, arg ...string) string {
	cost := kit.FmtDuration(time.Now().Sub(begin))
	m.Cmd(TOAST, mdb.MODIFY, mdb.HASH, h, cli.COST, cost, arg, ice.OptionSilent())
	return cost
}

var Icons = map[string]string{ice.PROCESS: "🕑", ice.FAILURE: "❌", ice.SUCCESS: "✅"}

func toastTitle(m *ice.Message) string {
	return kit.GetValid(
		func() string { return m.Option(ice.MSG_TITLE) },
		func() string {
			return kit.Keys(m.Option(ice.MSG_USERPOD0), m.Option(ice.MSG_USERPOD), ctx.ShortCmd(m.ShortKey()))
		},
	)
}
func toastContent(m *ice.Message, state string, arg ...ice.Any) string {
	if len(arg) == 0 {
		return kit.JoinWord(kit.Simple(Icons[state], m.Trans(kit.Select(ice.LIST, m.ActionKey()), ""), m.Trans(state, ""))...)
	} else {
		return kit.JoinWord(kit.Simple(Icons[state], arg)...)
	}
}
func ToastSuccess(m *ice.Message, arg ...ice.Any) {
	Toast(m, toastContent(m, ice.SUCCESS, arg...), "", cli.TIME_1s)
}
func ToastFailure(m *ice.Message, arg ...ice.Any) {
	Toast(m, toastContent(m, ice.FAILURE, arg...), "", m.Option(ice.TOAST_DURATION, "-1")).Sleep(cli.TIME_3s)
}
func ToastProcess(m *ice.Message, arg ...ice.Any) func(...ice.Any) {
	text := toastContent(m, ice.PROCESS, arg...)
	h, begin := toastCreate(m, mdb.TEXT, text)
	Toast(m, text, "", "-1", "", h)
	Count(m, kit.FuncName(1), toastTitle(m), text)
	return func(_arg ...ice.Any) {
		if m.IsErr() {
			return
		}
		kit.If(len(_arg) == 0, func() { _arg = arg })
		text := toastContent(m, ice.SUCCESS, _arg...)
		toastUpdate(m, h, begin, mdb.TEXT, text, mdb.STATUS, TOAST_DONE)
		Toast(m, text, "", cli.TIME_3s, "", h)
	}
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
	h, begin := toastCreate(m)
	icon, _total := Icons[ice.PROCESS], 1
	toast := func(name string, count, total int) {
		kit.If(total == 0, func() { total = 1 })
		text := kit.JoinWord(icon, kit.Select(ice.LIST, m.ActionKey()), name, strings.ReplaceAll(kit.FmtSize(count, total), "B", ""))
		cost := toastUpdate(m, h, begin, mdb.TEXT, text, mdb.STATUS, kit.Select("", TOAST_DONE, count == total))
		Toast(m, kit.JoinWord(text, cost), "", m.OptionDefault(ice.TOAST_DURATION, "-1"), count*100/total, h)
		_total = total
	}
	if list := cb(toast); len(list) > 0 {
		icon = Icons[ice.FAILURE]
		m.Option(ice.TOAST_DURATION, cli.TIME_30s)
		toast(kit.JoinWord(list...), len(list), _total)
		m.Sleep(m.Option(ice.TOAST_DURATION))
	} else {
		icon = Icons[ice.SUCCESS]
		m.Option(ice.TOAST_DURATION, cli.TIME_1s)
		toast(ice.SUCCESS, _total, _total)
	}
	Count(m, kit.FuncName(1), toastTitle(m), kit.FmtDuration(time.Now().Sub(begin)))
	return m
}
func Toast(m *ice.Message, text string, arg ...ice.Any) *ice.Message { // [title [duration [progress [hash]]]]
	if len(arg) > 1 {
		switch val := arg[1].(type) {
		case string:
			if value, err := time.ParseDuration(val); err == nil {
				arg[1] = int(value / time.Millisecond)
			}
		}
	}
	kit.If(len(arg) == 0, func() { arg = append(arg, "") })
	kit.If(len(arg) > 0 && arg[0] == "", func() { arg[0] = toastTitle(m) })
	PushNoticeToast(m, text, arg)
	return m
}
func PushNoticeGrowXterm(m *ice.Message, title string, cmd ...ice.Any) {
	PushCmdStream(m, title).Cmd(cli.SYSTEM, cmd)
}
func PushCmdStream(m *ice.Message, title string) *ice.Message {
	m.Options(ctx.DISPLAY, html.PLUGIN_XTERM, cli.CMD_OUTPUT, nfs.NewWriteCloser(func(buf []byte) (int, error) {
		PushNoticeGrow(m.Options(ice.MSG_TITLE, title, ice.MSG_COUNT, "0", ice.LOG_DEBUG, ice.FALSE, ice.LOG_DISABLE, ice.TRUE), strings.ReplaceAll(string(buf), lex.NL, "\r\n"))
		return len(buf), nil
	}, nil))
	return m
}
