.PHONY:
.SILENT:

buildServer:
	go build -o ./.bin/server src/cmd/server/main.go

runServer: buildServer
	./.bin/server
