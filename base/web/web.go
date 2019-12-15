package web

import (
	"github.com/gorilla/websocket"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bytes"
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"text/template"
	"time"
)

const (
	MSG_MAPS = 1
)

type WEB struct {
	*http.Client
	*http.Server
	*http.ServeMux
	m    *ice.Message
	send map[string]*ice.Message
}

func Cookie(msg *ice.Message, sessid string) string {
	w := msg.Optionv("response").(http.ResponseWriter)
	expire := time.Now().Add(kit.Duration(msg.Conf("aaa.sess", "meta.expire")))
	msg.Log("cookie", "expire %v sessid %s", kit.Format(expire), sessid)
	http.SetCookie(w, &http.Cookie{Name: "sessid", Value: sessid, Path: "/", Expires: expire})
	return sessid
}

func (web *WEB) Login(msg *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if msg.Options("sessid") {
		sub := msg.Cmd("aaa.sess", "check", msg.Option("sessid"))
		msg.Log("info", "user %s %s", msg.Option("userrole", sub.Append("userrole")),
			msg.Option("username", sub.Append("username")))
	}
	return true
}
func (web *WEB) HandleWSS(m *ice.Message, safe bool, c *websocket.Conn) {
	for {
		if t, b, e := c.ReadMessage(); e != nil {
			m.Log("warn", "space recv %d msg %v", t, e)
			break
		} else {
			switch t {
			case MSG_MAPS:
				socket, msg := c, m.Spawn(b)
				source := kit.Simple(msg.Optionv("_source"))
				target := kit.Simple(msg.Optionv("_target"))
				msg.Log("space", "recv %v %v->%v %v", t, source, target, msg.Formats("meta"))

				if len(target) > 0 {
					if s, ok := msg.Confv("web.space", "hash."+target[0]+".socket").(*websocket.Conn); ok {
						msg.Log("space", "route")
						// 转发报文
						socket, source, target = s, append(source, target[0]), target[1:]
					} else if call, ok := web.send[msg.Option("_target")]; len(target) == 1 && ok {
						msg.Log("space", "done")
						// 接收响应
						delete(web.send, msg.Option("_target"))
						call.Back(msg)
						break
					} else if msg.Option("_handle") == "true" {
						msg.Log("space", "miss")
						// 丢弃报文
						break
					} else {
						// 失败报文
						msg.Log("space", "error")
						msg.Echo("error")
						source, target = []string{source[len(source)-1]}, kit.Revert(source)[1:]
					}
				} else {
					msg.Log("space", "run")
					// 本地执行
					if safe {
						msg = msg.Cmd()
					} else {
						msg.Echo("no right")
					}
					msg.Optionv("_handle", "true")
					kit.Revert(source)
					source, target = []string{source[0]}, source[1:]
				}

				// 发送报文
				msg.Optionv("_source", source)
				msg.Optionv("_target", target)
				msg.Log("space", "send %v %v->%v %v", t, source, target, msg.Formats("meta"))
				socket.WriteMessage(t, []byte(msg.Format("meta")))
			}
		}
	}
}
func (web *WEB) HandleCGI(m *ice.Message, which string) *template.Template {
	cgi := template.FuncMap{
		"result": func(msg *ice.Message) string {
			return msg.Result()
		},
	}
	tmpl := template.New("render")
	for k, v := range m.Target().Commands {
		m.Log("info", "%v, %v", k, v.Name)
		if strings.HasPrefix(k, "/") || strings.HasPrefix(k, "_") {
			continue
		}

		func(k string, v *ice.Command) {
			cgi[k] = func(arg ...interface{}) (res interface{}) {
				m.TryCatch(m.Spawn(), true, func(msg *ice.Message) {
					msg.Option("render", "table")
					v.Hand(msg, m.Target(), k, kit.Simple(arg)...)

					buffer := bytes.NewBuffer([]byte{})
					m.Assert(tmpl.ExecuteTemplate(buffer, msg.Option("render"), msg))
					res = string(buffer.Bytes())
				})
				return
			}
		}(k, v)
	}
	tmpl = tmpl.Funcs(cgi)
	tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf("serve", "template.path"), "/*.tmpl")))
	tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf("serve", "template.path"), m.Target().Name, "/*.tmpl")))
	tmpl = template.Must(tmpl.ParseFiles(which))
	m.Confm("serve", "template.list", func(index int, value string) { tmpl = template.Must(tmpl.Parse(value)) })
	for i, v := range tmpl.Templates() {
		m.Log("info", "%v, %v", i, v.Name())
	}
	return tmpl
}
func (web *WEB) HandleCmd(m *ice.Message, key string, cmd *ice.Command) {
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		m.TryCatch(m.Spawns(), true, func(msg *ice.Message) {
			defer func() {
				msg.Log("cost", msg.Format("cost"))
			}()

			msg.Optionv("request", r)
			msg.Optionv("response", w)
			msg.Option("remote_ip", r.Header.Get("remote_ip"))
			msg.Option("agent", r.Header.Get("User-Agent"))
			msg.Option("referer", r.Header.Get("Referer"))
			msg.Option("accept", r.Header.Get("Accept"))
			msg.Option("method", r.Method)
			msg.Option("path", r.URL.Path)
			msg.Option("sessid", "")

			// 请求环境
			for _, v := range r.Cookies() {
				if v.Value != "" {
					msg.Option(v.Name, v.Value)
				}
			}

			// 请求参数
			r.ParseMultipartForm(4096)
			if r.ParseForm(); len(r.PostForm) > 0 {
				for k, v := range r.PostForm {
					msg.Log("info", "%s: %v", k, v)
				}
				msg.Log("info", "")
			}
			for k, v := range r.Form {
				for _, v := range v {
					msg.Add("option", k, v)
				}
			}

			// 请求数据
			switch r.Header.Get("Content-Type") {
			case "application/json":
				var data interface{}
				if e := json.NewDecoder(r.Body).Decode(&data); e != nil {
					msg.Log("warn", "%v", e)
				}
				msg.Optionv("content_data", data)
				msg.Log("info", "%v", kit.Formats(data))

				switch d := data.(type) {
				case map[string]interface{}:
					for k, v := range d {
						for _, v := range kit.Simple(v) {
							msg.Add("option", k, v)
						}
					}
				}
			}

			if web.Login(msg, w, r) {
				msg.Log("cmd", "%s %s", msg.Target().Name, key)
				cmd.Hand(msg, msg.Target(), msg.Option("path"), kit.Simple(msg.Optionv("cmds"))...)
				msg.Set("option")
				w.Write([]byte(msg.Formats("meta")))
			}
		})
	})
}
func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := web.m

	index := r.Header.Get("index.module") == ""
	if index {
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			r.Header.Set("remote_ip", ip)
		} else if ip := r.Header.Get("X-Real-Ip"); ip != "" {
			r.Header.Set("remote_ip", ip)
		} else if strings.HasPrefix(r.RemoteAddr, "[") {
			r.Header.Set("remote_ip", strings.Split(r.RemoteAddr, "]")[0][1:])
		} else {
			r.Header.Set("remote_ip", strings.Split(r.RemoteAddr, ":")[0])
		}
		m.Log("info", "").Log("info", "%v %s %s", r.Header.Get("remote_ip"), r.Method, r.URL)
		r.Header.Set("index.module", "some")
		r.Header.Set("index.url", r.URL.String())
		r.Header.Set("index.path", r.URL.Path)
	}

	web.ServeMux.ServeHTTP(w, r)
}

