LDFLAGS := -s -w

all: fmt build linux darwin windows

build: linux

fmt:
	go fmt ./...

linux:
	env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-linux .
	cp config/dev.yaml bin/config.yaml

darwin:
	env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-darwin .
	cp config/dev.yaml bin/config.yaml

windows:
	env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/clock-windows.exe .
	cp config/dev.yaml bin/config.yaml

clean:
	rm -f ./bin/clock