package web

import (
	"github.com/gorilla/websocket"
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
	"github.com/skip2/go-qrcode"

	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"
)

const (
	MSG_MAPS = 1
)

type Frame struct {
	*http.Client
	*http.Server
	*http.ServeMux
	m    *ice.Message
	send map[string]*ice.Message
}

func Count(m *ice.Message, cmd, key, name string) int {
	count := kit.Int(m.Conf(cmd, kit.Keys(key, name)))
	m.Conf(cmd, kit.Keys(key, name), count+1)
	return count
}
func Render(msg *ice.Message, cmd string, args ...interface{}) {
	msg.Log(ice.LOG_EXPORT, "%s: %v", cmd, args)
	switch arg := kit.Simple(args...); cmd {
	case ice.RENDER_OUTPUT:
	case "redirect":
		http.Redirect(msg.W, msg.R, kit.MergeURL(arg[0], arg[1:]), 307)

	case "refresh":
		arg = []string{"200", fmt.Sprintf(`<!DOCTYPE html><head><meta charset="utf-8"><meta http-equiv="Refresh" content="%d"></head><body>%s</body>`,
			kit.Int(kit.Select("3", arg, 0)), kit.Select("请稍后，系统初始化中...", arg, 1),
		)}
		fallthrough

	case "status":
		msg.W.WriteHeader(kit.Int(kit.Select("200", arg, 0)))
		msg.W.Write([]byte(kit.Select("", arg, 1)))

	case "cookie":
		expire := time.Now().Add(kit.Duration(msg.Conf(ice.AAA_SESS, "meta.expire")))
		http.SetCookie(msg.W, &http.Cookie{Value: arg[0], Name: kit.Select(ice.MSG_SESSID, arg, 1), Path: "/", Expires: expire})

	case ice.RENDER_DOWNLOAD:
		msg.W.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(path.Base(arg[0]), arg, 2)))
		msg.W.Header().Set("Content-Type", kit.Select("text/html", arg, 1))
		http.ServeFile(msg.W, msg.R, arg[0])

	case ice.RENDER_RESULT:
		if len(arg) > 0 {
			msg.W.Write([]byte(kit.Format(arg[0], args[1:]...)))
		} else {
			msg.W.Write([]byte(msg.Result()))
		}

	case ice.RENDER_QRCODE:
		if qr, e := qrcode.New(arg[0], qrcode.Medium); msg.Assert(e) {
			msg.W.Header().Set("Content-Type", "image/png")
			msg.Assert(qr.Write(kit.Int(kit.Select("256", arg, 1)), msg.W))
		}
	default:
		if cmd != "" {
			msg.Echo(kit.Format(cmd, args...))
		}
		msg.W.Header().Set("Content-Type", "application/json")
		fmt.Fprint(msg.W, msg.Formats("meta"))
	}
	msg.Append(ice.MSG_OUTPUT, ice.RENDER_OUTPUT)
}
func IsLocalIP(msg *ice.Message, ip string) (ok bool) {
	if ip == "::1" || strings.HasPrefix(ip, "127.") {
		return true
	}
	msg.Cmd("tcp.ifconfig").Table(func(index int, value map[string]string, head []string) {
		if value["ip"] == ip {
			ok = true
		}
	})
	return ok
}

func (web *Frame) Login(msg *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	msg.Option(ice.MSG_USERNAME, "")
	msg.Option(ice.MSG_USERROLE, "")

	if msg.Options(ice.MSG_SESSID) {
		// 会话认证
		sub := msg.Cmd(ice.AAA_SESS, "check", msg.Option(ice.MSG_SESSID))
		msg.Log(ice.LOG_AUTH, "role: %s user: %s", msg.Option(ice.MSG_USERROLE, sub.Append("userrole")),
			msg.Option(ice.MSG_USERNAME, sub.Append("username")))
	}

	if !msg.Options(ice.MSG_USERNAME) && IsLocalIP(msg, msg.Option(ice.MSG_USERIP)) {
		// 自动认证
		msg.Option(ice.MSG_USERNAME, msg.Conf(ice.CLI_RUNTIME, "boot.username"))
		msg.Option(ice.MSG_USERROLE, msg.Cmdx(ice.AAA_ROLE, "check", msg.Option(ice.MSG_USERNAME)))
		if strings.HasPrefix(msg.Option(ice.MSG_USERUA), "Mozilla/5.0") {
			msg.Option(ice.MSG_SESSID, msg.Cmdx(ice.AAA_SESS, "create", msg.Option(ice.MSG_USERNAME), msg.Option(ice.MSG_USERROLE)))
			msg.Render("cookie", msg.Option(ice.MSG_SESSID))
		}
		msg.Log(ice.LOG_AUTH, "user: %s role: %s sess: %s", msg.Option(ice.MSG_USERNAME), msg.Option(ice.MSG_USERROLE), msg.Option(ice.MSG_SESSID))
	}

	if s, ok := msg.Target().Commands[ice.WEB_LOGIN]; ok {
		// 权限检查
		msg.Target().Run(msg, s, ice.WEB_LOGIN, kit.Simple(msg.Optionv("cmds"))...)
	} else if strings.HasPrefix(msg.Option(ice.MSG_USERURL), "/static/") {
	} else if strings.HasPrefix(msg.Option(ice.MSG_USERURL), "/plugin/") {
	} else if strings.HasPrefix(msg.Option(ice.MSG_USERURL), "/login/") {
	} else if strings.HasPrefix(msg.Option(ice.MSG_USERURL), "/space/") {
	} else if strings.HasPrefix(msg.Option(ice.MSG_USERURL), "/route/") {
	} else if strings.HasPrefix(msg.Option(ice.MSG_USERURL), "/share/") {
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
func (web *Frame) HandleWSS(m *ice.Message, safe bool, c *websocket.Conn, name string) bool {
	for running := true; running; {
		if t, b, e := c.ReadMessage(); m.Warn(e != nil, "space recv %d msg %v", t, e) {
			break
		} else {
			switch t {
			case MSG_MAPS:
				// 接收报文
				socket, msg := c, m.Spawn(b)
				source := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name)
				target := kit.Simple(msg.Optionv(ice.MSG_TARGET))
				msg.Info("recv %v<-%v %v", target, source, msg.Format("meta"))

				if len(target) == 0 {
					// 本地执行
					msg.Option(ice.MSG_USERROLE, msg.Cmdx(ice.AAA_ROLE, "check", msg.Option(ice.MSG_USERNAME)))
					msg.Log("some", "%s: %s", msg.Option(ice.MSG_USERROLE), msg.Option(ice.MSG_USERNAME))
					if msg.Optionv(ice.MSG_HANDLE, "true"); !msg.Warn(!safe, "no right") {
						m.Option("_dev", name)
						msg = msg.Cmd()
					}
					if source, target = []string{}, kit.Revert(source)[1:]; msg.Detail() == "exit" {
						return true
					}

				} else if msg.Richs(ice.WEB_SPACE, nil, target[0], func(key string, value map[string]interface{}) {
					if s, ok := value["socket"].(*websocket.Conn); ok {
						socket, source, target = s, source, target[1:]
					} else {
						socket, source, target = s, source, target[1:]
					}
				}) != nil {
					// 转发报文
					msg.Info("space route")

				} else if call, ok := web.send[msg.Option(ice.MSG_TARGET)]; len(target) == 1 && ok {
					// 接收响应
					delete(web.send, msg.Option(ice.MSG_TARGET))
					call.Back(msg)
					break

				} else if msg.Warn(msg.Option("_handle") == "true", "space miss") {
					// 回复失败
					break

				} else {
					// 下发失败
					msg.Warn(true, "space error")
					source, target = []string{}, kit.Revert(source)[1:]
				}

				// 发送报文
				msg.Optionv(ice.MSG_SOURCE, source)
				msg.Optionv(ice.MSG_TARGET, target)
				socket.WriteMessage(t, []byte(msg.Format("meta")))
				msg.Info("send %v %v->%v %v", t, source, target, msg.Format("meta"))
				msg.Log("cost", "%s: ", msg.Format("cost"))
			}
		}
	}
	return false
}
func (web *Frame) HandleCGI(m *ice.Message, alias map[string]interface{}, which string) *template.Template {
	cgi := template.FuncMap{}

	tmpl := template.New(ice.WEB_TMPL)
	cb := func(k string, p []string, v *ice.Command) {
		cgi[k] = func(arg ...interface{}) (res interface{}) {
			m.TryCatch(m.Spawn(), true, func(msg *ice.Message) {
				msg.Target().Run(msg, v, k, kit.Simple(p, arg)...)

				buffer := bytes.NewBuffer([]byte{})
				m.Assert(tmpl.ExecuteTemplate(buffer, msg.Option(ice.WEB_TMPL), msg))
				res = string(buffer.Bytes())
			})
			return
		}
	}
	for k, v := range alias {
		list := kit.Simple(v)
		if v, ok := m.Target().Commands[list[0]]; ok {
			cb(k, list[1:], v)
		}
	}
	for k, v := range m.Target().Commands {
		if strings.HasPrefix(k, "/") || strings.HasPrefix(k, "_") {
			continue
		}
		cb(k, nil, v)
	}

	tmpl = tmpl.Funcs(cgi)
	// tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf(ice.WEB_SERVE, ice.Meta("template", "path")), "/*.tmpl")))
	// tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf(ice.WEB_SERVE, ice.Meta("template", "path")), m.Target().Name, "/*.tmpl")))
	tmpl, e := tmpl.ParseFiles(which)
	if e != nil {
	}
	// m.Confm(ice.WEB_SERVE, ice.Meta("template", "list"), func(index int, value string) { tmpl = template.Must(tmpl.Parse(value)) })
	return tmpl
}
func (web *Frame) HandleCmd(m *ice.Message, key string, cmd *ice.Command) {
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		m.TryCatch(m.Spawns(), true, func(msg *ice.Message) {
			defer func() { msg.Cost("%s %v %v", r.URL.Path, msg.Optionv("cmds"), msg.Format("append")) }()

			// 请求地址
			msg.Option(ice.MSG_USERWEB, m.Conf(ice.WEB_SHARE, "meta.domain"))
			msg.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
			msg.Option(ice.MSG_USERUA, r.Header.Get("User-Agent"))
			msg.Option(ice.MSG_USERURL, r.URL.Path)
			msg.Option(ice.MSG_USERPOD, "")
			msg.Option(ice.MSG_SESSID, "")
			msg.Option(ice.MSG_OUTPUT, "")
			msg.R, msg.W = r, w
			if r.Header.Get("X-Real-Port") != "" {
				msg.Option(ice.MSG_USERADDR, msg.Option(ice.MSG_USERIP)+":"+r.Header.Get("X-Real-Port"))
			} else {
				msg.Option(ice.MSG_USERADDR, r.RemoteAddr)
			}

			// 请求变量
			for _, v := range r.Cookies() {
				if v.Value != "" {
					msg.Option(v.Name, v.Value)
				}
			}

			// 解析引擎
			switch r.Header.Get("Content-Type") {
			case "application/json":
				var data interface{}
				if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e != nil, "%s", e) {
					msg.Optionv("content_data", data)
					msg.Info("%s", kit.Formats(data))
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
						msg.Info("%s: %v", k, v)
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
			if msg.Optionv("cmds") == nil {
				if p := strings.TrimPrefix(msg.Option(ice.MSG_USERURL), key); p != "" {
					msg.Optionv("cmds", strings.Split(p, "/"))
				}
			}
			cmds := kit.Simple(msg.Optionv("cmds"))

			if web.Login(msg, w, r) {
				// 登录成功
				msg.Option("_option", msg.Optionv(ice.MSG_OPTION))
				// 执行命令
				msg.Target().Run(msg, cmd, msg.Option(ice.MSG_USERURL), cmds...)
			}

			// 渲染引擎
			_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
			Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
		})
	})
}
func (web *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := web.m

	index := r.Header.Get("index.module") == ""
	if index {
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

		// 解析地址
		r.Header.Set("index.module", "some")
		r.Header.Set("index.path", r.URL.Path)
		r.Header.Set("index.url", r.URL.String())
	}

	if index && kit.Right(m.Conf(ice.WEB_SERVE, "meta.logheaders")) {
		// 请求参数
		for k, v := range r.Header {
			m.Info("%s: %v", k, kit.Format(v))
		}
		m.Info(" ")
	}

	if strings.HasPrefix(r.URL.Path, "/debug") {
		r.URL.Path = strings.Replace(r.URL.Path, "/debug", "/code", -1)
	}

	if r.URL.Path == "/" && m.Conf(ice.WEB_SERVE, "meta.init") != "true" {
		if _, e := os.Stat(m.Conf(ice.WEB_SERVE, "meta.volcanos.path")); e == nil {
			// 初始化成功
			m.Conf(ice.WEB_SERVE, "meta.init", "true")
		}
		m.W = w
		Render(m, "refresh")
		m.Event(ice.SYSTEM_INIT)
		m.W = nil
	} else if r.URL.Path == "/share" && r.Method == "GET" {
		http.ServeFile(w, r, m.Conf(ice.WEB_SERVE, "meta.page.share"))

	} else if r.URL.Path == "/" && r.Method == "GET" {
		http.ServeFile(w, r, m.Conf(ice.WEB_SERVE, "meta.page.index"))

	} else {
		web.ServeMux.ServeHTTP(w, r)
	}

	if index && kit.Right(m.Conf(ice.WEB_SERVE, "meta.logheaders")) {
		// 响应参数
		for k, v := range w.Header() {
			m.Info("%s: %v", k, kit.Format(v))
		}
		m.Info(" ")
	}
}

