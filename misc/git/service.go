package git

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"shylinux.com/x/go-git/v5/plumbing"
	"shylinux.com/x/go-git/v5/plumbing/protocol/packp"
	"shylinux.com/x/go-git/v5/plumbing/transport"
	"shylinux.com/x/go-git/v5/plumbing/transport/file"
	"shylinux.com/x/go-git/v5/plumbing/transport/server"
	"shylinux.com/x/go-git/v5/utils/ioutil"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _service_path(m *ice.Message, p string, arg ...string) string {
	return kit.Path(ice.USR_LOCAL_REPOS, kit.TrimExt(p, GIT), path.Join(arg...))
}
func _service_param(m *ice.Message, arg ...string) (string, string) {
	repos, service := arg[0], kit.Select(arg[len(arg)-1], m.Option(SERVICE))
	return _service_path(m, repos), strings.TrimPrefix(service, "git-")
}
func _service_repos(m *ice.Message, arg ...string) error {
	repos, service := _service_param(m, arg...)
	m.Cmd(web.COUNT, mdb.CREATE, service, strings.TrimPrefix(repos, kit.Path(ice.USR_LOCAL_REPOS)+nfs.PS), m.Option(ice.MSG_USERUA))
	if mdb.Conf(m, "web.code.git.service", "meta.cmd") == "git" {
		return _service_repos0(m, arg...)
	}
	m.Logs(m.R.Method, service, repos)
	info := false
	if m.Option(cli.CMD_DIR, repos); strings.HasSuffix(path.Join(arg...), INFO_REFS) {
		web.RenderType(m.W, "", kit.Format("application/x-git-%s-advertisement", service))
		_service_writer(m, "# service=git-"+service+lex.NL)
		info = true
	} else {
		web.RenderType(m.W, "", kit.Format("application/x-git-%s-result", service))
	}

	reader, err := _service_reader(m)
	if err != nil {
		return err
	}
	defer reader.Close()
	out := nfs.NewWriteCloser(func(buf []byte) (int, error) { return m.W.Write(buf) }, func() error { return nil })
	stream := ServerCommand{Stdin: reader, Stdout: out, Stderr: out}

	if service == RECEIVE_PACK {
		defer m.Cmd(Prefix(SERVICE), mdb.CREATE, mdb.NAME, path.Base(repos))
		return ServeReceivePack(info, stream, repos)
	} else {
		return ServeUploadPack(info, stream, repos)
	}
}
func _service_repos0(m *ice.Message, arg ...string) error {
	repos, service := _service_param(m, arg...)
	m.Logs(m.R.Method, service, repos)
	if m.Option(cli.CMD_DIR, repos); strings.HasSuffix(path.Join(arg...), INFO_REFS) {
		m.Option(ice.MSG_USERROLE, aaa.TECH)
		web.RenderType(m.W, "", kit.Format("application/x-git-%s-advertisement", service))
		_service_writer(m, "# service=git-"+service+lex.NL, _git_cmds(m, service, "--stateless-rpc", "--advertise-refs", nfs.PT))
		return nil
	}
	if service == RECEIVE_PACK {
		defer m.Cmd(Prefix(SERVICE), mdb.CREATE, mdb.NAME, path.Base(repos))
	}
	reader, err := _service_reader(m)
	if err != nil {
		return err
	}
	defer reader.Close()
	web.RenderType(m.W, "", kit.Format("application/x-git-%s-result", service))
	_git_cmd(m.Options(cli.CMD_INPUT, reader, cli.CMD_OUTPUT, m.W), service, "--stateless-rpc", nfs.PT)
	return nil
}
func _service_writer(m *ice.Message, cmd string, str ...string) {
	s := strconv.FormatInt(int64(len(cmd)+4), 16)
	kit.If(len(s)%4 != 0, func() { s = strings.Repeat("0", 4-len(s)%4) + s })
	m.W.Write([]byte(s + cmd + "0000" + strings.Join(str, "")))
}
func _service_reader(m *ice.Message) (io.ReadCloser, error) {
	switch m.R.Header.Get("content-encoding") {
	case "deflate":
		return flate.NewReader(m.R.Body), nil
	case "gzip":
		return gzip.NewReader(m.R.Body)
	}
	return m.R.Body, nil
}

