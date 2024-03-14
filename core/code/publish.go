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
	"shylinux.com/x/icebergs/base/tcp"
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
	defer m.Table(func(value ice.Maps) {
		if p := value[nfs.PATH]; strings.Contains(p, "ice.windows.") {
			m.PushDownload(mdb.LINK, "ice.exe", "/publish/"+p)
		} else {
			m.Push(mdb.LINK, kit.MergeURL2(web.UserHost(m), "/publish/"+p))
		}
	})
	return nfs.DirDeepAll(m, ice.USR_PUBLISH, nfs.PWD, nil)
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
	// m.Options(nfs.DIR_ROOT, "")
	m.OptionDefault(web.DOMAIN, tcp.PublishLocalhost(m, web.UserHost(m)))
	m.OptionDefault(cli.CTX_CLI, "temp=$(mktemp); if curl -h &>/dev/null; then curl -o $temp -fsSL $ctx_dev; else wget -O $temp -q $ctx_dev; fi; source $temp")
	m.OptionDefault(cli.CTX_ARG, kit.JoinCmdArgs(aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.LANGUAGE, m.Option(ice.MSG_LANGUAGE)))
	for _, k := range kit.Default(arg, ice.MISC) {
		env := []string{}
		switch k {
		case nfs.SOURCE, ice.DEV:
			if ice.Info.NodeType == web.SERVER {
				m.OptionDefault(nfs.SOURCE, ice.Info.Make.Remote)
			} else if m.Option(ice.MSG_USERPOD) == "" {
				m.OptionDefault(nfs.SOURCE, web.AdminCmd(m, cli.RUNTIME, "make.remote").Result())
			} else {
				m.OptionDefault(nfs.SOURCE, web.AdminCmd(m, web.SPACE, kit.KeyBase(m.Option(ice.MSG_USERPOD)), cli.RUNTIME, "make.remote").Result())
			}
			env = append(env, cli.CTX_REPOS, m.Option(nfs.SOURCE))
			fallthrough
		case nfs.BINARY, ice.APP:
			if host := web.HostPort(m, web.AdminCmd(m, tcp.HOST).Append(aaa.IP), web.AdminCmd(m, web.SERVE).Append(tcp.PORT)); m.Option(web.DOMAIN) != host {
				env = append(env, cli.CTX_DEV_IP, host)
			}
			// if ice.Info.NodeType == web.WORKER && m.Option(ice.MSG_USERPOD) != "" {
			if m.Option(ice.MSG_USERPOD) != "" {
				env = append(env, cli.CTX_POD, m.Option(ice.MSG_USERPOD))
			} else if name := web.AdminCmd(m, cli.RUNTIME, "boot.pathname").Result(); !kit.IsIn(name, path.Base(m.Option(nfs.SOURCE)), ice.CONTEXTS) {
				env = append(env, "ctx_name", name)
			}
		case cli.CURL, cli.WGET:
		}
		kit.If(len(env) > 0, func() { m.Options(cli.CTX_ENV, lex.SP+kit.JoinKV(mdb.EQ, lex.SP, env...)) })
		if template := strings.TrimSpace(nfs.Template(m, kit.Keys(k, SH))); m.Option("format") == "raw" {
			m.Echo(template)
		} else {
			m.EchoScript(template)
		}
	}
}

const PUBLISH = "publish"

func init() {
	web.Index.MergeCommands(ice.Commands{
		web.PP(ice.PUBLISH): {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
			web.ShareLocalFile(m, ice.USR_PUBLISH, path.Join(arg...))
			web.Count(m, PUBLISH, path.Join(arg...))
		}},
	})
	Index.MergeCommands(ice.Commands{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell", Help: "发布", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			ice.VOLCANOS: {Help: "火山架", Hand: func(m *ice.Message, arg ...string) { _publish_list(m, kit.ExtReg(HTML, CSS, JS)) }},
			ice.ICEBERGS: {Help: "冰山架", Hand: func(m *ice.Message, arg ...string) { _publish_bin_list(m).Cmdy("", ice.CONTEXTS) }},
			ice.INTSHELL: {Help: "神农架", Hand: func(m *ice.Message, arg ...string) { _publish_list(m, kit.ExtReg(SH, VIM, CONF)) }},
			ice.CONTEXTS: {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, arg...) }},
			nfs.SOURCE:   {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, nfs.SOURCE) }},
			nfs.BINARY:   {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, nfs.BINARY) }},
			cli.CURL:     {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, cli.CURL) }},
			cli.WGET:     {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, cli.WGET) }},
			"manual": {Hand: func(m *ice.Message, arg ...string) {
				host, args := web.UserHost(m), ""
				kit.If(m.Option(ice.MSG_USERPOD), func(p string) { args = "?pod=" + p })
				m.Cmdy("web.wiki.spark", "shell",
					cli.LINUX, kit.Format(`curl -fSL -O "%s/publish/ice.linux.amd64%s"`, host, args),
					cli.DARWIN, kit.Format(`curl -fSL -O "%s/publish/ice.darwin.amd64%s"`, host, args),
					cli.WINDOWS, kit.Format(`curl -fSL -O "%s/publish/ice.windows.amd64%s"`, host, args),
				)
			}},
			nfs.VERSION: {Hand: func(m *ice.Message, arg ...string) {
				defer m.Echo("<table>").Echo("</table>")
				kit.For([]string{cli.AMD64, cli.X86, cli.ARM}, func(cpu string) {
					defer m.Echo("<tr>").Echo("</tr>")
					kit.For([]string{cli.LINUX, cli.WINDOWS, cli.DARWIN}, func(sys string) {
						defer m.Echo("<td>").Echo("</td>")
						if file := fmt.Sprintf("ice.%s.%s", sys, cpu); nfs.Exists(m, ice.USR_PUBLISH+file) {
							m.EchoAnchor(file, "/publish/"+file)
						}
					})
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg[1:], nfs.DIR_CLI_FIELDS) }},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) { _publish_file(m, m.Option(nfs.PATH)) }},
			nfs.TRASH:  {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, path.Join(ice.USR_PUBLISH, m.Option(nfs.PATH))) }},
		}, ctx.ConfAction(mdb.FIELD, nfs.PATH)), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_ROOT, ice.USR_PUBLISH); len(arg) == 0 {
				_publish_list(m).Cmdy("", ice.CONTEXTS, ice.APP)
			} else {
				m.Cmdy(nfs.DIR, arg[0], "time,path,size,hash,link,action", ice.OptionFields(mdb.DETAIL))
				web.PushImages(m, web.P(PUBLISH, arg[0]))
			}
			m.PushAction(nfs.TRASH)
		}},
	})
}
