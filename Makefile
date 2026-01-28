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
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

.PHONY: build-static
build-static:
	cd ./cmd/server/ && $(GO) build \
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

# Database commands
.PHONY: backup-all
backup-all:
	@echo "+ Backing up ALL words from database to JSON"
	@$(GO) run cmd/dict-gen/main.go generate --all --output=backup-all.json
	@echo "Backup complete: backup-all.json"

.PHONY: backup
backup:
	@echo "+ Backing up 366 words (with day_index) from database to JSON"
	@$(GO) run cmd/dict-gen/main.go generate --output=backup-366.json
	@echo "Backup complete: backup-366.json"

.PHONY: migrate
migrate:
	@echo "+ Migrating dictionary.json to database"
	@$(GO) run cmd/dict-gen/main.go migrate --input=cmd/server/dictionary.json
	@echo "Migration complete"

.PHONY: validate
validate:
	@echo "+ Validating database"
	@$(GO) run cmd/dict-gen/main.go validate

.PHONY: help
help:
	@echo "Te Reo Bot - Makefile Commands"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build-static    - Build static binary"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Database Commands:"
	@echo "  make backup-all     - Export ALL words to backup-all.json (includes extras)"
	@echo "  make backup         - Export 366 words to backup-366.json (day_index only)"
	@echo "  make migrate        - Import dictionary.json into database"
	@echo "  make validate       - Validate database integrity"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-rmi     - Remove Docker images"
	@echo ""
	@echo "Other Commands:"
	@echo "  make version        - Update VERSION.txt"
	@echo "  make help           - Show this help message"
