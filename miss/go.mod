module miss

go 1.13

require github.com/shylinux/icebergs v0.1.0

replace (
	github.com/shylinux/icebergs => ../
	github.com/shylinux/toolkits => ../../toolkits
)
