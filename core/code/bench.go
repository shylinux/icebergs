package code

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

type Bench struct {
	*ice.Message
}

func (b *Bench) Get(key string) string {
	return b.Conf("hi", key)
}
func (b *Bench) GetInt(key string) int64 {
	return kit.Int64(b.Conf("hi", key))
}

func (b *Bench) Log(key string, arg ...interface{}) {
	b.Message.Info("%s: %v", key, kit.Simple(key))
}
func (b *Bench) Logf(key string, str string, arg ...interface{}) {
	b.Message.Info("%s: %v", key, kit.Format(str, arg...))
}
func (b *Bench) Info(arg ...interface{}) {
	b.Log("info", arg...)
}
func (b *Bench) Infof(str string, arg ...interface{}) {
	b.Logf("info", str, arg...)
}
func (b *Bench) Warn(arg ...interface{}) {
	b.Log("warn", arg...)
}
func (b *Bench) Warnf(str string, arg ...interface{}) {
	b.Logf("warn", str, arg...)
}
func (b *Bench) Error(arg ...interface{}) {
	b.Log("error", arg...)
}
func (b *Bench) Errorf(str string, arg ...interface{}) {
	b.Logf("error", str, arg...)
}
func (b *Bench) Debug(arg ...interface{}) {
	b.Log("debug", arg...)
}
func (b *Bench) Debugf(str string, arg ...interface{}) {
	b.Logf("debug", str, arg...)
}
