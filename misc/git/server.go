package git

import (
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
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
	return kit.Path(m.Config(nfs.PATH), REPOS, repos), strings.TrimPrefix(service, "git-")
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
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Render(ice.RENDER_VOID)
		}},
		"/repos/": {Name: "/repos/", Help: "代码库", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
					if p := r.URL.Path; strings.HasPrefix(p, "/x/") {
						r.URL.Path = strings.Replace(r.URL.Path, "/x/", "/code/git/repos/", -1)
						m.Info("rewrite %v -> %v", p, r.URL.Path)
					}
					return false
				})
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option("go-get") == "1" { // 下载地址
				p := kit.Split(m.MergeURL2("/x/"+path.Join(arg...)), "?")[0]
				m.RenderResult(kit.Format(`<meta name="%s" content="%s">`, "go-import", kit.Format(`%s git %s`, strings.TrimPrefix(p, "https://"), p)))
				return
			}

			switch repos, service := _server_param(m, arg...); service {
			case "receive-pack": // 上传代码
				if err := _server_login(m); err != nil {
					web.RenderHeader(m, "WWW-Authenticate", `Basic realm="git server"`)
					web.RenderStatus(m, 401, err.Error())
					return // 没有权限
				}
				if _, e := os.Stat(path.Join(repos)); os.IsNotExist(e) {
					m.Cmd(cli.SYSTEM, GIT, INIT, "--bare", repos) // 创建仓库
					m.Log_CREATE(REPOS, repos)
				}
			case "upload-pack": // 下载代码
			}

			if err := _server_repos(m, arg...); err != nil {
				web.RenderStatus(m, 500, err.Error())
			}
		}},
		SERVER: {Name: "server path auto create import", Help: "服务器", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create name", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, path.Join(ice.USR_LOCAL, REPOS))
				m.Cmdy(cli.SYSTEM, GIT, INIT, "--bare", m.Option(kit.MDB_NAME))
			}},
			mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, ice.OptionFields("time,name,path")).Table(func(index int, value map[string]string, head []string) {
					m.Option(cli.CMD_DIR, value[nfs.PATH])
					m.Cmd(cli.SYSTEM, GIT, PUSH, m.MergeURL2("/x/"+value[kit.MDB_NAME]), MASTER)
					m.Cmd(cli.SYSTEM, GIT, PUSH, "--tags", m.MergeURL2("/x/"+value[kit.MDB_NAME]), MASTER)
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option(nfs.DIR_ROOT, path.Join(ice.USR_LOCAL, REPOS)); len(arg) == 0 {
				m.Cmdy(nfs.DIR, ice.PWD, "time,path,size").Table(func(index int, value map[string]string, head []string) {
					m.PushScript("git clone " + m.MergeURL2("/x/"+value[nfs.PATH]))
				})
				m.StatusTimeCount()
				return
			}
			m.Cmdy("_sum", path.Join(m.Option(nfs.DIR_ROOT), arg[0]))
		}},
	}})
}
