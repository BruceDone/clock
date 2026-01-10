LDFLAGS := -s -w

all: fmt web build linux mac win

web:
	@echo "Building frontend..."
	cd server/webapp && npm install && npm run build
	@echo "Frontend built to server/webapp/dist"

build: linux

fmt:
	go fmt ./...

linux:
	env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-linux .
	cp config/dev.yaml bin/config.yaml

mac:
	env GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-darwin .
	cp config/dev.yaml bin/config.yaml

win:
	env GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-windows.exe .
	cp config/dev.yaml bin/config.yaml

clean:
	rm -f ./bin/clock
	rm -rf server/webapp/node_modules
	rm -rf server/webapp/dist

.PHONY: all build web clean fmt linux mac win