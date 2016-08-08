NAME=everdeen
PACKAGE=github.com/geckoboard/$(NAME)
SERVER_VERSION=0.1.0
GIT_SHA=$(shell git rev-parse --short HEAD)
RUBY_BINARIES_DIR=ruby_client/binaries
BUILD_DIR=build

FPM_VERSION = 1.3.3
DEB_S3_VERSION = 0.7.1
AWS_CLI_VERSION = 1.7.3
DEB_VERSION="0.0.0+git~$(GIT_SHA)"

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
		-ldflags "-X main.Version $(SERVER_VERSION) -X main.GitSHA $(GIT_SHA)" \
		-verbose \
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

install-ci-deps:
	go get -f -u github.com/jstemmer/go-junit-report
	(gem list | grep fpm > /dev/null) || gem install fpm --no-ri --no-rdoc --version $(FPM_VERSION)
	(gem list | grep deb-s3 > /dev/null) || gem install deb-s3 --no-ri --no-rdoc --version $(DEB_S3_VERSION)
	(pip freeze --user | grep "awscli==$(AWS_CLI_VERSION)" > /dev/null) || sudo pip install --ignore-installed --user awscli==$(AWS_CLI_VERSION)

package: build
	@mkdir -p pkg tmp/bin tmp/share/spreadsheets
	@cp bin/* tmp/bin
	@cp -r migrate/migrations tmp/share/spreadsheets/migrations
	fpm -C tmp -t deb -s dir --name $(NAME) --version $(DEB_VERSION) --prefix /usr/local/ --provides $(NAME) --force .
	@mv $(NAME)_$(DEB_VERSION)_amd64.deb pkg

import:
	@aws s3 cp $(APT_REPO) gpg-private-key
	gpg --import gpg-private-key
	@rm -f gpg-private-key

release: package
	@aws s3 cp pkg/$(NAME)_$(DEB_VERSION)_amd64.deb s3://gecko.pkg/$(NAME)/
	# Using the @ symbol to hide the command in make log after GPG passphrase interpolated
	@deb-s3 upload --bucket gecko.apt --arch amd64 pkg/$(NAME)_$(DEB_VERSION)_amd64.deb -v private --sign $(GPG_KEY_ID) --gpg-options "--passphrase $(GPG_PASSPHRASE)"
