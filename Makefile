NAME := te-reo-bot
PKG := github.com/wizact/$(NAME)
SHELL := /bin/bash
GO := go
BUILDTAGS :=
PREFIX?=$(shell pwd)
OUTDIR := ${PREFIX}/out

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
