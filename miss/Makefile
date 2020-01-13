all:
	@echo && date
	export CGO_ENABLED=0
	export GOPRIVATE=github.com
	export GOPROXY=https://goproxy.cn
	go build -o bin/ice.bin src/main.go && chmod u+x bin/ice.bin && ./bin/ice.sh restart
