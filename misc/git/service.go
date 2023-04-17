package git

import (
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"shylinux.com/x/go-git/v5/plumbing/transport/file"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _service_login(m *ice.Message) error {
	if ice.Info.Localhost && tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
		return nil
	} else if auth := strings.SplitN(m.R.Header.Get(web.Authorization), ice.SP, 2); strings.ToLower(auth[0]) != "basic" {
		return fmt.Errorf("Authentication type error")
	} else if data, err := base64.StdEncoding.DecodeString(auth[1]); err != nil {
		return err
	} else if auth := strings.SplitN(string(data), ice.DF, 2); m.Cmdv(Prefix(TOKEN), auth[0], TOKEN) != auth[1] {
		return fmt.Errorf("username or password error")
	} else if aaa.UserRole(m, auth[0]) == aaa.VOID {
		return fmt.Errorf("userrole has no right")
	} else {
		return nil
	}
}
func _service_path(m *ice.Message, p string, arg ...string) string {
	return kit.Path(ice.USR_LOCAL_REPOS, kit.TrimExt(p, GIT), path.Join(arg...))
}
func _service_param(m *ice.Message, arg ...string) (string, string) {
	repos, service := arg[0], kit.Select(arg[len(arg)-1], m.Option(SERVICE))
	return _service_path(m, repos), strings.TrimPrefix(service, "git-")
}
func _service_repos(m *ice.Message, arg ...string) error {
	repos, service := _service_param(m, arg...)
	m.Logs(m.R.Method, service, repos)
	if m.Option(cli.CMD_DIR, repos); strings.HasSuffix(path.Join(arg...), INFO_REFS) {
		m.Option(ice.MSG_USERROLE, aaa.TECH)
		web.RenderType(m.W, "", kit.Format("application/x-git-%s-advertisement", service))
		_service_writer(m, "# service=git-"+service+ice.NL, _git_cmds(m, service, "--stateless-rpc", "--advertise-refs", ice.PT))
		return nil
	}
	reader, err := _service_reader(m)
	if err != nil {
		return err
	}
	defer reader.Close()
	web.RenderType(m.W, "", kit.Format("application/x-git-%s-result", service))
	_git_cmd(m.Options(cli.CMD_INPUT, reader, cli.CMD_OUTPUT, m.W), service, "--stateless-rpc", ice.PT)
	return nil
}
func _service_writer(m *ice.Message, cmd string, str ...string) {
	s := strconv.FormatInt(int64(len(cmd)+4), 16)
	kit.If(len(s)%4 != 0, func() { s = strings.Repeat("0", 4-len(s)%4) + s })
	m.W.Write([]byte(s + cmd + "0000" + strings.Join(str, "")))
}
func _service_reader(m *ice.Message) (io.ReadCloser, error) {
	switch m.R.Header.Get("content-encoding") {
	case "deflate":
		return flate.NewReader(m.R.Body), nil
	case "gzip":
		return gzip.NewReader(m.R.Body)
	}
	return m.R.Body, nil
}

const (
	INFO_REFS    = "info/refs"
	RECEIVE_PACK = "receive-pack"
	UPLOAD_PACK  = "upload-pack"
)
const SERVICE = "service"

func init() {
	web.Index.MergeCommands(ice.Commands{"/x/": {Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
		if arg[0] == ice.LIST {
			m.Cmd(Prefix(SERVICE), func(value ice.Maps) { m.Push(nfs.REPOS, web.MergeLink(m, "/x/"+kit.Keys(value[nfs.REPOS], GIT))) })
			m.Sort(nfs.REPOS)
			return
		} else if m.RenderVoid(); m.Option("go-get") == "1" {
			p := _git_url(m, path.Join(arg...))
			m.RenderResult(kit.Format(`<meta name="go-import" content="%s">`, kit.Format(`%s git %s`, strings.TrimSuffix(strings.Split(p, "://")[1], nfs.PT+GIT), p)))
			return
		}
		switch repos, service := _service_param(m, arg...); service {
		case RECEIVE_PACK:
			if err := _service_login(m); m.Warn(err, ice.ErrNotLogin) {
				web.RenderHeader(m.W, "WWW-Authenticate", `Basic realm="git server"`)
				return
			} else if !nfs.Exists(m, repos) {
				_repos_init(m, repos)
			}
		case UPLOAD_PACK:
			if m.Warn(!nfs.Exists(m, repos), ice.ErrNotFound, arg[0]) {
				return
			}
		}
		m.Warn(_service_repos(m, arg...), ice.ErrNotValid)
	}}})
	Index.MergeCommands(ice.Commands{
		SERVICE: {Name: "service repos commit file auto", Help: "代码源", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, ice.USR_LOCAL_REPOS, func(value ice.Maps) { _repos_insert(m, value[nfs.PATH]) })
			}},
			mdb.CREATE: {Name: "create name*=demo", Hand: func(m *ice.Message, arg ...string) {
				_repos_init(m, _service_path(m, m.Option(mdb.NAME)))
				_repos_insert(m, _service_path(m, m.Option(mdb.NAME)))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(REPOS) != "")
				mdb.HashRemove(m, m.Option(REPOS))
				nfs.Trash(m, _service_path(m, m.Option(REPOS)))
			}},
			RECEIVE_PACK: {Hand: func(m *ice.Message, arg ...string) {
				if err := file.ServeReceivePack(arg[0]); err != nil {
					fmt.Fprintln(os.Stderr, "ERR:", err)
					os.Exit(128)
				}
			}},
			UPLOAD_PACK: {Hand: func(m *ice.Message, arg ...string) {
				if err := file.ServeUploadPack(arg[0]); err != nil {
					fmt.Fprintln(os.Stderr, "ERR:", err)
					os.Exit(128)
				}
			}},
			TOKEN:      {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(TOKEN, cli.MAKE) }},
			code.VIMER: {Hand: func(m *ice.Message, arg ...string) { _repos_vimer(m, _service_path, arg...) }},
		}, mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,commit"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...).Sort(REPOS).Action(mdb.CREATE, TOKEN)
				m.Echo(strings.ReplaceAll(m.Cmdx("web.code.publish", ice.CONTEXTS), "app username", "dev username"))
			} else if len(arg) == 1 {
				repos := _repos_open(m, arg[0])
				if branch, err := repos.Branch(arg[1]); !m.Warn(err) {
					_repos_log(m, branch, repos)
				}
			} else if len(arg) == 2 {
				if repos := _repos_open(m, arg[0]); arg[1] == INDEX {
					_repos_status(m, arg[0], repos)
				} else {
					_repos_stats(m, repos, arg[1])
				}
			} else {
				m.Cmdy("", code.VIMER, arg)
			}
		}},
	})
}
