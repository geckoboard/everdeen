NAME=everdeen
VERSION=0.1.0
ARCH=x86_64-linux-gnu

build: *.go
	go build -o build/$(NAME)_$(VERSION)_$(ARCH)

gem: build
	mkdir -p ruby_client/binaries
	cp build/* ruby_client/binaries
