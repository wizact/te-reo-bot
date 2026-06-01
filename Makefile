NAME := te-reo-bot
PKG := github.com/wizact/$(NAME)
SHELL := /bin/bash
GO := go
BUILDTAGS :=
PREFIX?=$(shell pwd)
OUTDIR := ${PREFIX}/out

.DEFAULT_GOAL := help

VERSION := $(shell cat VERSION.txt)
REGISTRY := "docker.pkg.github.com/wizact/te-reo-bot/"

GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
DOCKERIMAGEIDS := $(shell docker images --filter="REFERENCE=*${NAME}*" --filter="REFERENCE=${REGISTRY}${NAME}" -q)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
ifeq ($(GITCOMMIT),)
    GITCOMMIT := ${GITHUB_SHA}
endif

CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

# Detect OS for build target selection
UNAME_S := $(shell uname -s)

.PHONY: build-static
build-static:
	@mkdir -p $(OUTDIR)
	cd ./cmd/server/ && $(GO) build \
        -tags "$(BUILDTAGS) static_build" \
        ${GO_LDFLAGS_STATIC} -o $(OUTDIR)/$(NAME) .

.PHONY: build-dynamic
build-dynamic:
	@mkdir -p $(OUTDIR)
	cd ./cmd/server/ && $(GO) build \
        -tags "$(BUILDTAGS)" \
        ${GO_LDFLAGS} -o $(OUTDIR)/$(NAME) .

.PHONY: build
ifeq ($(UNAME_S),Darwin)
build: build-dynamic
	@echo "+ Built server (dynamic - macOS)"
else
build: build-static
	@echo "+ Built server (static - Linux)"
endif

.PHONY: clean
clean:
	@echo "+ $@"
	$(RM) -r $(OUTDIR)

.PHONY: version
version:
	@echo "+ $@"
	echo $(VERSION) > VERSION.txt

.PHONY: docker-build
docker-build:
	@echo "+ $@"
	@docker build --rm -t $(REGISTRY)$(NAME):$(GITCOMMIT) .

.PHONY: docker-rmi
docker-rmi:
	@echo "+ $@"
	@for DOCKERIMAGEID in $(DOCKERIMAGEIDS); do \
		echo "Image Id: $$DOCKERIMAGEID"; \
		DCIDS=$$(docker ps -q -a --filter "ancestor=$$DOCKERIMAGEID" $<); \
		if [[ ! -z $$DCIDS ]]; then \
			 echo "Docker containers found: $$DCIDS"; \
			 docker rm `docker ps -q -a --filter "ancestor=$$DOCKERIMAGEID"`; \
		fi; \
		docker rmi $$DOCKERIMAGEID; \
	done;

.PHONY: help
help:
	@echo "Te Reo Bot - Makefile Commands"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build          - Build server binary (auto-detects OS)"
	@echo "  make build-static   - Build static binary (Linux only)"
	@echo "  make build-dynamic  - Build dynamic binary (macOS/dev)"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-rmi     - Remove Docker images"
	@echo ""
	@echo "Other Commands:"
	@echo "  make version        - Update VERSION.txt"
	@echo "  make help           - Show this help message"
