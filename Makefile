NAME=everdeen
VERSION=0.1.0
ARCH=x86_64-linux-gnu

update-deps:
	go list -f '{{if not .Standard}}{{join .Deps "\n"}}{{end}}' ./... \
	  | sort -u \
	  | grep -v github.com/geckoboard/$(NAME) \
	  | xargs go get -f -u -d -v

build: *.go
	go build -o build/$(NAME)_$(VERSION)_$(ARCH)

gem: build
	mkdir -p ruby_client/binaries
	cp build/* ruby_client/binaries
