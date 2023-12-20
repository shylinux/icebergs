package aaa

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _sess_create(m *ice.Message, username string, arg ...string) {
	if msg := m.Cmd(USER, username); msg.Length() > 0 {
		mdb.HashCreate(m, msg.AppendSimple(USERNICK, USERNAME, USERROLE), arg)
	} else {
		mdb.HashCreate(m, m.OptionSimple(USERNICK, USERNAME, USERROLE), arg)
	}
}
func _sess_check(m *ice.Message, sessid string) {
	if val := mdb.HashSelectDetails(m, sessid, func(value ice.Map) bool { return !m.WarnTimeNotValid(value[mdb.TIME], sessid) }); len(val) > 0 {
		SessAuth(m, val)
	}
}

const (
	UA = "ua"
	IP = "ip"
)
const (
	CHECK  = "check"
	LOGIN  = "login"
	LOGOUT = "logout"
)
const SESS = "sess"

func init() {
	Index.MergeCommands(ice.Commands{
		SESS: {Help: "会话", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create username*", Hand: func(m *ice.Message, arg ...string) {
				_sess_create(m, m.Option(USERNAME), UA, m.Option(ice.MSG_USERUA), IP, m.Option(ice.MSG_USERIP))
			}},
			CHECK: {Name: "check sessid*", Hand: func(m *ice.Message, arg ...string) { _sess_check(m, m.Option(ice.MSG_SESSID)) }},
		}, mdb.ImportantHashAction(mdb.EXPIRE, mdb.MONTH, mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,usernick,username,userrole,ip,ua"))},
	})
}

func SessCreate(m *ice.Message, username string) string {
	return m.Option(ice.MSG_SESSID, m.Cmdx(SESS, mdb.CREATE, username))
}
func SessCheck(m *ice.Message, sessid string) bool {
	m.Options(ice.MSG_USERNICK, "", ice.MSG_USERNAME, "", ice.MSG_USERROLE, VOID, ice.MSG_CHECKER, logs.FileLine(-1))
	return sessid != "" && m.Cmdy(SESS, CHECK, sessid, logs.FileLineMeta(-1)).Option(ice.MSG_USERNAME) != ""
}
func SessAuth(m *ice.Message, value ice.Any, arg ...string) *ice.Message {
	language := kit.Select(m.Option(ice.MSG_LANGUAGE), kit.Format(kit.Value(value, LANGUAGE)))
	kit.If(language == "", func() {
		kit.If(kit.Format(kit.Value(value, USERNAME)), func(p string) { language = m.Cmdv(USER, p, LANGUAGE) })
	})
	kit.If(language == "", func() { language = kit.Select("", "zh-cn", strings.Contains(m.Option(ice.MSG_USERUA), "zh_CN")) })
	kit.If(language == "" && m.R != nil, func() { language = kit.Select("", kit.Split(m.R.Header.Get(html.AcceptLanguage), ",;"), 0) })
	kit.If(language == "", func() { language = ice.Info.Lang })
	language = strings.ReplaceAll(strings.ToLower(kit.Select("", kit.Split(language, " ."), 0)), "_", "-")
	return m.Auth(
		USERROLE, m.Option(ice.MSG_USERROLE, kit.Format(kit.Value(value, USERROLE))),
		USERNAME, m.Option(ice.MSG_USERNAME, kit.Format(kit.Value(value, USERNAME))),
		USERNICK, m.Option(ice.MSG_USERNICK, kit.Format(kit.Value(value, USERNICK))),
		LANGUAGE, m.Option(ice.MSG_LANGUAGE, language), arg,
		logs.FileLineMeta(kit.Select(logs.FileLine(-1), m.Option("aaa.checker"))),
	)
}
func SessLogout(m *ice.Message, arg ...string) {
	kit.If(m.Option(ice.MSG_SESSID), func(sessid string) { m.Cmd(SESS, mdb.REMOVE, mdb.HASH, sessid) })
}
