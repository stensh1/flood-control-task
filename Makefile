.PHONY:
.SILENT:

buildServer:
	go build -o ./.bin/server src/cmd/server/main.go

runServer: buildServer
	./.bin/server

buildClient:
	go build -o ./.bin/client src/cmd/client/main.go

runClient: buildClient
	./.bin/client