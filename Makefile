LDFLAGS := -s -w

all: fmt build linux mac win

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