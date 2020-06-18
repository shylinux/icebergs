package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func Login(msg *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	msg.Option(ice.MSG_USERNAME, "")
	msg.Option(ice.MSG_USERROLE, "")

	if msg.Options(ice.MSG_SESSID) {
		// 会话认证
		aaa.SessCheck(msg, msg.Option(ice.MSG_SESSID))
	}

	if !msg.Options(ice.MSG_USERNAME) && tcp.IPIsLocal(msg, msg.Option(ice.MSG_USERIP)) {
		// 自动认证
		if aaa.UserLogin(msg, cli.UserName, cli.PassWord) {
			if strings.HasPrefix(msg.Option(ice.MSG_USERUA), "Mozilla/5.0") {
				msg.Option(ice.MSG_SESSID, aaa.SessCreate(msg, msg.Option(ice.MSG_USERNAME), msg.Option(ice.MSG_USERROLE)))
				Render(msg, "cookie", msg.Option(ice.MSG_SESSID))
			}
		}
	}

	if s, ok := msg.Target().Commands[ice.WEB_LOGIN]; ok {
		// 权限检查
		msg.Target().Run(msg, s, ice.WEB_LOGIN, kit.Simple(msg.Optionv("cmds"))...)

	} else if ls := strings.Split(msg.Option(ice.MSG_USERURL), "/"); msg.Conf(SERVE, kit.Keys("meta.black", ls[1])) == "true" {
		return false // black

	} else if msg.Conf(SERVE, kit.Keys("meta.white", ls[1])) == "true" {
		return true // white

	} else {
		if msg.Warn(!msg.Options(ice.MSG_USERNAME), "not login %s", msg.Option(ice.MSG_USERURL)) {
			msg.Render("status", 401, "not login")
			return false
		}
		if !msg.Right(msg.Option(ice.MSG_USERURL)) {
			msg.Render("status", 403, "not auth")
			return false
		}
	}

	return msg.Option(ice.MSG_USERURL) != ""
}
func HandleCmd(key string, cmd *ice.Command, msg *ice.Message, w http.ResponseWriter, r *http.Request) {
	defer func() { msg.Cost("%s %v %v", r.URL.Path, msg.Optionv("cmds"), msg.Format("append")) }()
	if u, e := url.Parse(r.Header.Get("Referer")); e == nil {
		for k, v := range u.Query() {
			msg.Logs("refer", k, v)
			msg.Option(k, v)
		}
	}

	// 用户请求
	msg.Option(ice.MSG_USERWEB, msg.Conf(SHARE, "meta.domain"))
	msg.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
	msg.Option(ice.MSG_USERUA, r.Header.Get("User-Agent"))
	msg.Option(ice.MSG_USERURL, r.URL.Path)
	if msg.R, msg.W = r, w; r.Header.Get("X-Real-Port") != "" {
		msg.Option(ice.MSG_USERADDR, msg.Option(ice.MSG_USERIP)+":"+r.Header.Get("X-Real-Port"))
	} else {
		msg.Option(ice.MSG_USERADDR, r.RemoteAddr)
	}

	// 请求变量
	msg.Option(ice.MSG_SESSID, "")
	msg.Option(ice.MSG_OUTPUT, "")
	for _, v := range r.Cookies() {
		msg.Option(v.Name, v.Value)
	}

	// 解析引擎
	switch r.Header.Get("Content-Type") {
	case "application/json":
		var data interface{}
		if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e != nil, "%s", e) {
			msg.Optionv(ice.MSG_USERDATA, data)
			msg.Logs("json", "value", kit.Formats(data))
		}

		switch d := data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				msg.Optionv(k, v)
			}
		}
	default:
		r.ParseMultipartForm(kit.Int64(kit.Select(r.Header.Get("Content-Length"), "4096")))
		if r.ParseForm(); len(r.PostForm) > 0 {
			for k, v := range r.PostForm {
				msg.Logs("form", k, v)
			}
		}
	}

	// 请求参数
	for k, v := range r.Form {
		if msg.Optionv(k, v); k == ice.MSG_SESSID {
			msg.Render("cookie", v[0])
		}
	}

	// 请求命令
	if msg.Option(ice.MSG_USERPOD, msg.Option("pod")); msg.Optionv("cmds") == nil {
		if p := strings.TrimPrefix(msg.Option(ice.MSG_USERURL), key); p != "" {
			msg.Optionv("cmds", strings.Split(p, "/"))
		}
	}

	// 执行命令
	if cmds := kit.Simple(msg.Optionv("cmds")); Login(msg, w, r) {
		msg.Option("_option", msg.Optionv(ice.MSG_OPTION))
		msg.Target().Run(msg, cmd, msg.Option(ice.MSG_USERURL), cmds...)
	}

	// 渲染引擎
	_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
	Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
}
func (web *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := web.m

	if r.Header.Get("index.module") == "" {
		// 解析地址
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			r.Header.Set(ice.MSG_USERIP, ip)
		} else if ip := r.Header.Get("X-Real-Ip"); ip != "" {
			r.Header.Set(ice.MSG_USERIP, ip)
		} else if strings.HasPrefix(r.RemoteAddr, "[") {
			r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, "]")[0][1:])
		} else {
			r.Header.Set(ice.MSG_USERIP, strings.Split(r.RemoteAddr, ":")[0])
		}
		m.Info("").Info("%s %s %s", r.Header.Get(ice.MSG_USERIP), r.Method, r.URL)

		r.Header.Set("index.module", m.Target().Name)
		r.Header.Set("index.path", r.URL.Path)
		r.Header.Set("index.url", r.URL.String())

		if m.Conf(SERVE, "meta.logheaders") == "true" {
			// 请求参数
			for k, v := range r.Header {
				m.Info("%s: %v", k, kit.Format(v))
			}
			m.Info(" ")

			defer func() {
				// 响应参数
				for k, v := range w.Header() {
					m.Info("%s: %v", k, kit.Format(v))
				}
				m.Info(" ")
			}()
		}
	}

	if strings.HasPrefix(r.URL.Path, "/debug") {
		r.URL.Path = strings.Replace(r.URL.Path, "/debug", "/code", -1)
	}

	if r.URL.Path == "/" && m.Conf(SERVE, "meta.init") != "true" {
		if _, e := os.Stat(m.Conf(SERVE, "meta.volcanos.path")); e == nil {
			// 初始化成功
			m.Conf(SERVE, "meta.init", "true")
		}
		m.W = w
		Render(m, "refresh", m.Conf(SERVE, "meta.volcanos.refresh"))
		m.Event(ice.SYSTEM_INIT)
		m.W = nil
	} else if r.URL.Path == "/share" && r.Method == "GET" {
		http.ServeFile(w, r, m.Conf(SERVE, "meta.page.share"))
	} else {
		web.ServeMux.ServeHTTP(w, r)
	}
}

