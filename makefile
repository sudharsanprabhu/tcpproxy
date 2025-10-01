SERVER=proxyserver
CLIENT=proxyclient

build: clean copy-deps
	go build -o dist/${SERVER} ./cmd/server
	go build -o dist/${CLIENT} ./cmd/client

build-windows: clean copy-deps
	GOOS=windows GOARCH=amd64 go build -o dist/${SERVER}.exe ./cmd/server
	GOOS=windows GOARCH=amd64 go build -o dist/${CLIENT}.exe ./cmd/client

copy-deps:
	mkdir -p dist
	cp .env.template dist/.env
	cp config.toml.template dist/config.toml

clean:
	go clean
	rm -rf dist