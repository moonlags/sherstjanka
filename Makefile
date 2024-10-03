build:
	go build -o ./bin/sherstjanka -race ./cmd/main/

run:
	go run ./cmd/main/

windows-build:
	GOOS=windows GOARCH=amd64 go build -ldflags -H=windowsgui -o ./bin/sherstjanka.exe ./cmd/main/
