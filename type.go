package ice

import (
	"time"
)

type Cache struct {
	Value string
	Name  string
	Help  string
	Hand  func(m *Message, x *Cache, arg ...string) string
}
type Config struct {
	Value interface{}
	Name  string
	Help  string
	Hand  func(m *Message, x *Config, arg ...string) string
}
type Command struct {
	Form map[string]int
	Name string
	Help interface{}
	Auto func(m *Message, c *Context, key string, arg ...string) (ok bool)
	Hand func(m *Message, c *Context, key string, arg ...string) (e error)
}
type Context struct {
	Name string
	Help string

	Caches   map[string]*Cache
	Configs  map[string]*Config
	Commands map[string]*Command

	contexts map[string]*Context
	context  *Context
	root     *Context

	exit chan bool
	Server
}
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}

type Message struct {
	time time.Time
	code int

	Meta map[string][]string
	Data map[string]interface{}

	messages []*Message
	message  *Message
	root     *Message

	source *Context
	target *Context
	Hand   bool
}
