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
	if ls = strings.SplitN(string(data), ice.DF, 2); !aaa.UserLogin(m, ls[0], ls[1]) {
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
	web.Index.MergeCommands(ice.Commands{"/x/": {Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
		if m.RenderVoid(); m.Option("go-get") == "1" {
			p := _git_url(m, path.Join(arg...))
			m.RenderResult(kit.Format(`<meta name="go-import" content="%s">`, kit.Format(`%s git %s`, strings.Split(p, "://")[1], p)))
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
		SERVER: {Name: "server path commit auto create import", Help: "服务器", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name*", Hand: func(m *ice.Message, arg ...string) {
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
				m.Assert(m.Option(nfs.PATH) != "")
				nfs.Trash(m, path.Join(ice.USR_LOCAL_REPOS, m.Option(nfs.PATH)))
			}},
			web.DREAM_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.REPOS:
					m.Cmd("", func(value ice.Maps) { m.Push(nfs.PATH, _git_url(m, value[nfs.PATH])) })
				}
			}},
		}, gdb.EventAction(web.DREAM_INPUTS)), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_ROOT, ice.USR_LOCAL_REPOS); len(arg) == 0 {
				m.Cmdy(nfs.DIR, nfs.PWD, func(value ice.Maps) { m.PushScript("git clone " + _git_url(m, value[nfs.PATH])) }).Cut("time,path,size,script,action")
			}
		}},
	})
}
