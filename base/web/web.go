package web

import (
	"github.com/gorilla/websocket"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
	"github.com/skip2/go-qrcode"

	"bytes"
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

func Refresh(msg *ice.Message, n int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<head>
	<meta charset="utf-8">
	<meta http-equiv="Refresh" content="%d">
</head>
<body>
	请稍后，系统初始化中...
</body>
	`, n)
}

func Redirect(msg *ice.Message, url string, arg ...interface{}) *ice.Message {
	msg.Push("_output", "redirect")
	msg.Echo(kit.MergeURL(url, arg...))
	return msg
}
func Cookie(msg *ice.Message, sessid string) string {
	expire := time.Now().Add(kit.Duration(msg.Conf(ice.AAA_SESS, "meta.expire")))
	msg.Log("cookie", "expire:%v sessid:%s", kit.Format(expire), sessid)
	http.SetCookie(msg.W, &http.Cookie{Name: ice.WEB_SESS, Value: sessid, Path: "/", Expires: expire})
	return sessid
}
func Upload(m *ice.Message, name string) string {
	m.Cmdy(ice.WEB_STORY, "upload")
	if s, e := os.Stat(name); e == nil && s.IsDir() {
		name = path.Join(name, m.Append("name"))
	}
	m.Cmd(ice.WEB_STORY, ice.STORY_WATCH, m.Append("data"), name)
	return name
}
func Count(m *ice.Message, cmd, key, name string) int {
	count := kit.Int(m.Conf(cmd, kit.Keys(key, name)))
	m.Conf(cmd, kit.Keys(key, name), count+1)
	return count
}
func IsLocalIP(ip string) bool {
	return ip == "::1" || strings.HasPrefix(ip, "127.")
}

func (web *Frame) Login(msg *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if msg.Options(ice.WEB_SESS) {
		// 会话认证
		sub := msg.Cmd(ice.AAA_SESS, "check", msg.Option(ice.WEB_SESS))
		msg.Info("role: %s user: %s", msg.Option(ice.MSG_USERROLE, sub.Append("userrole")),
			msg.Option(ice.MSG_USERNAME, sub.Append("username")))
	}
	if strings.HasPrefix(msg.Option(ice.MSG_USERURL), "/space/") {
		return true
	}

	if (!msg.Options(ice.MSG_SESSID) || !msg.Options(ice.MSG_USERNAME)) && IsLocalIP(msg.Option(ice.MSG_USERIP)) {
		// 自动认证
		msg.Option(ice.MSG_USERNAME, msg.Conf(ice.CLI_RUNTIME, "boot.username"))
		msg.Option(ice.MSG_USERROLE, msg.Cmdx(ice.AAA_ROLE, "check", msg.Option(ice.MSG_USERNAME)))
		msg.Option(ice.MSG_SESSID, msg.Rich(ice.AAA_SESS, nil, kit.Dict(
			"username", msg.Option(ice.MSG_USERNAME), "userrole", msg.Option(ice.MSG_USERROLE),
		)))
		Cookie(msg, msg.Option(ice.MSG_SESSID))
		msg.Info("user: %s role: %s sess: %s", msg.Option(ice.MSG_USERNAME), msg.Option(ice.MSG_USERROLE), msg.Option(ice.MSG_SESSID))
	}

	if s, ok := msg.Target().Commands[ice.WEB_LOGIN]; ok {
		// 权限检查
		msg.Target().Run(msg, s, ice.WEB_LOGIN, kit.Simple(msg.Optionv("cmds"))...)
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
				msg.Info("recv %v %v->%v %v", t, source, target, msg.Format("meta"))

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
					msg.Info("space done")
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
			defer func() { msg.Cost("%s %v", r.URL.Path, msg.Optionv("cmds")) }()

			// 解析请求
			msg.Option(ice.MSG_USERUA, r.Header.Get("User-Agent"))
			msg.Option(ice.MSG_USERIP, r.Header.Get(ice.MSG_USERIP))
			msg.Option(ice.MSG_USERURL, r.URL.Path)
			msg.Option(ice.MSG_USERNAME, "")
			msg.Option(ice.WEB_SESS, "")
			msg.R, msg.W = r, w

			// 请求环境
			for _, v := range r.Cookies() {
				if v.Value != "" {
					msg.Option(v.Name, v.Value)
				}
			}

			// 请求数据
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
				msg.Optionv(k, v)
				if k == ice.MSG_SESSID {
					Cookie(msg, v[0])
				}
			}

			if !web.Login(msg, w, r) {
				// 登录失败
				w.WriteHeader(401)
				return
			}

			if msg.Optionv("cmds") == nil {
				msg.Optionv("cmds", strings.Split(msg.Option(ice.MSG_USERURL), "/")[2:])
			}
			cmds := kit.Simple(msg.Optionv("cmds"))

			// 执行命令
			if msg.Option("proxy") != "" {
				msg.Cmd(ice.WEB_PROXY, msg.Option("proxy"), msg.Option(ice.MSG_USERURL), cmds)
			} else {
				msg.Target().Run(msg, cmd, msg.Option(ice.MSG_USERURL), cmds...)
			}

			// 输出响应
			switch msg.Append("_output") {
			case "void":
			case "status":
				msg.Info("status %s", msg.Result())
				w.WriteHeader(kit.Int(kit.Select("200", msg.Result(0))))

			case "redirect":
				http.Redirect(w, r, msg.Result(), 302)

			case "file":
				msg.Info("_output: %s %s", msg.Append("_output"), msg.Append("file"))
				w.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(msg.Append("name"), msg.Append("story"))))
				w.Header().Set("Content-Type", kit.Select("text/html", msg.Append("type")))
				http.ServeFile(w, r, msg.Append("file"))

			case "qrcode":
				if qr, e := qrcode.New(msg.Result(), qrcode.Medium); m.Assert(e) {
					w.Header().Set("Content-Type", "image/png")
					m.Assert(qr.Write(256, w))
				}

			case "result":
				w.Header().Set("Content-Type", kit.Select("text/html", msg.Append("type")))
				fmt.Fprint(w, msg.Result())
			default:
				fmt.Fprint(w, msg.Formats("meta"))
			}
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

	if r.URL.Path == "/" && m.Conf(ice.WEB_SERVE, "meta.init") != "true" {
		if _, e := os.Stat(m.Conf(ice.WEB_SERVE, "meta.volcanos.path")); e == nil {
			// 初始化成功
			m.Conf(ice.WEB_SERVE, "meta.init", "true")
		}
		w.Write([]byte(Refresh(m, 5)))
		m.Event(ice.SYSTEM_INIT)
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
		m.Log("serve", "listen %s", web.Server.ListenAndServe())
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
			"static", map[string]interface{}{"/": "usr/volcanos/",
				"/static/volcanos/": "usr/volcanos/",
				"/publish/":         "usr/publish/",
			},
			"volcanos", kit.Dict("path", "usr/volcanos", "branch", "master",
				"repos", "https://github.com/shylinux/volcanos",
			),
			"template", map[string]interface{}{"path": "usr/template", "list": []interface{}{
				`{{define "raw"}}{{.Result}}{{end}}`,
			}},
			"logheaders", "false",
			"init", "false",
		)},
		ice.WEB_SPACE: {Name: "space", Help: "空间站", Value: kit.Data(kit.MDB_SHORT, "name",
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

		ice.WEB_ROUTE: {Name: "route", Help: "路由", Value: kit.Data()},
		ice.WEB_PROXY: {Name: "proxy", Help: "代理", Value: kit.Data()},
		ice.WEB_GROUP: {Name: "group", Help: "分组", Value: kit.Data()},
		ice.WEB_LABEL: {Name: "label", Help: "标签", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			if m.Richs(ice.WEB_SPIDE, nil, "self", nil) == nil {
				m.Cmd(ice.WEB_SPIDE, "add", "self", kit.Select("http://:9020", m.Conf(ice.CLI_RUNTIME, "conf.ctx_self")))
			}
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
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			p := m.Conf(ice.WEB_CACHE, "meta.store")
			m.Richs(ice.WEB_CACHE, nil, "*", func(key string, value map[string]interface{}) {
				if f, _, e := kit.Create(path.Join(p, key[:2], key)); e == nil {
					defer f.Close()
					f.WriteString(kit.Formats(value))
				}
			})
			// m.Conf(ice.WEB_CACHE, "hash", kit.Dict())
			m.Save(ice.WEB_FAVOR, ice.WEB_CACHE, ice.WEB_STORY, ice.WEB_SHARE,
				ice.WEB_SPIDE, ice.WEB_SERVE)

			m.Done()
			m.Richs(ice.WEB_SPACE, nil, "*", func(key string, value map[string]interface{}) {
				if kit.Format(value["type"]) == "master" {
					m.Done()
				}
			})
		}},

		ice.WEB_SPIDE: {Name: "spide", Help: "蜘蛛侠", List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 爬虫列表
				m.Richs(ice.WEB_SPIDE, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value["client"], []string{"name", "method", "url"})
				})
				m.Sort("name")
				return
			}
			if len(arg) == 1 {
				// 爬虫详情
				m.Richs(ice.WEB_SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Push("detail", value)
				})
				return
			}

			switch arg[0] {
			case "add":
				// 添加爬虫
				if uri, e := url.Parse(arg[2]); e == nil && arg[2] != "" {
					dir, file := path.Split(uri.EscapedPath())
					m.Rich(ice.WEB_SPIDE, nil, kit.Dict(
						"cookie", kit.Dict(),
						"header", kit.Dict(),
						"client", kit.Dict(
							"name", arg[1],
							"logheaders", false,
							"timeout", "100s",
							"method", "POST",
							"protocol", uri.Scheme,
							"hostname", uri.Host,
							"path", dir,
							"file", file,
							"query", uri.RawQuery,
							"url", arg[2],
						),
					))
					m.Log(ice.LOG_CREATE, "%s: %v", arg[1], arg[2:])
				}
			default:
				// spide shy [cache] [POST|GET] uri file|data|form|part|json arg...
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

					// 请求数据
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
						case "form":
							data := []string{}
							for i := 1; i < len(arg)-1; i += 2 {
								data = append(data, url.QueryEscape(arg[i])+"="+url.QueryEscape(arg[i+1]))
							}
							body = bytes.NewBufferString(strings.Join(data, "&"))
							head["Content-Type"] = "application/x-www-form-urlencoded"
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

					// 构造请求
					uri = kit.MergeURL2(kit.Format(client["url"]), uri, arg)
					req, e := http.NewRequest(method, uri, body)
					m.Info("%s %s", req.Method, req.URL)
					m.Assert(e)

					// 请求参数
					for k, v := range head {
						req.Header.Set(k, v)
					}
					if list, ok := m.Optionv("header").([]string); ok {
						for i := 0; i < len(list)-1; i += 2 {
							req.Header.Set(list[i], list[i+1])
						}
					}
					kit.Fetch(client["header"], func(key string, value string) {
						req.Header.Set(key, value)
					})
					kit.Fetch(client["cookie"], func(key string, value string) {
						req.AddCookie(&http.Cookie{Name: key, Value: value})
						m.Info("%s: %s", key, value)
					})
					if method == "POST" {
						m.Info("%s: %s", req.Header.Get("Content-Length"), req.Header.Get("Content-Type"))
					}

					// 请求代理
					web := m.Target().Server().(*Frame)
					if web.Client == nil {
						web.Client = &http.Client{Timeout: kit.Duration(kit.Format(client["timeout"]))}
					}

					// 发送请求
					res, e := web.Client.Do(req)
					if m.Warn(e != nil, "%s", e) {
						m.Set("result")
						return
					}

					// 验证结果
					if m.Cost("%s %s: %s", res.Status, res.Header.Get("Content-Length"), res.Header.Get("Content-Type")); m.Warn(res.StatusCode != http.StatusOK, "%s", res.Status) {
						if cache != "" {
							m.Set("result")
						}
						return
					}

					// 缓存参数
					for _, v := range res.Cookies() {
						kit.Value(client, "cookie."+v.Name, v.Value)
						m.Info("%s: %s", v.Name, v.Value)
					}

					switch cache {
					case "msg":
						var data map[string][]string
						m.Assert(json.NewDecoder(res.Body).Decode(&data))
						m.Info("res: %s", kit.Formats(data))
						if len(data[ice.MSG_APPEND]) > 0 {
							for i := range data[data[ice.MSG_APPEND][0]] {
								for _, k := range data[ice.MSG_APPEND] {
									m.Push(k, data[k][i])
								}
							}
						}
						m.Resultv(data[ice.MSG_RESULT])

					case "raw":
						if b, e := ioutil.ReadAll(res.Body); m.Assert(e) {
							m.Echo(string(b))
						}
					case "cache":
						// 缓存结果
						m.Optionv("response", res)
						m.Echo(m.Cmd(ice.WEB_CACHE, "download", res.Header.Get("Content-Type"), uri).Append("data"))
					default:
						if strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
							b, _ := ioutil.ReadAll(res.Body)
							m.Echo(string(b))
							break
						}

						// 解析结果
						var data interface{}
						m.Assert(json.NewDecoder(res.Body).Decode(&data))
						data = kit.KeyValue(map[string]interface{}{}, "", data)
						m.Info("res: %s", kit.Formats(data))
						kit.Fetch(data, func(key string, value interface{}) {
							switch value := value.(type) {
							case []interface{}:
								m.Push(key, value)
							default:
								m.Push(key, kit.Format(value))
							}
						})
					}
				})
			}
		}},
		ice.WEB_SERVE: {Name: "serve [shy|dev|self]", Help: "服务器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 节点信息
			m.Conf(ice.CLI_RUNTIME, "node.name", m.Conf(ice.CLI_RUNTIME, "boot.hostname"))
			m.Conf(ice.CLI_RUNTIME, "node.type", ice.WEB_SERVER)

			switch kit.Select("def", arg, 0) {
			case "shy":
				// 连接根服务
				m.Richs(ice.WEB_SPIDE, nil, "shy", func(key string, value map[string]interface{}) {
					m.Cmd(ice.WEB_SPACE, "connect", "shy")
				})
				fallthrough
			case "dev":
				// 连接上游服务
				m.Richs(ice.WEB_SPIDE, nil, "dev", func(key string, value map[string]interface{}) {
					m.Cmd(ice.WEB_SPACE, "connect", "dev")
				})
				fallthrough
			default:
				// 启动服务
				m.Target().Start(m, "self")
				m.Cmd(ice.WEB_SPACE, "connect", "self")
			}
		}},
		ice.WEB_SPACE: {Name: "space", Help: "空间站", Meta: kit.Dict("exports", []string{"pod", "name"}), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 空间列表
				m.Richs(ice.WEB_SPACE, nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "type", "name", "text"})
				})
				m.Sort("name")
				return
			}

			web := m.Target().Server().(*Frame)
			switch arg[0] {
			case "auth":
				m.Richs(ice.WEB_SPACE, nil, arg[1], func(key string, value map[string]interface{}) {
					sessid := kit.Format(kit.Value(value, "sessid"))
					if value["user"] = arg[2]; sessid == "" || m.Cmdx(ice.AAA_SESS, "check", sessid) != arg[1] {
						sessid = m.Cmdx(ice.AAA_SESS, "create", arg[2:])
						value["sessid"] = sessid
					}
					m.Cmd(ice.WEB_SPACE, arg[1], "sessid", sessid)
				})

			case "share":
				switch arg[1] {
				case "add":
					m.Cmdy(ice.WEB_SPIDE, "self", path.Join("/space/share/add", path.Join(arg[2:]...)))
				default:
					m.Richs(ice.WEB_SPIDE, nil, m.Option("_dev"), func(key string, value map[string]interface{}) {
						m.Log(ice.LOG_CREATE, "dev: %s share: %s", m.Option("_dev"), arg[1])
						value["share"] = arg[1]
					})
				}

			case "connect":
				// 基本信息
				dev := kit.Select("dev", arg, 1)
				node := m.Conf(ice.CLI_RUNTIME, "node.type")
				name := m.Conf(ice.CLI_RUNTIME, "node.name")
				user := m.Conf(ice.CLI_RUNTIME, "boot.username")

				m.Hold(1).Gos(m, func(msg *ice.Message) {
					msg.Richs(ice.WEB_SPIDE, nil, dev, func(key string, value map[string]interface{}) {
						proto := kit.Select("ws", "wss", kit.Format(kit.Value(value, "client.protocol")) == "https")
						host := kit.Format(kit.Value(value, "client.hostname"))

						for i := 0; i < msg.Confi(ice.WEB_SPACE, "meta.redial.c"); i++ {
							if u, e := url.Parse(kit.MergeURL(proto+"://"+host+"/space/", "node", node, "name", name, "user", user, "share", value["share"])); msg.Assert(e) {
								if s, e := net.Dial("tcp", host); !msg.Warn(e != nil, "%s", e) {
									if s, _, e := websocket.NewClient(s, u, nil, msg.Confi(ice.WEB_SPACE, "meta.buffer.r"), msg.Confi(ice.WEB_SPACE, "meta.buffer.w")); !msg.Warn(e != nil, "%s", e) {
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
								sleep := time.Duration(rand.Intn(msg.Confi(ice.WEB_SPACE, "meta.redial.a"))*i+i*msg.Confi(ice.WEB_SPACE, "meta.redial.b")) * time.Millisecond
								msg.Info("%d sleep: %s reconnect: %s", i, sleep, u)
								time.Sleep(sleep)
							}
						}
					})
					m.Done()
				})

			default:
				if len(arg) == 1 {
					// 空间空间
					list := []string{}
					m.Cmdy(ice.WEB_SPACE, arg[0], "space").Table(func(index int, value map[string]string, head []string) {
						list = append(list, arg[0]+"."+value["name"])
					})
					m.Append("name", list)
					break

					// 空间详情
					m.Richs(ice.WEB_SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
						m.Push("detail", value)
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
					if socket, ok := value["socket"].(*websocket.Conn); ok {
						// 复制选项
						for _, k := range kit.Simple(m.Optionv("_option")) {
							if m.Options(k) {
								switch k {
								case "detail", "cmds":
								default:
									if m.Option(k) != "" {
										m.Option(k, m.Option(k))
									}
								}
							}
						}

						// 构造路由
						id := kit.Format(c.ID())
						m.Set(ice.MSG_DETAIL, arg[1:]...)
						m.Optionv(ice.MSG_TARGET, target[1:])
						m.Optionv(ice.MSG_SOURCE, []string{id})
						m.Info("send %s %s", id, m.Format("meta"))

						// 下发命令
						m.Target().Server().(*Frame).send[id] = m
						socket.WriteMessage(MSG_MAPS, []byte(m.Format("meta")))
						t := time.AfterFunc(kit.Duration(m.Conf(ice.WEB_SPACE, "meta.timeout.c")), func() {
							m.TryCatch(m, true, func(m *ice.Message) {
								m.Log(ice.LOG_WARN, "timeout")
							})
						})
						m.Call(true, func(msg *ice.Message) *ice.Message {
							if msg != nil {
								m.Copy(msg)
							}
							// 返回结果
							m.Log("cost", "%s: %s %v", m.Format("cost"), arg[0], arg[1:])
							t.Stop()
							return nil
						})
					}
				}) == nil, "not found %s", arg[0])
			}
		}},
		ice.WEB_DREAM: {Name: "dream", Help: "梦想家", Meta: kit.Dict(
			"remote", "pod", "exports", []string{"you", "name"},
			"detail", []interface{}{"启动", "停止"},
		), List: kit.List(
			kit.MDB_INPUT, "text", "value", "", "name", "name",
			kit.MDB_INPUT, "button", "value", "创建", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				switch arg[1] {
				case "启动":
					arg = []string{arg[4]}
				case "停止", "stop":
					m.Cmd(ice.WEB_SPACE, kit.Select(m.Option("name"), arg, 4), "exit", "1")
					time.Sleep(time.Second * 3)
					m.Event(ice.DREAM_CLOSE, arg[4])
					arg = arg[:0]
				}
			}

			if len(arg) > 0 {
				// 规范命名
				if !strings.Contains(arg[0], "-") || !strings.HasPrefix(arg[0], "20") {
					arg[0] = m.Time("20060102-") + arg[0]
				}

				// 创建目录
				p := path.Join(m.Conf(ice.WEB_DREAM, "meta.path"), arg[0])
				if _, e := os.Stat(p); e != nil {
					os.MkdirAll(p, 0777)
				}

				if m.Richs(ice.WEB_SPACE, nil, arg[0], nil) == nil {
					// 启动任务
					m.Option("cmd_dir", p)
					m.Option("cmd_type", "daemon")
					m.Optionv("cmd_env",
						"ctx_log", "boot.log",
						"ctx_mod", "ctx,log,gdb,ssh",
						"ctx_dev", m.Conf(ice.CLI_RUNTIME, "conf.ctx_dev"),
						"PATH", kit.Path(path.Join(p, "bin"))+":"+os.Getenv("PATH"),
					)
					m.Cmd(m.Confv(ice.WEB_DREAM, "meta.cmd"), "self", arg[0])
					time.Sleep(time.Second * 1)
					m.Event(ice.DREAM_START, arg...)
				}
				m.Cmdy("nfs.dir", p)
				return
			}

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
		}},

		ice.WEB_FAVOR: {Name: "favor [path [type name [text [key value]....]]", Help: "收藏夹", Meta: kit.Dict(
			"remote", "pod", "exports", []string{"hot", "favor"},
			"detail", []string{"编辑", "收藏", "收录", "导出", "删除"},
		), List: kit.List(
			kit.MDB_INPUT, "text", "name", "favor", "action", "auto",
			kit.MDB_INPUT, "text", "name", "id", "action", "auto",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
			kit.MDB_INPUT, "button", "value", "渲染",
			kit.MDB_INPUT, "button", "value", "回放",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch m.Option("_action") {
			case "渲染":
				m.Option("render", "spide")
				m.Richs(ice.WEB_FAVOR, nil, kit.Select(m.Option("favor"), arg, 0), func(key string, value map[string]interface{}) {
					m.Option("render", kit.Select("spide", kit.Value(value, "meta.render")))
				})
				defer m.Render(m.Conf(ice.WEB_FAVOR, kit.Keys("meta.template", m.Option("render"))))

			case "回放":
				return
			}

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
			}
		}},
		ice.WEB_STORY: {Name: "story", Help: "故事会", Meta: kit.Dict("remote", "pod", "exports", []string{"top", "story"},
			"detail", []string{"共享", "更新", "推送"}), List: kit.List(
			kit.MDB_INPUT, "text", "name", "story", "action", "auto",
			kit.MDB_INPUT, "text", "name", "list", "action", "auto",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
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
					m.Echo("share: ")
					if list == "" {
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
				m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
					begin = kit.Select(kit.Format(kit.Value(val, kit.Keys("remote", kit.Select("dev", arg, 2), "pull", "head"))), arg, 3)
					end = kit.Format(kit.Value(val, kit.Keys("remote", kit.Select("dev", arg, 2), "pull", "list")))
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
								if kit.Int(node["count"]) > kit.Int(kit.Value(val, kit.Keys("remote", arg[2], "pull", "count"))) {
									// 更新分支
									m.Log(ice.LOG_IMPORT, "%v: %v", arg[2], pull)
									kit.Value(val, kit.Keys("remote", arg[2], "pull"), kit.Dict(
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

				// 查询索引
				prev, pull, some, list := "", "", "", ""
				m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
					prev = kit.Format(val["list"])
					pull = kit.Format(kit.Value(val, kit.Keys("remote", arg[2], "pull", "list")))
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

					if pull != "" || some != pull {
						// 合并节点
						local := m.Richs(ice.WEB_STORY, nil, prev, nil)
						remote := m.Richs(ice.WEB_STORY, nil, pull, nil)
						list = m.Rich(ice.WEB_STORY, nil, kit.Dict(
							"scene", val["scene"], "story", val["story"], "count", kit.Int(remote["count"])+1,
							"data", local["data"], "prev", pull, "push", prev,
						))
						m.Log(ice.LOG_CREATE, "merge: %s %s->%s", list, prev, pull)
						val["list"] = list
					}

					// 查询节点
					nodes := []string{}
					for list != "" && list != some {
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
					break
				}

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
						m.Push("list", key)
						m.Push(list, value, []string{"time", "count", "scene", "story"})
						m.Richs(ice.WEB_CACHE, nil, value["data"], func(key string, value map[string]interface{}) {
							m.Push("drama", value["text"])
							m.Push("data", key)
						})

						kit.Fetch(value["list"], func(key string, val string) {
							m.Richs(ice.WEB_STORY, nil, val, func(key string, value map[string]interface{}) {
								// 复合节点
								m.Push("list", key)
								m.Push(list, value, []string{"time", "count", "scene", "story"})
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
		ice.WEB_SHARE: {Name: "share", Help: "共享链", List: kit.List(
			kit.MDB_INPUT, "text", "name", "share", "action", "auto",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 共享列表
				m.Grows(ice.WEB_SHARE, nil, "", "", func(key int, value map[string]interface{}) {
					m.Push(kit.Format(key), value, []string{kit.MDB_TIME, "share", kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
					m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.link"), value["share"], value["share"]))
				})
				return
			}
			if len(arg) == 1 {
				// 共享详情
				m.Richs(ice.WEB_SHARE, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Push("detail", value)
					m.Push("key", "link")
					m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.link"), key, key))
					m.Push("key", "share")
					m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.share"), key))
					m.Push("key", "value")
					m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.value"), key))
				})
				return
			}

			switch arg[0] {
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

		ice.WEB_ROUTE: {Name: "route", Help: "路由", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(ice.WEB_SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
				switch value[kit.MDB_TYPE] {
				case ice.WEB_MASTER:
				case ice.WEB_SERVER:
				case ice.WEB_WORKER:
				}
			})
			m.Cmdy(ice.WEB_SPACE, arg[0], arg[1:])
		}},
		ice.WEB_PROXY: {Name: "proxy", Help: "代理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(ice.WEB_SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
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
		ice.WEB_GROUP: {Name: "group", Help: "分组", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(ice.WEB_PROXY, arg[0], arg[1:])
		}},
		ice.WEB_LABEL: {Name: "label", Help: "标签", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(ice.WEB_GROUP, arg[0], arg[1:])
		}},

		"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(ice.WEB_SHARE, nil, arg[0], func(key string, value map[string]interface{}) {
				m.Log(ice.LOG_EXPORT, "%s: %v", arg, kit.Format(value))

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
					if m.Append("_output", "file"); strings.HasPrefix(kit.Format(value["text"]), m.Conf(ice.WEB_CACHE, "meta.path")) {
						m.Append("type", value["type"])
						m.Append("name", value["name"])
						m.Append("file", value["text"])
					} else {
						m.Append("_output", "result")
						m.Echo("%s", value["text"])
					}
					return
				case "detail", "详情":
					m.Append("_output", "result")
					m.Echo(kit.Formats(value))
					return
				case "share", "共享码":
					m.Append("_output", "qrcode")
					m.Echo("%s/%s/", m.Conf(ice.WEB_SHARE, "meta.domain"), key)
					return
				case "value", "数据值":
					m.Append("_output", "qrcode")
					m.Echo("%s", value["text"])
					return
				}

				switch value["type"] {
				case ice.TYPE_RIVER:
					// 共享群组
					Redirect(m, "/", "share", key, "river", value["text"])

				case ice.TYPE_STORM:
					// 共享应用
					Redirect(m, "/", "share", key, "storm", value["text"], "river", kit.Value(value, "extra.river"))

				case ice.TYPE_ACTION:
					if len(arg) == 1 {
						// 跳转主页
						Redirect(m, "/share/"+arg[0]+"/", "title", value["name"])
						break
					}

					if arg[1] == "" {
						// 返回主页
						http.ServeFile(m.W, m.R, m.Conf(ice.WEB_SHARE, "meta.index"))
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

				case ice.TYPE_ACTIVE:
					// 扫码数据
					m.Append("_output", "qrcode")
					m.Echo(kit.Format(value))

				default:
					// 查看数据
					m.Option("type", value["type"])
					m.Option("name", value["name"])
					m.Option("text", value["text"])
					m.Render(m.Conf(ice.WEB_SHARE, "meta.template.simple"))
					m.Append("_output", "result")
				}
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
				m.Push("_output", kit.Select("file", "result", m.Append("file") == ""))
			}
		}},
		"/space/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			list := strings.Split(cmd, "/")

			switch list[2] {
			case "login":
				m.Option(ice.MSG_SESSID, Cookie(m, m.Cmdx(ice.AAA_USER, "login", m.Option("username"), m.Option("password"))))
				return
			case "share":
				m.Cmdy(ice.WEB_SHARE, list[3:])
				return
			}

			if s, e := websocket.Upgrade(m.W, m.R, nil, m.Confi(ice.WEB_SPACE, "meta.buffer.r"), m.Confi(ice.WEB_SPACE, "meta.buffer.w")); m.Assert(e) {
				m.Option("name", strings.Replace(m.Option("name"), ".", "_", -1))
				if !m.Options("name") {
					m.Option("name", kit.Hashs("uniq"))
				}

				// 共享空间
				share := m.Option("share")
				if m.Richs(ice.WEB_SHARE, nil, share, nil) == nil {
					share = m.Cmdx(ice.WEB_SHARE, "add", m.Option("node"), m.Option("name"), m.Option("user"))
				}

				// 添加节点
				h := m.Rich(ice.WEB_SPACE, nil, kit.Dict(
					kit.MDB_TYPE, m.Option("node"),
					kit.MDB_NAME, m.Option("name"),
					kit.MDB_TEXT, m.Option("user"),
					"sessid", m.Option("sessid"),
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
		"/static/volcanos/plugin/github.com/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			file := strings.TrimPrefix(cmd, "/static/volcanos/")
			if _, e := os.Stat(path.Join("usr/volcanos", file)); e != nil {
				m.Cmd("cli.system", "git", "clone", "https://"+strings.Join(strings.Split(cmd, "/")[4:7], "/"),
					path.Join("usr/volcanos", strings.Join(strings.Split(cmd, "/")[3:7], "/")))
			}

			m.Push("_output", "void")
			http.ServeFile(m.W, m.R, path.Join("usr/volcanos", file))
		}},
		"/plugin/github.com/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if _, e := os.Stat(path.Join("usr/volcanos", cmd)); e != nil {
				m.Cmd("cli.system", "git", "clone", "https://"+strings.Join(strings.Split(cmd, "/")[2:5], "/"),
					path.Join("usr/volcanos", strings.Join(strings.Split(cmd, "/")[1:5], "/")))
			}

			m.Push("_output", "void")
			http.ServeFile(m.W, m.R, path.Join("usr/volcanos", cmd))
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
