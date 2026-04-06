run:
	go run cmd/main.go

build:
	go build -o bin/homelab cmd/main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/homelab.exe cmd/main.go