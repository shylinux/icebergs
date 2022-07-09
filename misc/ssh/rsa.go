package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"path"

	"golang.org/x/crypto/ssh"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	PUBLIC  = "public"
	PRIVATE = "private"
)
const RSA = "rsa"

func init() {
	aaa.Index.Merge(&ice.Context{Configs: ice.Configs{
		RSA: {Name: RSA, Help: "角色", Value: kit.Data(mdb.SHORT, mdb.HASH, mdb.FIELD, "time,hash,title,public,private")},
	}, Commands: ice.Commands{
		RSA: {Name: "rsa hash auto", Help: "公钥", Actions: ice.MergeAction(ice.Actions{
			mdb.IMPORT: {Name: "import key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(m.PrefixKey(), kit.Keys(mdb.HASH, path.Base(m.Option("key"))), kit.Data(mdb.TIME, m.Time(),
					"title", kit.Format("%s@%s", ice.Info.UserName, ice.Info.HostName),
					PRIVATE, m.Cmdx(nfs.CAT, kit.HomePath(m.Option("key"))),
					PUBLIC, m.Cmdx(nfs.CAT, kit.HomePath(m.Option("pub"))),
				))
			}},
			mdb.EXPORT: {Name: "export key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(m.PrefixKey(), m.Option(mdb.HASH)).Table(func(index int, value ice.Maps, head []string) {
					m.Cmdx(nfs.SAVE, kit.HomePath(m.Option("key")), value[PRIVATE])
					m.Cmdx(nfs.SAVE, kit.HomePath(m.Option("pub")), value[PUBLIC])
				})
			}},
			mdb.CREATE: {Name: "create bits=2048,4096 title=some", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if key, err := rsa.GenerateKey(rand.Reader, kit.Int(m.Option("bits"))); m.Assert(err) {
					if pub, err := ssh.NewPublicKey(key.Public()); m.Assert(err) {
						m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple("title"), PUBLIC, string(ssh.MarshalAuthorizedKey(pub)),
							PRIVATE, string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
						)
					}
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction(mdb.EXPORT, mdb.REMOVE); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.IMPORT)
			}
		}},
	}})
}
