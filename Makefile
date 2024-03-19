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
        @mkdir -p $(BUILD_DIR)/$(Version)/usr/local/bin/
        @mkdir -p $(BUILD_DIR)/$(Version)/etc/sysconfig/
        @mkdir -p $(BUILD_DIR)/$(Version)/usr/lib/systemd/system/
        @mkdir -p $(BUILD_DIR)/$(Version)/var/log/$(BINARY_NAME)/
        @mkdir -p $(BUILD_DIR)/$(Version)/etc/rsyslog.d/
build:
        $(MAKE) pre_build
        @CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -ldflags $(FLAGS) -o $(BUILD_DIR)/${Version}/usr/local/bin/$(BINARY_NAME)
        @cp $(BINARY_NAME).service $(BUILD_DIR)/$(Version)/usr/lib/systemd/system/$(BINARY_NAME)
        @cp rsyslog                $(BUILD_DIR)/$(Version)/etc/rsyslog.d/$(BINARY_NAME)
        @cp sysconfig              $(BUILD_DIR)/$(Version)/etc/sysconfig/$(BINARY_NAME)

clean:
        @rm -rf $(BUILD_DIR)

dep:
        export GOPROXY=https://goproxy.io,direct
        go mod download