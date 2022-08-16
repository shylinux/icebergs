package git

import (
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _server_rewrite(m *ice.Message, p string, r *http.Request) {
	if ua := r.Header.Get(web.UserAgent); strings.HasPrefix(ua, "Mozilla") {
		r.URL.Path = strings.Replace(r.URL.Path, "/x/", "/chat/pod/", 1)
		m.Info("rewrite %v -> %v", p, r.URL.Path) // 访问服务

	} else {
		r.URL.Path = strings.Replace(r.URL.Path, "/x/", "/code/git/repos/", 1)
		m.Info("rewrite %v -> %v", p, r.URL.Path) // 下载源码
	}
}
func _server_login(m *ice.Message) error {
	if m.Conf("web.serve", kit.Keym(tcp.LOCALHOST)) != ice.FALSE {
		if tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
			return nil // 本机请求
		}
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
		return fmt.Errorf("username or password error") // 登录失败
	}
	if aaa.UserRole(m, ls[0]) == aaa.VOID {
		return fmt.Errorf("userrole has no right") // 没有权限
	}
	return nil
}
func _server_param(m *ice.Message, arg ...string) (string, string) {
	repos, service := path.Join(arg...), kit.Select(arg[len(arg)-1], m.Option("service"))
	switch {
	case strings.HasSuffix(repos, "info/refs"):
		repos = strings.TrimSuffix(repos, "info/refs")
	default:
		repos = strings.TrimSuffix(repos, service)
	}
	return kit.Path(m.Config(nfs.PATH), REPOS, strings.TrimSuffix(repos, ".git/")), strings.TrimPrefix(service, "git-")
}
func _server_repos(m *ice.Message, arg ...string) error {
	repos, service := _server_param(m, arg...)

	if m.Option(cli.CMD_DIR, repos); strings.HasSuffix(path.Join(arg...), "info/refs") {
		web.RenderType(m.W, "", kit.Format("application/x-git-%s-advertisement", service))
		msg := _git_cmd(m, service, "--stateless-rpc", "--advertise-refs", ice.PT)
		_server_writer(m, "# service=git-"+service+ice.NL, msg.Result())
		return nil
	}

	reader, err := _server_reader(m)
	if err != nil {
		return err
	}
	defer reader.Close()

	m.Option(cli.CMD_OUTPUT, m.W)
	m.Option(cli.CMD_INPUT, reader)
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

const SERVER = "server"

func init() {
	Index.MergeCommands(ice.Commands{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) { m.Render(ice.RENDER_VOID) }},
		"/repos/": {Name: "/repos/", Help: "代码库", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddRewrite(func(p string, w http.ResponseWriter, r *http.Request) bool {
					if strings.HasPrefix(p, "/x/") {
						_server_rewrite(m, p, r)
					}
					return false
				})
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option("go-get") == "1" { // 下载地址
				p := web.MergeLink(m, "/x/"+path.Join(arg...))
				m.RenderResult(kit.Format(`<meta name="%s" content="%s">`, "go-import", kit.Format(`%s git %s`, strings.Split(p, "://")[1], p)))
				return
			}

			switch repos, service := _server_param(m, arg...); service {
			case "receive-pack": // 上传代码
				if err := _server_login(m); err != nil {
					web.RenderHeader(m.W, "WWW-Authenticate", `Basic realm="git server"`)
					web.RenderStatus(m.W, http.StatusUnauthorized, err.Error())
					return // 没有权限
				}
				if !nfs.ExistsFile(m, repos) { // 创建仓库
					_git_cmd(m, INIT, "--bare", repos)
					m.Logs(mdb.CREATE, REPOS, repos)
				}
			case "upload-pack": // 下载代码
				if !nfs.ExistsFile(m, repos) {
					web.RenderStatus(m.W, http.StatusNotFound, kit.Format("not found: %s", arg[0]))
					return
				}
			}

			if err := _server_repos(m, arg...); err != nil {
				web.RenderStatus(m.W, http.StatusInternalServerError, err.Error())
			}
		}},
		SERVER: {Name: "server path auto create import", Help: "服务器", Actions: ice.Actions{
			mdb.CREATE: {Name: "create name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, ice.USR_LOCAL_REPOS)
				_git_cmd(m, INIT, "--bare", m.Option(mdb.NAME))
			}},
			mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, ice.OptionFields("time,name,path"), func(value ice.Maps) {
					m.Option(cli.CMD_DIR, value[nfs.PATH])
					remote := web.MergeLink(m, "/x/"+value[REPOS])
					_git_cmd(m, PUSH, remote, MASTER)
					_git_cmd(m, PUSH, "--tags", remote, MASTER)
				})
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(nfs.PATH) != "")
				m.Cmd(nfs.TRASH, path.Join(ice.USR_LOCAL_REPOS, m.Option(nfs.PATH)))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_ROOT, ice.USR_LOCAL_REPOS); len(arg) == 0 {
				m.Cmdy(nfs.DIR, nfs.PWD, func(value ice.Maps) {
					m.PushScript("git clone " + web.MergeLink(m, "/x/"+path.Clean(value[nfs.PATH])))
				}).Cut("time,path,size,script,action")
				return
			}
			m.Cmdy("_sum", path.Join(m.Option(nfs.DIR_ROOT), arg[0]))
		}},
	})
}