const SERVE = "serve"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_SERVE: {Name: "serve", Help: "服务器", Value: kit.Data(
				"init", "false", "logheaders", "false",
				"black", kit.Dict(),
				"white", kit.Dict(
					"login", true,
					"share", true,
					"space", true,
					"route", true,
					"static", true,
					"plugin", true,
					"publish", true,
				),

				"title", "github.com/shylinux/contexts",
				"legal", []interface{}{`<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`},

				"static", kit.Dict("/", "usr/volcanos/"),
				"volcanos", kit.Dict("path", "usr/volcanos", "branch", "master",
					"repos", "https://github.com/shylinux/volcanos",
					"require", ".ish/pluged",
					"refresh", "5",
				), "page", kit.Dict(
					"index", "usr/volcanos/page/index.html",
					"share", "usr/volcanos/page/share.html",
				), "publish", "usr/publish/",

				"template", kit.Dict("path", "usr/template", "list", []interface{}{
					`{{define "raw"}}{{.Result}}{{end}}`,
				}),
			)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_SERVE: {Name: "serve [random] [ups...]", Help: "服务器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if cli.NodeType(m, ice.WEB_SERVER, cli.HostName); len(arg) > 0 && arg[0] == "random" {
					cli.NodeType(m, ice.WEB_SERVER, cli.PathName)
					// 随机端口
					SpideCreate(m, "self", "http://random")
					arg = arg[1:]
				}

				// 启动服务
				m.Target().Start(m, "self")
				defer m.Cmd(ice.WEB_SPACE, "connect", "self")
				m.Sleep("1s")

				// 连接服务
				for _, k := range arg {
					m.Cmd(ice.WEB_SPACE, "connect", k)
				}
			}},
		}}, nil)
}
