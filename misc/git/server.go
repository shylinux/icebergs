package git

import (
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _server_login(m *ice.Message) error {
	if tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) && m.Conf("web.serve", kit.Keym(tcp.LOCALHOST)) == ice.TRUE {
		return nil
	}
	ls := strings.SplitN(m.R.Header.Get(web.Authorization), ice.SP, 2)
	if strings.ToLower(ls[0]) != "basic" {
		return fmt.Errorf("Authentication '%s' was not of 'Basic' type", ls[0])
	}
	data, err := base64.StdEncoding.DecodeString(ls[1])
	if err != nil {
		return err
	}
	if ls = strings.SplitN(string(data), ice.DF, 2); m.Cmd("web.code.git.token", ls[0]).Append(TOKEN) != ls[1] && !aaa.UserLogin(m.Spawn(), ls[0], ls[1]) {
		return fmt.Errorf("username or password error")
	}
	if aaa.UserRole(m, ls[0]) == aaa.VOID {
		return fmt.Errorf("userrole has no right")
	}
	return nil
}
func _server_param(m *ice.Message, arg ...string) (string, string) {
	repos, service := path.Join(arg...), kit.Select(arg[len(arg)-1], m.Option("service"))
	switch {
	case strings.HasSuffix(repos, INFO_REFS):
		repos = strings.TrimSuffix(repos, INFO_REFS)
	default:
		repos = strings.TrimSuffix(repos, service)
	}
	return kit.Path(ice.USR_LOCAL_REPOS, strings.TrimSuffix(repos, ".git/")), strings.TrimPrefix(service, "git-")
}
func _server_repos(m *ice.Message, arg ...string) error {
	repos, service := _server_param(m, arg...)
	if m.Option(cli.CMD_DIR, repos); strings.HasSuffix(path.Join(arg...), INFO_REFS) {
		web.RenderType(m.W, "", kit.Format("application/x-git-%s-advertisement", service))
		_server_writer(m, "# service=git-"+service+ice.NL, _git_cmds(m, service, "--stateless-rpc", "--advertise-refs", ice.PT))
		return nil
	}
	reader, err := _server_reader(m)
	if err != nil {
		return err
	}
	defer reader.Close()
	m.Options(cli.CMD_INPUT, reader, cli.CMD_OUTPUT, m.W)
	web.RenderType(m.W, "", kit.Format("application/x-git-%s-result", service))
	_git_cmd(m, service, "--stateless-rpc", ice.PT)
	return nil
}
func _server_writer(m *ice.Message, cmd string, str ...string) {
	s := strconv.FormatInt(int64(len(cmd)+4), 16)
	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}
	m.W.Write([]byte(s + cmd + "0000" + strings.Join(str, "")))
}
func _server_reader(m *ice.Message) (io.ReadCloser, error) {
	switch m.R.Header.Get("content-encoding") {
	case "deflate":
		return flate.NewReader(m.R.Body), nil
	case "gzip":
		return gzip.NewReader(m.R.Body)
	}
	return m.R.Body, nil
}

const (
	INFO_REFS = "info/refs"
)
const SERVER = "server"