const (
	INFO_REFS    = "info/refs"
	RECEIVE_PACK = "receive-pack"
	UPLOAD_PACK  = "upload-pack"
)
const SERVICE = "service"

func init() {
	web.Index.MergeCommands(ice.Commands{"/info/refs": {Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
		m.RenderRedirect(kit.MergeURL(ice.Info.Make.Remote+"/info/refs", m.OptionSimple(SERVICE)))
	}}})
	web.Index.MergeCommands(ice.Commands{"/x/": {Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
		if len(arg) == 0 {
			return
		}
		if arg[0] == ice.LIST {
			m.Cmd(Prefix(SERVICE), func(value ice.Maps) { m.Push(nfs.REPOS, web.MergeLink(m, "/x/"+kit.Keys(value[nfs.REPOS], GIT))) })
			m.Sort(nfs.REPOS)
			return
		} else if m.RenderVoid(); m.Option("go-get") == "1" {
			p := _git_url(m, path.Join(arg...))
			m.RenderResult(kit.Format(`<meta name="go-import" content="%s">`, kit.Format(`%s git %s`, strings.TrimSuffix(strings.Split(p, "://")[1], nfs.PT+GIT), p)))
			return
		}
		switch repos, service := _service_param(m, arg...); service {
		case RECEIVE_PACK:
			if !web.BasicCheck(m, "git server") {
				return
			}
			if !nfs.Exists(m, repos) {
				m.Cmd(Prefix(SERVICE), mdb.CREATE, mdb.NAME, path.Base(repos))
			}

		case UPLOAD_PACK:
			if mdb.Conf(m, Prefix(SERVICE), kit.Keym(aaa.AUTH)) == aaa.PRIVATE {
				if !web.BasicCheck(m, "git server") {
					return
				}
			}
			if m.Warn(!nfs.Exists(m, repos), ice.ErrNotFound, arg[0]) {
				return
			}
		}
		m.Warn(_service_repos(m, arg...), ice.ErrNotValid)
	}}})
	Index.MergeCommands(ice.Commands{
		SERVICE: {Name: "service repos branch commit file auto", Help: "代码源", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, ice.USR_LOCAL_REPOS, func(value ice.Maps) { _repos_insert(m, value[nfs.PATH]) })
			}},
			mdb.CREATE: {Name: "create name*=demo", Hand: func(m *ice.Message, arg ...string) {
				_repos_init(m, _service_path(m, m.Option(mdb.NAME)))
				_repos_insert(m, _service_path(m, m.Option(mdb.NAME)))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(REPOS) != "")
				mdb.HashRemove(m, m.Option(REPOS))
				nfs.Trash(m, _service_path(m, m.Option(REPOS)))
			}},
			RECEIVE_PACK: {Hand: func(m *ice.Message, arg ...string) {
				if err := file.ServeReceivePack(arg[0]); err != nil {
					fmt.Fprintln(os.Stderr, "ERR:", err)
					os.Exit(128)
				}
			}},
			UPLOAD_PACK: {Hand: func(m *ice.Message, arg ...string) {
				if err := file.ServeUploadPack(arg[0]); err != nil {
					fmt.Fprintln(os.Stderr, "ERR:", err)
					os.Exit(128)
				}
			}},
			code.INNER: {Hand: func(m *ice.Message, arg ...string) { _repos_inner(m, _service_path, arg...) }},
			web.DREAM_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(arg[0] == REPOS, func() {
					mdb.HashSelect(m).Sort(REPOS).Cut("repos,version,time")
					m.Cmd(mdb.SEARCH, nfs.REPOS).Table(func(value ice.Maps) {
						m.Push(nfs.REPOS, value["html_url"])
						m.Push(nfs.VERSION, "")
						m.Push(mdb.TIME, value["updated_at"])
					})
				})
			}},
		}, gdb.EventsAction(web.DREAM_INPUTS), mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,version,comment"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
					m.Push(nfs.SIZE, kit.Split(m.Cmdx(cli.SYSTEM, "du", "-sh", path.Join(ice.USR_LOCAL_REPOS, value[REPOS])))[0])
					m.PushScript(kit.Format("git clone %s", tcp.PublishLocalhost(m, kit.Split(web.MergeURL2(m, "/x/"+value[REPOS]+".git"), mdb.QS)[0])))
				}).Sort(REPOS).Cmdy("web.code.publish", ice.CONTEXTS, ice.DEV)
				kit.If(mdb.Config(m, aaa.AUTH) == aaa.PRIVATE, func() { m.StatusTimeCount(aaa.AUTH, aaa.PRIVATE) })
			} else if len(arg) == 1 {
				_repos_branch(m, _repos_open(m, arg[0]))
				m.EchoScript(tcp.PublishLocalhost(m, kit.Split(web.MergeURL2(m, "/x/"+arg[0]), mdb.QS)[0]))
			} else if len(arg) == 2 {
				repos := _repos_open(m, arg[0])
				if iter, err := repos.Branches(); err == nil {
					iter.ForEach(func(refer *plumbing.Reference) error {
						if refer.Name().Short() == arg[1] {
							_repos_log(m, refer.Hash(), repos)
						}
						return nil
					})
				}
			} else if len(arg) == 3 {
				if repos := _repos_open(m, arg[0]); arg[2] == INDEX {
					_repos_status(m, arg[0], repos)
				} else {
					_repos_stats(m, repos, arg[2])
				}
			} else {
				m.Cmdy("", code.INNER, arg)
			}
		}},
	})
}

