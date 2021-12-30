package web

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _spide_create(m *ice.Message, name, address string) {
	if uri, e := url.Parse(address); e == nil && address != "" {
		if m.Richs(SPIDE, nil, name, func(key string, value map[string]interface{}) {
			kit.Value(value, "client.hostname", uri.Host)
			kit.Value(value, "client.url", address)
		}) == nil {
			dir, file := path.Split(uri.EscapedPath())
			m.Echo(m.Rich(SPIDE, nil, kit.Dict(
				SPIDE_COOKIE, kit.Dict(), SPIDE_HEADER, kit.Dict(), SPIDE_CLIENT, kit.Dict(
					kit.MDB_NAME, name, SPIDE_METHOD, SPIDE_POST, "url", address,
					tcp.PROTOCOL, uri.Scheme, tcp.HOSTNAME, uri.Host,
					nfs.PATH, dir, nfs.FILE, file, "query", uri.RawQuery,
					kit.MDB_TIMEOUT, "600s", LOGHEADERS, ice.FALSE,
				),
			)))
		}
		m.Log_CREATE(SPIDE, name, ADDRESS, address)
	}
}
func _spide_list(m *ice.Message, arg ...string) {
	m.Richs(SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
		// 缓存方式
		cache, save := "", ""
		switch arg[1] {
		case SPIDE_RAW:
			cache, arg = arg[1], arg[1:]
		case SPIDE_MSG:
			cache, arg = arg[1], arg[1:]
		case SPIDE_SAVE:
			cache, save, arg = arg[1], arg[2], arg[2:]
		case SPIDE_CACHE:
			cache, arg = arg[1], arg[1:]
		}

		// 请求方法
		client := value[SPIDE_CLIENT].(map[string]interface{})
		method := kit.Select(SPIDE_POST, client[SPIDE_METHOD])
		switch arg = arg[1:]; arg[0] {
		case SPIDE_GET:
			method, arg = SPIDE_GET, arg[1:]
		case SPIDE_PUT:
			method, arg = SPIDE_PUT, arg[1:]
		case SPIDE_POST:
			method, arg = SPIDE_POST, arg[1:]
		case SPIDE_DELETE:
			method, arg = SPIDE_DELETE, arg[1:]
		}

		// 请求地址
		uri, arg := arg[0], arg[1:]

		// 请求参数
		body, head, arg := _spide_body(m, method, arg...)

		// 构造请求
		req, e := http.NewRequest(method, kit.MergeURL2(kit.Format(client["url"]), uri, arg), body)
		if m.Warn(e, ice.ErrNotFound, uri) {
			return
		}

		// 请求配置
		_spide_head(m, req, head, value)

		// 发送请求
		res, e := _spide_send(m, req, kit.Format(client[kit.MDB_TIMEOUT]))
		if m.Warn(e, ice.ErrNotFound, uri) {
			return
		}

		if m.Config(LOGHEADERS) == ice.TRUE {
			for k, v := range res.Header {
				m.Debug("%v: %v", k, v)
			}
		}

		// 检查结果
		defer res.Body.Close()
		m.Cost(kit.MDB_STATUS, res.Status, kit.MDB_SIZE, res.Header.Get(ContentLength), kit.MDB_TYPE, res.Header.Get(ContentType))

		// 缓存配置
		for _, v := range res.Cookies() {
			kit.Value(value, kit.Keys(SPIDE_COOKIE, v.Name), v.Value)
			m.Log_IMPORT(v.Name, v.Value)
		}

		// 错误信息
		if m.Warn(res.StatusCode != http.StatusOK, ice.ErrNotFound, uri, "status", res.Status) {
			switch m.Set(ice.MSG_RESULT); res.StatusCode {
			case http.StatusNotFound:
				m.Warn(true, ice.ErrNotFound, uri)
				return
			case http.StatusUnauthorized:
				m.Warn(true, ice.ErrNotRight, uri)
				return
			}
		}

		// 解析结果
		_spide_save(m, cache, save, uri, res)
	})
}
func _spide_body(m *ice.Message, method string, arg ...string) (io.Reader, map[string]string, []string) {
	head := map[string]string{}
	body, ok := m.Optionv(SPIDE_BODY).(io.Reader)
	if !ok && len(arg) > 0 && method != SPIDE_GET {
		if len(arg) == 1 {
			arg = []string{SPIDE_DATA, arg[0]}
		}
		switch arg[0] {
		case SPIDE_FORM:
			data := []string{}
			for i := 1; i < len(arg)-1; i += 2 {
				data = append(data, url.QueryEscape(arg[i])+"="+url.QueryEscape(arg[i+1]))
			}
			body = bytes.NewBufferString(strings.Join(data, "&"))
			head[ContentType] = ContentFORM

		case SPIDE_PART:
			body, head[ContentType] = _spide_part(m, arg...)

		case SPIDE_DATA:
			if len(arg) == 1 {
				arg = append(arg, "{}")
			}
			body, arg = bytes.NewBufferString(arg[1]), arg[2:]
			head[ContentType] = ContentJSON

		case SPIDE_FILE:
			if f, e := os.Open(arg[1]); m.Assert(e) {
				defer f.Close()
				body, arg = f, arg[2:]
			}

		case SPIDE_JSON:
			arg = arg[1:]
			fallthrough
		default:
			data := map[string]interface{}{}
			for i := 0; i < len(arg)-1; i += 2 {
				kit.Value(data, arg[i], arg[i+1])
			}
			if b, e := json.Marshal(data); m.Assert(e) {
				head[ContentType] = ContentJSON
				body = bytes.NewBuffer(b)
			}
			m.Log_EXPORT(SPIDE_JSON, kit.Format(data))
		}
		arg = arg[:0]
	} else {
		body = bytes.NewBuffer([]byte{})
	}
	return body, head, arg
}
func _spide_part(m *ice.Message, arg ...string) (io.Reader, string) {
	buf := &bytes.Buffer{}
	mp := multipart.NewWriter(buf)
	defer mp.Close()

	cache := time.Now().Add(-time.Hour * 240000)
	for i := 1; i < len(arg)-1; i += 2 {
		if arg[i] == SPIDE_CACHE {
			if t, e := time.ParseInLocation(ice.MOD_TIME, arg[i+1], time.Local); e == nil {
				cache = t
			}
		}
		if strings.HasPrefix(arg[i+1], "@") {
			if s, e := os.Stat(arg[i+1][1:]); e == nil {
				m.Debug("local: %s cache: %s", s.ModTime(), cache)
				if s.ModTime().Before(cache) {
					break
				}
			}
			if f, e := os.Open(arg[i+1][1:]); m.Assert(e) {
				defer f.Close()
				if p, e := mp.CreateFormFile(arg[i], path.Base(arg[i+1][1:])); m.Assert(e) {
					if n, e := io.Copy(p, f); m.Assert(e) {
						m.Debug("upload: %s %d", arg[i+1], n)
					}
				}
			}
		} else {
			mp.WriteField(arg[i], arg[i+1])
		}
	}
	return buf, mp.FormDataContentType()
}
func _spide_head(m *ice.Message, req *http.Request, head map[string]string, value map[string]interface{}) {
	m.Info("%s %s", req.Method, req.URL)
	kit.Fetch(value[SPIDE_HEADER], func(key string, value string) {
		req.Header.Set(key, value)
	})
	kit.Fetch(value[SPIDE_COOKIE], func(key string, value string) {
		req.AddCookie(&http.Cookie{Name: key, Value: value})
		m.Logs(key, value)
	})
	list := kit.Simple(m.Optionv(SPIDE_HEADER))
	for i := 0; i < len(list)-1; i += 2 {
		req.Header.Set(list[i], list[i+1])
		m.Logs(list[i], list[i+1])
	}
	for k, v := range head {
		req.Header.Set(k, v)
	}
	if req.Method == SPIDE_POST {
		m.Logs(req.Header.Get(ContentLength), req.Header.Get(ContentType))
	}
}
func _spide_send(m *ice.Message, req *http.Request, timeout string) (*http.Response, error) {
	web := m.Target().Server().(*Frame)
	if web.Client == nil {
		web.Client = &http.Client{Timeout: kit.Duration(timeout)}
	}
	return web.Client.Do(req)
}
func _spide_save(m *ice.Message, cache, save, uri string, res *http.Response) {
	switch cache {
	case SPIDE_RAW:
		b, _ := ioutil.ReadAll(res.Body)
		if strings.HasPrefix(res.Header.Get(ContentType), ContentJSON) {
			m.Echo(kit.Formats(kit.UnMarshal(string(b))))
		} else {
			m.Echo(string(b))
		}

	case SPIDE_MSG:
		var data map[string][]string
		m.Assert(json.NewDecoder(res.Body).Decode(&data))
		for _, k := range data[ice.MSG_APPEND] {
			for i := range data[k] {
				m.Push(k, data[k][i])
			}
		}
		m.Resultv(data[ice.MSG_RESULT])

	case SPIDE_SAVE:
		if f, p, e := kit.Create(save); m.Assert(e) {
			defer f.Close()

			if n, e := io.Copy(f, res.Body); m.Assert(e) {
				m.Log_EXPORT(kit.MDB_SIZE, n, nfs.FILE, p)
				m.Echo(p)
			}
		}

	case SPIDE_CACHE:
		m.Optionv(RESPONSE, res)
		m.Cmdy(CACHE, DOWNLOAD, res.Header.Get(ContentType), uri)
		m.Echo(m.Append(DATA))

	default:
		b, _ := ioutil.ReadAll(res.Body)

		var data interface{}
		if e := json.Unmarshal(b, &data); e != nil {
			m.Echo(string(b))
			break
		}

		m.Optionv(SPIDE_RES, data)
		data = kit.KeyValue(map[string]interface{}{}, "", data)
		m.Push("", data)
	}
}

