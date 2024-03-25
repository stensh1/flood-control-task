.PHONY:
.SILENT:

buildServer:
	go build -o ./.bin/server main.go

runServer: buildServer
	./.bin/server