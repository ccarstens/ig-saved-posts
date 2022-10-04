PHONY: build

build:
	go build -o bin/IgSavedPosts src/main.go && GOOS=windows GOARCH=amd64 go build -o bin/IgSavedPosts-amd64.exe src/main.go
