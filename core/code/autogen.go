package code

import (
	"bytes"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _autogen_list(m *ice.Message) string {
	return m.OptionDefault(mdb.LIST, ice.Maps{
		"Hash": "hash auto",
		"Zone": "zone id auto",
		"Data": "path auto",
		"Lang": "path auto",
		"Code": "port path auto start build download",
	}[m.Option(mdb.TYPE)])
}
func _autogen_source(m *ice.Message, main, file string) {
	m.Cmd(nfs.DEFS, main, m.Cmdx(nfs.CAT, ice.SRC_MAIN_SHY))
	m.Cmd(nfs.PUSH, main, lex.NL+ssh.SOURCE+lex.SP+strings.TrimPrefix(file, nfs.SRC)+lex.NL)
	ReposAddFile(m, "", ice.SRC_MAIN_SHY)
}
func _autogen_script(m *ice.Message, file string) {
	m.Cmd(nfs.DEFS, file, nfs.Template(m, DEMO_SHY))
	ReposAddFile(m, "", file)
}
func _autogen_module(m *ice.Message, file string) {
	m.Cmd(nfs.DEFS, file, nfs.Template(m, DEMO_GO))
	ReposAddFile(m, "", file)
}
func _autogen_defs(m *ice.Message, arg ...string) {
	kit.For(arg, func(p string) { m.Cmd(nfs.DEFS, p, m.Cmdx(nfs.CAT, p)); ReposAddFile(m, "", p) })
}
func _autogen_import(m *ice.Message, main string, ctx string, mod string) string {
	m.Cmd(nfs.DEFS, ice.ETC_MISS_SH, m.Template("miss.sh"))
	// _autogen_defs(m, ice.README_MD, ice.MAKEFILE, ice.LICENSE)
	_autogen_defs(m, ice.SRC_MAIN_GO, ice.ETC_MISS_SH, ice.README_MD, ice.MAKEFILE, ice.LICENSE)
	begin, done, list := false, false, []string{}
	imports := kit.Format(`_ "%s/src/%s"`, mod, ctx)
	m.Cmd(nfs.CAT, main, func(line string, index int) {
		if strings.HasSuffix(line, imports) {
			done = true
		} else if begin && !done && strings.HasPrefix(line, ")") {
			done, list = true, append(list, "", kit.Format(`	%s`, imports))
		}
		if list = append(list, line); done {
			return
		}
		if strings.HasPrefix(line, "import (") {
			begin = true
		} else if strings.HasPrefix(line, IMPORT) {
			done, list = true, append(list, kit.Format(`import %s`, imports))
		}
	})
	m.Cmd(nfs.SAVE, main, kit.Join(list, lex.NL))
	GoImports(m, main)
	ReposAddFile(m, "", main)
	return path.Join(mod, "src", ctx)
}
func _autogen_version(m *ice.Message) string {
	if mod := _autogen_mod(m, ice.GO_MOD); !nfs.Exists(m, ".git") {
		m.Cmd(REPOS, INIT, nfs.ORIGIN, strings.Split(kit.MergeURL2(kit.Select(m.Option(ice.MSG_USERWEB), ice.Info.Make.Remote), web.X(path.Base(mod))), mdb.QS)[0], mdb.NAME, path.Base(mod), nfs.PATH, nfs.PWD)
		defer m.Cmd(REPOS, ADD, kit.Dict(nfs.REPOS, path.Base(mod), nfs.FILE, ice.GO_MOD))
		defer m.Cmd(REPOS, ADD, kit.Dict(nfs.REPOS, path.Base(mod), nfs.FILE, nfs.SRC))
	}
	m.Cmd(nfs.DEFS, ".gitignore", nfs.Template(m, "gitignore"))
	m.Cmd(nfs.DEFS, ice.SRC_BINPACK_USR_GO, nfs.Template(m, "binpack.go"))
	m.Cmd(nfs.DEFS, ice.SRC_BINPACK_GO, nfs.Template(m, "binpack.go"))
	m.Cmd(nfs.SAVE, ice.SRC_VERSION_GO, kit.Format(nfs.Template(m, "version.go"), _autogen_gits(m)))
	m.Cmdy(nfs.DIR, ice.SRC_BINPACK_USR_GO)
	m.Cmdy(nfs.DIR, ice.SRC_BINPACK_GO)
	m.Cmdy(nfs.DIR, ice.SRC_VERSION_GO)
	m.Cmdy(nfs.DIR, ice.SRC_MAIN_GO)
	GoFmt(m, ice.SRC_VERSION_GO)
	return ice.SRC_VERSION_GO
}
func _autogen_gits(m *ice.Message, arg ...string) string {
	res := []string{}
	kit.For(_autogen_git(m, arg...), func(k, v string) {
		res = append(res, kit.Format(`		%s: "%s",`, kit.Capital(k), strings.TrimSpace(v)))
	})
	return kit.Join(res, lex.NL)
}
func _autogen_git(m *ice.Message, arg ...string) ice.Map {
	msg := m.Cmd(REPOS, REMOTE)
	m.Cmd(MOD, mdb.RENDER, MOD, "go.mod", "./").Table(func(value ice.Maps) {
		if value["require"] == "shylinux.com/x/ice" {
			msg.Append("release", value["version"])
		}
		if value["require"] == "shylinux.com/x/icebergs" {
			msg.Append("icebergs", value["version"])
		}
	})
	return kit.Dict(arg, aaa.USERNAME, m.Option(ice.MSG_USERNAME), tcp.HOSTNAME, ice.Info.Hostname, nfs.PATH, kit.Path("")+nfs.PS, mdb.TIME, m.Time(),
		GIT, GitVersion(m), GO, GoVersion(m), nfs.MODULE, _autogen_mod(m, ice.GO_MOD),
		msg.AppendSimple("remote,branch,version,forword,author,email,hash,when,message,release,icebergs"),
		web.DOMAIN, m.Spawn(kit.Dict(ice.MSG_USERWEB, web.UserHost(m), ice.MSG_USERPOD, m.Option(ice.MSG_USERPOD))).MergePod(""),
		cli.SYSTEM, ice.Info.System,
	)
}
func _autogen_mod(m *ice.Message, file string) (mod string) {
	host := kit.Select(ice.Info.Make.Remote, m.Cmdx(REPOS, REMOTE_URL))
	if host == "" {
		host = m.Option(ice.MSG_USERHOST)
		if host == "" || !strings.HasPrefix(host, ice.HTTP) {
			host = path.Base(kit.Path(""))
		} else {
			host = strings.Split(host, "://")[1]
			host = path.Join(host, "x", path.Base(kit.Path("")))
		}
	} else {
		host = strings.Split(host, "://")[1]
	}
	kit.If(strings.Contains(host, ":"), func() { host = path.Base(host) })
	m.Cmd(nfs.DEFS, file, kit.Format(nfs.Template(m, ice.GO_MOD), host))
	// ReposAddFile(m, "", ice.GO_MOD)
	m.Cmd(nfs.CAT, file, func(line string) {
		kit.If(strings.HasPrefix(line, nfs.MODULE), func() { mod = kit.Split(line, lex.SP)[1] })
	})
	return
}

const AUTOGEN = "autogen"

func init() {
	Index.MergeCommands(ice.Commands{
		AUTOGEN: {Name: "autogen path auto script module devpack webpack binpack version", Help: "生成", Actions: ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case cli.MAIN:
					m.Cmdy(nfs.DIR, nfs.PWD, nfs.PATH, kit.Dict(nfs.DIR_ROOT, ice.SRC, nfs.DIR_REG, kit.ExtReg(GO)))
				case mdb.NAME:
					m.Cmdy(nfs.DIR, nfs.PWD, mdb.NAME, kit.Dict(nfs.DIR_ROOT, ice.SRC, nfs.DIR_TYPE, nfs.DIR))
				case mdb.ZONE:
					m.Push(arg[0], path.Dir(m.Option(nfs.FILE)))
					m.Cmd(nfs.DIR, nfs.PWD, nfs.PATH, kit.Dict(nfs.DIR_ROOT, ice.SRC, nfs.DIR_TYPE, nfs.DIR, nfs.DIR_DEEP, ice.TRUE)).Table(func(value ice.Maps) {
						if !kit.HasPrefix(value[nfs.PATH], "template/", "document/") {
							m.Push(arg[0], path.Join(value[nfs.PATH]))
						}
					})
				case mdb.KEY:
					kit.For([]string{"code", "wiki", "chat", "team", "mall"}, func(p string) {
						m.Push(arg[0], kit.Keys("web", p, m.Option(mdb.ZONE), m.Option(mdb.NAME)))
					})
				case "top":
					list := map[string]bool{}
					m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
						if !list[c.Prefix()] {
							m.Push(arg[0], c.Prefix())
							list[c.Prefix()] = true
						}
					})
				}
			}},
			nfs.SCRIPT: {Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, ice.ETC_MISS_SH, m.Cmdx(nfs.CAT, ice.ETC_MISS_SH))
				m.Cmdy(nfs.DIR, ice.ETC_MISS_SH).Cmdy(nfs.CAT, ice.ETC_MISS_SH)
			}},
			nfs.REPOS: {Help: "仓库", Hand: func(m *ice.Message, arg ...string) {
				_autogen_defs(m, ice.SRC_MAIN_GO, ice.SRC_MAIN_SHY, ice.ETC_MISS_SH, ice.README_MD, ice.MAKEFILE, ice.LICENSE)
				_autogen_mod(m, ice.GO_MOD)
				ReposAddFile(m, "", ice.GO_MOD)
			}},
			nfs.MODULE: {Name: "module name*=hi help type*=Hash,Zone,Data,Lang,Code main*=main.go@key zone top", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				if m.WarnNotFound(!nfs.Exists(m, kit.Path(".git")), "未初始化代码库") {
					return
				}
				m.OptionDefault(mdb.ZONE, m.Option(mdb.NAME), mdb.HELP, m.Option(mdb.NAME), "top", "web.chat")
				m.OptionDefault(mdb.KEY, kit.Keys(m.Option("top"), strings.ReplaceAll(m.Option(mdb.ZONE), "/", "."), m.Option(mdb.NAME)))
				m.Option(nfs.FILE, path.Join(m.Option(mdb.ZONE), kit.Keys(m.Option(mdb.NAME), GO)))
				m.Option(mdb.TEXT, kit.Format("`name:\"list %s\" help:\"%s\"`", _autogen_list(m), m.Option(mdb.HELP)))
				defer m.Go(func() { _autogen_version(m.Spawn()) })
				if p := path.Join(ice.SRC, m.Option(nfs.FILE)); !nfs.Exists(m, p) {
					_autogen_import(m, path.Join(ice.SRC, m.Option(cli.MAIN)), m.Option(mdb.ZONE), _autogen_mod(m, ice.GO_MOD))
					m.Option(mdb.ZONE, path.Base(m.Option(mdb.ZONE)))
					_autogen_module(m, p)
				}
				return
				if p := path.Join(ice.SRC, kit.ExtChange(m.Option(nfs.FILE), SHY)); !nfs.Exists(m, p) {
					_autogen_source(m, kit.ExtChange(path.Join(ice.SRC, m.Option(cli.MAIN)), SHY), p)
					_autogen_script(m, p)
				}
			}},
			DEVPACK: {Help: "开发", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(WEBPACK, mdb.REMOVE)
				if ice.Info.NodeType == web.SERVER {
					m.Cmd(web.DREAM, nfs.GOWORK)
				}
			}},
			WEBPACK: {Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(WEBPACK, mdb.CREATE)
				if ice.Info.NodeType == web.SERVER {
					nfs.Trash(m, ice.GO_WORK_SUM)
					nfs.Trash(m, ice.GO_WORK)
				}
			}},
			IMPORT: {Hand: func(m *ice.Message, arg ...string) {
				_autogen_import(m, ice.SRC_MAIN_GO, arg[0], kit.Select(_autogen_mod(m, ice.GO_MOD), arg, 1))
			}},
			BINPACK: {Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				kit.For([]string{"intshell", "volcanos", "node_modules", "learning", "icons"}, func(p string) {
					if _, e := os.Stat("usr/" + p); os.IsNotExist(e) {
						m.Cmd("web.code.git.repos", "clone", "https://shylinux.com/x/"+p, p, "usr/"+p+"/")
					}
				})
				const (
					USR_RELEASE_CONF_GO    = "usr/release/conf.go"
					USR_RELEASE_BINPACK_GO = "usr/release/binpack.go"
				)
				if m.Cmd(BINPACK, mdb.CREATE); isReleaseContexts(m) {
					nfs.CopyFile(m, USR_RELEASE_BINPACK_GO, ice.SRC_BINPACK_GO, func(buf []byte, offset int) []byte {
						kit.If(offset == 0, func() { buf = bytes.Replace(buf, []byte("package main"), []byte("package ice"), 1) })
						return buf
					})
					m.Cmd(nfs.COPY, USR_RELEASE_CONF_GO, ice.USR_ICEBERGS+"conf.go")
					m.Cmdy(nfs.DIR, USR_RELEASE_BINPACK_GO)
					m.Cmdy(nfs.DIR, USR_RELEASE_CONF_GO)
				}
				m.Cmdy(nfs.CAT, _autogen_version(m))
			}},
			VERSION: {Help: "版本", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, _autogen_version(m)) }},
		}},
	})
}
func isReleaseContexts(m *ice.Message) bool {
	return ice.Info.Make.Module == "shylinux.com/x/ContextOS" && nfs.Exists(m, ice.USR_RELEASE) && nfs.Exists(m, ice.USR_VOLCANOS) && nfs.Exists(m, ice.USR_INTSHELL)
}
func AutogenMod(m *ice.Message) string {
	return _autogen_mod(m, ice.GO_MOD)
}
func AutogenImport(m *ice.Message, zone string) string {
	return _autogen_import(m, ice.SRC_MAIN_GO, zone, _autogen_mod(m, ice.GO_MOD))
}
