NAME := te-reo-bot
PKG := github.com/wizact/$(NAME)
SHELL := /bin/bash
GO := go
BUILDTAGS :=
PREFIX?=$(shell pwd)
OUTDIR := ${PREFIX}/out

VERSION := $(shell cat VERSION.txt)

GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
ifeq ($(GITCOMMIT),)
    GITCOMMIT := ${GITHUB_SHA}
endif

CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

.PHONY: build-static
build-static:
	$(GO) build \
        -tags "$(BUILDTAGS) static_build" \
        ${GO_LDFLAGS_STATIC} -o $(OUTDIR)/$(NAME) .

.PHONY: clean
clean:
	@echo "+ $@"
	$(RM) -r $(OUTDIR)

.PHONY: version
version:
	@echo "+ $@"
	echo $(VERSION) > VERSION.txt


