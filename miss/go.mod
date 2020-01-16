module miss

go 1.13

require github.com/shylinux/icebergs v0.1.9

replace (
	github.com/shylinux/icebergs => ../
	github.com/shylinux/toolkits => ../../toolkits
)
