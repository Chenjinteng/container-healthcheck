# Copyright 2024 Chenjinteng

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BINARY_NAME=container-healthcheck
Author=Chenjinteng
BUILD_DIR=binary
Version :=$(shell cat VERSION)
GoVersion := $(shell go version)
BuildTime := $(shell date +'%Y-%m-%d %H:%M:%S')
GitCommit := $(shell git rev-parse --short HEAD)
OSARCH := $(shell uname -m)

FLAGS="-X 'main.Author=$(Author)' -X 'main.Version=$(Version)'  -X 'main.GitCommit=$(GitCommit)'  -X 'main.GoVersion=$(GoVersion)' -X 'main.BuildTime=$(BuildTime)'"


build_x86: ARCH=amd64
build_x86:
	@go version || exit $$?;
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/usr/local/bin/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/etc/sysconfig/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/usr/lib/systemd/system/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/var/log/$(BINARY_NAME)/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/etc/rsyslog.d/
	CGO_ENABLED=0 \
		GOOS="linux" \
		GOARCH="$(ARCH)" \
		go build \
		-ldflags $(FLAGS) \
		-o $(BUILD_DIR)/${Version}/$(ARCH)/usr/local/bin/$(BINARY_NAME)
	chmod +x $(BUILD_DIR)/$(Version)/$(ARCH)/usr/local/bin/$(BINARY_NAME)
	@cp config/$(BINARY_NAME).service $(BUILD_DIR)/$(Version)/$(ARCH)/usr/lib/systemd/system/$(BINARY_NAME).service
	@cp config/rsyslog                $(BUILD_DIR)/$(Version)/$(ARCH)/etc/rsyslog.d/$(BINARY_NAME)
	@cp config/sysconfig              $(BUILD_DIR)/$(Version)/$(ARCH)/etc/sysconfig/$(BINARY_NAME)


build_arm7: ARCH=arm64
build_arm7:
	@go version || exit $$?;
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/usr/local/bin/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/etc/sysconfig/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/usr/lib/systemd/system/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/var/log/$(BINARY_NAME)/
	@mkdir -p $(BUILD_DIR)/$(Version)/$(ARCH)/etc/rsyslog.d/
	CGO_ENABLED=0 \
		GOOS="linux" \
		GOARCH="$(ARCH)" \
		GOARM=7 \
		go build \
		-ldflags $(FLAGS) \
		-o $(BUILD_DIR)/${Version}/$(ARCH)/usr/local/bin/$(BINARY_NAME)
	chmod +x $(BUILD_DIR)/$(Version)/$(ARCH)/usr/local/bin/$(BINARY_NAME)
	@cp config/$(BINARY_NAME).service $(BUILD_DIR)/$(Version)/$(ARCH)/usr/lib/systemd/system/$(BINARY_NAME).service
	@cp config/rsyslog                $(BUILD_DIR)/$(Version)/$(ARCH)/etc/rsyslog.d/$(BINARY_NAME)
	@cp config/sysconfig              $(BUILD_DIR)/$(Version)/$(ARCH)/etc/sysconfig/$(BINARY_NAME)

build:
	$(MAKE) build_x86
	$(MAKE) build_arm7

install:
ifeq ("$(OSARCH)", "x86_64")
	cp -r $(BUILD_DIR)/$(Version)/amd64/* /
else 
	cp -r $(BUILD_DIR)/$(Version)/arm64/* /
endif

package:
	mkdir -p $(Workdir)/$(BUILD_DIR)/packages
	cd $(Workdir)/$(BUILD_DIR)/$(Version)/arm64/; tar zcvf $(Workdir)/$(BUILD_DIR)/packages/$(BINARY_NAME)-$(Version)-arm64.tar.gz *
	cd $(Workdir)/$(BUILD_DIR)/$(Version)/amd64/; tar zcvf $(Workdir)/$(BUILD_DIR)/packages/$(BINARY_NAME)-$(Version)-amd64.tar.gz *

clean:
	rm -rf $(BUILD_DIR)

uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)
	rm -f /etc/sysconfig/$(BINARY_NAME)
	rm -f /usr/lib/systemd/system/$(BINARY_NAME).service

dep:
	GOPROXY=https://goproxy.io,direct go mod download
