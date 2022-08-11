package git

import (
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func requestReader(m *ice.Message) (io.ReadCloser, error) {
	switch m.R.Header.Get("content-encoding") {
	case "deflate":
		return flate.NewReader(m.R.Body), nil
	case "gzip":
		return gzip.NewReader(m.R.Body)
	}
	return m.R.Body, nil
}
func packetWrite(m *ice.Message, cmd string, str ...string) {
	s := strconv.FormatInt(int64(len(cmd)+4), 16)
	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}
	m.W.Write([]byte(s + cmd + "0000" + strings.Join(str, "")))
}

var basicAuthRegex = regexp.MustCompile("^([^:]*):(.*)$")

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
	if m.Conf("web.serve", "meta.localhost") != ice.FALSE {
		if tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
			return nil
		}
	}
	parts := strings.SplitN(m.R.Header.Get("Authorization"), ice.SP, 2)
	if len(parts) < 2 {
		return fmt.Errorf("Invalid authorization header, not enought parts")
	}

	authType, authData := parts[0], parts[1]
	if strings.ToLower(authType) != "basic" {
		return fmt.Errorf("Authentication '%s' was not of 'Basic' type", authType)
	}

	data, err := base64.StdEncoding.DecodeString(authData)
	if err != nil {
		return err
	}

	matches := basicAuthRegex.FindStringSubmatch(string(data))
	if matches == nil {
		return fmt.Errorf("Authorization data '%s' did not match auth regexp", data)
	}

	username, password := matches[1], matches[2]
	if !aaa.UserLogin(m, username, password) {
		return fmt.Errorf("username or password error")
	}
	if aaa.UserRole(m, username) == aaa.VOID {
		return fmt.Errorf("userrole has no right")
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
		msg := m.Cmd(cli.SYSTEM, GIT, service, "--stateless-rpc", "--advertise-refs", ice.PT)
		packetWrite(m, "# service=git-"+service+ice.NL, msg.Result())
		return nil
	}

	reader, err := requestReader(m)
	if err != nil {
		return err
	}
	defer reader.Close()

	m.Option(cli.CMD_OUTPUT, m.W)
	m.Option(cli.CMD_INPUT, reader)
	web.RenderType(m.W, "", kit.Format("application/x-git-%s-result", service))
	m.Cmd(cli.SYSTEM, GIT, service, "--stateless-rpc", ice.PT)
	return nil
}

const SERVER = "server"

func init() {
	Index.MergeCommands(ice.Commands{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) { m.Render(ice.RENDER_VOID) }},
		"/repos/": {Name: "/repos/", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddRewrite(func(p string, w http.ResponseWriter, r *http.Request) bool {
					if strings.HasPrefix(p, "/x/") {
						_server_rewrite(m, p, r)
					} else if strings.HasPrefix(p, "/chat/pod/") {
						_server_rewrite(m, p, r)
					}
					return false
				})
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if !m.IsCliUA() {
				p := kit.Split(web.MergeURL2(m, "/x/"+path.Join(arg...)), "?")[0]
				m.RenderResult("git clone %v", p)
				return
			}
			if m.Option("go-get") == "1" { // 下载地址
				p := kit.Split(web.MergeURL2(m, "/x/"+path.Join(arg...)), "?")[0]
				m.RenderResult(kit.Format(`<meta name="%s" content="%s">`, "go-import", kit.Format(`%s git %s`, strings.TrimPrefix(p, "https://"), p)))
				return
			}

			switch repos, service := _server_param(m, arg...); service {
			case "receive-pack": // 上传代码
				if err := _server_login(m); err != nil {
					web.RenderHeader(m.W, "WWW-Authenticate", `Basic realm="git server"`)
					web.RenderStatus(m.W, 401, err.Error())
					return // 没有权限
				}
				if !kit.FileExists(path.Join(repos)) {
					m.Cmd(cli.SYSTEM, GIT, INIT, "--bare", repos) // 创建仓库
					m.Logs(mdb.CREATE, REPOS, repos)
				}
			case "upload-pack": // 下载代码
				aaa.UserRoot(m)
				if kit.Select("", arg, 1) == "info" && m.Cmd(web.DREAM, arg[0]).Length() > 0 {
					m.Cmd(web.SPACE, arg[0], "web.code.git.status", "submit", web.MergeURL2(m, "/x/")+arg[0])
				}
				if !kit.FileExists(path.Join(repos)) {
					web.RenderStatus(m.W, 404, kit.Format("not found: %s", arg[0]))
					return
				}
			}

			if err := _server_repos(m, arg...); err != nil {
				web.RenderStatus(m.W, 500, err.Error())
			}
		}},
		SERVER: {Name: "server path auto create import", Help: "服务器", Actions: ice.Actions{
			mdb.CREATE: {Name: "create name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, path.Join(ice.USR_LOCAL, REPOS))
				m.Cmdy(cli.SYSTEM, GIT, INIT, "--bare", m.Option(mdb.NAME))
			}},
			mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, ice.OptionFields("time,name,path")).Tables(func(value ice.Maps) {
					remote := strings.Split(web.MergeURL2(m, "/x/"+value[REPOS]), "?")[0]
					m.Option(cli.CMD_DIR, value[nfs.PATH])
					m.Cmd(cli.SYSTEM, GIT, PUSH, remote, MASTER)
					m.Cmd(cli.SYSTEM, GIT, PUSH, "--tags", remote, MASTER)
				})
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(nfs.PATH) != "")
				m.Cmd(nfs.TRASH, path.Join(ice.USR_LOCAL_REPOS, m.Option(nfs.PATH)))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(nfs.DIR_ROOT, ice.USR_LOCAL_REPOS); len(arg) == 0 {
				m.Cmdy(nfs.DIR, nfs.PWD).Tables(func(value ice.Maps) {
					m.PushScript("git clone " + web.MergeURL2(m, "/x/"+strings.TrimSuffix(value[nfs.PATH], ice.PS)))
				})
				m.Cut("time,path,size,script,action")
				m.StatusTimeCount()
				return
			}
			m.Cmdy("_sum", path.Join(m.Option(nfs.DIR_ROOT), arg[0]))
		}},
	})
}
