NAME=everdeen
PACKAGE=github.com/geckoboard/$(NAME)
SERVER_VERSION=0.1.0
RUBY_BINARIES_DIR=ruby_client/binaries
BUILD_DIR=build

update-deps:
	go get github.com/mitchellh/gox
	go list -f '{{if not .Standard}}{{join .Deps "\n"}}{{end}}' ./... \
	  | sort -u \
	  | grep -v github.com/geckoboard/$(NAME) \
	  | xargs go get -f -u -d -v

build: *.go
	rm -rf $(BUILD_DIR)
	gox -osarch="linux/386 linux/amd64 darwin/amd64" \
		-output="$(BUILD_DIR)/$(NAME)_$(SERVER_VERSION)_{{.OS}}-{{.Arch}}" \
		$(PACKAGE)

gem: build
	rm -rf $(RUBY_BINARIES_DIR)
	mkdir -p $(RUBY_BINARIES_DIR)
	cp $(BUILD_DIR)/* $(RUBY_BINARIES_DIR)

test: test-server test-client

test-server:
	go test -v ./...

test-client: gem
	cd ruby_client && bundle exec rspec
