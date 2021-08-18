package git

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	githttp "github.com/AaronO/go-git-http"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

var basicAuthRegex = regexp.MustCompile("^([^:]*):(.*)$")

func _server_login(m *ice.Message) error {
	parts := strings.SplitN(m.R.Header.Get("Authorization"), " ", 2)
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
		return fmt.Errorf("userrole error")
	}
	return nil
}
func _server_repos(m *ice.Message, arg ...string) {
	m.Option(cli.CMD_DIR, path.Join(m.Conf(SERVER, kit.META_PATH), REPOS))
	p := strings.TrimSuffix(path.Join(arg...), "info/refs")
	if _, e := os.Stat(path.Join(m.Option(cli.CMD_DIR), p)); os.IsNotExist(e) {
		m.Cmd(cli.SYSTEM, GIT, INIT, "--bare", p) // 创建仓库
	}
}

const SERVER = "server"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVER: {Name: SERVER, Help: "服务器", Value: kit.Data(kit.MDB_PATH, ice.USR_LOCAL)},
		},
		Commands: map[string]*ice.Command{
			web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Render(ice.RENDER_VOID)
			}},
			"/repos/": {Name: "/repos/", Help: "代码库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch m.Option("service") {
				case "git-receive-pack": // 上传代码
					if err := _server_login(m); err != nil {
						m.W.Header().Set("WWW-Authenticate", `Basic realm="git server"`)
						http.Error(m.W, err.Error(), 401) // 认证失败
						return
					}
					_server_repos(m, arg...)

				case "git-upload-pack": // 下载代码

				}

				githttp.New(kit.Path(m.Conf(SERVER, kit.META_PATH))).ServeHTTP(m.W, m.R)
			}},
			SERVER: {Name: "server path auto create", Help: "服务器", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, path.Join(m.Conf(SERVER, kit.META_PATH), REPOS))
					m.Cmd(cli.SYSTEM, GIT, INIT, "--bare", m.Option(kit.MDB_NAME))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(nfs.DIR_ROOT, path.Join(m.Conf(SERVER, kit.META_PATH), REPOS)); len(arg) == 0 {
					m.Cmdy(nfs.DIR, "./")
					return
				}

				m.Cmdy("_sum", path.Join(m.Option(nfs.DIR_ROOT), arg[0]))
			}},
		},
	})
}
