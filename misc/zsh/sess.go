package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"io/ioutil"
	"strings"
)

const (
	SID = "sid"
	ARG = "arg"
	SUB = "sub"
	PWD = "pwd"
	PID = "pid"
)
const SESS = "sess"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SESS: {Name: SESS, Help: "会话流", Value: kit.Data(
				kit.MDB_FIELD, "time,hash,status,username,hostname,pid,pwd",
				"contexts", `export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp`,
			)},
		},
		Commands: map[string]*ice.Command{
			SESS: {Name: "sess hash auto 清理", Help: "会话流", Action: map[string]*ice.Action{
				"contexts": {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, strings.Split(u.Host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(u.Host, ":"), 1)))

					if buf, err := kit.Render(m.Conf(m.Prefix(SESS), "meta.contexts"), m); m.Assert(err) {
						m.Cmdy("web.wiki.spark", "shell", string(buf))
					}
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, m.Prefix(SESS), "", mdb.HASH, kit.MDB_STATUS, "logout")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select(m.Conf(m.Prefix(SESS), kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, arg)
			}},
			"/sess": {Name: "/sess", Help: "会话", Action: map[string]*ice.Action{
				"logout": {Name: "logout", Help: "退出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, m.Option(SID), kit.MDB_STATUS, "logout")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, m.Conf(m.Prefix(SESS), kit.META_FIELD))
				if strings.TrimSpace(m.Option(SID)) == "" {
					m.Option(SID, m.Cmdx(mdb.INSERT, m.Prefix(SESS), "", mdb.HASH, kit.MDB_STATUS, "login",
						aaa.USERNAME, m.Option(aaa.USERNAME), aaa.HOSTNAME, m.Option(aaa.HOSTNAME), PID, m.Option(PID), PWD, m.Option(PWD)))
				} else {
					m.Cmdy(mdb.MODIFY, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, m.Option(SID), kit.MDB_STATUS, "login")
				}
				m.Echo(m.Option(SID))
			}},
			web.LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if f, _, e := m.R.FormFile(SUB); e == nil {
					defer f.Close()
					// 文件参数
					if b, e := ioutil.ReadAll(f); e == nil {
						m.Option(SUB, string(b))
					}
				}

				m.Option(mdb.FIELDS, m.Conf(m.Prefix(SESS), kit.META_FIELD))
				if strings.TrimSpace(m.Option(SID)) != "" {
					msg := m.Cmd(mdb.SELECT, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, strings.TrimSpace(m.Option(SID)))
					if m.Option(SID, msg.Append(kit.MDB_HASH)) != "" {
						m.Option(aaa.USERNAME, msg.Append(aaa.USERNAME))
						m.Option(aaa.HOSTNAME, msg.Append(aaa.HOSTNAME))
					}
				}
				m.Render(ice.RENDER_RESULT)
			}},
		},
	}, nil)
}