func (web *WEB) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &WEB{}
}
func (web *WEB) Begin(m *ice.Message, arg ...string) ice.Server {
	web.send = map[string]*ice.Message{}
	return web
}
func (web *WEB) Start(m *ice.Message, arg ...string) bool {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if w, ok := s.Server().(*WEB); ok {
			if w.ServeMux != nil {
				return
			}
			w.ServeMux = http.NewServeMux()
			msg := m.Spawns(s)

			// 级联路由
			route := "/" + s.Name + "/"
			if n, ok := p.Server().(*WEB); ok && n.ServeMux != nil {
				msg.Log("route", "%s <- %s", p.Name, route)
				n.Handle(route, http.StripPrefix(path.Dir(route), w))
			}

			// 静态路由
			m.Confm("web.serve", "static", func(key string, value string) {
				msg.Log("route", "%s <- %s <- %s", s.Name, key, value)
				w.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 命令路由
			for k, x := range s.Commands {
				if k[0] == '/' {
					msg.Log("route", "%s <- %s", s.Name, k)
					w.HandleCmd(msg, k, x)
				}
			}
		}
	})

	port := kit.Select(m.Conf("spide", "self.port"), arg, 0)
	web.m = m
	web.Server = &http.Server{Addr: port, Handler: web}
	m.Log("serve", "listen %s", port)
	m.Log("serve", "listen %s", web.Server.ListenAndServe())
	return true
}
func (web *WEB) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "web", Help: "网页模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"spide": {Name: "客户端", Value: map[string]interface{}{
			"self": map[string]interface{}{"port": ":9020"},
		}},
		"serve": {Name: "服务端", Value: map[string]interface{}{
			"static": map[string]interface{}{"/": "usr/volcanos/",
				"/static/volcanos/": "usr/volcanos/",
			},
			"template": map[string]interface{}{
				"path": "usr/template",
				"list": []interface{}{
					`{{define "raw"}}{{.|result}}{{end}}`,
					`{{define "title"}}{{.|result}}{{end}}`,
					`{{define "chapter"}}{{.|result}}{{end}}`,
					`{{define "section"}}{{.|result}}{{end}}`,
					`{{define "block"}}<div>{{.|result}}<div>{{end}}`,
				},
			},
		}},
		"space": {Name: "空间端", Value: map[string]interface{}{
			"meta": map[string]interface{}{"buffer": 4096, "redial": 3000},
			"hash": map[string]interface{}{},
			"list": map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		"_init": {Name: "_init", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
		}},
		"serve": {Name: "hi", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf("cli.runtime", "node.type", "server")
			m.Run(arg...)
		}},
		"/space": &ice.Command{Name: "/space", Help: "", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)
			if s, e := websocket.Upgrade(w, r, nil, m.Confi("web.space", "meta.buffer"), m.Confi("web.space", "meta.buffer")); m.Assert(e) {
				h := m.Option("name")

				meta := map[string]interface{}{
					"create_time": m.Time(),
					"socket":      s,
					"type":        m.Option("node"),
					"name":        m.Option("name"),
				}
				m.Confv("space", []string{"hash", h}, meta)
				m.Log("space", "conn %v %v", h, kit.Formats(m.Confv("space")))

				web := m.Target().Server().(*WEB)
				m.Gos(m, func(m *ice.Message) {
					web.HandleWSS(m, false, s)
				})
			}
		}},
		"space": &ice.Command{Name: "space", Help: "", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			web := m.Target().Server().(*WEB)
			switch arg[0] {
			case "connect":
				node, name := m.Conf("cli.runtime", "node.type"), m.Conf("cli.runtime", "boot.hostname")
				if node == "worker" {
					name = m.Conf("cli.runtime", "boot.pathname")
				}
				host := kit.Select(m.Conf("web.spide", "self.port"), arg, 1)
				p := "ws://" + host + kit.Select("/space", arg, 2) + "?node=" + node + "&name=" + name

				if u, e := url.Parse(p); m.Assert(e) {
					m.TryCatch(m, true, func(m *ice.Message) {
						for {
							if s, e := net.Dial("tcp", host); e == nil {
								if s, _, e := websocket.NewClient(s, u, nil, m.Confi("web.space", "meta.buffer"), m.Confi("web.space", "meta.buffer")); e == nil {
									id := m.Option("_source", []string{kit.Format(c.ID()), "some"})
									web.send[id] = m
									s.WriteMessage(MSG_MAPS, []byte(m.Format("meta")))

									web.HandleWSS(m, true, s)
								}
							}
							time.Sleep(time.Duration(rand.Intn(m.Confi("web.space", "meta.redial"))) * time.Millisecond)
							m.Log("info", "reconnect %v", host)
						}
					})
				}
			default:
				if arg[0] == "" {
					m.Cmdy(arg[1:])
					break
				}
				target := strings.Split(arg[0], ".")
				if socket, ok := m.Confv("space", "hash."+target[0]+".socket").(*websocket.Conn); !ok {
					m.Echo("error").Echo("not found")
				} else {
					id := kit.Format(c.ID())
					m.Optionv("_source", []string{id, target[0]})
					m.Optionv("_target", target[1:])

					web := m.Target().Server().(*WEB)
					web.send[id] = m
					m.Add("detail", arg[1:]...)
					socket.WriteMessage(MSG_MAPS, []byte(m.Format("meta")))
					m.Call(true, func(msg *ice.Message) *ice.Message {
						m.Copy(msg)
						return nil
					})
				}
			}
		}},
	},
}

func init() { ice.Index.Register(Index, &WEB{}) }