func init() {
	web.Index.MergeCommands(ice.Commands{"/x/": {Actions: ice.MergeActions(ctx.CmdAction(), aaa.WhiteAction(ctx.COMMAND, ice.RUN)), Hand: func(m *ice.Message, arg ...string) {
		if arg[0] == ice.LIST {
			m.Cmd("web.code.git.server", func(value ice.Maps) { m.Push(nfs.REPOS, web.MergeLink(m, "/x/"+value[nfs.REPOS]+".git")) })
			m.Sort(nfs.REPOS)
			return
		}
		if !m.IsCliUA() || len(arg) > 0 && strings.Contains(arg[0], ice.AT) || len(arg) > 1 && arg[1] == ice.SRC {
			if len(arg) > 0 && strings.Contains(arg[0], ice.AT) {
				ls := strings.Split(arg[0], ice.AT)
				_repos_cat(m, path.Join(ice.USR_LOCAL_REPOS, ls[0]), "master", ls[1], path.Join(arg[1:]...))
				m.RenderResult()
			} else if len(arg) > 1 && strings.HasPrefix(arg[1], "v") && strings.Contains(arg[1], ice.PT) {
				_repos_cat(m, path.Join(ice.USR_LOCAL_REPOS, arg[0]), "master", arg[1], path.Join(arg[2:]...))
				m.RenderResult()
			} else if len(arg) > 1 && arg[1] == ice.SRC {
				_repos_cat(m, path.Join(ice.USR_LOCAL_REPOS, arg[0]), "master", "", path.Join(arg[1:]...))
				m.RenderResult()
			} else {
				web.RenderCmds(m, kit.Dict(ctx.DISPLAY, "/plugin/local/code/repos.js", ctx.INDEX, "web.code.git.inner",
					ctx.ARGS, kit.List(strings.TrimSuffix(arg[0], ".git"), arg[1], "pwd", kit.Select("README.md", path.Join(kit.Slice(arg, 2)...)))))
			}
			return
		}
		if m.RenderVoid(); m.Option("go-get") == "1" {
			p := _git_url(m, path.Join(arg...))
			m.RenderResult(kit.Format(`<meta name="go-import" content="%s">`, kit.Format(`%s git %s`, strings.TrimSuffix(strings.Split(p, "://")[1], ".git"), p)))
			return
		}
		switch repos, service := _server_param(m, arg...); service {
		case "receive-pack":
			if err := _server_login(m); m.Warn(err, ice.ErrNotLogin) {
				web.RenderHeader(m.W, "WWW-Authenticate", `Basic realm="git server"`)
				return
			} else if !nfs.ExistsFile(m, repos) {
				m.Logs(mdb.CREATE, REPOS, repos)
				_repos_init(m, repos)
			}
		case "upload-pack":
			if m.Warn(!nfs.ExistsFile(m, repos), ice.ErrNotFound, arg[0]) {
				return
			}
		}
		m.Warn(_server_repos(m, arg...), ice.ErrNotValid)
	}}})
	Index.MergeCommands(ice.Commands{
		"inner": {Name: "inner repos branch commit path auto token", Help: "服务器", Actions: ice.MergeActions(ice.Actions{}, aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_ROOT, ice.USR_LOCAL_REPOS); len(arg) == 0 {
			} else if dir := path.Join(m.Option(nfs.DIR_ROOT), arg[0]); len(arg) == 1 {
			} else if len(arg) == 2 {
			} else if len(arg) == 3 || strings.HasSuffix(arg[3], nfs.PS) {
				_repos_dir(m, dir, arg[1], arg[2], kit.Select("", arg, 3), nil)
			} else {
				m.Option(nfs.FILE, kit.Select("", arg, 3))
				_repos_cat(m, dir, arg[1], arg[2], kit.Select("", arg, 3))
			}
		}},
		SERVER: {Name: "server repos branch commit path auto create import token", Help: "代码源", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name*=demo", Hand: func(m *ice.Message, arg ...string) {
				_repos_init(m, path.Join(ice.USR_LOCAL_REPOS, m.Option(mdb.NAME)))
			}},
			mdb.IMPORT: {Hand: func(m *ice.Message, arg ...string) {
				ReposList(m).Tables(func(value ice.Maps) {
					m.Option(cli.CMD_DIR, value[nfs.PATH])
					remote := _git_url(m, value[REPOS])
					_git_cmd(m, PUSH, remote, MASTER)
					_git_cmd(m, PUSH, "--tags", remote, MASTER)
				})
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(nfs.REPOS) != "")
				nfs.Trash(m, path.Join(ice.USR_LOCAL_REPOS, m.Option(nfs.REPOS)))
			}},
			web.DREAM_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.REPOS:
					m.Cmd("", func(value ice.Maps) { m.Push(nfs.PATH, _git_url(m, value[nfs.PATH])) })
				}
			}},
			"inner": {Help: "编辑器", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[0] != ice.RUN {
					arg = []string{path.Join(ice.USR_LOCAL_REPOS, arg[0]), kit.Select("README.md", arg, 3)}
				} else if kit.Select("", arg, 1) != ctx.ACTION {
					if dir := path.Join(ice.USR_LOCAL_REPOS, m.Option(REPOS)); len(arg) < 3 {
						_repos_dir(m, dir, m.Option(BRANCH), m.Option(COMMIT), kit.Select("", arg, 1), nil)
					} else {
						_repos_cat(m, dir, m.Option(BRANCH), m.Option(COMMIT), arg[2])
						ctx.DisplayLocal(m, "code/inner.js")
					}
					return
				}
				ctx.ProcessField(m, "", arg, arg...)
			}},
			TOKEN: {Hand: func(m *ice.Message, arg ...string) {
				token := kit.Hashs("uniq")
				m.Cmd(TOKEN, mdb.CREATE, aaa.USERNAME, m.Option(ice.MSG_USERNAME), TOKEN, token)
				m.EchoScript(kit.Format("echo %s >> ~/.git-credentials", strings.Replace(m.Option(ice.MSG_USERHOST), "://", kit.Format("://%s:%s@", m.Option(ice.MSG_USERNAME), token), 1)))
			}},
		}, gdb.EventAction(web.DREAM_INPUTS)), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_ROOT, ice.USR_LOCAL_REPOS); len(arg) == 0 {
				m.Option(ice.MSG_USERROLE, aaa.TECH)
				m.Cmdy(nfs.DIR, nfs.PWD, "time,name,size,action", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_DIR), func(value ice.Maps) {
					m.PushScript("git clone " + _git_url(m, value[mdb.NAME]))
				}).Cut("time,name,size,script,action").RenameAppend(mdb.NAME, nfs.REPOS).SortTimeR(mdb.TIME)
				m.Echo(strings.ReplaceAll(m.Cmdx("web.code.publish", ice.CONTEXTS), "app username", "dev username"))
			} else if dir := path.Join(m.Option(nfs.DIR_ROOT), arg[0]); len(arg) == 1 {
				_repos_branch(m, dir)
			} else if len(arg) == 2 {
				_repos_commit(m, dir, arg[1], nil)
			} else if len(arg) == 3 || arg[3] == "" || strings.HasSuffix(arg[3], ice.PS) {
				_repos_dir(m, dir, arg[1], arg[2], kit.Select("", arg, 3), nil)
			} else {
				m.Cmdy("", "inner", arg)
			}
			m.StatusTimeCount()
		}},
	})
}
