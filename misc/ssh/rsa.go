package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	PUBLIC  = "public"
	PRIVATE = "private"
)
const RSA = "rsa"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		RSA: {Name: RSA, Help: "角色", Value: kit.Data(mdb.SHORT, mdb.HASH, mdb.FIELD, "time,hash,public,private")},
	}, Commands: map[string]*ice.Command{
		RSA: {Name: "rsa hash auto create import", Help: "公钥", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				// m.Cmd(m.PrefixKey(), mdb.IMPORT)
			}},
			mdb.IMPORT: {Name: "import key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(m.PrefixKey(), kit.Keys(mdb.HASH, "id_rsa"), kit.Data(mdb.TIME, m.Time(),
					PRIVATE, m.Cmdx("nfs.cat", kit.HomePath(m.Option("key"))),
					PUBLIC, m.Cmdx("nfs.cat", kit.HomePath(m.Option("pub"))),
				))
			}},
			mdb.EXPORT: {Name: "export key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(m.PrefixKey(), m.Option(mdb.HASH)).Table(func(index int, value map[string]string, head []string) {
					m.Cmdx("nfs.save", kit.HomePath(m.Option("key")), value[PRIVATE])
					m.Cmdx("nfs.save", kit.HomePath(m.Option("pub")), value[PUBLIC])
				})
			}},
			mdb.CREATE: {Name: "create bits=2048,4096", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if key, err := rsa.GenerateKey(rand.Reader, kit.Int(m.Option("bits"))); m.Assert(err) {
					if pub, err := ssh.NewPublicKey(key.Public()); m.Assert(err) {
						m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH,
							PRIVATE, string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
							PUBLIC, string(ssh.MarshalAuthorizedKey(pub)),
						)
					}
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.PushAction(mdb.EXPORT, mdb.REMOVE)
		}},
	}})
}