func (web *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (web *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	web.send = map[string]*ice.Message{}
	return web
}
func (web *Frame) Start(m *ice.Message, arg ...string) bool {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if w, ok := s.Server().(*Frame); ok {
			if w.ServeMux != nil {
				return
			}
			w.ServeMux = http.NewServeMux()

			// 静态路由
			msg := m.Spawns(s)
			m.Confm(ice.WEB_SERVE, "meta.static", func(key string, value string) {
				m.Log("route", "%s <- %s <- %s", s.Name, key, value)
				w.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 级联路由
			route := "/" + s.Name + "/"
			if n, ok := p.Server().(*Frame); ok && n.ServeMux != nil {
				msg.Log("route", "%s <= %s", p.Name, route)
				n.Handle(route, http.StripPrefix(path.Dir(route), w))
			}

			// 命令路由
			m.Travel(func(p *ice.Context, sub *ice.Context, k string, x *ice.Command) {
				if s == sub && k[0] == '/' {
					msg.Log("route", "%s <- %s", s.Name, k)
					w.HandleCmd(msg, k, x)
				}
			})
		}
	})

	// TODO simple
	m.Richs(ice.WEB_SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
		client := value["client"].(map[string]interface{})

		// 服务地址
		port := m.Cap(ice.CTX_STREAM, client["hostname"])
		m.Log("serve", "listen %s %s %v", arg[0], port, m.Conf(ice.CLI_RUNTIME, "node"))

		// 启动服务
		web.m, web.Server = m, &http.Server{Addr: port, Handler: web}
		m.Event(ice.SERVE_START, arg[0])
		m.Warn(true, "listen %s", web.Server.ListenAndServe())
		m.Event(ice.SERVE_CLOSE, arg[0])
	})
	return true
}
func (web *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "web", Help: "网络模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.WEB_SPIDE: {Name: "spide", Help: "蜘蛛侠", Value: kit.Data(kit.MDB_SHORT, "client.name")},
		ice.WEB_SERVE: {Name: "serve", Help: "服务器", Value: kit.Data(
			"page", kit.Dict(
				"index", "usr/volcanos/page/index.html",
				"share", "usr/volcanos/page/share.html",
			),
			"static", kit.Dict("/", "usr/volcanos/",
				"/static/volcanos/", "usr/volcanos/",
				"/publish/", "usr/publish/",
			),
			"volcanos", kit.Dict("path", "usr/volcanos", "branch", "master",
				"repos", "https://github.com/shylinux/volcanos",
				"require", "usr/local",
			),
			"template", kit.Dict("path", "usr/template", "list", []interface{}{
				`{{define "raw"}}{{.Result}}{{end}}`,
			}),
			"logheaders", "false", "init", "false",
		)},
		ice.WEB_SPACE: {Name: "space", Help: "空间站", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
			"redial.a", 3000, "redial.b", 1000, "redial.c", 1000,
			"buffer.r", 4096, "buffer.w", 4096,
			"timeout.c", "30s",
		)},
		ice.WEB_DREAM: {Name: "dream", Help: "梦想家", Value: kit.Data("path", "usr/local/work",
			// "cmd", []interface{}{ice.CLI_SYSTEM, "ice.sh", "start", ice.WEB_SPACE, "connect"},
			"cmd", []interface{}{ice.CLI_SYSTEM, "ice.bin", ice.WEB_SPACE, "connect"},
		)},

		ice.WEB_FAVOR: {Name: "favor", Help: "收藏夹", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, "template", favor_template,
			"proxy", "",
		)},
		ice.WEB_CACHE: {Name: "cache", Help: "缓存池", Value: kit.Data(
			kit.MDB_SHORT, "text", "path", "var/file", "store", "var/data", "fsize", "100000", "limit", "50", "least", "30",
		)},
		ice.WEB_STORY: {Name: "story", Help: "故事会", Value: kit.Dict(
			kit.MDB_META, kit.Dict(kit.MDB_SHORT, "data"),
			"head", kit.Data(kit.MDB_SHORT, "story"),
			"mime", kit.Dict("md", "txt"),
		)},
		ice.WEB_SHARE: {Name: "share", Help: "共享链", Value: kit.Data("index", "usr/volcanos/share.html", "template", share_template)},

		ice.WEB_ROUTE: {Name: "route", Help: "路由", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		ice.WEB_PROXY: {Name: "proxy", Help: "代理", Value: kit.Data(kit.MDB_SHORT, "proxy")},
		ice.WEB_GROUP: {Name: "group", Help: "分组", Value: kit.Data(kit.MDB_SHORT, "group")},
		ice.WEB_LABEL: {Name: "label", Help: "标签", Value: kit.Data(kit.MDB_SHORT, "label")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			m.Cmd(ice.WEB_SPIDE, "add", "self", kit.Select("http://:9020", m.Conf(ice.CLI_RUNTIME, "conf.ctx_self")))
			if m.Richs(ice.WEB_SPIDE, nil, "dev", nil) == nil {
				m.Cmd(ice.WEB_SPIDE, "add", "dev", kit.Select("http://:9020", m.Conf(ice.CLI_RUNTIME, "conf.ctx_dev")))
			}
			if m.Richs(ice.WEB_SPIDE, nil, "shy", nil) == nil {
				m.Cmd(ice.WEB_SPIDE, "add", "shy", kit.Select("https://shylinux.com:443", m.Conf(ice.CLI_RUNTIME, "conf.ctx_shy")))
			}
			m.Rich(ice.WEB_SPACE, nil, kit.Dict(
				kit.MDB_TYPE, ice.WEB_BETTER, kit.MDB_NAME, "tmux",
				kit.MDB_TEXT, m.Conf(ice.CLI_RUNTIME, "boot.username"),
			))
			m.Watch(ice.SYSTEM_INIT, "web.code.git.repos", "volcanos", m.Conf(ice.WEB_SERVE, "meta.volcanos.path"),
				m.Conf(ice.WEB_SERVE, "meta.volcanos.repos"), m.Conf(ice.WEB_SERVE, "meta.volcanos.branch"))
			m.Conf(ice.WEB_FAVOR, "meta.template", favor_template)
			m.Conf(ice.WEB_SHARE, "meta.template", share_template)

			m.Cmd(ice.APP_SEARCH, "add", "favor", "base", m.AddCmd(&ice.Command{Name: "search word", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "set":
					m.Richs(ice.WEB_FAVOR, nil, arg[1], func(key string, value map[string]interface{}) {
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", arg[2], func(index int, value map[string]interface{}) {
							if cmd := m.Conf(ice.WEB_FAVOR, kit.Keys("meta.render", value["type"])); cmd != "" {
								m.Optionv("value", value)
								m.Cmdy(cmd, arg[1:])
							} else {
								m.Push("detail", value)
							}
						})
					})
					return
				}

				m.Option("cache.limit", -2)
				wg := &sync.WaitGroup{}
				m.Richs(ice.WEB_FAVOR, nil, "*", func(key string, val map[string]interface{}) {
					favor := kit.Format(kit.Value(val, "meta.name"))
					wg.Add(1)
					m.Info("routine %v", favor)
					m.Gos(m, func(m *ice.Message) {
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							if favor == arg[0] || value["type"] == arg[0] ||
								strings.Contains(kit.Format(value["name"]), arg[0]) || strings.Contains(kit.Format(value["text"]), arg[0]) {
								m.Push("pod", m.Option(ice.MSG_USERPOD))
								m.Push("engine", "favor")
								m.Push("favor", favor)
								m.Push("", value, []string{"id", "type", "name", "text"})
							}
						})
						wg.Done()
					})
				})
				wg.Wait()
			}}))

			m.Cmd(ice.APP_SEARCH, "add", "story", "base", m.AddCmd(&ice.Command{Name: "search word", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "set":
					m.Cmdy(ice.WEB_STORY, "index", arg[2])
					return
				}

				m.Richs(ice.WEB_STORY, "head", "*", func(key string, val map[string]interface{}) {
					if val["story"] == arg[0] {
						m.Push("pod", m.Option(ice.MSG_USERPOD))
						m.Push("engine", "story")
						m.Push("favor", val["story"])
						m.Push("id", val["list"])

						m.Push("type", val["scene"])
						m.Push("name", val["story"])
						m.Push("text", val["count"])
					}
				})
			}}))

			m.Cmd(ice.APP_SEARCH, "add", "share", "base", m.AddCmd(&ice.Command{Name: "search word", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "set":
					m.Cmdy(ice.WEB_SHARE, arg[2])
					return
				}

				m.Option("cache.limit", -2)
				m.Grows(ice.WEB_SHARE, nil, "", "", func(index int, value map[string]interface{}) {
					if value["share"] == arg[0] || value["type"] == arg[0] ||
						strings.Contains(kit.Format(value["name"]), arg[0]) || strings.Contains(kit.Format(value["text"]), arg[0]) {
						m.Push("pod", m.Option(ice.MSG_USERPOD))
						m.Push("engine", "share")
						m.Push("favor", value["type"])
						m.Push("id", value["share"])

						m.Push("type", value["type"])
						m.Push("name", value["name"])
						m.Push("text", value["text"])
					}
				})
			}}))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(ice.WEB_SPIDE, ice.WEB_SERVE, ice.WEB_GROUP, ice.WEB_LABEL,
				ice.WEB_FAVOR, ice.WEB_CACHE, ice.WEB_STORY, ice.WEB_SHARE,
			)

			m.Done()
			m.Richs(ice.WEB_SPACE, nil, "*", func(key string, value map[string]interface{}) {
				if kit.Format(value["type"]) == "master" {
					m.Done()
				}
			})
		}},

		ice.WEB_SPIDE: {Name: "spide name=auto [action:select=msg|raw|cache] [method:select=POST|GET] url [format:select=json|form|part|data|file] arg... auto", Help: "蜘蛛侠", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				// 爬虫列表
				m.Richs(ice.WEB_SPIDE, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["client"], []string{"name", "share", "login", "method", "url"})
				})
				m.Sort("name")
				return
			}
			if len(arg) == 1 || len(arg) > 3 && arg[3] == "" {
				// 爬虫详情
				m.Richs(ice.WEB_SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Push("detail", value)
					if kit.Value(value, "client.share") != nil {
						m.Push("key", "share")
						m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.text"), m.Conf(ice.WEB_SHARE, "meta.domain"), kit.Value(value, "client.share")))
					}
				})
				return
			}

			switch arg[0] {
			case "add":
				// 添加爬虫
				if uri, e := url.Parse(arg[2]); e == nil && arg[2] != "" {
					if uri.Host == "random" {
						uri.Host = ":" + m.Cmdx("tcp.getport")
						arg[2] = strings.Replace(arg[2], "random", uri.Host, -1)
					}
					dir, file := path.Split(uri.EscapedPath())
					if m.Richs(ice.WEB_SPIDE, nil, arg[1], func(key string, value map[string]interface{}) {
						// kit.Value(value, "client.name", arg[1])
						// kit.Value(value, "client.text", arg[2])
						kit.Value(value, "client.hostname", uri.Host)
						kit.Value(value, "client.url", arg[2])
					}) == nil {
						m.Rich(ice.WEB_SPIDE, nil, kit.Dict(
							"cookie", kit.Dict(), "header", kit.Dict(), "client", kit.Dict(
								"share", m.Cmdx(ice.WEB_SHARE, "add", ice.TYPE_SPIDE, arg[1], arg[2]),
								// "type", "POST", "name", arg[1], "text", arg[2],
								"name", arg[1], "url", arg[2], "method", "POST",
								"protocol", uri.Scheme, "hostname", uri.Host,
								"path", dir, "file", file, "query", uri.RawQuery,
								"timeout", "100s", "logheaders", false,
							),
						))
					}
					m.Log(ice.LOG_CREATE, "%s: %v", arg[1], arg[2:])
				}
				return
			case "login":
				m.Richs(ice.WEB_SPIDE, nil, arg[1], func(key string, value map[string]interface{}) {
					msg := m.Cmd(ice.WEB_SPIDE, arg[1], "msg", "/route/login", "login")
					if msg.Append(ice.MSG_USERNAME) != "" {
						m.Echo(msg.Append(ice.MSG_USERNAME))
						return
					}
					if msg.Result() != "" {
						kit.Value(value, "client.login", msg.Result())
						kit.Value(value, "client.share", m.Cmdx(ice.WEB_SHARE, "add", ice.TYPE_SPIDE, arg[1],
							kit.Format("%s?sessid=%s", kit.Value(value, "client.url"), kit.Value(value, "cookie.sessid"))))
					}
					m.Render(ice.RENDER_QRCODE, kit.Dict(
						kit.MDB_TYPE, "login", kit.MDB_NAME, arg[1],
						kit.MDB_TEXT, kit.Value(value, "cookie.sessid"),
					))
				})
				return
			}

			m.Richs(ice.WEB_SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
				client := value["client"].(map[string]interface{})
				// 缓存数据
				cache := ""
				switch arg[1] {
				case "raw":
					cache, arg = arg[1], arg[1:]
				case "msg":
					cache, arg = arg[1], arg[1:]
				case "cache":
					cache, arg = arg[1], arg[1:]
				}

				// 请求方法
				method := kit.Select("POST", client["method"])
				switch arg = arg[1:]; arg[0] {
				case "POST":
					method, arg = "POST", arg[1:]
				case "GET":
					method, arg = "GET", arg[1:]
				}

				// 请求地址
				uri, arg := arg[0], arg[1:]

				// 渲染引擎
				head := map[string]string{}
				body, ok := m.Optionv("body").(io.Reader)
				if !ok && len(arg) > 0 && method != "GET" {
					switch arg[0] {
					case "file":
						if f, e := os.Open(arg[1]); m.Warn(e != nil, "%s", e) {
							return
						} else {
							defer f.Close()
							body, arg = f, arg[2:]
						}
					case "data":
						body, arg = bytes.NewBufferString(arg[1]), arg[2:]
					case "part":
						buf := &bytes.Buffer{}
						mp := multipart.NewWriter(buf)
						for i := 1; i < len(arg)-1; i += 2 {
							if strings.HasPrefix(arg[i+1], "@") {
								if f, e := os.Open(arg[i+1][1:]); m.Assert(e) {
									defer f.Close()
									if p, e := mp.CreateFormFile(arg[i], path.Base(arg[i+1][1:])); m.Assert(e) {
										io.Copy(p, f)
									}
								}
							} else {
								mp.WriteField(arg[i], arg[i+1])
							}
						}
						mp.Close()
						head["Content-Type"] = mp.FormDataContentType()
						body = buf
					case "form":
						data := []string{}
						for i := 1; i < len(arg)-1; i += 2 {
							data = append(data, url.QueryEscape(arg[i])+"="+url.QueryEscape(arg[i+1]))
						}
						body = bytes.NewBufferString(strings.Join(data, "&"))
						head["Content-Type"] = "application/x-www-form-urlencoded"
					case "json":
						arg = arg[1:]
						fallthrough
					default:
						data := map[string]interface{}{}
						for i := 0; i < len(arg)-1; i += 2 {
							kit.Value(data, arg[i], arg[i+1])
						}
						if b, e := json.Marshal(data); m.Assert(e) {
							head["Content-Type"] = "application/json"
							body = bytes.NewBuffer(b)
						}
						m.Log(ice.LOG_EXPORT, "json: %s", kit.Format(data))
					}
					arg = arg[:0]
				} else {
					body = bytes.NewBuffer([]byte{})
				}

				// 请求地址
				uri = kit.MergeURL2(kit.Format(client["url"]), uri, arg)
				req, e := http.NewRequest(method, uri, body)
				m.Info("%s %s", req.Method, req.URL)
				m.Assert(e)

				// 请求变量
				kit.Fetch(value["cookie"], func(key string, value string) {
					req.AddCookie(&http.Cookie{Name: key, Value: value})
					m.Info("%s: %s", key, value)
				})
				kit.Fetch(value["header"], func(key string, value string) {
					req.Header.Set(key, value)
				})
				list := kit.Simple(m.Optionv("header"))
				for i := 0; i < len(list)-1; i += 2 {
					req.Header.Set(list[i], list[i+1])
				}
				for k, v := range head {
					req.Header.Set(k, v)
				}

				// 请求代理
				web := m.Target().Server().(*Frame)
				if web.Client == nil {
					web.Client = &http.Client{Timeout: kit.Duration(kit.Format(client["timeout"]))}
				}
				if method == "POST" {
					m.Info("%s: %s", req.Header.Get("Content-Length"), req.Header.Get("Content-Type"))
				}

				// 发送请求
				res, e := web.Client.Do(req)
				if m.Warn(e != nil, "%s", e) {
					return
				}

				// 检查结果
				if m.Cost("%s %s: %s", res.Status, res.Header.Get("Content-Length"), res.Header.Get("Content-Type")); m.Warn(res.StatusCode != http.StatusOK, "%s", res.Status) {
					return
				}

				// 缓存变量
				for _, v := range res.Cookies() {
					kit.Value(value, "cookie."+v.Name, v.Value)
					m.Log(ice.LOG_IMPORT, "%s: %s", v.Name, v.Value)
				}

				// 解析引擎
				switch cache {
				case "cache":
					m.Optionv("response", res)
					m.Cmdy(ice.WEB_CACHE, "download", res.Header.Get("Content-Type"), uri)
					m.Echo(m.Append("data"))
				case "raw":
					if b, e := ioutil.ReadAll(res.Body); m.Assert(e) {
						m.Echo(string(b))
					}
				case "msg":
					var data map[string][]string
					m.Assert(json.NewDecoder(res.Body).Decode(&data))
					m.Info("res: %s", kit.Formats(data))
					for _, k := range data[ice.MSG_APPEND] {
						for i := range data[k] {
							m.Push(k, data[k][i])
						}
					}
					m.Resultv(data[ice.MSG_RESULT])
				default:
					if strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
						b, _ := ioutil.ReadAll(res.Body)
						m.Echo(string(b))
						break
					}

					var data interface{}
					m.Assert(json.NewDecoder(res.Body).Decode(&data))
					data = kit.KeyValue(map[string]interface{}{}, "", data)
					m.Info("res: %s", kit.Formats(data))
					m.Push("", data)
				}
			})
		}},
		ice.WEB_SERVE: {Name: "serve [random] [ups...]", Help: "服务器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf(ice.CLI_RUNTIME, "node.name", m.Conf(ice.CLI_RUNTIME, "boot.hostname"))
			m.Conf(ice.CLI_RUNTIME, "node.type", ice.WEB_SERVER)

			if len(arg) > 0 && arg[0] == "random" {
				// 随机端口
				m.Conf(ice.CLI_RUNTIME, "node.name", m.Conf(ice.CLI_RUNTIME, "boot.pathname"))
				m.Cmd(ice.WEB_SPIDE, "add", "self", "http://random")
				arg = arg[1:]
			}

			// 启动服务
			m.Target().Start(m, "self")
			m.Sleep("1s")

			// 连接服务
			m.Cmd(ice.WEB_SPACE, "connect", "self")
			for _, k := range arg {
				m.Cmd(ice.WEB_SPACE, "connect", k)
			}
		}},
		ice.WEB_SPACE: {Name: "space name auto", Help: "空间站", Meta: kit.Dict(
			"exports", []string{"pod", "name"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 空间列表
				m.Richs(ice.WEB_SPACE, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "type", "name", "text"})
					if m.Option(ice.MSG_USERUA) != "" {
						m.Push("link", fmt.Sprintf(`<a target="_blank" href="%s?pod=%s">%s</a>`,
							kit.Select(m.Conf(ice.WEB_SHARE, "meta.domain"), m.Option(ice.MSG_USERWEB)),
							kit.Keys(m.Option(ice.MSG_USERPOD), value["name"]), value["name"]))
					}
				})
				m.Sort("name")
				return
			}

			web := m.Target().Server().(*Frame)
			switch arg[0] {
			case "share":
				m.Richs(ice.WEB_SPIDE, nil, m.Option("_dev"), func(key string, value map[string]interface{}) {
					m.Log(ice.LOG_CREATE, "dev: %s share: %s", m.Option("_dev"), arg[1])
					value["share"] = arg[1]
				})

			case "connect":
				// 基本信息
				dev := kit.Select("dev", arg, 1)
				node := m.Conf(ice.CLI_RUNTIME, "node.type")
				name := kit.Select(m.Conf(ice.CLI_RUNTIME, "node.name"), arg, 2)
				user := m.Conf(ice.CLI_RUNTIME, "boot.username")

				m.Hold(1).Gos(m, func(msg *ice.Message) {
					msg.Richs(ice.WEB_SPIDE, nil, dev, func(key string, value map[string]interface{}) {
						proto := kit.Select("ws", "wss", kit.Format(kit.Value(value, "client.protocol")) == "https")
						host := kit.Format(kit.Value(value, "client.hostname"))

						for i := 0; i < kit.Int(msg.Conf(ice.WEB_SPACE, "meta.redial.c")); i++ {
							if u, e := url.Parse(kit.MergeURL(proto+"://"+host+"/space/", "node", node, "name", name, "user", user,
								"proxy", kit.Select("master", arg, 3), "group", kit.Select("worker", arg, 4), "share", value["share"])); msg.Assert(e) {
								if s, e := net.Dial("tcp", host); !msg.Warn(e != nil, "%s", e) {
									if s, _, e := websocket.NewClient(s, u, nil, kit.Int(msg.Conf(ice.WEB_SPACE, "meta.buffer.r")), kit.Int(msg.Conf(ice.WEB_SPACE, "meta.buffer.w"))); !msg.Warn(e != nil, "%s", e) {
										msg = m.Spawn()

										// 连接成功
										msg.Rich(ice.WEB_SPACE, nil, kit.Dict(
											kit.MDB_TYPE, ice.WEB_MASTER, kit.MDB_NAME, dev, kit.MDB_TEXT, kit.Value(value, "client.hostname"),
											"socket", s,
										))
										msg.Log(ice.LOG_CMDS, "%d conn %s success %s", i, dev, u)
										if i = 0; web.HandleWSS(msg, true, s, dev) {
											break
										}
									}
								}

								// 断线重连
								sleep := time.Duration(rand.Intn(kit.Int(msg.Conf(ice.WEB_SPACE, "meta.redial.a"))*i+2)+kit.Int(msg.Conf(ice.WEB_SPACE, "meta.redial.b"))) * time.Millisecond
								msg.Info("%d sleep: %s reconnect: %s", i, sleep, u)
								time.Sleep(sleep)
							}
						}
					})
					m.Done()
				})

			default:
				if len(arg) == 1 {
					// 空间详情
					m.Richs(ice.WEB_SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
						m.Push("detail", value)
						m.Push("key", "link")
						m.Push("value", fmt.Sprintf(`<a target="_blank" href="%s?pod=%s">%s</a>`, m.Conf(ice.WEB_SHARE, "meta.domain"), value["name"], value["name"]))
					})
					break
				}

				if arg[0] == "" {
					// 本地命令
					m.Cmdy(arg[1:])
					break
				}

				target := strings.Split(arg[0], ".")
				m.Warn(m.Richs(ice.WEB_SPACE, nil, target[0], func(key string, value map[string]interface{}) {
					if socket, ok := value["socket"].(*websocket.Conn); !m.Warn(!ok, "socket err") {
						// 复制选项
						for _, k := range kit.Simple(m.Optionv("_option")) {
							switch k {
							case "detail", "cmds":
							default:
								m.Optionv(k, m.Optionv(k))
							}
						}
						m.Optionv("_option", m.Optionv("_option"))

						// 构造路由
						id := kit.Format(c.ID())
						m.Set(ice.MSG_DETAIL, arg[1:]...)
						m.Optionv(ice.MSG_TARGET, target[1:])
						m.Optionv(ice.MSG_SOURCE, []string{id})
						m.Info("send %s->%v %s", id, target, m.Format("meta"))

						// 下发命令
						m.Target().Server().(*Frame).send[id] = m
						socket.WriteMessage(MSG_MAPS, []byte(m.Format("meta")))

						m.Call(m.Option("_async") == "", func(res *ice.Message) *ice.Message {
							if res != nil && m != nil {
								// 返回结果
								m.Log("cost", "%s: %s %v %v", m.Format("cost"), arg[0], arg[1:], res.Format("append"))
								return m.Copy(res)
							}
							return nil
						})
					}
				}) == nil, "not found %s", arg[0])
			}
		}},
		ice.WEB_DREAM: {Name: "dream name auto", Help: "梦想家", Meta: kit.Dict(
			"exports", []string{"you", "name"}, "detail", []interface{}{"启动", "停止"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 && arg[0] == "action" {
				switch arg[1] {
				case "启动", "start":
					arg = []string{arg[4]}
				case "停止", "stop":
					m.Cmd(ice.WEB_SPACE, kit.Select(m.Option("name"), arg, 4), "exit", "1")
					m.Event(ice.DREAM_CLOSE, arg[4])
					return
				}
			}

			if len(arg) == 0 {
				// 任务列表
				m.Cmdy("nfs.dir", m.Conf(ice.WEB_DREAM, "meta.path"), "time name")
				m.Table(func(index int, value map[string]string, head []string) {
					if m.Richs(ice.WEB_SPACE, nil, value["name"], func(key string, value map[string]interface{}) {
						m.Push("type", value["type"])
						m.Push("status", "start")
					}) == nil {
						m.Push("type", "none")
						m.Push("status", "stop")
					}
				})
				m.Sort("name")
				m.Sort("status")
				return
			}

			// 规范命名
			if !strings.Contains(arg[0], "-") || !strings.HasPrefix(arg[0], "20") {
				arg[0] = m.Time("20060102-") + arg[0]
			}

			// 创建目录
			p := path.Join(m.Conf(ice.WEB_DREAM, "meta.path"), arg[0])
			os.MkdirAll(p, 0777)

			if b, e := ioutil.ReadFile(path.Join(p, m.Conf(ice.GDB_SIGNAL, "meta.pid"))); e == nil {
				if s, e := os.Stat("/proc/" + string(b)); e == nil && s.IsDir() {
					m.Info("already exists %v", string(b))
					return
				}
			}

			if m.Richs(ice.WEB_SPACE, nil, arg[0], nil) == nil {
				// 启动任务
				m.Option("cmd_dir", p)
				m.Option("cmd_type", "daemon")
				m.Optionv("cmd_env",
					"ctx_dev", m.Conf(ice.CLI_RUNTIME, "conf.ctx_dev"),
					"ctx_log", "boot.log", "ctx_mod", "ctx,log,gdb,ssh",
					"PATH", kit.Path(path.Join(p, "bin"))+":"+os.Getenv("PATH"),
				)
				m.Cmd(m.Confv(ice.WEB_DREAM, "meta.cmd"), "self", arg[0])
				time.Sleep(time.Second * 1)
				m.Event(ice.DREAM_START, arg...)
			}
			m.Cmdy("nfs.dir", p)
		}},

		ice.WEB_FAVOR: {Name: "favor favor=auto id=auto auto", Help: "收藏夹", Meta: kit.Dict(
			"exports", []string{"hot", "favor"}, "detail", []string{"编辑", "收藏", "收录", "导出", "删除"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 && arg[0] == "action" {
				favor, id := m.Option("favor"), m.Option("id")
				switch arg[2] {
				case "favor":
					favor = arg[3]
				case "id":
					id = arg[3]
				}

				switch arg[1] {
				case "modify", "编辑":
					m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
						if id == "" {
							m.Log(ice.LOG_MODIFY, "favor: %s value: %v->%v", key, kit.Value(value, kit.Keys("meta", arg[2])), arg[3])
							m.Echo("%s->%s", kit.Value(value, kit.Keys("meta", arg[2])), arg[3])
							kit.Value(value, kit.Keys("meta", arg[2]), arg[3])
							return
						}
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", id, func(index int, value map[string]interface{}) {
							m.Log(ice.LOG_MODIFY, "favor: %s index: %d value: %v->%v", key, index, value[arg[2]], arg[3])
							m.Echo("%s->%s", value[arg[2]], arg[3])
							kit.Value(value, arg[2], arg[3])
						})
					})
				case "commit", "收录":
					m.Echo("list: ")
					m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", id, func(index int, value map[string]interface{}) {
							m.Cmdy(ice.WEB_STORY, "add", value["type"], value["name"], value["text"])
						})
					})
				case "export", "导出":
					m.Echo("list: ")
					if favor == "" {
						m.Cmdy(ice.MDB_EXPORT, ice.WEB_FAVOR, kit.MDB_HASH, kit.MDB_HASH, "favor")
					} else {
						m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
							m.Cmdy(ice.MDB_EXPORT, ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_LIST, favor)
						})
					}
				case "delete", "删除":
					m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
						m.Cmdy(ice.MDB_DELETE, ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_DICT)
					})
				case "import", "导入":
					if favor == "" {
						m.Cmdy(ice.MDB_IMPORT, ice.WEB_FAVOR, kit.MDB_HASH, kit.MDB_HASH, "favor")
					} else {
						m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
							m.Cmdy(ice.MDB_IMPORT, ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_LIST, favor)
						})
					}
				}
				return
			}

			if len(arg) == 0 {
				// 收藏门类
				m.Richs(ice.WEB_FAVOR, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["meta"], []string{"time", "count"})
					m.Push("render", kit.Select("spide", kit.Value(value, "meta.render")))
					m.Push("favor", kit.Value(value, "meta.name"))
				})
				m.Sort("favor")
				return
			}

			switch arg[0] {
			case "save":
				f, p, e := kit.Create(arg[1])
				m.Assert(e)
				defer f.Close()
				w := csv.NewWriter(f)

				w.Write([]string{"favor", "type", "name", "text", "extra"})

				n := 0
				m.Option("cache.offend", 0)
				m.Option("cache.limit", -2)
				for _, favor := range arg[2:] {
					m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, val map[string]interface{}) {
						if m.Conf(ice.WEB_FAVOR, kit.Keys("meta.skip", kit.Value(val, "meta.name"))) == "true" {
							return
						}
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							w.Write(kit.Simple(kit.Value(val, "meta.name"), value["type"], value["name"], value["text"], kit.Format(value["extra"])))
							n++
						})
					})
				}
				w.Flush()
				m.Echo("%s: %d", p, n)
				return

			case "load":
				f, e := os.Open(arg[1])
				m.Assert(e)
				defer f.Close()
				r := csv.NewReader(f)

				head, e := r.Read()
				m.Assert(e)
				m.Info("head: %v", head)

				for {
					line, e := r.Read()
					if e != nil {
						break
					}
					m.Cmd(ice.WEB_FAVOR, line)
				}
				return

			case "sync":
				m.Richs(ice.WEB_FAVOR, nil, arg[1], func(key string, val map[string]interface{}) {
					remote := kit.Keys("meta.remote", arg[2], arg[3])
					count := kit.Int(kit.Value(val, kit.Keys("meta.count")))

					pull := kit.Int(kit.Value(val, kit.Keys(remote, "pull")))
					m.Cmd(ice.WEB_SPIDE, arg[2], "msg", "/favor/pull", "favor", arg[3], "begin", pull+1).Table(func(index int, value map[string]string, head []string) {
						m.Cmd(ice.WEB_FAVOR, arg[1], value["type"], value["name"], value["text"], value["extra"])
						pull = kit.Int(value["id"])
					})

					m.Option("cache.limit", count-kit.Int(kit.Value(val, kit.Keys(remote, "push"))))
					m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
						m.Cmd(ice.WEB_SPIDE, arg[2], "msg", "/favor/push", "favor", arg[3],
							"type", value["type"], "name", value["name"], "text", value["text"],
							"extra", kit.Format(value["extra"]),
						)
						pull++
					})
					kit.Value(val, kit.Keys(remote, "pull"), pull)
					kit.Value(val, kit.Keys(remote, "push"), kit.Value(val, "meta.count"))
					m.Echo("%d", kit.Value(val, "meta.count"))
					return
				})
				return
			}

			m.Option("favor", arg[0])
			fields := []string{kit.MDB_TIME, kit.MDB_ID, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT}
			if len(arg) > 1 && arg[1] == "extra" {
				fields, arg = append(fields, arg[2:]...), arg[:1]
			}
			if len(arg) < 3 {
				m.Richs(ice.WEB_FAVOR, nil, arg[0], func(key string, value map[string]interface{}) {
					if len(arg) < 2 {
						// 收藏列表
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							m.Push("", value, fields)
						})
						return
					}
					// 收藏详情
					m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", arg[1], func(index int, value map[string]interface{}) {
						m.Push("detail", value)
					})
				})
				return
			}

			favor := ""
			if m.Richs(ice.WEB_FAVOR, nil, arg[0], func(key string, value map[string]interface{}) {
				favor = key
			}) == nil {
				// 创建收藏
				favor = m.Rich(ice.WEB_FAVOR, nil, kit.Data(kit.MDB_NAME, arg[0]))
				m.Log(ice.LOG_CREATE, "favor: %s name: %s", favor, arg[0])
			}

			if len(arg) == 3 {
				arg = append(arg, "")
			}

			// 添加收藏
			index := m.Grow(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, favor), kit.Dict(
				kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
				"extra", kit.Dict(arg[4:]),
			))
			m.Richs(ice.WEB_FAVOR, nil, favor, func(key string, value map[string]interface{}) {
				kit.Value(value, "meta.time", m.Time())
			})
			m.Log(ice.LOG_INSERT, "favor: %s index: %d name: %s text: %s", favor, index, arg[2], arg[3])
			m.Echo("%d", index)

			// 分发数据
			if p := kit.Select(m.Conf(ice.WEB_FAVOR, "meta.proxy"), m.Option("you")); p != "" {
				m.Option("you", "")
				m.Cmdy(ice.WEB_PROXY, p, ice.WEB_FAVOR, arg)
			}
		}},
		ice.WEB_CACHE: {Name: "cache", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 缓存记录
				m.Grows(ice.WEB_CACHE, nil, "", "", func(index int, value map[string]interface{}) {
					m.Push("", value)
				})
				return
			}

			switch arg[0] {
			case "catch":
				// 导入文件
				if f, e := os.Open(arg[2]); m.Assert(e) {
					defer f.Close()

					// 创建文件
					h := kit.Hashs(f)
					if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, "meta.path"), h[:2], h)); m.Assert(e) {
						defer o.Close()
						f.Seek(0, os.SEEK_SET)

						// 导入数据
						if n, e := io.Copy(o, f); m.Assert(e) {
							m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(n), p)
							arg = kit.Simple(arg[0], arg[1], path.Base(arg[2]), p, p, n)
						}
					}
				}
				fallthrough
			case "upload", "download":
				if m.R != nil {
					// 上传文件
					if f, h, e := m.R.FormFile(kit.Select("upload", arg, 1)); e == nil {
						defer f.Close()

						// 创建文件
						file := kit.Hashs(f)
						if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
							defer o.Close()
							f.Seek(0, os.SEEK_SET)

							// 导入数据
							if n, e := io.Copy(o, f); m.Assert(e) {
								m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(n), p)
								arg = kit.Simple(arg[0], h.Header.Get("Content-Type"), h.Filename, p, p, n)
							}
						}
					}
				} else if r, ok := m.Optionv("response").(*http.Response); ok {
					// 下载文件
					if buf, e := ioutil.ReadAll(r.Body); m.Assert(e) {
						defer r.Body.Close()

						// 创建文件
						file := kit.Hashs(string(buf))
						if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
							defer o.Close()

							// 导入数据
							if n, e := o.Write(buf); m.Assert(e) {
								m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(int64(n)), p)
								arg = kit.Simple(arg[0], arg[1], arg[2], p, p, n)
							}
						}
					}
				}
				fallthrough
			case "add":
				size := kit.Int(kit.Select(kit.Format(len(arg[3])), arg, 5))
				if arg[0] == "add" && size > 512 {
					// 创建文件
					file := kit.Hashs(arg[3])
					if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
						defer o.Close()

						// 导入数据
						if n, e := o.WriteString(arg[3]); m.Assert(e) {
							m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(int64(n)), p)
							arg = kit.Simple(arg[0], arg[1], arg[2], p, p, n)
						}
					}
				}

				// 添加数据
				h := m.Rich(ice.WEB_CACHE, nil, kit.Dict(
					kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
					kit.MDB_FILE, kit.Select("", arg, 4), kit.MDB_SIZE, size,
				))
				m.Log(ice.LOG_CREATE, "cache: %s %s: %s", h, arg[1], arg[2])

				// 添加记录
				m.Grow(ice.WEB_CACHE, nil, kit.Dict(
					kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
					kit.MDB_SIZE, size, "data", h,
				))

				// 返回结果
				m.Push("time", m.Time())
				m.Push("type", arg[1])
				m.Push("name", arg[2])
				m.Push("text", arg[3])
				m.Push("size", size)
				m.Push("data", h)
			case "watch":
				if m.Richs(ice.WEB_CACHE, nil, arg[1], func(key string, value map[string]interface{}) {
					if value["file"] == "" {
						if f, _, e := kit.Create(arg[2]); m.Assert(e) {
							defer f.Close()
							f.WriteString(kit.Format(value["text"]))
						}
					} else {
						os.MkdirAll(path.Dir(arg[2]), 0777)
						os.Remove(arg[2])
						os.Link(kit.Format(value["file"]), arg[2])
					}
				}) == nil {
					m.Cmdy(ice.WEB_SPIDE, "dev", "cache", "/cache/"+arg[1])
					os.MkdirAll(path.Dir(arg[2]), 0777)
					os.Remove(arg[2])
					os.Link(m.Append("file"), arg[2])
				}
				m.Echo(arg[2])
			}
		}},
		ice.WEB_STORY: {Name: "story story=auto key=auto auto", Help: "故事会", Meta: kit.Dict(
			"exports", []string{"top", "story"}, "detail", []string{"共享", "更新", "推送"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 && arg[0] == "action" {
				story, list := m.Option("story"), m.Option("list")
				switch arg[2] {
				case "story":
					story = arg[3]
				case "list":
					list = arg[3]
				}

				switch arg[1] {
				case "share", "共享":
					if m.Echo("share: "); list == "" {
						msg := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, story)
						m.Cmdy(ice.WEB_SHARE, "add", "story", story, msg.Append("list"))
					} else {
						msg := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, list)
						m.Cmdy(ice.WEB_SHARE, "add", msg.Append("scene"), msg.Append("story"), msg.Append("text"))
					}
				}
				return
			}

			if len(arg) == 0 {
				// 故事列表
				m.Richs(ice.WEB_STORY, "head", "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "story", "count"})
				})
				m.Sort("time", "time_r")
				return
			}

			switch arg[0] {
			case ice.STORY_PULL: // story [spide [story]]
				// 起止节点
				prev, begin, end := "", arg[3], ""
				repos := kit.Keys("remote", arg[2], arg[3])
				m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
					end = kit.Format(kit.Value(val, kit.Keys(repos, "pull", "list")))
					prev = kit.Format(val["list"])
				})

				pull := end
				var first map[string]interface{}
				for begin != "" && begin != end {
					if m.Cmd(ice.WEB_SPIDE, arg[2], "msg", "/story/pull", "begin", begin, "end", end).Table(func(index int, value map[string]string, head []string) {
						if m.Richs(ice.WEB_CACHE, nil, value["data"], nil) == nil {
							m.Log(ice.LOG_IMPORT, "%v: %v", value["data"], value["save"])
							if node := kit.UnMarshal(value["save"]); kit.Format(kit.Value(node, "file")) != "" {
								// 下载文件
								m.Cmd(ice.WEB_SPIDE, arg[2], "cache", "GET", "/story/download/"+value["data"])
							} else {
								// 导入缓存
								m.Conf(ice.WEB_CACHE, kit.Keys("hash", value["data"]), node)
							}
						}

						node := kit.UnMarshal(value["node"]).(map[string]interface{})
						if m.Richs(ice.WEB_STORY, nil, value["list"], nil) == nil {
							// 导入节点
							m.Log(ice.LOG_IMPORT, "%v: %v", value["list"], value["node"])
							m.Conf(ice.WEB_STORY, kit.Keys("hash", value["list"]), node)
						}

						if first == nil {
							if m.Richs(ice.WEB_STORY, "head", arg[1], nil) == nil {
								// 自动创建
								h := m.Rich(ice.WEB_STORY, "head", kit.Dict(
									"scene", node["scene"], "story", arg[1],
									"count", node["count"], "list", value["list"],
								))
								m.Log(ice.LOG_CREATE, "%v: %v", h, node["story"])
							}

							pull, first = kit.Format(value["list"]), node
							m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
								prev = kit.Format(val["list"])
								if kit.Int(node["count"]) > kit.Int(kit.Value(val, kit.Keys(repos, "pull", "count"))) {
									// 更新分支
									m.Log(ice.LOG_IMPORT, "%v: %v", arg[2], pull)
									kit.Value(val, kit.Keys(repos, "pull"), kit.Dict(
										"head", arg[3], "time", node["time"], "list", pull, "count", node["count"],
									))
								}
							})
						}

						if prev == kit.Format(node["prev"]) || prev == kit.Format(node["push"]) {
							// 快速合并
							m.Log(ice.LOG_IMPORT, "%v: %v", pull, arg[2])
							m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
								val["count"] = first["count"]
								val["time"] = first["time"]
								val["list"] = pull
							})
							prev = pull
						}

						begin = kit.Format(node["prev"])
					}).Appendv("list") == nil {
						break
					}
				}

			case ice.STORY_PUSH:
				// 更新分支
				m.Cmdx(ice.WEB_STORY, "pull", arg[1:])

				repos := kit.Keys("remote", arg[2], arg[3])
				// 查询索引
				prev, pull, some, list := "", "", "", ""
				m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
					prev = kit.Format(val["list"])
					pull = kit.Format(kit.Value(val, kit.Keys(repos, "pull", "list")))
					for some = pull; prev != some && some != ""; {
						local := m.Richs(ice.WEB_STORY, nil, prev, nil)
						remote := m.Richs(ice.WEB_STORY, nil, some, nil)
						if diff := kit.Time(kit.Format(remote["time"])) - kit.Time(kit.Format(local["time"])); diff > 0 {
							some = kit.Format(remote["prev"])
						} else if diff < 0 {
							prev = kit.Format(local["prev"])
						}
					}

					if prev = kit.Format(val["list"]); prev == pull {
						// 相同节点
						return
					}

					if some != pull {
						// 合并节点
						local := m.Richs(ice.WEB_STORY, nil, prev, nil)
						remote := m.Richs(ice.WEB_STORY, nil, pull, nil)
						list = m.Rich(ice.WEB_STORY, nil, kit.Dict(
							"scene", val["scene"], "story", val["story"], "count", kit.Int(remote["count"])+1,
							"data", local["data"], "prev", pull, "push", prev,
						))
						m.Log(ice.LOG_CREATE, "merge: %s %s->%s", list, prev, pull)
						val["list"] = list
						prev = list
						val["count"] = kit.Int(remote["count"]) + 1
					}

					// 查询节点
					nodes := []string{}
					for list = prev; list != some; {
						m.Richs(ice.WEB_STORY, nil, list, func(key string, value map[string]interface{}) {
							nodes, list = append(nodes, list), kit.Format(value["prev"])
						})
					}

					for _, v := range kit.Revert(nodes) {
						m.Richs(ice.WEB_STORY, nil, v, func(list string, node map[string]interface{}) {
							m.Richs(ice.WEB_CACHE, nil, node["data"], func(data string, save map[string]interface{}) {
								if kit.Format(save["file"]) != "" {
									// 推送缓存
									m.Cmd(ice.WEB_SPIDE, arg[2], "/story/upload",
										"part", "upload", "@"+kit.Format(save["file"]),
									)
								}

								// 推送节点
								m.Log(ice.LOG_EXPORT, "%s: %s", v, kit.Format(node))
								m.Cmd(ice.WEB_SPIDE, arg[2], "/story/push",
									"story", arg[3], "list", v, "node", kit.Format(node),
									"data", node["data"], "save", kit.Format(save),
								)
							})
						})
					}
				})

				// 更新分支
				m.Cmd(ice.WEB_STORY, "pull", arg[1:])

			case "commit":
				// 查询索引
				head, prev, value, count := "", "", map[string]interface{}{}, 0
				m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
					head, prev, value, count = key, kit.Format(val["list"]), val, kit.Int(val["count"])
					m.Log("info", "head: %v prev: %v count: %v", head, prev, count)
				})

				// 提交信息
				arg[2] = m.Cmdx(ice.WEB_STORY, "add", "submit", arg[2], "hostname,username")

				// 节点信息
				menu := map[string]string{}
				for i := 3; i < len(arg); i++ {
					menu[arg[i]] = m.Cmdx(ice.WEB_STORY, ice.STORY_INDEX, arg[i])
				}

				// 添加节点
				list := m.Rich(ice.WEB_STORY, nil, kit.Dict(
					"scene", "commit", "story", arg[1], "count", count+1, "data", arg[2], "list", menu, "prev", prev,
				))
				m.Log(ice.LOG_CREATE, "commit: %s %s: %s", list, arg[1], arg[2])
				m.Push("list", list)

				if head == "" {
					// 添加索引
					m.Rich(ice.WEB_STORY, "head", kit.Dict("scene", "commit", "story", arg[1], "count", count+1, "list", list))
				} else {
					// 更新索引
					value["count"] = count + 1
					value["time"] = m.Time()
					value["list"] = list
				}
				m.Echo(list)

			case ice.STORY_TRASH:
				bak := kit.Select(kit.Keys(arg[1], "bak"), arg, 2)
				os.Remove(bak)
				os.Rename(arg[1], bak)

			case ice.STORY_WATCH:
				// 备份文件
				name := kit.Select(arg[1], arg, 2)
				m.Cmd(ice.WEB_STORY, ice.STORY_TRASH, name)

				if msg := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, arg[1]); msg.Append("file") != "" {
					p := path.Dir(name)
					os.MkdirAll(p, 0777)

					// 导出文件
					os.Link(msg.Append("file"), name)
					m.Log(ice.LOG_EXPORT, "%s: %s", msg.Append("file"), name)
				} else {
					if f, p, e := kit.Create(name); m.Assert(e) {
						defer f.Close()
						// 导出数据
						f.WriteString(msg.Append("text"))
						m.Log(ice.LOG_EXPORT, "%s: %s", msg.Append("text"), p)
					}
				}
				m.Echo(name)

			case ice.STORY_CATCH:
				if last := m.Richs(ice.WEB_STORY, "head", arg[2], nil); last != nil {
					if t, e := time.ParseInLocation(ice.ICE_TIME, kit.Format(last["time"]), time.Local); e == nil {
						// 文件对比
						if s, e := os.Stat(arg[2]); e == nil && s.ModTime().Before(t) {
							m.Info("%s last: %s", arg[2], kit.Format(t))
							m.Echo("%s", last["list"])
							break
						}
					}
				}
				fallthrough
			case "add", ice.STORY_UPLOAD, ice.STORY_DOWNLOAD:
				if m.Richs(ice.WEB_CACHE, nil, kit.Select("", arg, 3), func(key string, value map[string]interface{}) {
					// 复用缓存
					arg[3] = key
				}) == nil {
					// 添加缓存
					m.Cmdy(ice.WEB_CACHE, arg)
					arg = []string{arg[0], m.Append("type"), m.Append("name"), m.Append("data")}
				}

				// 查询索引
				head, prev, value, count := "", "", map[string]interface{}{}, 0
				m.Richs(ice.WEB_STORY, "head", arg[2], func(key string, val map[string]interface{}) {
					head, prev, value, count = key, kit.Format(val["list"]), val, kit.Int(val["count"])
					m.Log("info", "head: %v prev: %v count: %v", head, prev, count)
				})

				if last := m.Richs(ice.WEB_STORY, nil, prev, nil); prev != "" && last != nil && last["data"] == arg[3] {
					// 重复提交
					m.Echo(prev)
				} else {
					// 添加节点
					list := m.Rich(ice.WEB_STORY, nil, kit.Dict(
						"scene", arg[1], "story", arg[2], "count", count+1, "data", arg[3], "prev", prev,
					))
					m.Log(ice.LOG_CREATE, "story: %s %s: %s", list, arg[1], arg[2])
					m.Push("list", list)

					if head == "" {
						// 添加索引
						m.Rich(ice.WEB_STORY, "head", kit.Dict("scene", arg[1], "story", arg[2], "count", count+1, "list", list))
					} else {
						// 更新索引
						value["count"] = count + 1
						value["time"] = m.Time()
						value["list"] = list
					}
					m.Echo(list)
				}

				// 分发数据
				if p := kit.Select(m.Conf(ice.WEB_FAVOR, "meta.proxy"), m.Option("you")); p != "" {
					m.Info("what %v", p)
					m.Option("you", "")
					m.Cmd(ice.WEB_PROXY, p, ice.WEB_STORY, ice.STORY_PULL, arg[2], "dev", arg[2])
				}

			case ice.STORY_INDEX:
				m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, value map[string]interface{}) {
					// 查询索引
					arg[1] = kit.Format(value["list"])
				})

				m.Richs(ice.WEB_STORY, nil, arg[1], func(key string, value map[string]interface{}) {
					// 查询节点
					m.Push("list", key)
					m.Push(key, value, []string{"scene", "story"})
					arg[1] = kit.Format(value["data"])
				})

				m.Richs(ice.WEB_CACHE, nil, arg[1], func(key string, value map[string]interface{}) {
					// 查询数据
					m.Push("data", key)
					m.Push(key, value, []string{"text", "time", "size", "type", "name", "file"})
					m.Echo("%s", value["text"])
				})

			case ice.STORY_HISTORY:
				// 历史记录
				list := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, arg[1]).Append("list")
				for i := 0; i < kit.Int(kit.Select("30", m.Option("cache.limit"))) && list != ""; i++ {

					m.Richs(ice.WEB_STORY, nil, list, func(key string, value map[string]interface{}) {
						// 直连节点
						m.Push(key, value, []string{"time", "key", "count", "scene", "story"})
						m.Richs(ice.WEB_CACHE, nil, value["data"], func(key string, value map[string]interface{}) {
							m.Push("drama", value["text"])
							m.Push("data", key)
						})

						kit.Fetch(value["list"], func(key string, val string) {
							m.Richs(ice.WEB_STORY, nil, val, func(key string, value map[string]interface{}) {
								// 复合节点
								m.Push(key, value, []string{"time", "key", "count", "scene", "story"})
								m.Richs(ice.WEB_CACHE, nil, value["data"], func(key string, value map[string]interface{}) {
									m.Push("drama", value["text"])
									m.Push("data", key)
								})
							})
						})

						// 切换节点
						list = kit.Format(value["prev"])
					})
				}

			default:
				if len(arg) == 1 {
					// 故事记录
					m.Cmdy(ice.WEB_STORY, "history", arg)
					break
				}
				// 故事详情
				m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, arg[1]).Table(func(index int, value map[string]string, head []string) {
					for k, v := range value {
						m.Push("key", k)
						m.Push("value", v)
					}
					m.Sort("key")
				})
			}
		}},
		ice.WEB_SHARE: {Name: "share share auto", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 共享列表
				m.Grows(ice.WEB_SHARE, nil, "", "", func(index int, value map[string]interface{}) {
					m.Push("", value, []string{kit.MDB_TIME, "share", kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
					m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.link"), m.Conf(ice.WEB_SHARE, "meta.domain"), value["share"], value["share"]))
				})
				return
			}
			if len(arg) == 1 {
				// 共享详情
				if m.Richs(ice.WEB_SHARE, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Push("detail", value)
					m.Push("key", "link")
					m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.link"), m.Conf(ice.WEB_SHARE, "meta.domain"), key, key))
					m.Push("key", "share")
					m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.share"), m.Conf(ice.WEB_SHARE, "meta.domain"), key))
					m.Push("key", "value")
					m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.value"), m.Conf(ice.WEB_SHARE, "meta.domain"), key))
				}) != nil {
					return
				}
			}

			switch arg[0] {
			case "invite":
				arg = []string{arg[0], m.Cmdx(ice.WEB_SHARE, "add", "invite", kit.Select("tech", arg, 1), kit.Select("miss", arg, 2))}

				fallthrough
			case "check":
				m.Richs(ice.WEB_SHARE, nil, arg[1], func(key string, value map[string]interface{}) {
					m.Render(ice.RENDER_QRCODE, kit.Format(kit.Dict(
						kit.MDB_TYPE, "share", kit.MDB_NAME, value["type"], kit.MDB_TEXT, key,
					)))
				})

			case "auth":
				m.Richs(ice.WEB_SHARE, nil, arg[1], func(key string, value map[string]interface{}) {
					switch value["type"] {
					case "active":
						m.Cmdy(ice.WEB_SPACE, value["name"], "sessid", m.Cmdx(ice.AAA_SESS, "create", arg[2]))
					case "user":
						m.Cmdy(ice.AAA_ROLE, arg[2], value["name"])
					default:
						m.Cmdy(ice.AAA_SESS, "auth", value["text"], arg[2])
					}
				})

			case "add":
				arg = arg[1:]
				fallthrough
			default:
				if len(arg) == 2 {
					arg = append(arg, "")
				}
				extra := kit.Dict(arg[3:])

				// 创建共享
				h := m.Rich(ice.WEB_SHARE, nil, kit.Dict(
					kit.MDB_TIME, m.Time("10m"),
					kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2],
					"extra", extra,
				))
				// 创建列表
				m.Grow(ice.WEB_SHARE, nil, kit.Dict(
					kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2],
					"share", h,
				))
				m.Log(ice.LOG_CREATE, "share: %s extra: %s", h, kit.Format(extra))
				m.Echo(h)
			}
		}},

		ice.WEB_ROUTE: {Name: "route name auto", Help: "路由", Meta: kit.Dict("detail", []string{"分组"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 && arg[0] == "action" {
				switch arg[1] {
				case "group", "分组":
					if m.Option("grp") != "" && m.Option("name") != "" {
						m.Cmdy(ice.WEB_GROUP, m.Option("grp"), "add", m.Option("name"))
					}
				}
				return
			}

			target, rest := "*", ""
			if len(arg) > 0 {
				ls := strings.SplitN(arg[0], ".", 2)
				if target = ls[0]; len(ls) > 1 {
					rest = ls[1]
				}
			}
			m.Richs(ice.WEB_SPACE, nil, target, func(key string, value map[string]interface{}) {
				if len(arg) > 1 {
					ls := []interface{}{ice.WEB_SPACE, value[kit.MDB_NAME]}
					m.Call(false, func(res *ice.Message) *ice.Message { return res })
					// 发送命令
					if rest != "" {
						ls = append(ls, ice.WEB_SPACE, rest)
					}
					m.Cmdy(ls, arg[1:])
					return
				}

				switch value[kit.MDB_TYPE] {
				case ice.WEB_SERVER:
					if value[kit.MDB_NAME] == m.Conf(ice.CLI_RUNTIME, "node.name") {
						// 避免循环
						return
					}

					// 远程查询
					m.Cmd(ice.WEB_SPACE, value[kit.MDB_NAME], ice.WEB_ROUTE).Table(func(index int, val map[string]string, head []string) {
						m.Push(kit.MDB_TYPE, val[kit.MDB_TYPE])
						m.Push(kit.MDB_NAME, kit.Keys(value[kit.MDB_NAME], val[kit.MDB_NAME]))
					})
					fallthrough
				case ice.WEB_WORKER:
					// 本机查询
					m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
					m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
				}
			})
		}},
		ice.WEB_PROXY: {Name: "proxy", Help: "代理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(ice.WEB_SPACE, nil, kit.Select(m.Conf(ice.WEB_FAVOR, "meta.proxy"), arg[0]), func(key string, value map[string]interface{}) {
				if value[kit.MDB_TYPE] == ice.WEB_BETTER {
					switch value[kit.MDB_NAME] {
					case "tmux":
						m.Cmd("web.code.tmux.session").Table(func(index int, value map[string]string, head []string) {
							if value["tag"] == "1" {
								m.Log(ice.LOG_SELECT, "space: %s", value["session"])
								arg[0] = value["session"]
							}
						})
					}
				}
			})

			m.Cmdy(ice.WEB_ROUTE, arg[0], arg[1:])
		}},
		ice.WEB_GROUP: {Name: "group group name auto", Help: "分组", Meta: kit.Dict(
			"exports", []string{"grp", "group"}, "detail", []string{"标签", "退还"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 && arg[0] == "action" {
				switch arg[1] {
				case "label", "标签":
					if m.Option("lab") != "" && m.Option("group") != "" {
						m.Cmdy(ice.WEB_LABEL, m.Option("lab"), "add", m.Option("group"), m.Option("name"))
					}
				case "del", "退还":
					if m.Option("group") != "" && m.Option("name") != "" {
						m.Cmdy(ice.WEB_GROUP, m.Option("group"), "del", m.Option("name"))
					}
				}
				return
			}

			if len(arg) < 2 {
				// 分组列表
				m.Richs(cmd, nil, kit.Select("*", arg, 0), func(key string, value map[string]interface{}) {
					if len(arg) < 1 {
						m.Push(key, value[kit.MDB_META])
						return
					}
					m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), "*", func(key string, value map[string]interface{}) {
						m.Push(key, value)
					})
				})
				m.Logs(ice.LOG_SELECT, cmd, m.Format(ice.MSG_APPEND))
				return
			}

			if m.Richs(cmd, nil, arg[0], nil) == nil {
				// 添加分组
				n := m.Rich(cmd, nil, kit.Data(kit.MDB_SHORT, kit.MDB_NAME, cmd, arg[0]))
				m.Logs(ice.LOG_CREATE, cmd, n)
			}

			m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
				if m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[1], func(key string, value map[string]interface{}) {
					// 分组详情
					m.Push("detail", value)
				}) != nil {
					return
				}

				switch arg[1] {
				case "add":
					if m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[2], func(key string, value map[string]interface{}) {
						if value[kit.MDB_STATUS] == "void" {
							value[kit.MDB_STATUS] = "free"
						}
						m.Logs(ice.LOG_MODIFY, cmd, key, kit.MDB_NAME, arg[2], kit.MDB_STATUS, value[kit.MDB_STATUS])
					}) == nil {
						m.Logs(ice.LOG_INSERT, cmd, key, kit.MDB_NAME, arg[2])
						m.Rich(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(kit.MDB_NAME, arg[2], kit.MDB_STATUS, "free"))
					}
					m.Echo(arg[0])
				case "del":
					m.Logs(ice.LOG_MODIFY, cmd, key, kit.MDB_NAME, arg[2], kit.MDB_STATUS, "void")
					m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[2], func(sub string, value map[string]interface{}) {
						value[kit.MDB_STATUS] = "void"
					})
				case "get":
					m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), kit.Select("%", arg, 2), func(sub string, value map[string]interface{}) {
						m.Logs(ice.LOG_MODIFY, cmd, key, kit.MDB_NAME, value[kit.MDB_NAME], kit.MDB_STATUS, "busy")
						value[kit.MDB_STATUS] = "busy"
						m.Echo("%s", value[kit.MDB_NAME])
					})
				case "put":
					m.Logs(ice.LOG_MODIFY, cmd, key, kit.MDB_NAME, arg[2], kit.MDB_STATUS, "free")
					m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[2], func(sub string, value map[string]interface{}) {
						value[kit.MDB_STATUS] = "free"
					})
				default:
					m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), "*", func(key string, value map[string]interface{}) {
						// 执行命令
						m.Cmdy(ice.WEB_PROXY, value["name"], arg[1:])
					})
				}
			})
		}},
		ice.WEB_LABEL: {Name: "label label name auto", Help: "标签", Meta: kit.Dict(
			"exports", []string{"lab", "label"}, "detail", []string{"归还"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 && arg[0] == "action" {
				switch arg[1] {
				case "del", "归还":
					if m.Option("label") != "" && m.Option("name") != "" {
						m.Cmdy(ice.WEB_LABEL, m.Option("label"), "del", m.Option("name"))
					}
				}
				return
			}

			if len(arg) < 3 {
				m.Richs(cmd, nil, kit.Select("*", arg, 0), func(key string, value map[string]interface{}) {
					if len(arg) == 0 {
						// 分组列表
						m.Push(key, value[kit.MDB_META])
						return
					}
					m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), kit.Select("*", arg, 1), func(key string, value map[string]interface{}) {
						if len(arg) == 1 {
							// 设备列表
							m.Push(key, value)
							return
						}
						// 设备详情
						m.Push("detail", value)
					})
				})
				m.Logs(ice.LOG_SELECT, cmd, m.Format(ice.MSG_APPEND))
				return
			}

			if m.Richs(cmd, nil, arg[0], nil) == nil {
				// 添加分组
				m.Logs(ice.LOG_CREATE, cmd, m.Rich(cmd, nil, kit.Data(kit.MDB_SHORT, kit.MDB_NAME, cmd, arg[0])))
			}

			m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
				switch arg[1] {
				case "add":
					if pod := m.Cmdx(ice.WEB_GROUP, arg[2], "get", arg[3:]); pod != "" {
						if m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), pod, func(key string, value map[string]interface{}) {
							if value[kit.MDB_STATUS] == "void" {
								// 更新设备
								value[kit.MDB_STATUS] = "free"
								m.Logs(ice.LOG_MODIFY, cmd, key, kit.MDB_NAME, pod, kit.MDB_STATUS, value[kit.MDB_STATUS])
							}
						}) == nil {
							// 获取设备
							m.Logs(ice.LOG_INSERT, cmd, key, kit.MDB_NAME, pod)
							m.Rich(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(kit.MDB_NAME, pod, "group", arg[2], kit.MDB_STATUS, "free"))
						}
					}
					m.Echo(arg[0])

				case "del":
					m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[2], func(sub string, value map[string]interface{}) {
						// 归还设备
						m.Cmdx(ice.WEB_GROUP, value["group"], "put", arg[2])
						m.Logs(ice.LOG_MODIFY, cmd, key, kit.MDB_NAME, arg[2], kit.MDB_STATUS, "void")
						value[kit.MDB_STATUS] = "void"
						m.Echo(arg[2])
					})
				default:
					wg := &sync.WaitGroup{}
					m.Option("_async", "true")
					m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[1], func(key string, value map[string]interface{}) {
						wg.Add(1)
						// 远程命令
						m.Option("user.pod", value["name"])
						m.Cmd(ice.WEB_PROXY, value["name"], arg[2:]).Call(false, func(res *ice.Message) *ice.Message {
							if wg.Done(); res != nil && m != nil {
								m.Copy(res)
							}
							return nil
						})
					})
					wg.Wait()
				}
			})
		}},

		"/route/": {Name: "/route/", Help: "路由器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "login":
				if m.Option(ice.MSG_USERNAME) != "" {
					m.Push(ice.MSG_USERNAME, m.Option(ice.MSG_USERNAME))
					m.Info("username: %v", m.Option(ice.MSG_USERNAME))
					break
				}
				if m.Option(ice.MSG_SESSID) != "" && m.Cmdx(ice.AAA_SESS, "check", m.Option(ice.MSG_SESSID)) != "" {
					m.Info("sessid: %v", m.Option(ice.MSG_SESSID))
					break
				}

				sessid := m.Cmdx(ice.AAA_SESS, "create", "")
				share := m.Cmdx(ice.WEB_SHARE, "add", "login", m.Option(ice.MSG_USERIP), sessid)
				Render(m, "cookie", sessid)
				m.Render(share)
			}
		}},
		"/space/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if s, e := websocket.Upgrade(m.W, m.R, nil, m.Confi(ice.WEB_SPACE, "meta.buffer.r"), m.Confi(ice.WEB_SPACE, "meta.buffer.w")); m.Assert(e) {
				m.Option("name", strings.Replace(kit.Select(m.Option(ice.MSG_USERADDR), m.Option("name")), ".", "_", -1))
				m.Option("node", kit.Select("worker", m.Option("node")))

				// 共享空间
				share := m.Option("share")
				if m.Richs(ice.WEB_SHARE, nil, share, nil) == nil {
					share = m.Cmdx(ice.WEB_SHARE, "add", m.Option("node"), m.Option("name"), m.Option("user"))
				}
				// m.Cmd(ice.WEB_GROUP, m.Option("group"), "add", m.Option("name"))

				// 添加节点
				h := m.Rich(ice.WEB_SPACE, nil, kit.Dict(
					kit.MDB_TYPE, m.Option("node"),
					kit.MDB_NAME, m.Option("name"),
					kit.MDB_TEXT, m.Option("user"),
					"share", share, "socket", s,
				))
				m.Log(ice.LOG_CREATE, "space: %s share: %s", m.Option(kit.MDB_NAME), share)

				m.Gos(m, func(m *ice.Message) {
					// 监听消息
					m.Event(ice.SPACE_START, m.Option("node"), m.Option("name"))
					m.Target().Server().(*Frame).HandleWSS(m, false, s, m.Option("name"))
					m.Log(ice.LOG_CLOSE, "%s: %s", m.Option(kit.MDB_NAME), kit.Format(m.Confv(ice.WEB_SPACE, kit.Keys(kit.MDB_HASH, h))))
					m.Event(ice.SPACE_CLOSE, m.Option("node"), m.Option("name"))
					m.Confv(ice.WEB_SPACE, kit.Keys(kit.MDB_HASH, h), "")
				})

				// 共享空间
				if share != m.Option("share") {
					m.Cmd(ice.WEB_SPACE, m.Option("name"), ice.WEB_SPACE, "share", share)
				}
				m.Echo(share)
			}
		}},

		"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "local":
				m.Render(ice.RENDER_DOWNLOAD, m.Cmdx(arg[1], path.Join(arg[2:]...)))
				return
			}

			m.Richs(ice.WEB_SHARE, nil, arg[0], func(key string, value map[string]interface{}) {
				m.Log(ice.LOG_EXPORT, "%s: %v", arg, kit.Format(value))
				if m.Option(ice.MSG_USERROLE) != ice.ROLE_ROOT && kit.Time(kit.Format(value[kit.MDB_TIME])) < kit.Time(m.Time()) {
					m.Echo("invalid")
					return
				}

				switch value["type"] {
				case ice.TYPE_SPACE:
				case ice.TYPE_STORY:
					// 查询数据
					msg := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, value["text"])
					if msg.Append("text") == "" && kit.Value(value, "extra.pod") != "" {
						msg = m.Cmd(ice.WEB_SPACE, kit.Value(value, "extra.pod"), ice.WEB_STORY, ice.STORY_INDEX, value["text"])
					}
					value = kit.Dict("type", msg.Append("scene"), "name", msg.Append("story"), "text", msg.Append("text"), "file", msg.Append("file"))
					m.Log(ice.LOG_EXPORT, "%s: %v", arg, kit.Format(value))
				}

				switch kit.Select("", arg, 1) {
				case "download", "下载":
					if strings.HasPrefix(kit.Format(value["text"]), m.Conf(ice.WEB_CACHE, "meta.path")) {
						m.Render(ice.RENDER_DOWNLOAD, value["text"], value["type"], value["name"])
					} else {
						m.Render("%s", value["text"])
					}
					return
				case "detail", "详情":
					m.Render(kit.Formats(value))
					return
				case "share", "共享码":
					m.Render(ice.RENDER_QRCODE, kit.Format("%s/share/%s/", m.Conf(ice.WEB_SHARE, "meta.domain"), key))
					return
				case "check", "安全码":
					m.Render(ice.RENDER_QRCODE, kit.Format(kit.Dict(
						kit.MDB_TYPE, "share", kit.MDB_NAME, value["type"], kit.MDB_TEXT, key,
					)))
					return
				case "value", "数据值":
					m.Render(ice.RENDER_QRCODE, kit.Format(value), kit.Select("256", arg, 2))
					return
				case "text":
					m.Render(ice.RENDER_QRCODE, kit.Format(value["text"]))
					return
				}

				switch value["type"] {
				case ice.TYPE_RIVER:
					// 共享群组
					m.Render("redirect", "/", "share", key, "river", kit.Format(value["text"]))

				case ice.TYPE_STORM:
					// 共享应用
					m.Render("redirect", "/", "share", key, "storm", kit.Format(value["text"]), "river", kit.Format(kit.Value(value, "extra.river")))

				case ice.TYPE_ACTION:
					if len(arg) == 1 {
						// 跳转主页
						m.Render("redirect", "/share/"+arg[0]+"/", "title", kit.Format(value["name"]))
						break
					}

					if arg[1] == "" {
						// 返回主页
						Render(m, ice.RENDER_DOWNLOAD, m.Conf(ice.WEB_SERVE, "meta.page.share"))
						break
					}

					if len(arg) == 2 {
						// 应用列表
						value["count"] = kit.Int(value["count"]) + 1
						kit.Fetch(kit.Value(value, "extra.tool"), func(index int, value map[string]interface{}) {
							m.Push("river", arg[0])
							m.Push("storm", arg[1])
							m.Push("action", index)

							m.Push("node", value["pod"])
							m.Push("group", value["ctx"])
							m.Push("index", value["cmd"])
							m.Push("args", value["args"])

							msg := m.Cmd(m.Space(value["pod"]), ice.CTX_COMMAND, value["ctx"], value["cmd"])
							m.Push("name", value["cmd"])
							m.Push("help", kit.Select(msg.Append("help"), kit.Format(value["help"])))
							m.Push("inputs", msg.Append("list"))
							m.Push("feature", msg.Append("meta"))
						})
						break
					}

					// 默认参数
					meta := kit.Value(value, kit.Format("extra.tool.%s", arg[2])).(map[string]interface{})
					if meta["single"] == "yes" && kit.Select("", arg, 3) != "action" {
						arg = append(arg[:3], kit.Simple(kit.UnMarshal(kit.Format(meta["args"])))...)
						for i := len(arg) - 1; i >= 0; i-- {
							if arg[i] != "" {
								break
							}
							arg = arg[:i]
						}
					}

					// 执行命令
					cmds := kit.Simple(m.Space(meta["pod"]), kit.Keys(meta["ctx"], meta["cmd"]), arg[3:])
					m.Cmdy(cmds).Option("cmds", cmds)
					m.Option("title", value["name"])

				default:
					// 查看数据
					m.Option("type", value["type"])
					m.Option("name", value["name"])
					m.Option("text", value["text"])
					m.Render(ice.RENDER_TEMPLATE, m.Conf(ice.WEB_SHARE, "meta.template.simple"))
					m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
				}
			})
		}},
		"/cache/": {Name: "/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(ice.WEB_CACHE, nil, arg[0], func(key string, value map[string]interface{}) {
				m.Render(ice.RENDER_DOWNLOAD, value["file"])
			})
		}},
		"/story/": {Name: "/story/", Help: "故事会", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			switch arg[0] {
			case ice.STORY_PULL:
				// 下载节点

				list := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, m.Option("begin")).Append("list")
				for i := 0; i < 10 && list != "" && list != m.Option("end"); i++ {
					if m.Richs(ice.WEB_STORY, nil, list, func(key string, value map[string]interface{}) {
						// 节点信息
						m.Push("list", key)
						m.Push("node", kit.Format(value))
						m.Push("data", value["data"])
						m.Push("save", kit.Format(m.Richs(ice.WEB_CACHE, nil, value["data"], nil)))
						list = kit.Format(value["prev"])
					}) == nil {
						break
					}
				}
				m.Log(ice.LOG_EXPORT, "%s %s", m.Option("begin"), m.Format("append"))

			case ice.STORY_PUSH:
				// 上传节点

				if m.Richs(ice.WEB_CACHE, nil, m.Option("data"), nil) == nil {
					// 导入缓存
					m.Log(ice.LOG_IMPORT, "%v: %v", m.Option("data"), m.Option("save"))
					m.Conf(ice.WEB_CACHE, kit.Keys("hash", m.Option("data")), kit.UnMarshal(m.Option("save")))
				}

				node := kit.UnMarshal(m.Option("node")).(map[string]interface{})
				if m.Richs(ice.WEB_STORY, nil, m.Option("list"), nil) == nil {
					// 导入节点
					m.Log(ice.LOG_IMPORT, "%v: %v", m.Option("list"), m.Option("node"))
					m.Conf(ice.WEB_STORY, kit.Keys("hash", m.Option("list")), node)
				}

				if head := m.Richs(ice.WEB_STORY, "head", m.Option("story"), nil); head == nil {
					// 自动创建
					h := m.Rich(ice.WEB_STORY, "head", kit.Dict(
						"scene", node["scene"], "story", m.Option("story"),
						"count", node["count"], "list", m.Option("list"),
					))
					m.Log(ice.LOG_CREATE, "%v: %v", h, m.Option("story"))
				} else if head["list"] == kit.Format(node["prev"]) || head["list"] == kit.Format(node["pull"]) {
					// 快速合并
					head["list"] = m.Option("list")
					head["count"] = node["count"]
					head["time"] = node["time"]
				} else {
					// 推送失败
				}

			case ice.STORY_UPLOAD:
				// 上传数据
				m.Cmdy(ice.WEB_CACHE, "upload")

			case ice.STORY_DOWNLOAD:
				// 下载数据
				m.Cmdy(ice.WEB_STORY, ice.STORY_INDEX, arg[1])
				m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append("file") == ""), m.Append("text"))
			}
		}},
		"/favor/": {Name: "/story/", Help: "收藏夹", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			switch arg[0] {
			case "pull":
				m.Richs(ice.WEB_FAVOR, nil, m.Option("favor"), func(key string, value map[string]interface{}) {
					m.Option("cache.limit", kit.Int(kit.Value(value, "meta.count"))+1-kit.Int(m.Option("begin")))
					m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
						m.Log(ice.LOG_EXPORT, "%v", value)
						m.Push("", value, []string{"id", "type", "name", "text"})
						m.Push("extra", kit.Format(value["extra"]))
					})
				})
			case "push":
				m.Cmdy(ice.WEB_FAVOR, m.Option("favor"), m.Option("type"), m.Option("name"), m.Option("text"), m.Option("extra"))
			}
		}},

		"/plugin/github.com/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := m.Conf(ice.WEB_SERVE, "meta.volcanos.require")
			if _, e := os.Stat(path.Join(prefix, cmd)); e != nil {
				m.Cmd(ice.CLI_SYSTEM, "git", "clone", "https://"+strings.Join(strings.Split(cmd, "/")[2:5], "/"),
					path.Join(prefix, strings.Join(strings.Split(cmd, "/")[1:5], "/")))
			}
			m.Render(ice.RENDER_DOWNLOAD, path.Join(prefix, cmd))
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
