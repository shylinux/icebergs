package web

import (
	"github.com/gorilla/websocket"
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
	"github.com/skip2/go-qrcode"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"
)

var SERVE = ice.Name("serve", Index)

type Frame struct {
	*http.Client
	*http.Server
	*http.ServeMux
	m *ice.Message

	send map[string]*ice.Message
}

func Count(m *ice.Message, cmd, key, name string) int {
	count := kit.Int(m.Conf(cmd, kit.Keys(key, name)))
	m.Conf(cmd, kit.Keys(key, name), count+1)
	return count
}
func Format(key string, arg ...interface{}) string {
	switch args := kit.Simple(arg); key {
	case "a":
		return fmt.Sprintf("<a href='%s' target='_blank'>%s</a>", kit.Format(args[0]), kit.Select(kit.Format(args[0]), args, 1))
	}
	return ""
}
func Render(msg *ice.Message, cmd string, args ...interface{}) {
	if cmd != "" {
		msg.Log(ice.LOG_EXPORT, "%s: %v", cmd, args)
	}
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

	msg.Log_AUTH("ip", ip)
	if msg.Richs(SERVE, kit.Keys("meta.white"), ip, nil) != nil {
		msg.Log_AUTH("ip", ip)
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
		msg.Logs(ice.LOG_AUTH, "role", msg.Option(ice.MSG_USERROLE, sub.Append("userrole")),
			"user", msg.Option(ice.MSG_USERNAME, sub.Append("username")))
	}

	if !msg.Options(ice.MSG_USERNAME) && IsLocalIP(msg, msg.Option(ice.MSG_USERIP)) {
		// 自动认证
		msg.Option(ice.MSG_USERNAME, msg.Conf(ice.CLI_RUNTIME, "boot.username"))
		msg.Option(ice.MSG_USERROLE, msg.Cmdx(ice.AAA_ROLE, "check", msg.Option(ice.MSG_USERNAME)))
		if strings.HasPrefix(msg.Option(ice.MSG_USERUA), "Mozilla/5.0") {
			msg.Option(ice.MSG_SESSID, msg.Cmdx(ice.AAA_SESS, "create", msg.Option(ice.MSG_USERNAME), msg.Option(ice.MSG_USERROLE)))
			msg.Render("cookie", msg.Option(ice.MSG_SESSID))
		}
		msg.Logs(ice.LOG_AUTH, "role", msg.Option(ice.MSG_USERROLE), "user", msg.Option(ice.MSG_USERNAME), "sess", msg.Option(ice.MSG_SESSID))
	}

	if s, ok := msg.Target().Commands[ice.WEB_LOGIN]; ok {
		// 权限检查
		msg.Target().Run(msg, s, ice.WEB_LOGIN, kit.Simple(msg.Optionv("cmds"))...)
	} else if ls := strings.Split(msg.Option(ice.MSG_USERURL), "/"); kit.IndexOf([]string{
		"static", "plugin", "login", "space", "route", "share",
		"publish",
	}, ls[1]) > -1 {

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
			// 解析失败
			break
		} else {
			socket, msg := c, m.Spawns(b)
			target := kit.Simple(msg.Optionv(ice.MSG_TARGET))
			source := kit.Simple(msg.Optionv(ice.MSG_SOURCE), name)
			msg.Info("recv %v<-%v %s %v", target, source, msg.Detailv(), msg.Format("meta"))

			if len(target) == 0 {
				msg.Option(ice.MSG_USERROLE, msg.Cmdx(ice.AAA_ROLE, "check", msg.Option(ice.MSG_USERNAME)))
				msg.Logs(ice.LOG_AUTH, "role", msg.Option(ice.MSG_USERROLE), "user", msg.Option(ice.MSG_USERNAME))
				if msg.Optionv(ice.MSG_HANDLE, "true"); !msg.Warn(!safe, "no right") {
					// 本地执行
					m.Option("_dev", name)
					msg = msg.Cmd()
				}
				if source, target = []string{}, kit.Revert(source)[1:]; msg.Detail() == "exit" {
					// 重启进程
					return true
				}
			} else if msg.Richs(ice.WEB_SPACE, nil, target[0], func(key string, value map[string]interface{}) {
				// 查询节点
				if s, ok := value["socket"].(*websocket.Conn); ok {
					socket, source, target = s, source, target[1:]
				} else {
					socket, source, target = s, source, target[1:]
				}
			}) != nil {
				// 转发报文

			} else if res, ok := web.send[msg.Option(ice.MSG_TARGET)]; len(target) == 1 && ok {
				// 接收响应
				delete(web.send, msg.Option(ice.MSG_TARGET))
				res.Cost("%v->%v %v %v", target, source, res.Detailv(), msg.Format("append"))
				res.Back(msg)
				continue

			} else if msg.Warn(msg.Option("_handle") == "true", "space miss") {
				// 回复失败
				continue

			} else {
				// 下发失败
				msg.Warn(true, "space error")
				source, target = []string{}, kit.Revert(source)[1:]
			}

			// 发送报文
			msg.Optionv(ice.MSG_SOURCE, source)
			msg.Optionv(ice.MSG_TARGET, target)
			socket.WriteMessage(t, []byte(msg.Format("meta")))
			target = append([]string{name}, target...)
			msg.Info("send %v %v->%v %v %v", t, source, target, msg.Detailv(), msg.Format("meta"))
			msg.Cost("%v->%v %v %v", source, target, msg.Detailv(), msg.Format("append"))
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
			if u, e := url.Parse(r.Header.Get("Referer")); e == nil {
				for k, v := range u.Query() {
					msg.Info("%s: %v", k, v)
					msg.Option(k, v)
				}
			}

			// 用户请求
			msg.Option(ice.MSG_USERWEB, m.Conf(ice.WEB_SHARE, "meta.domain"))
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
			if msg.Option(ice.MSG_USERPOD, msg.Option("pod")); msg.Optionv("cmds") == nil {
				if p := strings.TrimPrefix(msg.Option(ice.MSG_USERURL), key); p != "" {
					msg.Optionv("cmds", strings.Split(p, "/"))
				}
			}

			// 执行命令
			if cmds := kit.Simple(msg.Optionv("cmds")); web.Login(msg, w, r) {
				msg.Option("_option", msg.Optionv(ice.MSG_OPTION))
				msg.Target().Run(msg, cmd, msg.Option(ice.MSG_USERURL), cmds...)
			}

			// 渲染引擎
			_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
			Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
		})
	})
}
func (web *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m, index := web.m, r.Header.Get("index.module") == ""
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
		Render(m, "refresh", m.Conf(ice.WEB_SERVE, "meta.volcanos.refresh"))
		m.Event(ice.SYSTEM_INIT)
		m.W = nil
	} else if r.URL.Path == "/share" && r.Method == "GET" {
		http.ServeFile(w, r, m.Conf(ice.WEB_SERVE, "meta.page.share"))

		// } else if r.URL.Path == "/" && r.Method == "GET" {
		// 	http.ServeFile(w, r, m.Conf(ice.WEB_SERVE, "meta.page.index"))
		//
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
			"title", "github.com/shylinux/contexts",
			"legal", []interface{}{`<a href="mailto:shylinuxc@gmail.com">shylinuxc@gmail.com</a>`},
			"page", kit.Dict(
				"index", "usr/volcanos/page/index.html",
				"share", "usr/volcanos/page/share.html",
			),
			"static", kit.Dict("/", "usr/volcanos/"),
			"publish", "usr/publish/",
			"volcanos", kit.Dict("path", "usr/volcanos", "branch", "master",
				"repos", "https://github.com/shylinux/volcanos",
				"require", ".ish/pluged",
				"refresh", "5",
			),
			"template", kit.Dict("path", "usr/template", "list", []interface{}{
				`{{define "raw"}}{{.Result}}{{end}}`,
			}),
			"logheaders", "false", "init", "false",
			"black", kit.Dict(),
		)},
		ice.WEB_DREAM: {Name: "dream", Help: "梦想家", Value: kit.Data("path", "usr/local/work",
			// "cmd", []interface{}{ice.CLI_SYSTEM, "ice.sh", "start", ice.WEB_SPACE, "connect"},
			"cmd", []interface{}{ice.CLI_SYSTEM, "ice.bin", ice.WEB_SPACE, "connect"},
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			m.Cmd(ice.WEB_SPIDE, "add", "self", kit.Select("http://:9020", m.Conf(ice.CLI_RUNTIME, "conf.ctx_self")))
			m.Cmd(ice.WEB_SPIDE, "add", "dev", kit.Select("http://:9020", m.Conf(ice.CLI_RUNTIME, "conf.ctx_dev")))
			m.Cmd(ice.WEB_SPIDE, "add", "shy", kit.Select("https://shylinux.com:443", m.Conf(ice.CLI_RUNTIME, "conf.ctx_shy")))

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
					m.Gos(m, func(m *ice.Message) {
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							if favor == arg[0] || value["type"] == arg[0] ||
								strings.Contains(kit.Format(value["name"]), arg[0]) || strings.Contains(kit.Format(value["text"]), arg[0]) {
								m.Push("pod", m.Option(ice.MSG_USERPOD))
								m.Push("engine", "favor")
								m.Push("favor", favor)
								m.Push("", value, []string{"id", "time", "type", "name", "text"})
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

						m.Push("time", val["time"])
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

						m.Push("time", value["time"])
						m.Push("type", value["type"])
						m.Push("name", value["name"])
						m.Push("text", value["text"])
					}
				})
			}}))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(ice.WEB_SPIDE, ice.WEB_SERVE, ice.WEB_GROUP, ice.WEB_LABEL,
				ice.WEB_FAVOR, ice.WEB_CACHE, ice.WEB_STORY, ice.WEB_SHARE)

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

			m.Watch(ice.SYSTEM_INIT, "web.code.git.repos", "volcanos", m.Conf(ice.WEB_SERVE, "meta.volcanos.path"),
				m.Conf(ice.WEB_SERVE, "meta.volcanos.repos"), m.Conf(ice.WEB_SERVE, "meta.volcanos.branch"))
			m.Conf(ice.WEB_FAVOR, "meta.template", favor_template)
			m.Conf(ice.WEB_SHARE, "meta.template", share_template)
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
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
