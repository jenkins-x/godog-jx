#
# Copyright (C) 2015 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

SHELL := /bin/bash
NAME := godog-jenkins
GO := GO15VENDOREXPERIMENT=1 go
ROOT_PACKAGE := $(shell $(GO) list .)
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/)

REV        := $(shell git rev-parse --short HEAD 2> /dev/null  || echo 'unknown')
BRANCH     := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null  || echo 'unknown')
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
BUILDFLAGS := -ldflags \
  " -X $(ROOT_PACKAGE)/version.Version=$(VERSION)\
		-X $(ROOT_PACKAGE)/version.Revision='$(REV)'\
		-X $(ROOT_PACKAGE)/version.Branch='$(BRANCH)'\
		-X $(ROOT_PACKAGE)/version.BuildDate='$(BUILD_DATE)'\
		-X $(ROOT_PACKAGE)/version.GoVersion='$(GO_VERSION)'"
CGO_ENABLED = 0

VENDOR_DIR=vendor

GITEA_USER ?= testuser
GITEA_PASSWORD ?= testuser
GITEA_EMAIL ?= testuser@acme.com

GIT_PROVIDER_URL ?= https://github.beescloud.com
GHE_USER ?= dev1
GHE_TOKEN ?= changeme
GHE_EMAIL ?= testuser@acme.com

all: test

check: fmt test

#build: *.go */*.go
#	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(BUILDFLAGS) -o build/$(NAME) $(NAME).go

test:
	cd jx && godog
	cd github && godog
	cd jenkins && godog


bdd-all: *.go
	$(GO) test $(PACKAGE_DIRS) -test.run TestMain  -test.v -godog.feature=.

jx-spring: *.go
	$(GO) test -test.v -godog.feature=spring.feature

jx-import: *.go
	$(GO) test -test.v -godog.feature=import.feature

jx-import-url: *.go
	$(GO) test -test.v -godog.feature=importurl.feature


configure-gitea:
	echo "Installing gitea addon with user $(GITEA_USER) email: $(GITEA_EMAIL)"
	jx create addon gitea -b --headless --username $(GITEA_USER) --password $(GITEA_PASSWORD) --email $(GITEA_EMAIL)

configure-ghe:
	echo "Setting up GitHub Enterprise support for user $(GHE_USER) email: $(GITEA_EMAIL)"
	jx create git server github $(GIT_PROVIDER_URL) -n GHE
	jx delete git server GitHub
	jx create git token -n GHE $(GHE_USER) -t $(GHE_TOKEN)

bdd-init: 
	git config --global --add user.name JenkinsXBot
	git config --global --add user.email jenkins-x@googlegroups.com
	
bdd-tests: jx-spring

fmt:
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

bootstrap: vendoring

vendoring:
	$(GO) get -u github.com/Masterminds/glide
	GO15VENDOREXPERIMENT=1 glide update --strip-vendor


clean:
	rm -rf build

.PHONY: release clean test
