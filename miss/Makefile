all:
	@echo && date
	export GOPRIVATE=github.com
	go build -o bin/ice.bin src/main.go && chmod u+x bin/ice.bin && ./bin/ice.sh restart
