package code

import (
	"os"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _publish_bin_list(m *ice.Message) *ice.Message {
	defer m.SortTimeR(mdb.TIME)
	m.Option(cli.CMD_DIR, ice.USR_PUBLISH)
	for _, ls := range strings.Split(cli.SystemCmds(m, "ls |xargs file |grep executable"), ice.NL) {
		if file := strings.TrimSpace(strings.Split(ls, ice.DF)[0]); file != "" {
			if s, e := nfs.StatFile(m, path.Join(ice.USR_PUBLISH, file)); e == nil {
				m.Push(mdb.TIME, s.ModTime()).Push(nfs.SIZE, kit.FmtSize(s.Size())).Push(nfs.PATH, file)
				m.PushDownload(mdb.LINK, file, path.Join(ice.USR_PUBLISH, file)).PushButton(nfs.TRASH)
			}
		}
	}
	return m
}
func _publish_list(m *ice.Message, arg ...string) *ice.Message {
	defer m.SortTimeR(mdb.TIME)
	m.Option(nfs.DIR_REG, kit.Select("", arg, 0))
	return nfs.DirDeepAll(m, ice.USR_PUBLISH, nfs.PWD, nil, nfs.DIR_WEB_FIELDS)
}
func _publish_file(m *ice.Message, file string, arg ...string) string {
	if strings.HasSuffix(file, ice.ICE_BIN) {
		file, arg = cli.SystemFind(m, os.Args[0]), kit.Simple(kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH))
	} else if s, e := nfs.StatFile(m, file); m.Assert(e) && s.IsDir() {
		file = m.Cmdx(nfs.TAR, mdb.IMPORT, path.Base(file), file)
		defer func() { nfs.Remove(m, file) }()
	}
	return m.Cmdx(nfs.LINK, path.Join(ice.USR_PUBLISH, kit.Select(path.Base(file), arg, 0)), file)
}
func _publish_contexts(m *ice.Message, arg ...string) {
	m.Option(nfs.DIR_ROOT, "")
	for _, k := range kit.Default(arg, ice.MISC) {
		m.Options(web.DOMAIN, m.Option(ice.MSG_USERHOST), cli.CTX_ENV, kit.Select("", ice.SP+kit.JoinKV(ice.EQ, ice.SP, cli.CTX_POD, m.Option(ice.MSG_USERPOD)), m.Option(ice.MSG_USERPOD) != ""))
		switch k {
		case INSTALL:
			m.Echo(strings.TrimSpace(nfs.Template(m, kit.Keys(ice.MISC, SH))))
			return
		case ice.BASE:
			m.Option(web.DOMAIN, m.Cmd(web.SPIDE, ice.SHY).Append(web.CLIENT_ORIGIN))
		case ice.CORE:
			m.Option(web.DOMAIN, m.Cmd(web.SPIDE, ice.DEV).Append(web.CLIENT_ORIGIN))
		default:
			_publish_file(m, ice.ICE_BIN)
		}
		m.EchoScript(strings.TrimSpace(nfs.Template(m, kit.Keys(k, SH))))
	}
}

const PUBLISH = "publish"

func init() {
	Index.MergeCommands(ice.Commands{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell", Help: "发布", Actions: ice.MergeActions(ice.Actions{
			ice.VOLCANOS: {Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				_publish_list(m, kit.ExtReg(HTML, CSS, JS)).Cmdy("", ice.CONTEXTS, ice.MISC).Echo(ice.NL).EchoQRCode(m.Option(ice.MSG_USERWEB))
			}},
			ice.ICEBERGS: {Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				_publish_bin_list(m).Cmdy("", ice.CONTEXTS, ice.CORE)
			}},
			ice.INTSHELL: {Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				_publish_list(m, kit.ExtReg(SH, VIM, CONF)).Cmdy("", ice.CONTEXTS, ice.BASE)
			}},
			ice.CONTEXTS: {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, arg...) }},
			mdb.INPUTS:   {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg[1:], nfs.DIR_CLI_FIELDS) }},
			mdb.CREATE:   {Hand: func(m *ice.Message, arg ...string) { _publish_file(m, m.Option(nfs.PATH)) }},
			nfs.TRASH:    {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, path.Join(ice.USR_PUBLISH, m.Option(nfs.PATH))) }},
		}, ctx.ConfAction(mdb.FIELD, nfs.PATH), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_ROOT, ice.USR_PUBLISH); len(arg) == 0 {
				_publish_list(m)
			} else {
				m.Cmdy(nfs.DIR, arg[0], "time,size,path,hash,link,action", ice.OptionFields(mdb.DETAIL))
			}
		}},
	})
}
