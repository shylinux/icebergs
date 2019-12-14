module github.com/shylinux/icebergs/demo

go 1.13

require (
	github.com/shylinux/icebergs v0.0.0-20191212145348-fe6226481eaa
	github.com/shylinux/toolkits v0.0.0-20191212145555-d32eaba90a9e
)

replace (
	github.com/shylinux/icebergs => ../
	github.com/shylinux/toolkits => ../../toolkits
)
