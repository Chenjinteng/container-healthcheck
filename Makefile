.PHONY: build pre_make clean dep

BINARY_NAME=container-healthcheck
Author=Chenjinteng
BUILD_DIR="build"
Version :=$(shell cat VERSION)
GoVersion := $(shell go version)
BuildTime := $(shell date +'%Y-%m-%d %H:%M:%S')
GitCommit := $(shell git rev-parse --short HEAD)

FLAGS="-X 'main.Author=$(Author)' -X 'main.Version=$(Version)'  -X 'main.GitCommit=$(GitCommit)'  -X 'main.GoVersion=$(GoVersion)' -X 'main.BuildTime=$(BuildTime)'"

pre_build:
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/usr/local/bin/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/etc/sysconfig/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/usr/lib/systemd/system/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/var/log/$(BINARY_NAME)/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/etc/rsyslog.d/
	@cp config/$(BINARY_NAME).service $(BUILD_DIR)/$(Version)/$(ARCH)/usr/lib/systemd/system/$(BINARY_NAME).service
	@cp config/rsyslog                $(BUILD_DIR)/$(Version)/$(ARCH)/etc/rsyslog.d/$(BINARY_NAME)
	@cp config/sysconfig              $(BUILD_DIR)/$(Version)/$(ARCH)/etc/sysconfig/$(BINARY_NAME)


build_x86:
	@go version || exit $$?;
	ARCH="amd64"
	$(MAKE) pre_build
	@CGO_ENABLED=0 GOOS="linux" GOARCH="$(ARCH)" go build -ldflags $(FLAGS) -o $(BUILD_DIR)/${Version}/usr/local/bin/$(BINARY_NAME)
	@chmod +x $(BUILD_DIR)/$(Version)$(ARCH)/usr/local/bin/$(BINARY_NAME)

build_arm7:
	@go version || exit $$?;
	ARCH="arm64"
	$(MAKE) pre_build
	@CGO_ENABLED=0 GOOS="linux" GOARCH="$(ARCH)" GOARM=7 go build -ldflags $(FLAGS) -o $(BUILD_DIR)/${Version}/usr/local/bin/$(BINARY_NAME)
	@chmod +x $(BUILD_DIR)/$(Version)$(ARCH)/usr/local/bin/$(BINARY_NAME)

build:
	$(MAKE) build_x86
	$(MAKE) build_arm7

install:
	OSARCH := $(shell uname -m)
	ifeq "$(OSARCH)" "x86_64"; then \
		cp -r $(BUILD_DIR)/$(Version)/amd64/* /
	else \
		cp -r $(BUILD_DIR)/$(Version)/arm64/* /
	endif

clean:
	@rm -rf $(BUILD_DIR)

dep:
	GOPROXY=https://goproxy.io,direct go mod download