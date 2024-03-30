package ssh

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	ETC_CERT = "etc/cert/"
	PEM      = "pem"
	KEY      = "key"

	SIGN    = "sign"
	VERIFY  = "verify"
	ENCRYPT = "encrypt"
	DECRYPT = "decrypt"
)
const CERT = "cert"

func init() {
	aaa.Index.MergeCommands(ice.Commands{
		CERT: {Name: "cert path auto", Help: "证书", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name* country province city street postal company year month=1 day", Hand: func(m *ice.Message, arg ...string) {
				if nfs.Exists(m, CertPath(m, m.Option(mdb.NAME)), func(p string) {
					m.Push(PEM, p).Push(KEY, kit.ExtChange(p, KEY))
				}) {
					return
				}
				cert := &x509.Certificate{
					SerialNumber: big.NewInt(time.Now().Unix()),
					Subject: pkix.Name{
						CommonName: m.Option(mdb.NAME), Organization: []string{m.Option(aaa.COMPANY)},
						Country: []string{m.Option(aaa.COUNTRY)}, Province: []string{m.Option(aaa.PROVINCE)}, Locality: []string{m.Option(aaa.CITY)},
						StreetAddress: []string{m.Option("street")}, PostalCode: []string{m.Option("postal")},
					},
					NotBefore: time.Now(), NotAfter: time.Now().AddDate(kit.Int(m.Option("year")), kit.Int(m.Option("month")), kit.Int(m.Option("day"))),
					KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
				}
				certKey, _ := rsa.GenerateKey(rand.Reader, 4096)
				ca, caKey, err := LoadCertKey(m)
				kit.If(err != nil, func() { ca, caKey, cert.KeyUsage = cert, certKey, x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign })
				if certBuf, err := x509.CreateCertificate(rand.Reader, cert, ca, certKey.Public(), caKey); !m.Warn(err) {
					SaveCertKey(m, m.Option(mdb.NAME)+nfs.PT+KEY, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(certKey))
					SaveCertKey(m, m.Option(mdb.NAME)+nfs.PT+PEM, "CERTIFICATE", certBuf)
				}
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, m.Option(nfs.PATH)) }},
			ENCRYPT: {Name: "encrypt text", Help: "加密", Hand: func(m *ice.Message, arg ...string) {
				if cert, err := loadCert(m, m.Option(nfs.PATH)); err == nil {
					data, _ := rsa.EncryptPKCS1v15(rand.Reader, cert.PublicKey.(*rsa.PublicKey), []byte(m.Option(mdb.TEXT)))
					m.Echo(base64.StdEncoding.EncodeToString(data)).ProcessInner()
				}
			}},
			DECRYPT: {Name: "decrypt text", Help: "解密", Hand: func(m *ice.Message, arg ...string) {
				if key, err := loadKey(m, m.Option(nfs.PATH)); err == nil {
					text, _ := base64.StdEncoding.DecodeString(m.Option(mdb.TEXT))
					if data, err := rsa.DecryptPKCS1v15(rand.Reader, key, text); !m.Warn(err) {
						m.Echo(string(data)).ProcessInner()
					}
				}
			}},
			SIGN: {Name: "sign text", Help: "签名", Hand: func(m *ice.Message, arg ...string) {
				if key, err := loadKey(m, m.Option(nfs.PATH)); err == nil {
					hash := sha256.New()
					hash.Write([]byte(strings.TrimSpace(m.Option(mdb.TEXT))))
					data, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash.Sum(nil))
					m.Echo(base64.StdEncoding.EncodeToString(data) + lex.SP + m.Option(mdb.TEXT)).ProcessInner()
				}
			}},
			VERIFY: {Name: "verify text", Help: "验签", Hand: func(m *ice.Message, arg ...string) {
				if cert, err := loadCert(m, m.Option(nfs.PATH)); err == nil {
					ls := strings.SplitN(strings.TrimSpace(m.Option(mdb.TEXT)), lex.SP, 2)
					hash := sha256.New()
					hash.Write([]byte(ls[1]))
					sign, _ := base64.StdEncoding.DecodeString(ls[0])
					if !m.Warn(rsa.VerifyPKCS1v15(cert.PublicKey.(*rsa.PublicKey), crypto.SHA256, hash.Sum(nil), sign)) {
						m.Echo(ice.OK).ProcessInner()
					}
				}
			}},
		}, mdb.HashAction(PEM, nfs.ETC_CERT_PEM, KEY, nfs.ETC_CERT_KEY)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(nfs.DIR, ETC_CERT).Table(func(value ice.Maps) {
					switch kit.Ext(value[nfs.PATH]) {
					case PEM:
						m.PushButton(ENCRYPT, VERIFY, nfs.TRASH)
					case KEY:
						m.PushButton(DECRYPT, SIGN, nfs.TRASH)
					default:
						m.PushButton(nfs.TRASH)
					}
				}).Action(mdb.CREATE)
			} else {
				switch block, _ := pem.Decode([]byte(m.Cmdx(nfs.CAT, arg[0]))); block.Type {
				case "CERTIFICATE":
					cert, _ := x509.ParseCertificate(block.Bytes)
					m.Push("NotAfter", cert.NotAfter.Format(ice.MOD_TIME)).Push("Subject", cert.Subject.CommonName)
					m.Push("Issuer", kit.Select(cert.Issuer.CommonName, cert.Issuer.Organization, 0))
					m.Echo(kit.Formats(cert))
				case "RSA PRIVATE KEY":
					key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
					m.Echo(kit.Formats(key))
				}
			}
		}},
	})
}
func CertPath(m *ice.Message, domain string) string {
	return path.Join(ETC_CERT+domain) + nfs.PT + PEM
}
func loadBlock(m *ice.Message, p string) []byte {
	block, _ := pem.Decode([]byte(m.Cmdx(nfs.CAT, p)))
	return block.Bytes
}
func loadCert(m *ice.Message, p string) (*x509.Certificate, error) {
	if cert, err := x509.ParseCertificate(loadBlock(m, p)); m.Warn(err) {
		return nil, err
	} else {
		return cert, nil
	}
}
func loadKey(m *ice.Message, p string) (*rsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(loadBlock(m, p)); m.Warn(err) {
		return nil, err
	} else {
		return key, nil
	}
}
func LoadCertKey(m *ice.Message) (*x509.Certificate, *rsa.PrivateKey, error) {
	if cert, err := loadCert(m, mdb.Config(m, PEM)); m.Warn(err) {
		return nil, nil, err
	} else if key, err := loadKey(m, mdb.Config(m, KEY)); m.Warn(err) {
		return nil, nil, err
	} else {
		return cert, key, nil
	}
}
func SaveCertKey(m *ice.Message, file, Type string, Bytes []byte) {
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{Type: Type, Bytes: Bytes})
	p := path.Join(ETC_CERT + file)
	m.Push(kit.Ext(file), p).Cmd(nfs.SAVE, p, certPEM.String())
}