const (
	SPIDE_RAW   = "raw"
	SPIDE_MSG   = "msg"
	SPIDE_SAVE  = "save"
	SPIDE_CACHE = "cache"

	SPIDE_GET    = "GET"
	SPIDE_PUT    = "PUT"
	SPIDE_POST   = "POST"
	SPIDE_DELETE = "DELETE"

	SPIDE_BODY = "body"
	SPIDE_FORM = "form"
	SPIDE_PART = "part"
	SPIDE_DATA = "data"
	SPIDE_FILE = "file"
	SPIDE_JSON = "json"

	SPIDE_RES = "content_data"

	ContentType   = "Content-Type"
	ContentLength = "Content-Length"

	ContentFORM = "application/x-www-form-urlencoded"
	ContentJSON = "application/json"
	ContentHTML = "text/html"
	ContentPNG  = "image/png"
)
const (
	SPIDE_CLIENT = "client"
	SPIDE_METHOD = "method"
	SPIDE_HEADER = "header"
	SPIDE_COOKIE = "cookie"

	ADDRESS  = "address"
	REQUEST  = "request"
	RESPONSE = "response"

	CLIENT_NAME = "client.name"
	LOGHEADERS  = "logheaders"
)
const SPIDE = "spide"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SPIDE: {Name: SPIDE, Help: "蜘蛛侠", Value: kit.Data(
			kit.MDB_SHORT, CLIENT_NAME, kit.MDB_FIELD, "time,client.name,client.url",
			LOGHEADERS, ice.FALSE,
		)},
	}, Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(SPIDE, mdb.CREATE, ice.OPS, kit.Select("http://:9020", m.Conf(cli.RUNTIME, "conf.ctx_ops")))
			m.Cmd(SPIDE, mdb.CREATE, ice.DEV, kit.Select("http://:9020", m.Conf(cli.RUNTIME, "conf.ctx_dev")))
			m.Cmd(SPIDE, mdb.CREATE, ice.SHY, kit.Select("https://shylinux.com:443", m.Conf(cli.RUNTIME, "conf.ctx_shy")))
		}},
		SPIDE: {Name: "spide client.name action=raw,msg,save,cache method=GET,PUT,POST,DELETE url format=form,part,json,data,file arg run:button create", Help: "蜘蛛侠", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create name address", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_spide_create(m, m.Option(kit.MDB_NAME), m.Option(ADDRESS))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 || arg[0] == "" || (len(arg) > 3 && arg[3] == "") {
				mdb.HashSelect(m, kit.Slice(arg, 0, 1)...)
				m.Sort("client.name")
				return
			}
			_spide_list(m, arg...)
		}},

		SPIDE_GET: {Name: "GET url key value run:button", Help: "蜘蛛侠", Action: map[string]*ice.Action{
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, SPIDE, "", mdb.HASH, m.OptionSimple(CLIENT_NAME))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, SPIDE_GET, arg[0], arg[1:]))))
		}},
		SPIDE_POST: {Name: "POST url key value run:button", Help: "蜘蛛侠", Action: map[string]*ice.Action{
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, SPIDE, "", mdb.HASH, m.OptionSimple(CLIENT_NAME))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, SPIDE_POST, arg[0], arg[1:]))))
		}},
	}})
}
