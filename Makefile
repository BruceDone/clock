LDFLAGS := -s -w

all: fmt build linux darwin windows

build: linux

fmt:
	go fmt ./...

linux:
	env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-linux .
	cp config/dev.yaml bin/config.yaml

darwin:
	env GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-darwin .
	cp config/dev.yaml bin/config.yaml

windows:
	env GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-windows.exe .
	cp config/dev.yaml bin/config.yaml

clean:
	rm -f ./bin/clock