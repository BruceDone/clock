LDFLAGS := -s -w
CMD_PATH := ./cmd/clock
WEB_SRC := server/webapp
WEB_DEST := cmd/clock/web/dist

all: fmt web build linux mac win

web:
	@echo "Building frontend..."
	cd $(WEB_SRC) && npm install && npm run build
	@echo "Copying frontend to $(WEB_DEST)..."
	mkdir -p cmd/clock/web
	cp -r $(WEB_SRC)/dist $(WEB_DEST)/..
	@echo "Frontend built"

build: linux

fmt:
	go fmt ./...

linux:
	env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-linux $(CMD_PATH)
	cp configs/config.toml bin/config.toml

mac:
	env GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-darwin $(CMD_PATH)
	cp configs/config.toml bin/config.toml

win:
	env GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-windows.exe $(CMD_PATH)
	cp configs/config.toml bin/config.toml

test:
	go test ./...

clean:
	rm -f ./bin/clock*
	rm -rf $(WEB_SRC)/node_modules
	rm -rf $(WEB_SRC)/dist
	rm -rf $(WEB_DEST)

.PHONY: all build web clean fmt linux mac win test