func ServeReceivePack(info bool, srvCmd ServerCommand, path string) error {
	ep, err := transport.NewEndpoint(path)
	if err != nil {
		return err
	}
	s, err := server.DefaultServer.NewReceivePackSession(ep, nil)
	if err != nil {
		return fmt.Errorf("error creating session: %s", err)
	}
	return serveReceivePack(info, srvCmd, s)
}
func ServeUploadPack(info bool, srvCmd ServerCommand, path string) error {
	ep, err := transport.NewEndpoint(path)
	if err != nil {
		return err
	}
	s, err := server.DefaultServer.NewUploadPackSession(ep, nil)
	if err != nil {
		return fmt.Errorf("error creating session: %s", err)
	}
	return serveUploadPack(info, srvCmd, s)
}

type ServerCommand struct {
	Stderr io.Writer
	Stdout io.WriteCloser
	Stdin  io.Reader
}

func serveReceivePack(info bool, cmd ServerCommand, s transport.ReceivePackSession) error {
	if info {
		ar, err := s.AdvertisedReferences()
		if err != nil {
			return fmt.Errorf("internal error in advertised references: %s", err)
		}
		if err := ar.Encode(cmd.Stdout); err != nil {
			return fmt.Errorf("error in advertised references encoding: %s", err)
		}
		return nil
	}
	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(cmd.Stdin); err != nil {
		return fmt.Errorf("error decoding: %s", err)
	}
	rs, err := s.ReceivePack(context.TODO(), req)
	if rs != nil {
		if err := rs.Encode(cmd.Stdout); err != nil {
			return fmt.Errorf("error in encoding report status %s", err)
		}
	}
	if err != nil {
		return fmt.Errorf("error in receive pack: %s", err)
	}
	return nil
}
func serveUploadPack(info bool, cmd ServerCommand, s transport.UploadPackSession) (err error) {
	ioutil.CheckClose(cmd.Stdout, &err)
	if info {
		ar, err := s.AdvertisedReferences()
		if err != nil {
			return err
		}
		if err := ar.Encode(cmd.Stdout); err != nil {
			return err
		}
		return nil
	}
	req := packp.NewUploadPackRequest()
	if err := req.Decode(cmd.Stdin); err != nil {
		return err
	}
	resp, err := s.UploadPack(context.TODO(), req)
	if err != nil {
		return err
	}
	return resp.Encode(cmd.Stdout)
}
