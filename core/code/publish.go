package code

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _publish_bin_list(m *ice.Message) *ice.Message {
	defer m.SortStrR(mdb.TIME)
	return m.Cmdy(nfs.DIR, nfs.PWD, nfs.DIR_WEB_FIELDS, kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN, nfs.DIR_DEEP, ice.TRUE, nfs.DIR_ROOT, ice.USR_PUBLISH))
}
func _publish_list(m *ice.Message, arg ...string) *ice.Message {
	defer m.SortStrR(mdb.TIME)
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
	m.OptionDefault(ice.MSG_USERNAME, "demo")
	for _, k := range kit.Default(arg, ice.MISC) {
		m.Options(web.DOMAIN, web.UserHost(m), cli.CTX_ENV, kit.Select("", lex.SP+kit.JoinKV(mdb.EQ, lex.SP, cli.CTX_POD, m.Option(ice.MSG_USERPOD)), m.Option(ice.MSG_USERPOD) != ""))
		switch k {
		case INSTALL:
			m.Echo(strings.TrimSpace(nfs.Template(m, kit.Keys(ice.MISC, SH))))
			return
		case ice.BASE:
			m.Option(web.DOMAIN, m.Cmd(web.SPIDE, ice.SHY).Append(web.CLIENT_ORIGIN))
		case ice.CORE:
			m.Option(web.DOMAIN, m.Cmd(web.SPIDE, ice.DEV).Append(web.CLIENT_ORIGIN))
		case nfs.SOURCE, "dev":
			m.Options(nfs.SOURCE, ice.Info.Make.Remote)
		case nfs.BINARY, "app":

		case "wget", "curl":
		case "manual":
			m.Options(nfs.BINARY, "ice.linux.amd64")
		default:
			// _publish_file(m, ice.ICE_BIN)
		}
		if m.Option("format") == "raw" {
			m.Echo(strings.TrimSpace(nfs.Template(m, kit.Keys(k, SH))))
		} else {
			m.EchoScript(strings.TrimSpace(nfs.Template(m, kit.Keys(k, SH))))
		}
	}
}

const PUBLISH = "publish"

func init() {
	web.Index.MergeCommands(ice.Commands{
		web.PP(ice.PUBLISH): {Name: "/publish/", Help: "定制化", Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			web.ShareLocalFile(m, ice.USR_PUBLISH, path.Join(arg...))
		}},
	})
	Index.MergeCommands(ice.Commands{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell", Help: "发布", Actions: ice.MergeActions(ice.Actions{
			ice.VOLCANOS: {Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				_publish_list(m, kit.ExtReg(HTML, CSS, JS)).EchoQRCode(m.Option(ice.MSG_USERWEB))
			}},
			ice.ICEBERGS: {Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				_publish_bin_list(m).Cmdy("", ice.CONTEXTS)
			}},
			ice.INTSHELL: {Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				_publish_list(m, kit.ExtReg(SH, VIM, CONF))
			}},
			ice.CONTEXTS: {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, arg...) }},
			nfs.SOURCE:   {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, nfs.SOURCE) }},
			nfs.BINARY:   {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, nfs.BINARY) }},
			"manual": {Hand: func(m *ice.Message, arg ...string) {
				host := web.UserHost(m)
				args := ""
				if m.Option(ice.MSG_USERPOD) != "" {
					args = "?pod=" + m.Option(ice.MSG_USERPOD)
				}
				m.Cmdy("web.wiki.spark", "shell",
					cli.LINUX, kit.Format(`curl -fSL -O "%s/publish/ice.linux.amd64%s"`, host, args),
					cli.DARWIN, kit.Format(`curl -fSL -O "%s/publish/ice.darwin.amd64%s"`, host, args),
					cli.WINDOWS, kit.Format(`curl -fSL -O "%s/publish/ice.windows.amd64%s"`, host, args),
				)
			}},
			"wget": {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, "wget") }},
			"curl": {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, "curl") }},
			"version": {Hand: func(m *ice.Message, arg ...string) {
				defer m.Echo("<table>").Echo("</table>")
				for _, cpu := range []string{cli.AMD64, cli.X86, cli.ARM} {
					m.Echo("<tr>")
					for _, sys := range []string{cli.LINUX, cli.WINDOWS, cli.DARWIN} {
						m.Echo("<td>")
						if file := fmt.Sprintf("ice.%s.%s", sys, cpu); nfs.Exists(m, "usr/publish/"+file) {
							m.EchoAnchor(file, "/publish/"+file)
						}
						m.Echo("</td>")
					}
					m.Echo("</tr>")
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg[1:], nfs.DIR_CLI_FIELDS) }},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) { _publish_file(m, m.Option(nfs.PATH)) }},
			nfs.TRASH:  {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, path.Join(ice.USR_PUBLISH, m.Option(nfs.PATH))) }},
		}, ctx.ConfAction(mdb.FIELD, nfs.PATH), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_ROOT, ice.USR_PUBLISH); len(arg) == 0 {
				_publish_list(m).Cmdy("", ice.CONTEXTS, "app")
			} else {
				m.Cmdy(nfs.DIR, arg[0], "time,path,size,hash,link,action", ice.OptionFields(mdb.DETAIL))
				web.PushImages(m, web.P(PUBLISH, arg[0]))
			}
		}},
	})
}
